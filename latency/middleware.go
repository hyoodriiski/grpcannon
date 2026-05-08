package latency

import (
	"context"
	"time"
)

// InvokerFn is the function signature used by grpcannon invokers.
type InvokerFn func(ctx context.Context) error

// Guard wraps an InvokerFn, recording the call latency into t on every
// invocation regardless of whether the call succeeds or fails.
func Guard(t *Tracker, next InvokerFn) InvokerFn {
	return func(ctx context.Context) error {
		start := time.Now()
		err := next(ctx)
		t.Record(time.Since(start))
		return err
	}
}
