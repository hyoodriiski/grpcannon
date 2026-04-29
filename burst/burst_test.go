package burst_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/your-org/grpcannon/burst"
)

func TestNew_ClampsInvalidCap(t *testing.T) {
	l := burst.New(0, 10)
	if l == nil {
		t.Fatal("expected non-nil limiter")
	}
	// Should have at least 1 token available.
	if l.Tokens() < 1 {
		t.Errorf("expected at least 1 token, got %f", l.Tokens())
	}
}

func TestAcquire_WithinBurst_Succeeds(t *testing.T) {
	l := burst.New(3, 1)
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		if err := l.Acquire(ctx); err != nil {
			t.Fatalf("acquire %d: unexpected error: %v", i, err)
		}
	}
}

func TestAcquire_ExceedsBurst_ReturnsErr(t *testing.T) {
	l := burst.New(2, 1)
	ctx := context.Background()
	_ = l.Acquire(ctx)
	_ = l.Acquire(ctx)
	if err := l.Acquire(ctx); err != burst.ErrExceeded {
		t.Fatalf("expected ErrExceeded, got %v", err)
	}
}

func TestAcquire_CancelledContext_ReturnsErr(t *testing.T) {
	l := burst.New(5, 10)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := l.Acquire(ctx); err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestTokens_RefillsOverTime(t *testing.T) {
	// Use a real small sleep to confirm refill logic runs without panic.
	l := burst.New(2, 100) // 100 tokens/sec
	ctx := context.Background()
	_ = l.Acquire(ctx)
	_ = l.Acquire(ctx)
	time.Sleep(20 * time.Millisecond)
	if l.Tokens() <= 0 {
		t.Error("expected tokens to refill after sleep")
	}
}

func TestAcquire_ConcurrentSafe(t *testing.T) {
	l := burst.New(50, 1000)
	ctx := context.Background()
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = l.Acquire(ctx)
		}()
	}
	wg.Wait()
}
