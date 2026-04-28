// Package shedder implements adaptive load shedding based on a rolling
// error-rate window. When the observed error rate exceeds a configurable
// threshold the shedder rejects new requests until the rate recovers.
package shedder

import (
	"context"
	"errors"
	"sync"
	"time"
)

// ErrShed is returned when a request is rejected by the shedder.
var ErrShed = errors.New("shedder: request shed due to high error rate")

// Shedder tracks a rolling error rate and sheds load when the rate is too high.
type Shedder struct {
	mu        sync.Mutex
	threshold float64   // 0–1; shed when error rate exceeds this
	window    time.Duration
	samples   []sample
}

type sample struct {
	at  time.Time
	err bool
}

// New returns a Shedder that sheds requests when the error rate inside
// window exceeds threshold (0.0–1.0). A threshold of 0 disables shedding.
func New(threshold float64, window time.Duration) *Shedder {
	if threshold < 0 {
		threshold = 0
	}
	if threshold > 1 {
		threshold = 1
	}
	if window <= 0 {
		window = 10 * time.Second
	}
	return &Shedder{threshold: threshold, window: window}
}

// Allow returns nil when the request should proceed, or ErrShed when it
// should be rejected. ctx is reserved for future deadline-aware behaviour.
func (s *Shedder) Allow(_ context.Context) error {
	if s.threshold == 0 {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.trim(time.Now())
	if rate := s.rate(); rate > s.threshold {
		return ErrShed
	}
	return nil
}

// Record registers the outcome of a completed request.
func (s *Shedder) Record(err bool) {
	s.mu.Lock()
	s.samples = append(s.samples, sample{at: time.Now(), err: err})
	s.mu.Unlock()
}

// Rate returns the current error rate (0–1) over the configured window.
func (s *Shedder) Rate() float64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.trim(time.Now())
	return s.rate()
}

func (s *Shedder) rate() float64 {
	if len(s.samples) == 0 {
		return 0
	}
	var errs int
	for _, sp := range s.samples {
		if sp.err {
			errs++
		}
	}
	return float64(errs) / float64(len(s.samples))
}

func (s *Shedder) trim(now time.Time) {
	cutoff := now.Add(-s.window)
	i := 0
	for i < len(s.samples) && s.samples[i].at.Before(cutoff) {
		i++
	}
	s.samples = s.samples[i:]
}
