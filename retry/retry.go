package retry

import (
	"context"
	"time"
)

// Policy defines retry behaviour for failed gRPC calls.
type Policy struct {
	MaxAttempts int
	Delay       time.Duration
}

// Default returns a Policy with sensible defaults.
func Default() Policy {
	return Policy{
		MaxAttempts: 3,
		Delay:       50 * time.Millisecond,
	}
}

// Do executes fn up to p.MaxAttempts times, returning the last error on
// exhaustion. It respects context cancellation between attempts.
func (p Policy) Do(ctx context.Context, fn func() error) error {
	if p.MaxAttempts <= 0 {
		return fn()
	}
	var err error
	for i := 0; i < p.MaxAttempts; i++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		err = fn()
		if err == nil {
			return nil
		}
		if i < p.MaxAttempts-1 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(p.Delay):
			}
		}
	}
	return err
}
