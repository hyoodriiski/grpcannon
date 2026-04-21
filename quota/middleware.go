package quota

import (
	"context"
	"time"
)

// InvokerFn is a function that performs a single gRPC invocation.
type InvokerFn func(ctx context.Context) (time.Duration, error)

// Guard wraps an InvokerFn and enforces the quota before each call.
// If the quota is exceeded the call is skipped and ErrQuotaExceeded is returned.
func Guard(q *Quota, fn InvokerFn) InvokerFn {
	return func(ctx context.Context) (time.Duration, error) {
		if err := q.Acquire(); err != nil {
			return 0, err
		}
		return fn(ctx)
	}
}
