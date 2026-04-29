package hedge_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"grpcannon/hedge"
)

func TestNew_ZeroDelay_NoHedge(t *testing.T) {
	h := hedge.New(0)
	calls := int32(0)
	err := h.Run(context.Background(), func(_ context.Context) error {
		atomic.AddInt32(&calls, 1)
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestRun_NilInvoker_ReturnsErr(t *testing.T) {
	h := hedge.New(10 * time.Millisecond)
	if err := h.Run(context.Background(), nil); !errors.Is(err, hedge.ErrNoInvoker) {
		t.Fatalf("expected ErrNoInvoker, got %v", err)
	}
}

func TestRun_FirstSucceedsFast_NoHedgeFired(t *testing.T) {
	h := hedge.New(50 * time.Millisecond)
	calls := int32(0)
	err := h.Run(context.Background(), func(_ context.Context) error {
		atomic.AddInt32(&calls, 1)
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	time.Sleep(80 * time.Millisecond)
	if n := atomic.LoadInt32(&calls); n != 1 {
		t.Fatalf("hedge should not have fired; got %d calls", n)
	}
}

func TestRun_HedgeFires_WhenSlowFirstAttempt(t *testing.T) {
	h := hedge.New(20 * time.Millisecond)
	calls := int32(0)
	err := h.Run(context.Background(), func(_ context.Context) error {
		n := atomic.AddInt32(&calls, 1)
		if n == 1 {
			time.Sleep(60 * time.Millisecond)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if atomic.LoadInt32(&calls) < 2 {
		t.Fatal("expected hedge to fire a second attempt")
	}
}

func TestRun_BothFail_ReturnsError(t *testing.T) {
	h := hedge.New(10 * time.Millisecond)
	sentinel := errors.New("rpc error")
	err := h.Run(context.Background(), func(_ context.Context) error {
		return sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}

func TestRun_ContextCancelled_ReturnsErr(t *testing.T) {
	h := hedge.New(5 * time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := h.Run(ctx, func(c context.Context) error {
		return c.Err()
	})
	if err == nil {
		t.Fatal("expected an error for cancelled context")
	}
}

func TestRun_HedgeFires_SecondSucceeds_FirstStillSlow(t *testing.T) {
	// Verify that Run returns as soon as the hedge attempt succeeds,
	// without waiting for the still-running first attempt to finish.
	h := hedge.New(15 * time.Millisecond)
	calls := int32(0)
	start := time.Now()
	err := h.Run(context.Background(), func(_ context.Context) error {
		n := atomic.AddInt32(&calls, 1)
		if n == 1 {
			time.Sleep(200 * time.Millisecond) // first attempt is very slow
		}
		return nil
	})
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The hedge fires at ~15 ms and succeeds immediately, so total time
	// should be well under the 200 ms first-attempt sleep.
	if elapsed >= 150*time.Millisecond {
		t.Fatalf("Run took too long (%v); hedge should have short-circuited", elapsed)
	}
}
