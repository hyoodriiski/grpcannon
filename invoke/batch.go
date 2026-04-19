package invoke

import (
	"context"

	"google.golang.org/grpc"
)

// Result holds the outcome of a single RPC call.
type Result struct {
	Latency time.Duration
	Err     error
}

// RunBatch executes n RPC calls sequentially using the provided connection and
// method, returning one Result per call. It is intentionally simple so that
// higher-level workers can parallelise across multiple RunBatch invocations.
func RunBatch(ctx context.Context, conn *grpc.ClientConn, method string, n int) []Result {
	invoker := New(conn)
	results := make([]Result, 0, n)
	for i := 0; i < n; i++ {
		select {
		case <-ctx.Done():
			results = append(results, Result{Err: ctx.Err()})
			continue
		default:
		}
		latency, err := invoker.Call(ctx, method)
		results = append(results, Result{Latency: latency, Err: err})
	}
	return results
}
