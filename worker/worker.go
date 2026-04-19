package worker

import (
	"context"
	"sync"
	"time"

	"google.golang.org/grpc"
)

// Task represents a single gRPC call to be executed.
type Task struct {
	Method string
	Payload []byte
}

// Result holds the outcome of a single task execution.
type Result struct {
	Latency time.Duration
	Err     error
}

// Invoker is a function that performs a gRPC call.
type Invoker func(ctx context.Context, conn *grpc.ClientConn, task Task) Result

// Pool manages a fixed number of concurrent workers.
type Pool struct {
	conn      *grpc.ClientConn
	concurrency int
	invoker   Invoker
}

// NewPool creates a new worker pool.
func NewPool(conn *grpc.ClientConn, concurrency int, invoker Invoker) *Pool {
	if concurrency < 1 {
		concurrency = 1
	}
	return &Pool{conn: conn, concurrency: concurrency, invoker: invoker}
}

// Run distributes tasks across workers and collects results.
func (p *Pool) Run(ctx context.Context, tasks <-chan Task) <-chan Result {
	results := make(chan Result, p.concurrency)
	var wg sync.WaitGroup
	for i := 0; i < p.concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range tasks {
				select {
				case <-ctx.Done():
					return
				default:
					results <- p.invoker(ctx, p.conn, task)
				}
			}
		}()
	}
	go func() {
		wg.Wait()
		close(results)
	}()
	return results
}
