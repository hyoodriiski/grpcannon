package shed

import (
	"context"
)

// Invoker is a function that performs a single gRPC call.
type Invoker func(ctx context.Context) error

// Guard wraps an Invoker with load-shedding logic. If the shed rejects the
// request the invoker is not called and ErrShed is returned immediately.
func Guard(s *Shed, next Invoker) Invoker {
	return func(ctx context.Context) error {
		if err := s.Acquire(ctx); err != nil {
			return err
		}
		defer s.Release()
		return next(ctx)
	}
}
