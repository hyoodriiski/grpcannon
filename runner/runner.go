package runner

import (
	"context"
	"sync"
	"time"

	"github.com/grpcannon/config"
)

// Result holds the outcome of a single RPC call.
type Result struct {
	Latency time.Duration
	Err     error
}

// Runner executes load against a gRPC target.
type Runner struct {
	cfg     *config.Config
	results []Result
	mu      sync.Mutex
}

// New creates a Runner from the given config.
func New(cfg *config.Config) *Runner {
	return &Runner{cfg: cfg}
}

// Run spawns cfg.Concurrency workers and collects results.
func (r *Runner) Run(ctx context.Context, call func(ctx context.Context) error) []Result {
	jobs := make(chan struct{}, r.cfg.TotalRequests)
	for i := 0; i < r.cfg.TotalRequests; i++ {
		jobs <- struct{}{}
	}
	close(jobs)

	var wg sync.WaitGroup
	for i := 0; i < r.cfg.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range jobs {
				tctx, cancel := context.WithTimeout(ctx, r.cfg.Timeout)
				start := time.Now()
				err := call(tctx)
				latency := time.Since(start)
				cancel()
				r.mu.Lock()
				r.results = append(r.results, Result{Latency: latency, Err: err})
				r.mu.Unlock()
			}
		}()
	}
	wg.Wait()
	return r.results
}
