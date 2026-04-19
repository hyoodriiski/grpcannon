package invoke

import (
	"context"
	"sync"
)

// BatchConfig controls how many calls are made in parallel.
type BatchConfig struct {
	Concurrency int
	Total       int
}

// RunBatch executes Total calls using Concurrency goroutines and returns all Results.
func RunBatch(ctx context.Context, inv *Invoker, cfg BatchConfig) []Result {
	if cfg.Concurrency <= 0 {
		cfg.Concurrency = 1
	}

	results := make([]Result, cfg.Total)
	sem := make(chan struct{}, cfg.Concurrency)
	var wg sync.WaitGroup

	for i := 0; i < cfg.Total; i++ {
		if ctx.Err() != nil {
			break
		}
		wg.Add(1)
		sem <- struct{}{}
		go func(idx int) {
			defer wg.Done()
			defer func() { <-sem }()
			results[idx] = inv.Call(ctx)
		}(i)
	}

	wg.Wait()
	return results
}
