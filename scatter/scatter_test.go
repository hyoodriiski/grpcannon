package scatter_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/grpcannon/scatter"
)

func TestScatter_ZeroN_ReturnsNil(t *testing.T) {
	results := scatter.Scatter(context.Background(), 0, func(_ context.Context) error { return nil })
	if results != nil {
		t.Fatalf("expected nil, got %v", results)
	}
}

func TestScatter_NilFn_ReturnsNil(t *testing.T) {
	results := scatter.Scatter(context.Background(), 3, nil)
	if results != nil {
		t.Fatalf("expected nil, got %v", results)
	}
}

func TestScatter_AllSuccess(t *testing.T) {
	var calls int64
	fn := func(_ context.Context) error {
		atomic.AddInt64(&calls, 1)
		return nil
	}
	results := scatter.Scatter(context.Background(), 5, fn)
	if len(results) != 5 {
		t.Fatalf("expected 5 results, got %d", len(results))
	}
	if atomic.LoadInt64(&calls) != 5 {
		t.Fatalf("expected 5 calls, got %d", calls)
	}
	if sc := scatter.SuccessCount(results); sc != 5 {
		t.Fatalf("expected 5 successes, got %d", sc)
	}
}

func TestScatter_SomeErrors(t *testing.T) {
	sentinel := errors.New("boom")
	fn := func(_ context.Context) error { return sentinel }
	results := scatter.Scatter(context.Background(), 4, fn)
	errs := scatter.Errors(results)
	if len(errs) != 4 {
		t.Fatalf("expected 4 errors, got %d", len(errs))
	}
	for _, e := range errs {
		if !errors.Is(e, sentinel) {
			t.Fatalf("unexpected error: %v", e)
		}
	}
}

func TestScatter_LatencyPopulated(t *testing.T) {
	fn := func(_ context.Context) error {
		time.Sleep(5 * time.Millisecond)
		return nil
	}
	results := scatter.Scatter(context.Background(), 2, fn)
	for _, r := range results {
		if r.Latency < 5*time.Millisecond {
			t.Fatalf("expected latency >= 5ms, got %v", r.Latency)
		}
	}
}

func TestScatter_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	fn := func(c context.Context) error {
		return c.Err()
	}
	results := scatter.Scatter(ctx, 3, fn)
	for _, r := range results {
		if !errors.Is(r.Err, context.Canceled) {
			t.Fatalf("expected Canceled, got %v", r.Err)
		}
	}
}

func TestSuccessCount_Mixed(t *testing.T) {
	results := []scatter.Result{
		{Err: nil},
		{Err: errors.New("x")},
		{Err: nil},
	}
	if sc := scatter.SuccessCount(results); sc != 2 {
		t.Fatalf("expected 2, got %d", sc)
	}
}
