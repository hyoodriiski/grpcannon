package drift

import (
	"context"
	"time"
)

// Invoker is a function that performs a single gRPC-style call.
type Invoker func(ctx context.Context) error

// Guard wraps an Invoker, recording each call's latency in the Detector.
// If the detector reports drift, the underlying error (or ErrDrifted) is
// returned to the caller.
func Guard(d *Detector, next Invoker) Invoker {
	return func(ctx context.Context) error {
		start := time.Now()
		err := next(ctx)
		lat := time.Since(start)

		if driftErr := d.Record(lat); driftErr != nil {
			if err != nil {
				return err
			}
			return driftErr
		}
		return err
	}
}
