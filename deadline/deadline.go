// Package deadline provides per-request deadline enforcement for gRPC calls.
package deadline

import (
	"context"
	"errors"
	"time"
)

// ErrDeadlineExceeded is returned when a call exceeds its allowed duration.
var ErrDeadlineExceeded = errors.New("deadline: request deadline exceeded")

// Enforcer wraps an invocation function with a per-call deadline.
type Enforcer struct {
	timeout time.Duration
}

// New creates an Enforcer with the given per-call timeout.
// A zero or negative timeout means no deadline is applied.
func New(timeout time.Duration) *Enforcer {
	return &Enforcer{timeout: timeout}
}

// InvokeFunc is the signature of a function that performs a single gRPC call.
type InvokeFunc func(ctx context.Context) error

// Run executes fn with an optional deadline derived from the parent context.
// If the enforcer timeout is positive, a child context with that deadline is
// created. If fn does not return before the deadline, ErrDeadlineExceeded is
// returned and the child context is cancelled.
func (e *Enforcer) Run(parent context.Context, fn InvokeFunc) error {
	if e.timeout <= 0 {
		return fn(parent)
	}

	ctx, cancel := context.WithTimeout(parent, e.timeout)
	defer cancel()

	type result struct {
		err error
	}

	ch := make(chan result, 1)
	go func() {
		ch <- result{err: fn(ctx)}
	}()

	select {
	case r := <-ch:
		return r.err
	case <-ctx.Done():
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return ErrDeadlineExceeded
		}
		return ctx.Err()
	}
}
