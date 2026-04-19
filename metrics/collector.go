package metrics

import (
	"sync"
	"time"
)

// Collector accumulates per-request latencies and error counts in a
// thread-safe manner during a load test run.
type Collector struct {
	mu        sync.Mutex
	latencies []time.Duration
	errors    int
	total     int
}

// NewCollector returns an initialised Collector.
func NewCollector() *Collector {
	return &Collector{}
}

// Record adds a single observation. errored indicates the request failed.
func (c *Collector) Record(d time.Duration, errored bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.total++
	if errored {
		c.errors++
		return
	}
	c.latencies = append(c.latencies, d)
}

// Snapshot returns an immutable copy of the collected data.
func (c *Collector) Snapshot() Snapshot {
	c.mu.Lock()
	defer c.mu.Unlock()
	lats := make([]time.Duration, len(c.latencies))
	copy(lats, c.latencies)
	return Snapshot{
		Latencies: lats,
		Errors:    c.errors,
		Total:     c.total,
	}
}

// Snapshot is an immutable view of collected metrics.
type Snapshot struct {
	Latencies []time.Duration
	Errors    int
	Total     int
}

// ErrorRate returns the fraction of requests that errored.
func (s Snapshot) ErrorRate() float64 {
	if s.Total == 0 {
		return 0
	}
	return float64(s.Errors) / float64(s.Total)
}
