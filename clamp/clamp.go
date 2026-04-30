// Package clamp provides a concurrency limiter that enforces a minimum and
// maximum bound on the number of concurrent in-flight requests, shedding work
// that falls outside the allowed range.
package clamp

import (
	"context"
	"errors"
	"sync/atomic"
)

// ErrBelow is returned when the current concurrency is below the minimum floor.
var ErrBelow = errors.New("clamp: concurrency below minimum floor")

// ErrAbove is returned when the current concurrency exceeds the maximum ceiling.
var ErrAbove = errors.New("clamp: concurrency above maximum ceiling")

// Clamp enforces a [min, max] concurrency window.
type Clamp struct {
	min int64
	max int64
	inflight atomic.Int64
}

// New creates a Clamp with the given min and max bounds.
// If min < 0 it is clamped to 0. If max < min it is set equal to min.
func New(min, max int64) *Clamp {
	if min < 0 {
		min = 0
	}
	if max < min {
		max = min
	}
	return &Clamp{min: min, max: max}
}

// Acquire increments the in-flight counter and checks the bounds.
// It returns ErrAbove when the ceiling is exceeded, or ErrBelow when the
// current count (after increment) is still below the floor.
// On success the caller must call Release when the work is complete.
func (c *Clamp) Acquire(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	n := c.inflight.Add(1)
	if n > c.max {
		c.inflight.Add(-1)
		return ErrAbove
	}
	if n < c.min {
		c.inflight.Add(-1)
		return ErrBelow
	}
	return nil
}

// Release decrements the in-flight counter. It is a no-op if the counter is
// already zero.
func (c *Clamp) Release() {
	if c.inflight.Load() > 0 {
		c.inflight.Add(-1)
	}
}

// InFlight returns the current number of in-flight requests.
func (c *Clamp) InFlight() int64 {
	return c.inflight.Load()
}
