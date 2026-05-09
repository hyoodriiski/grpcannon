// Package cache provides a simple TTL-based in-memory cache for
// storing and retrieving arbitrary values with automatic expiry.
package cache

import (
	"sync"
	"time"
)

// ErrExpired is returned when a key exists but has passed its TTL.
var ErrExpired = errExpired("cache: entry expired")

type errExpired string

func (e errExpired) Error() string { return string(e) }

// ErrNotFound is returned when a key does not exist in the cache.
var ErrNotFound = errNotFound("cache: key not found")

type errNotFound string

func (e errNotFound) Error() string { return string(e) }

type entry struct {
	value     interface{}
	expiresAt time.Time
}

// Cache is a thread-safe TTL cache.
type Cache struct {
	mu      sync.RWMutex
	items   map[string]entry
	default TTL time.Duration
}

// New creates a Cache with the given default TTL.
// A zero TTL means entries never expire.
func New(defaultTTL time.Duration) *Cache {
	return &Cache{
		items:      make(map[string]entry),
		defaultTTL: defaultTTL,
	}
}

// Set stores value under key using the default TTL.
func (c *Cache) Set(key string, value interface{}) {
	c.SetTTL(key, value, c.defaultTTL)
}

// SetTTL stores value under key with an explicit TTL.
func (c *Cache) SetTTL(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	var exp time.Time
	if ttl > 0 {
		exp = time.Now().Add(ttl)
	}
	c.items[key] = entry{value: value, expiresAt: exp}
}

// Get retrieves a value by key. Returns ErrNotFound or ErrExpired on failure.
func (c *Cache) Get(key string) (interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.items[key]
	if !ok {
		return nil, ErrNotFound
	}
	if !e.expiresAt.IsZero() && time.Now().After(e.expiresAt) {
		return nil, ErrExpired
	}
	return e.value, nil
}

// Delete removes a key from the cache.
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

// Flush removes all entries from the cache.
func (c *Cache) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]entry)
}

// Len returns the number of entries currently stored (including expired).
func (c *Cache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}
