// Package sampler provides request sampling for load tests,
// allowing a fraction of requests to be captured for inspection.
package sampler

import (
	"math/rand"
	"sync"
	"time"
)

// Sample holds a single captured request/response record.
type Sample struct {
	Method    string
	LatencyMs float64
	Error     error
	Timestamp time.Time
}

// Sampler captures a random fraction of call results.
type Sampler struct {
	mu      sync.Mutex
	rate    float64 // 0.0–1.0
	rng     *rand.Rand
	samples []Sample
}

// New creates a Sampler that captures approximately rate*100 % of calls.
// rate is clamped to [0.0, 1.0].
func New(rate float64) *Sampler {
	if rate < 0 {
		rate = 0
	}
	if rate > 1 {
		rate = 1
	}
	return &Sampler{
		rate: rate,
		rng:  rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Record conditionally stores a sample based on the configured rate.
func (s *Sampler) Record(method string, latencyMs float64, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.rate == 0 {
		return
	}
	if s.rng.Float64() <= s.rate {
		s.samples = append(s.samples, Sample{
			Method:    method,
			LatencyMs: latencyMs,
			Error:     err,
			Timestamp: time.Now(),
		})
	}
}

// Samples returns a copy of all captured samples.
func (s *Sampler) Samples() []Sample {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Sample, len(s.samples))
	copy(out, s.samples)
	return out
}

// Len returns the number of captured samples.
func (s *Sampler) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.samples)
}

// Reset clears all captured samples.
func (s *Sampler) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.samples = s.samples[:0]
}
