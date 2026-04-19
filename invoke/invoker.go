package invoke

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
)

// Invoker performs a single gRPC unary call and returns the latency.
type Invoker struct {
	conn    *grpc.ClientConn
	method  string
	payload interface{}
}

// Result holds the outcome of a single invocation.
type Result struct {
	Latency time.Duration
	Err     error
}

// New creates a new Invoker.
func New(conn *grpc.ClientConn, method string, payload interface{}) *Invoker {
	return &Invoker{conn: conn, method: method, payload: payload}
}

// Call executes the gRPC method and returns a Result.
func (inv *Invoker) Call(ctx context.Context) Result {
	if inv.conn == nil {
		return Result{Err: fmt.Errorf("invoker: nil connection")}
	}
	if inv.method == "" {
		return Result{Err: fmt.Errorf("invoker: empty method")}
	}

	var reply interface{}
	start := time.Now()
	err := inv.conn.Invoke(ctx, inv.method, inv.payload, &reply)
	return Result{
		Latency: time.Since(start),
		Err:     err,
	}
}
