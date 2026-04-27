// Package warmup provides a pre-load warm-up phase before the main run.
package warmup

import (
	"context"
	"sync"
	"time"
)

// Config holds warm-up configuration.
type Config struct {
	Duration    time.Duration
	Concurrency int
}

// Invoker is the function signature used to invoke a single RPC.
type Invoker func(ctx context.Context) error

// Run executes the warm-up phase by invoking fn concurrently for the
// specified duration. Errors during warm-up are silently discarded.
func Run(ctx context.Context, cfg Config, fn Invoker) {
	if cfg.Duration <= 0 || cfg.Concurrency <= 0 {
		return
	}

	deadline, cancel := context.WithTimeout(ctx, cfg.Duration)
	defer cancel()

	work := make(chan struct{})
	go func() {
		defer close(work)
		for {
			select {
			case <-deadline.Done():
				return
			case work <- struct{}{}:
			}
		}
	}()

	var wg sync.WaitGroup
	for i := 0; i < cfg.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range work {
				_ = fn(deadline)
			}
		}()
	}

	wg.Wait()
}
