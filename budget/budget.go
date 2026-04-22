// Package budget provides an error budget tracker for load testing.
// It tracks the ratio of errors to total requests and signals when
// the configured error threshold has been exceeded.
package budget

import (
	"errors"
	"sync"
	"sync/atomic"
)

// ErrExhausted is returned when the error budget has been exhausted.
var ErrExhausted = errors.New("budget: error budget exhausted")

// Budget tracks errors against a configurable threshold ratio.
type Budget struct {
	mu        sync.Mutex
	threshold float64 // 0.0–1.0: max allowed error rate
	total     atomic.Int64
	errors    atomic.Int64
	exhausted atomic.Bool
}

// New creates a Budget with the given error rate threshold (0.0–1.0).
// A threshold of 0.05 means 5% errors are tolerated before exhaustion.
func New(threshold float64) *Budget {
	if threshold < 0 {
		threshold = 0
	}
	if threshold > 1 {
		threshold = 1
	}
	return &Budget{threshold: threshold}
}

// Record records a request outcome. If errored is true, the error
// counter is incremented. Returns ErrExhausted if the budget is spent.
func (b *Budget) Record(errored bool) error {
	b.total.Add(1)
	if errored {
		b.errors.Add(1)
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	total := b.total.Load()
	errs := b.errors.Load()
	if total > 0 && float64(errs)/float64(total) > b.threshold {
		b.exhausted.Store(true)
		return ErrExhausted
	}
	return nil
}

// Exhausted reports whether the error budget has been exhausted.
func (b *Budget) Exhausted() bool {
	return b.exhausted.Load()
}

// Rate returns the current error rate as a fraction (0.0–1.0).
func (b *Budget) Rate() float64 {
	total := b.total.Load()
	if total == 0 {
		return 0
	}
	return float64(b.errors.Load()) / float64(total)
}

// Reset clears all counters and the exhausted flag.
func (b *Budget) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.total.Store(0)
	b.errors.Store(0)
	b.exhausted.Store(false)
}
