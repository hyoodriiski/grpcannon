// Package snapshot captures a point-in-time view of metrics collected
// during a load test run and exposes helpers to diff successive snapshots.
package snapshot

import (
	"sync"
	"time"
)

// Snapshot holds a frozen view of key counters at a moment in time.
type Snapshot struct {
	CapturedAt  time.Time
	Requests    int64
	Errors      int64
	TotalLatency time.Duration
}

// ErrorRate returns the fraction of requests that failed (0–1).
func (s Snapshot) ErrorRate() float64 {
	if s.Requests == 0 {
		return 0
	}
	return float64(s.Errors) / float64(s.Requests)
}

// AvgLatency returns the mean latency across all recorded requests.
func (s Snapshot) AvgLatency() time.Duration {
	if s.Requests == 0 {
		return 0
	}
	return time.Duration(int64(s.TotalLatency) / s.Requests)
}

// Recorder accumulates raw counters and can produce Snapshots on demand.
type Recorder struct {
	mu           sync.Mutex
	requests     int64
	errors       int64
	totalLatency time.Duration
}

// Record adds a single observation.
func (r *Recorder) Record(latency time.Duration, err bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.requests++
	r.totalLatency += latency
	if err {
		r.errors++
	}
}

// Take returns a Snapshot of the current state.
func (r *Recorder) Take() Snapshot {
	r.mu.Lock()
	defer r.mu.Unlock()
	return Snapshot{
		CapturedAt:   time.Now(),
		Requests:     r.requests,
		Errors:       r.errors,
		TotalLatency: r.totalLatency,
	}
}

// Delta returns the difference between two snapshots (b − a).
func Delta(a, b Snapshot) Snapshot {
	return Snapshot{
		CapturedAt:   b.CapturedAt,
		Requests:     b.Requests - a.Requests,
		Errors:       b.Errors - a.Errors,
		TotalLatency: b.TotalLatency - a.TotalLatency,
	}
}
