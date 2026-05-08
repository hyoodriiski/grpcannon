package pacing_test

import (
	"context"
	"testing"
	"time"

	"grpcannon/pacing"
)

func TestNew_ZeroRPS_NoDelay(t *testing.T) {
	p := pacing.New(0)
	if p.Interval() != 0 {
		t.Fatalf("expected zero interval, got %v", p.Interval())
	}
}

func TestNew_PositiveRPS_SetsInterval(t *testing.T) {
	p := pacing.New(10) // 100ms between requests
	if p.Interval() != 100*time.Millisecond {
		t.Fatalf("expected 100ms, got %v", p.Interval())
	}
}

func TestWait_NoLimit_ReturnsImmediately(t *testing.T) {
	p := pacing.New(0)
	start := time.Now()
	if err := p.Wait(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if elapsed := time.Since(start); elapsed > 10*time.Millisecond {
		t.Fatalf("expected immediate return, took %v", elapsed)
	}
}

func TestWait_ContextCancelled_ReturnsErr(t *testing.T) {
	p := pacing.New(1) // 1 rps → 1s interval
	ctx, cancel := context.WithCancel(context.Background())
	// Seed the pacer so the next call will have to wait.
	_ = p.Wait(ctx)
	cancel()
	err := p.Wait(ctx)
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}

func TestWait_RateApproximate(t *testing.T) {
	const rps = 50.0
	p := pacing.New(rps)
	ctx := context.Background()

	const calls = 5
	start := time.Now()
	for i := 0; i < calls; i++ {
		if err := p.Wait(ctx); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	elapsed := time.Since(start)
	expected := time.Duration(float64(time.Second)/rps) * (calls - 1)
	if elapsed < expected/2 {
		t.Fatalf("finished too fast: elapsed=%v expected>=%v", elapsed, expected/2)
	}
}

func TestGuard_PacesInvocations(t *testing.T) {
	p := pacing.New(0) // no pacing — just verify the chain works
	invoked := false
	next := func(ctx context.Context) (time.Duration, error) {
		invoked = true
		return time.Millisecond, nil
	}
	guarded := pacing.Guard(p, next)
	_, err := guarded(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !invoked {
		t.Fatal("expected next to be invoked")
	}
}

func TestGuard_CancelledContext_SkipsNext(t *testing.T) {
	p := pacing.New(1) // 1 rps → forces a wait after first call
	ctx, cancel := context.WithCancel(context.Background())
	next := func(ctx context.Context) (time.Duration, error) {
		return time.Millisecond, nil
	}
	guarded := pacing.Guard(p, next)
	// First call seeds the pacer.
	_, _ = guarded(ctx)
	cancel()
	_, err := guarded(ctx)
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}
