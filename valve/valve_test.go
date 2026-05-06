package valve

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestNew_OpenState(t *testing.T) {
	v := New(true)
	if !v.IsOpen() {
		t.Fatal("expected valve to be open")
	}
}

func TestNew_ClosedState(t *testing.T) {
	v := New(false)
	if v.IsOpen() {
		t.Fatal("expected valve to be closed")
	}
}

func TestWait_OpenValve_ReturnsImmediately(t *testing.T) {
	v := New(true)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	if err := v.Wait(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWait_ClosedValve_BlocksUntilOpen(t *testing.T) {
	v := New(false)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(20 * time.Millisecond)
		v.Open()
	}()

	if err := v.Wait(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wg.Wait()
}

func TestWait_ContextCancelled_ReturnsErr(t *testing.T) {
	v := New(false)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := v.Wait(ctx); err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestClose_BlocksAfterOpen(t *testing.T) {
	v := New(true)
	v.Close()
	if v.IsOpen() {
		t.Fatal("expected valve to be closed")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	if err := v.Wait(ctx); err == nil {
		t.Fatal("expected timeout on closed valve")
	}
}

func TestGuard_BlocksAndProceedsOnOpen(t *testing.T) {
	v := New(false)
	called := false
	next := InvokerFn(func(ctx context.Context, method string, payload []byte) (time.Duration, error) {
		called = true
		return time.Millisecond, nil
	})
	guarded := Guard(v, next)

	go func() {
		time.Sleep(20 * time.Millisecond)
		v.Open()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, err := guarded(ctx, "svc/Method", nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected next to be called")
	}
}
