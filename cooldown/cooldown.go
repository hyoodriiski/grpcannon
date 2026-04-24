// Package cooldown provides a per-key cooldown mechanism that prevents
// an action from being taken more than once within a configurable window.
package cooldown

import (
	"sync"
	"time"
)

// Cooldown tracks per-key cooldown state.
type Cooldown struct {
	mu       sync.Mutex
	window   time.Duration
	entries  map[string]time.Time
	nowFn    func() time.Time
}

// New creates a Cooldown with the given window duration.
// A zero or negative window means actions are never throttled.
func New(window time.Duration) *Cooldown {
	return &Cooldown{
		window:  window,
		entries: make(map[string]time.Time),
		nowFn:   time.Now,
	}
}

// Allow reports whether the action identified by key is allowed right now.
// If allowed, the key's cooldown timer is reset. If not allowed, the call
// is a no-op and false is returned.
func (c *Cooldown) Allow(key string) bool {
	if c.window <= 0 {
		return true
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	now := c.nowFn()
	if last, ok := c.entries[key]; ok {
		if now.Sub(last) < c.window {
			return false
		}
	}
	c.entries[key] = now
	return true
}

// Reset removes the cooldown record for key, allowing the next call
// to Allow to succeed regardless of when it last fired.
func (c *Cooldown) Reset(key string) {
	c.mu.Lock()
	delete(c.entries, key)
	c.mu.Unlock()
}

// Remaining returns how much cooldown time is left for key.
// Returns 0 if the key is not in cooldown.
func (c *Cooldown) Remaining(key string) time.Duration {
	if c.window <= 0 {
		return 0
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	last, ok := c.entries[key]
	if !ok {
		return 0
	}
	remaining := c.window - c.nowFn().Sub(last)
	if remaining < 0 {
		return 0
	}
	return remaining
}
