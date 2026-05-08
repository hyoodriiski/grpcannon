// Package evict provides a simple LRU-based eviction policy for
// capping the number of tracked keys in a load test run.
package evict

import (
	"container/list"
	"errors"
	"sync"
)

// ErrEvicted is returned when a key is removed due to capacity pressure.
var ErrEvicted = errors.New("evict: key evicted")

type entry struct {
	key   string
	value any
}

// Cache is a concurrency-safe LRU cache with a fixed capacity.
type Cache struct {
	mu       sync.Mutex
	cap      int
	items    map[string]*list.Element
	order    *list.List
	onEvict  func(key string, value any)
}

// New creates a Cache with the given capacity. If cap <= 0 it defaults to 1.
// onEvict is called (if non-nil) whenever an entry is evicted.
func New(cap int, onEvict func(key string, value any)) *Cache {
	if cap <= 0 {
		cap = 1
	}
	return &Cache{
		cap:     cap,
		items:   make(map[string]*list.Element),
		order:   list.New(),
		onEvict: onEvict,
	}
}

// Set inserts or updates key with value, evicting the least-recently-used
// entry when the cache is at capacity.
func (c *Cache) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if el, ok := c.items[key]; ok {
		c.order.MoveToFront(el)
		el.Value.(*entry).value = value
		return
	}

	if c.order.Len() >= c.cap {
		c.evictOldest()
	}

	el := c.order.PushFront(&entry{key: key, value: value})
	c.items[key] = el
}

// Get returns the value for key and marks it as recently used.
// The second return value reports whether the key was present.
func (c *Cache) Get(key string) (any, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	el, ok := c.items[key]
	if !ok {
		return nil, false
	}
	c.order.MoveToFront(el)
	return el.Value.(*entry).value, true
}

// Len returns the current number of entries in the cache.
func (c *Cache) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.order.Len()
}

func (c *Cache) evictOldest() {
	el := c.order.Back()
	if el == nil {
		return
	}
	e := el.Value.(*entry)
	c.order.Remove(el)
	delete(c.items, e.key)
	if c.onEvict != nil {
		c.onEvict(e.key, e.value)
	}
}
