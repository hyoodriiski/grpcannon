// Package tag provides a simple key-value tagging system for annotating
// load test requests and results with arbitrary metadata.
package tag

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

// Tags holds an immutable snapshot of key-value pairs.
type Tags map[string]string

// String returns a deterministic, human-readable representation.
func (t Tags) String() string {
	keys := make([]string, 0, len(t))
	for k := range t {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, t[k]))
	}
	return strings.Join(parts, ",")
}

// Bag is a concurrency-safe mutable collection of tags.
type Bag struct {
	mu   sync.RWMutex
	data map[string]string
}

// New returns an empty Bag.
func New() *Bag {
	return &Bag{data: make(map[string]string)}
}

// Set adds or overwrites a tag. Empty keys are ignored.
func (b *Bag) Set(key, value string) {
	if key == "" {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.data[key] = value
}

// Get returns the value for key and whether it was found.
func (b *Bag) Get(key string) (string, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	v, ok := b.data[key]
	return v, ok
}

// Delete removes a tag by key.
func (b *Bag) Delete(key string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.data, key)
}

// Snapshot returns an immutable copy of the current tags.
func (b *Bag) Snapshot() Tags {
	b.mu.RLock()
	defer b.mu.RUnlock()
	out := make(Tags, len(b.data))
	for k, v := range b.data {
		out[k] = v
	}
	return out
}

// Len returns the number of tags currently stored.
func (b *Bag) Len() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.data)
}
