package ticker_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/grpcannon/ticker"
)

func TestNew_DefaultsNegativeInterval(t *testing.T) {
	// Should not panic and should use 1s fallback.
	tk := ticker.New(-1*time.Second, func() {})
	if tk == nil {
		t.Fatal("expected non-nil ticker")
	}
}

func TestRun_CallsFnPeriodically(t *testing.T) {
	var count int64
	tk := ticker.New(20*time.Millisecond, func() {
		atomic.AddInt64(&count, 1)
	})

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Millisecond)
	defer cancel()

	go tk.Run(ctx)
	<-ctx.Done()
	tk.Stop()

	got := atomic.LoadInt64(&count)
	if got < 2 {
		t.Fatalf("expected at least 2 ticks, got %d", got)
	}
}

func TestStop_HaltsLoop(t *testing.T) {
	var count int64
	tk := ticker.New(10*time.Millisecond, func() {
		atomic.AddInt64(&count, 1)
	})

	ctx := context.Background()
	go tk.Run(ctx)
	time.Sleep(35 * time.Millisecond)
	tk.Stop()

	before := atomic.LoadInt64(&count)
	time.Sleep(30 * time.Millisecond)
	after := atomic.LoadInt64(&count)

	if after != before {
		t.Fatalf("ticker continued after Stop: before=%d after=%d", before, after)
	}
}

func TestStop_Idempotent(t *testing.T) {
	tk := ticker.New(50*time.Millisecond, func() {})
	ctx, cancel := context.WithCancel(context.Background())
	go tk.Run(ctx)
	cancel()
	// Calling Stop multiple times must not panic.
	tk.Stop()
	tk.Stop()
}

func TestRun_ContextCancelled_Exits(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	tk := ticker.New(10*time.Millisecond, func() {})

	done := make(chan struct{})
	go func() {
		tk.Run(ctx)
		close(done)
	}()

	cancel()
	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Run did not exit after context cancellation")
	}
}
