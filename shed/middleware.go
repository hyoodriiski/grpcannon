package shed

import (
	"context"
)

// InvokerFunc is a function that performs a gRPC invocation.
type InvokerFunc func(ctx context.Context) error

// Guard wraps an InvokerFunc with load-shedding logic. If the shed rejects
// the request (capacity exceeded), the next invoker is never called and
// ErrShed is returned immediately. On success or downstream error the slot
// is released automatically.
func Guard(s *Shed, next InvokerFunc) InvokerFunc {
	return func(ctx context.Context) error {
		if err := s.Acquire(ctx); err != nil {
			return err
		}
		defer s.Release()
		return next(ctx)
	}
}
