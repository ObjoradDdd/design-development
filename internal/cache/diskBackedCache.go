package cache

import (
	"bytes"
	"encoding/gob"
	"log/slog"
	"time"

	"go.etcd.io/bbolt"
)

type diskBackedCache struct {
	db *bbolt.DB
}

func NewDiskBackedCache(ttl time.Duration) Cache {
	db, err := bbolt.Open("cache.db", 0600, nil)

	if err != nil {
		panic(err)
	}

	c := &diskBackedCache{
		db: db,
	}

	go c.janitor(ttl)

	return c
}

func (c *diskBackedCache) Get(key string) (CacheEntry, bool) {

	slog.Debug("getting from disk cache for key", "key", key)
	var entry CacheEntry
	found := false

	c.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("CacheEntries"))
		if b == nil {
			return nil
		}
		v := b.Get([]byte(key))
		if v == nil {
			return nil
		}

		var err error
		entry, err = decode(v)
		if err != nil {
			return err
		}

		if time.Now().After(entry.ExpiresAt) {
			return nil
		}

		found = true
		return nil
	})

	return entry, found
}

func (c *diskBackedCache) Set(key string, value CacheEntry, ttl time.Duration) {
	slog.Debug("setting disk cache for key", "key", key)
	bucketName := []byte("CacheEntries")
	value.ExpiresAt = time.Now().Add(ttl)

	err := c.db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(bucketName)
		if err != nil {
			return err
		}

		data, err := encode(value)
		if err != nil {
			return err
		}

		err = b.Put([]byte(key), data)
		return err
	})

	if err != nil {
		slog.Error("Error writing to cache:", err)
	}
}

func (c *diskBackedCache) Close() {
	c.db.Close()
}

func encode(entry CacheEntry) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	if err := enc.Encode(entry); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func decode(data []byte) (CacheEntry, error) {
	var entry CacheEntry

	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)

	if err := dec.Decode(&entry); err != nil {
		return CacheEntry{}, err
	}

	return entry, nil
}

func (c *diskBackedCache) janitor(ttl time.Duration) {
	ticker := time.NewTicker(ttl)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()

		c.db.Update(func(tx *bbolt.Tx) error {
			b := tx.Bucket([]byte("CacheEntries"))
			if b == nil {
				return nil
			}

			cursor := b.Cursor()

			for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
				entry, err := decode(v)

				if err != nil || now.After(entry.ExpiresAt) {
					cursor.Delete()
				}
			}
			return nil
		})
	}
}
