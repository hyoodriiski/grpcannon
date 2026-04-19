package ratelimit_test

import (
	"context"
	"testing"
	"time"

	"github.com/example/grpcannon/ratelimit"
)

func TestNew_NoLimit(t *testing.T) {
	l := ratelimit.New(0)
	defer l.Stop()
	ctx := context.Background()
	if err := l.Wait(ctx); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestNew_WithLimit_Allows(t *testing.T) {
	l := ratelimit.New(100)
	defer l.Stop()
	ctx := context.Background()
	if err := l.Wait(ctx); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestWait_ContextCancelled(t *testing.T) {
	l := ratelimit.New(1) // 1 rps — slow
	defer l.Stop()
	// drain first tick
	_ = l.Wait(context.Background())

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := l.Wait(ctx)
	if err == nil {
		t.Fatal("expected context error, got nil")
	}
}

func TestWait_RateApproximate(t *testing.T) {
	rps := 200
	l := ratelimit.New(rps)
	defer l.Stop()

	ctx := context.Background()
	start := time.Now()
	for i := 0; i < 10; i++ {
		if err := l.Wait(ctx); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	elapsed := time.Since(start)
	expected := time.Duration(9) * (time.Second / time.Duration(rps))
	if elapsed < expected/2 {
		t.Errorf("rate too fast: elapsed %v, expected >= %v", elapsed, expected/2)
	}
}
