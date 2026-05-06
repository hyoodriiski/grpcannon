// Package ceiling provides a concurrency limiter that enforces a hard upper
// bound on the number of simultaneous in-flight operations.
package ceiling

import (
	"context"
	"errors"
	"sync/atomic"
)

// ErrAbove is returned when the ceiling has been reached.
var ErrAbove = errors.New("ceiling: limit reached")

// Ceiling enforces a hard cap on concurrent operations.
type Ceiling struct {
	max     int64
	active  atomic.Int64
}

// New creates a Ceiling with the given maximum concurrency.
// If max is zero or negative, it is treated as unlimited (always allows).
func New(max int) *Ceiling {
	if max < 0 {
		max = 0
	}
	return &Ceiling{max: int64(max)}
}

// Acquire attempts to claim one slot. It returns ErrAbove when the ceiling
// is reached, or ctx.Err() if the context is already done.
func (c *Ceiling) Acquire(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if c.max == 0 {
		c.active.Add(1)
		return nil
	}
	for {
		cur := c.active.Load()
		if cur >= c.max {
			return ErrAbove
		}
		if c.active.CompareAndSwap(cur, cur+1) {
			return nil
		}
	}
}

// Release decrements the active counter. Each successful Acquire must be
// paired with exactly one Release.
func (c *Ceiling) Release() {
	if c.active.Load() > 0 {
		c.active.Add(-1)
	}
}

// Active returns the current number of in-flight operations.
func (c *Ceiling) Active() int64 {
	return c.active.Load()
}

// Max returns the configured ceiling. Zero means unlimited.
func (c *Ceiling) Max() int64 {
	return c.max
}
