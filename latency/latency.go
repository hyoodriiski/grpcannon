// Package latency provides a fixed-size ring buffer for tracking
// per-request latency samples with percentile computation.
package latency

import (
	"math"
	"sort"
	"sync"
	"time"
)

// Tracker records latency samples in a bounded ring buffer.
type Tracker struct {
	mu      sync.Mutex
	buf     []time.Duration
	cap     int
	pos     int
	full    bool
}

// New returns a Tracker that retains at most cap samples.
// If cap is <= 0 it is set to 1024.
func New(cap int) *Tracker {
	if cap <= 0 {
		cap = 1024
	}
	return &Tracker{
		buf: make([]time.Duration, cap),
		cap: cap,
	}
}

// Record adds a latency sample.
func (t *Tracker) Record(d time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.buf[t.pos] = d
	t.pos = (t.pos + 1) % t.cap
	if t.pos == 0 {
		t.full = true
	}
}

// Samples returns a sorted copy of all recorded samples.
func (t *Tracker) Samples() []time.Duration {
	t.mu.Lock()
	defer t.mu.Unlock()
	size := t.pos
	if t.full {
		size = t.cap
	}
	out := make([]time.Duration, size)
	copy(out, t.buf[:size])
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

// Percentile returns the p-th percentile (0–100) of recorded samples.
// Returns 0 if no samples have been recorded.
func (t *Tracker) Percentile(p float64) time.Duration {
	samples := t.Samples()
	if len(samples) == 0 {
		return 0
	}
	idx := int(math.Ceil(p/100.0*float64(len(samples)))) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(samples) {
		idx = len(samples) - 1
	}
	return samples[idx]
}

// Reset clears all recorded samples.
func (t *Tracker) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.pos = 0
	t.full = false
}
