package coalesce_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"grpcannon/coalesce"
)

func TestDo_SingleCaller_ReturnsValue(t *testing.T) {
	g := coalesce.New()
	v, err := g.Do(context.Background(), "k", func(_ context.Context) (interface{}, error) {
		return 42, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.(int) != 42 {
		t.Fatalf("expected 42, got %v", v)
	}
}

func TestDo_SingleCaller_ReturnsError(t *testing.T) {
	g := coalesce.New()
	sentinel := errors.New("boom")
	_, err := g.Do(context.Background(), "k", func(_ context.Context) (interface{}, error) {
		return nil, sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}

func TestDo_ConcurrentCallers_OnlyOneInvocation(t *testing.T) {
	g := coalesce.New()
	var invocations atomic.Int64
	start := make(chan struct{})

	const n = 20
	var wg sync.WaitGroup
	wg.Add(n)
	results := make([]interface{}, n)
	errs := make([]error, n)

	for i := 0; i < n; i++ {
		i := i
		go func() {
			defer wg.Done()
			<-start
			v, err := g.Do(context.Background(), "shared", func(_ context.Context) (interface{}, error) {
				invocations.Add(1)
				time.Sleep(20 * time.Millisecond)
				return 99, nil
			})
			results[i] = v
			errs[i] = err
		}()
	}

	close(start)
	wg.Wait()

	if inv := invocations.Load(); inv > 3 {
		// Allow a small number due to timing but not all 20.
		t.Fatalf("too many invocations: %d", inv)
	}
	for i, v := range results {
		if errs[i] != nil {
			t.Errorf("goroutine %d got error: %v", i, errs[i])
		}
		if v.(int) != 99 {
			t.Errorf("goroutine %d got %v, want 99", i, v)
		}
	}
}

func TestDo_ContextCancelled_ReturnsErr(t *testing.T) {
	g := coalesce.New()
	blocked := make(chan struct{})

	// Start a long-running call.
	go func() {
		g.Do(context.Background(), "slow", func(_ context.Context) (interface{}, error) { //nolint
			close(blocked)
			time.Sleep(200 * time.Millisecond)
			return nil, nil
		})
	}()

	<-blocked

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := g.Do(ctx, "slow", func(_ context.Context) (interface{}, error) {
		return nil, nil
	})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected DeadlineExceeded, got %v", err)
	}
}

func TestInFlight_CountsActiveKeys(t *testing.T) {
	g := coalesce.New()
	ready := make(chan struct{})
	done := make(chan struct{})

	go func() {
		g.Do(context.Background(), "x", func(_ context.Context) (interface{}, error) { //nolint
			close(ready)
			<-done
			return nil, nil
		})
	}()

	<-ready
	if n := g.InFlight(); n != 1 {
		t.Fatalf("expected 1 in-flight, got %d", n)
	}
	close(done)
	time.Sleep(20 * time.Millisecond)
	if n := g.InFlight(); n != 0 {
		t.Fatalf("expected 0 in-flight after completion, got %d", n)
	}
}
