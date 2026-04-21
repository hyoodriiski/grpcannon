// Package window provides a sliding time-window counter for tracking
// request rates and error counts over a rolling duration.
package window

import (
	"sync"
	"time"
)

// bucket holds counts for a discrete time slice.
type bucket struct {
	at     time.Time
	count  int64
	errors int64
}

// Window is a sliding-window counter with a fixed number of buckets.
type Window struct {
	mu       sync.Mutex
	buckets  []bucket
	size     int
	resol    time.Duration // resolution per bucket
}

// New creates a Window that spans duration d divided into n buckets.
// Panics if n < 1 or d <= 0.
func New(d time.Duration, n int) *Window {
	if n < 1 {
		panic("window: bucket count must be >= 1")
	}
	if d <= 0 {
		panic("window: duration must be positive")
	}
	return &Window{
		buckets: make([]bucket, n),
		size:    n,
		resol:   d / time.Duration(n),
	}
}

// Record increments the request count (and optionally the error count) for now.
func (w *Window) Record(isErr bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	now := time.Now()
	b := w.currentBucket(now)
	b.count++
	if isErr {
		b.errors++
	}
}

// Counts returns the total requests and errors within the active window.
func (w *Window) Counts() (requests, errors int64) {
	w.mu.Lock()
	defer w.mu.Unlock()
	cutoff := time.Now().Add(-w.resol * time.Duration(w.size))
	for i := range w.buckets {
		if w.buckets[i].at.After(cutoff) {
			requests += w.buckets[i].count
			errors += w.buckets[i].errors
		}
	}
	return
}

// currentBucket returns a pointer to the bucket for now, resetting stale ones.
func (w *Window) currentBucket(now time.Time) *bucket {
	slot := int(now.UnixNano()/int64(w.resol)) % w.size
	b := &w.buckets[slot]
	// reset if the bucket belongs to an earlier window rotation
	if now.Sub(b.at) >= w.resol {
		*b = bucket{at: now}
	}
	return b
}
