package flow_test

import (
	"context"
	"testing"
	"time"

	"github.com/example/grpcannon/flow"
)

func TestNew_ZeroRPS_NoDelay(t *testing.T) {
	f := flow.New(0)
	ctx := context.Background()
	start := time.Now()
	if err := f.Wait(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if elapsed := time.Since(start); elapsed > 10*time.Millisecond {
		t.Fatalf("expected no delay, got %v", elapsed)
	}
}

func TestNew_NegativeRPS_NoDelay(t *testing.T) {
	f := flow.New(-5)
	ctx := context.Background()
	start := time.Now()
	if err := f.Wait(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if elapsed := time.Since(start); elapsed > 10*time.Millisecond {
		t.Fatalf("expected no delay, got %v", elapsed)
	}
}

func TestWait_ContextCancelled_ReturnsErr(t *testing.T) {
	f := flow.New(0.1) // very slow: 10 s interval
	// prime the last timestamp so next call must wait
	_ = f.Wait(context.Background())

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	err := f.Wait(ctx)
	if err == nil {
		t.Fatal("expected context error, got nil")
	}
}

func TestClose_ReturnsErrClosed(t *testing.T) {
	f := flow.New(10)
	f.Close()
	err := f.Wait(context.Background())
	if err != flow.ErrClosed {
		t.Fatalf("expected ErrClosed, got %v", err)
	}
}

func TestWait_RateApproximate(t *testing.T) {
	const rps = 50.0
	f := flow.New(rps)
	ctx := context.Background()

	const calls = 5
	start := time.Now()
	for i := 0; i < calls; i++ {
		if err := f.Wait(ctx); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	elapsed := time.Since(start)

	expected := time.Duration(float64(time.Second)/rps) * (calls - 1)
	lo := expected - 20*time.Millisecond
	hi := expected + 60*time.Millisecond
	if elapsed < lo || elapsed > hi {
		t.Fatalf("elapsed %v outside expected range [%v, %v]", elapsed, lo, hi)
	}
}

func TestClose_Idempotent(t *testing.T) {
	f := flow.New(10)
	f.Close()
	f.Close() // must not panic
}
