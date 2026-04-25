package bulkhead

import (
	"context"
	"time"
)

// InvokeFunc represents a single gRPC invocation.
type InvokeFunc func(ctx context.Context, method string, payload []byte) (time.Duration, error)

// Guard wraps next with bulkhead protection. If the bulkhead is full the call
// is rejected immediately without invoking next.
func Guard(b *Bulkhead, next InvokeFunc) InvokeFunc {
	return func(ctx context.Context, method string, payload []byte) (time.Duration, error) {
		if err := b.Acquire(ctx); err != nil {
			return 0, err
		}
		defer b.Release()
		return next(ctx, method, payload)
	}
}
