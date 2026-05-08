// Package mirror duplicates each recorded result to one or more secondary
// collectors, enabling side-by-side comparison of load test runs.
package mirror

import (
	"sync"
	"time"
)

// Result holds the outcome of a single RPC invocation.
type Result struct {
	Latency time.Duration
	Err     error
}

// Sink is any type that can receive a Result.
type Sink interface {
	Record(r Result)
}

// Mirror fans out every recorded Result to a primary Sink and zero or more
// secondary Sinks.  All writes are protected by a mutex so Mirror is safe
// for concurrent use.
type Mirror struct {
	mu        sync.Mutex
	primary   Sink
	secondary []Sink
}

// New returns a Mirror that writes to primary and any additional sinks.
// A nil primary is accepted; Record becomes a no-op for the primary slot.
func New(primary Sink, secondary ...Sink) *Mirror {
	filtered := make([]Sink, 0, len(secondary))
	for _, s := range secondary {
		if s != nil {
			filtered = append(filtered, s)
		}
	}
	return &Mirror{primary: primary, secondary: filtered}
}

// Add appends a Sink to the secondary list.  Nil sinks are ignored.
func (m *Mirror) Add(s Sink) {
	if s == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.secondary = append(m.secondary, s)
}

// Record delivers r to the primary Sink and every secondary Sink.
func (m *Mirror) Record(r Result) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.primary != nil {
		m.primary.Record(r)
	}
	for _, s := range m.secondary {
		s.Record(r)
	}
}

// Len returns the total number of sinks (primary + secondary).
func (m *Mirror) Len() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	n := len(m.secondary)
	if m.primary != nil {
		n++
	}
	return n
}
