package stagger_test

import (
	"context"
	"testing"
	"time"

	"grpcannon/stagger"
)

func TestGuard_DelaysFirstCall(t *testing.T) {
	s, _ := stagger.New(100*time.Millisecond, 10) // 10ms per step
	calls := 0
	next := func(ctx context.Context) (time.Duration, error) {
		calls++
		return time.Millisecond, nil
	}
	guarded := stagger.Guard(s, 2, next)
	start := time.Now()
	_, err := guarded(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if time.Since(start) < 15*time.Millisecond {
		t.Fatal("expected stagger delay on first call")
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestGuard_SubsequentCalls_NoDelay(t *testing.T) {
	s, _ := stagger.New(10*time.Second, 1)
	next := func(ctx context.Context) (time.Duration, error) {
		return time.Millisecond, nil
	}
	guarded := stagger.Guard(s, 0, next) // index 0 → no delay
	for i := 0; i < 3; i++ {
		start := time.Now()
		_, err := guarded(context.Background())
		if err != nil {
			t.Fatalf("call %d: unexpected error: %v", i, err)
		}
		if time.Since(start) > 30*time.Millisecond {
			t.Fatalf("call %d took too long — unexpected delay", i)
		}
	}
}

func TestGuard_CancelledContext_ReturnsErr(t *testing.T) {
	s, _ := stagger.New(10*time.Second, 1)
	next := func(ctx context.Context) (time.Duration, error) {
		return time.Millisecond, nil
	}
	guarded := stagger.Guard(s, 1, next)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := guarded(ctx)
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}
