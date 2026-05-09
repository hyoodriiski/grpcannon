// Package stagger introduces a configurable delay between successive
// invocations to spread load across time rather than firing all at once.
package stagger

import (
	"context"
	"errors"
	"time"
)

// ErrZeroConcurrency is returned when concurrency is less than one.
var ErrZeroConcurrency = errors.New("stagger: concurrency must be >= 1")

// Stagger holds the configuration for staggered invocation.
type Stagger struct {
	delay time.Duration
}

// New creates a Stagger that spaces workers by (window / concurrency).
// If concurrency is less than 1 an error is returned.
func New(window time.Duration, concurrency int) (*Stagger, error) {
	if concurrency < 1 {
		return nil, ErrZeroConcurrency
	}
	delay := window / time.Duration(concurrency)
	return &Stagger{delay: delay}, nil
}

// Delay returns the per-worker delay.
func (s *Stagger) Delay() time.Duration {
	return s.delay
}

// Wait blocks for index * delay or until ctx is cancelled.
// index 0 returns immediately.
func (s *Stagger) Wait(ctx context.Context, index int) error {
	if index <= 0 || s.delay == 0 {
		return nil
	}
	timer := time.NewTimer(time.Duration(index) * s.delay)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
