// Package quota provides a token-based request quota limiter.
// It tracks how many requests have been made and rejects new ones
// once the configured maximum is reached.
package quota

import (
	"errors"
	"sync"
	"sync/atomic"
)

// ErrQuotaExceeded is returned when the quota has been exhausted.
var ErrQuotaExceeded = errors.New("quota: limit exceeded")

// Quota tracks and enforces a maximum number of allowed operations.
type Quota struct {
	mu      sync.Mutex
	max     int64
	used    atomic.Int64
	closed  bool
}

// New creates a Quota that allows at most max operations.
// A max of zero means unlimited.
func New(max int64) *Quota {
	if max < 0 {
		max = 0
	}
	return &Quota{max: max}
}

// Acquire attempts to consume one unit of quota.
// Returns ErrQuotaExceeded if the limit has been reached.
func (q *Quota) Acquire() error {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.closed {
		return ErrQuotaExceeded
	}
	if q.max == 0 {
		q.used.Add(1)
		return nil
	}
	if q.used.Load() >= q.max {
		return ErrQuotaExceeded
	}
	q.used.Add(1)
	return nil
}

// Used returns the number of quota units consumed so far.
func (q *Quota) Used() int64 {
	return q.used.Load()
}

// Remaining returns how many units are left, or -1 if unlimited.
func (q *Quota) Remaining() int64 {
	if q.max == 0 {
		return -1
	}
	r := q.max - q.used.Load()
	if r < 0 {
		return 0
	}
	return r
}

// Close prevents any further acquisitions.
func (q *Quota) Close() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.closed = true
}
