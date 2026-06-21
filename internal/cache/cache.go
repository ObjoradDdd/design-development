package cache

import (
	"net/http"
	"time"
)

type lruItem struct {
	key   string
	entry CacheEntry
}

type CacheEntry struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
	ExpiresAt  time.Time
}

type Cache interface {
	Get(key string) (CacheEntry, bool)
	Set(key string, value CacheEntry, ttl time.Duration)
}