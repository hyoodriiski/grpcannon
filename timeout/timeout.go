// Package timeout provides per-request deadline enforcement for gRPC calls.
package timeout

import (
	"context"
	"fmt"
	"time"
)

// ErrDeadlineExceeded is returned when a call exceeds its allowed duration.
var ErrDeadlineExceeded = fmt.Errorf("timeout: deadline exceeded")

// Enforcer wraps a timeout duration and applies it to contexts.
type Enforcer struct {
	duration time.Duration
}

// New creates an Enforcer with the given timeout.
// If d is zero or negative, no deadline is applied.
func New(d time.Duration) *Enforcer {
	return &Enforcer{duration: d}
}

// Apply returns a child context with the configured deadline applied.
// If the Enforcer has no positive duration, the original context is returned
// along with a no-op cancel function.
func (e *Enforcer) Apply(ctx context.Context) (context.Context, context.CancelFunc) {
	if e.duration <= 0 {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, e.duration)
}

// Wrap executes fn inside a context bounded by the enforcer's deadline.
// It returns ErrDeadlineExceeded if the context times out.
func (e *Enforcer) Wrap(ctx context.Context, fn func(ctx context.Context) error) error {
	ctx, cancel := e.Apply(ctx)
	defer cancel()

	ch := make(chan error, 1)
	go func() {
		ch <- fn(ctx)
	}()

	select {
	case err := <-ch:
		return err
	case <-ctx.Done():
		return ErrDeadlineExceeded
	}
}
