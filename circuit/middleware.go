package circuit

import (
	"context"
	"time"
)

// InvokeFn is the signature of a single gRPC invocation.
type InvokeFn func(ctx context.Context) error

// Guard wraps an InvokeFn with circuit breaker logic.
// If the breaker is open the call is skipped and ErrOpen is returned.
// Successes close the circuit; failures are recorded against the threshold.
func Guard(b *Breaker, fn InvokeFn) InvokeFn {
	return func(ctx context.Context) error {
		if err := b.Allow(); err != nil {
			return err
		}
		err := fn(ctx)
		if err != nil {
			b.RecordFailure()
		} else {
			b.RecordSuccess()
		}
		return err
	}
}

// RunWithBreaker executes fn through a temporary breaker with the given
// threshold and reset window. Useful for one-shot protected calls.
func RunWithBreaker(ctx context.Context, threshold int, resetAfter time.Duration, fn InvokeFn) error {
	b := New(threshold, resetAfter)
	return Guard(b, fn)(ctx)
}
