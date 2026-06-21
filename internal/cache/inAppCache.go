package cache

import (
	"container/list"
	"log/slog"
	"sync"
	"time"
)

type inAppCache struct {
	mutex        sync.RWMutex
	store        map[string]*list.Element
	ll           *list.List
	maxBytes     int64
	currentBytes int64
}

func NewInAppCache(cleanupInterval time.Duration, maxMemoryMB int) Cache {
	c := &inAppCache{
		store:    make(map[string]*list.Element),
		ll:       list.New(),
		maxBytes: int64(maxMemoryMB) * 1024 * 1024,
	}

	go c.janitor(cleanupInterval)

	return c
}

func getSize(entry CacheEntry) int64 {
	return int64(len(entry.Body) + 200)
}

func (c *inAppCache) Get(key string) (CacheEntry, bool) {
	slog.Debug("getting from in-app cache for key", "key", key)
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if element, exists := c.store[key]; exists {
		item := element.Value.(*lruItem)

		if time.Now().After(item.entry.ExpiresAt) {
			c.removeElement(element)
			return CacheEntry{}, false
		}

		c.ll.MoveToFront(element)
		return item.entry, true
	}

	return CacheEntry{}, false
}

func (c *inAppCache) Set(key string, value CacheEntry, ttl time.Duration) {
	slog.Debug("setting in-app cache for key", "key", key)
	c.mutex.Lock()
	defer c.mutex.Unlock()

	value.ExpiresAt = time.Now().Add(ttl)
	itemSize := getSize(value)

	if element, exists := c.store[key]; exists {
		c.ll.MoveToFront(element)
		oldItem := element.Value.(*lruItem)

		c.currentBytes -= getSize(oldItem.entry)
		oldItem.entry = value
		c.currentBytes += itemSize
	} else {
		item := &lruItem{key: key, entry: value}
		element := c.ll.PushFront(item)
		c.store[key] = element
		c.currentBytes += itemSize
	}

	for c.currentBytes > c.maxBytes && c.ll.Len() > 0 {
		c.removeOldest()
	}
}

func (c *inAppCache) removeOldest() {
	back := c.ll.Back()
	if back != nil {
		c.removeElement(back)
	}
}

func (c *inAppCache) removeElement(e *list.Element) {
	c.ll.Remove(e)
	item := e.Value.(*lruItem)
	delete(c.store, item.key)
	c.currentBytes -= getSize(item.entry)
}

func (c *inAppCache) janitor(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()

		c.mutex.Lock()
		for e := c.ll.Back(); e != nil; {
			prev := e.Prev()
			item := e.Value.(*lruItem)

			if now.After(item.entry.ExpiresAt) {
				c.removeElement(e)
			}
			e = prev
		}
		c.mutex.Unlock()
	}
}