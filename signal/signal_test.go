package signal

import (
	"context"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

func TestNew_DefaultSignals(t *testing.T) {
	w := New()
	if len(w.signals) != 2 {
		t.Fatalf("expected 2 default signals, got %d", len(w.signals))
	}
}

func TestRegister_NilHandlerIgnored(t *testing.T) {
	w := New()
	w.Register(nil)
	if len(w.handlers) != 0 {
		t.Fatal("nil handler should not be registered")
	}
}

func TestWait_ContextCancelled_CallsHandlers(t *testing.T) {
	w := New(syscall.SIGUSR1)

	var called int32
	w.Register(func(ctx context.Context) {
		atomic.AddInt32(&called, 1)
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	w.Wait(ctx)

	if atomic.LoadInt32(&called) != 1 {
		t.Fatal("handler should have been called once")
	}
}

func TestWait_SignalReceived_CallsAllHandlers(t *testing.T) {
	w := New(syscall.SIGUSR1)

	var count int32
	for i := 0; i < 3; i++ {
		w.Register(func(ctx context.Context) {
			atomic.AddInt32(&count, 1)
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	done := make(chan struct{})
	go func() {
		w.Wait(ctx)
		close(done)
	}()

	// Give the goroutine time to register the signal channel.
	time.Sleep(20 * time.Millisecond)
	syscall.Kill(syscall.Getpid(), syscall.SIGUSR1) //nolint:errcheck

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Wait did not return after signal")
	}

	if atomic.LoadInt32(&count) != 3 {
		t.Fatalf("expected 3 handler calls, got %d", count)
	}
}

func TestGuard_CancelledContext_SkipsNext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var invoked bool
	next := InvokerFn(func(_ context.Context) error {
		invoked = true
		return nil
	})

	guarded := Guard(ctx, next)
	err := guarded(context.Background())

	if err == nil {
		t.Fatal("expected non-nil error when context cancelled")
	}
	if invoked {
		t.Fatal("next should not have been invoked")
	}
}

func TestGuard_ActiveContext_CallsNext(t *testing.T) {
	ctx := context.Background()

	var invoked bool
	next := InvokerFn(func(_ context.Context) error {
		invoked = true
		return nil
	})

	guarded := Guard(ctx, next)
	if err := guarded(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !invoked {
		t.Fatal("next should have been invoked")
	}
}
