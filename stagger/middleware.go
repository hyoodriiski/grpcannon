package stagger

import (
	"context"
	"time"
)

// InvokerFn is the signature used throughout grpcannon for a single call.
type InvokerFn func(ctx context.Context) (time.Duration, error)

// Guard wraps next with a staggered delay based on the worker index before
// the first invocation. Subsequent calls pass through without delay.
func Guard(s *Stagger, index int, next InvokerFn) InvokerFn {
	var fired bool
	return func(ctx context.Context) (time.Duration, error) {
		if !fired {
			fired = true
			if err := s.Wait(ctx, index); err != nil {
				return 0, err
			}
		}
		return next(ctx)
	}
}
