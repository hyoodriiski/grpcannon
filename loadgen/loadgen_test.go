package loadgen_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/yourorg/grpcannon/loadgen"
)

func successInvoker(_ context.Context) error { return nil }

func errorInvoker(_ context.Context) error { return errors.New("rpc error") }

func TestRun_AllSuccess(t *testing.T) {
	cfg := loadgen.Config{
		Concurrency: 4,
		TotalCalls:  20,
		RatePerSec:  0,
		Timeout:     5 * time.Second,
		Invoker:     successInvoker,
	}
	res, err := loadgen.Run(context.Background(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Errors != 0 {
		t.Errorf("expected 0 errors, got %d", res.Errors)
	}
	if res.Duration <= 0 {
		t.Error("expected positive duration")
	}
}

func TestRun_AllErrors(t *testing.T) {
	cfg := loadgen.Config{
		Concurrency: 2,
		TotalCalls:  10,
		Invoker:     errorInvoker,
	}
	res, err := loadgen.Run(context.Background(), cfg)
	if err != nil {
		t.Fatalf("unexpected orchestration error: %v", err)
	}
	if res.Errors == 0 {
		t.Error("expected some errors to be recorded")
	}
}

func TestRun_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	var calls int64
	invoker := func(ctx context.Context) error {
		atomic.AddInt64(&calls, 1)
		if atomic.LoadInt64(&calls) >= 3 {
			cancel()
		}
		return nil
	}

	cfg := loadgen.Config{
		Concurrency: 1,
		TotalCalls:  1000,
		Invoker:     invoker,
	}
	res, err := loadgen.Run(ctx, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Fewer than 1000 calls should have completed.
	if atomic.LoadInt64(&calls) >= 1000 {
		t.Error("expected run to terminate early after context cancellation")
	}
	_ = res
}

func TestRun_DefaultConcurrency(t *testing.T) {
	cfg := loadgen.Config{
		Concurrency: 0, // should default to 1
		TotalCalls:  5,
		Invoker:     successInvoker,
	}
	res, err := loadgen.Run(context.Background(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil result")
	}
}
