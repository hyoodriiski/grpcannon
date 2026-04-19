package warmup_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/example/grpcannon/warmup"
)

func TestRun_ZeroDuration_NoInvocations(t *testing.T) {
	var count int64
	warmup.Run(context.Background(), warmup.Config{Duration: 0, Concurrency: 4}, func(_ context.Context) error {
		atomic.AddInt64(&count, 1)
		return nil
	})
	if count != 0 {
		t.Fatalf("expected 0 invocations, got %d", count)
	}
}

func TestRun_ZeroConcurrency_NoInvocations(t *testing.T) {
	var count int64
	warmup.Run(context.Background(), warmup.Config{Duration: 50 * time.Millisecond, Concurrency: 0}, func(_ context.Context) error {
		atomic.AddInt64(&count, 1)
		return nil
	})
	if count != 0 {
		t.Fatalf("expected 0 invocations, got %d", count)
	}
}

func TestRun_InvokesAtLeastOnce(t *testing.T) {
	var count int64
	warmup.Run(context.Background(), warmup.Config{Duration: 100 * time.Millisecond, Concurrency: 2}, func(_ context.Context) error {
		atomic.AddInt64(&count, 1)
		return nil
	})
	if count == 0 {
		t.Fatal("expected at least one invocation")
	}
}

func TestRun_ErrorsIgnored(t *testing.T) {
	warmup.Run(context.Background(), warmup.Config{Duration: 80 * time.Millisecond, Concurrency: 2}, func(_ context.Context) error {
		return context.DeadlineExceeded
	})
	// should return without panic
}

func TestRun_RespectsParentCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var count int64
	warmup.Run(ctx, warmup.Config{Duration: 200 * time.Millisecond, Concurrency: 2}, func(_ context.Context) error {
		atomic.AddInt64(&count, 1)
		return nil
	})
	// parent already cancelled; invocations should be zero or very few
	if count > 5 {
		t.Fatalf("expected near-zero invocations with cancelled context, got %d", count)
	}
}
