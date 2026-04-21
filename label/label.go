// Package label provides key-value metadata tagging for load test requests.
package label

import (
	"fmt"
	"strings"
	"sync"
)

// Set holds an immutable collection of string key-value labels.
type Set struct {
	mu     sync.RWMutex
	labels map[string]string
}

// New creates an empty label Set.
func New() *Set {
	return &Set{labels: make(map[string]string)}
}

// Add inserts or overwrites a label. Empty keys are silently ignored.
func (s *Set) Add(key, value string) {
	key = strings.TrimSpace(key)
	if key == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.labels[key] = value
}

// Get returns the value for key and whether it was found.
func (s *Set) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.labels[key]
	return v, ok
}

// All returns a shallow copy of all labels.
func (s *Set) All() map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string]string, len(s.labels))
	for k, v := range s.labels {
		out[k] = v
	}
	return out
}

// String returns labels formatted as "k=v,k=v" sorted by key.
func (s *Set) String() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	parts := make([]string, 0, len(s.labels))
	for k, v := range s.labels {
		parts = append(parts, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(parts, ",")
}

// Len returns the number of labels in the set.
func (s *Set) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.labels)
}
