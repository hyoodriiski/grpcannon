package bulkhead_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"grpcannon/bulkhead"
)

func TestNew_ZeroMax_Unlimited(t *testing.T) {
	b := bulkhead.New(0)
	for i := 0; i < 100; i++ {
		if err := b.Acquire(context.Background()); err != nil {
			t.Fatalf("unexpected error on iteration %d: %v", i, err)
		}
	}
}

func TestAcquire_WithinLimit_Succeeds(t *testing.T) {
	b := bulkhead.New(3)
	for i := 0; i < 3; i++ {
		if err := b.Acquire(context.Background()); err != nil {
			t.Fatalf("expected success, got %v", err)
		}
	}
	if got := b.Active(); got != 3 {
		t.Fatalf("expected active=3, got %d", got)
	}
}

func TestAcquire_ExceedsLimit_ReturnsFull(t *testing.T) {
	b := bulkhead.New(2)
	_ = b.Acquire(context.Background())
	_ = b.Acquire(context.Background())

	if err := b.Acquire(context.Background()); err != bulkhead.ErrFull {
		t.Fatalf("expected ErrFull, got %v", err)
	}
}

func TestRelease_RestoresCapacity(t *testing.T) {
	b := bulkhead.New(1)
	_ = b.Acquire(context.Background())
	b.Release()

	if err := b.Acquire(context.Background()); err != nil {
		t.Fatalf("expected success after release, got %v", err)
	}
}

func TestAcquire_CancelledContext_ReturnsErr(t *testing.T) {
	b := bulkhead.New(5)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := b.Acquire(ctx); err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestClose_PreventsAcquire(t *testing.T) {
	b := bulkhead.New(10)
	b.Close()
	if err := b.Acquire(context.Background()); err == nil {
		t.Fatal("expected error after Close")
	}
}

func TestAcquire_ConcurrentSafe(t *testing.T) {
	const limit = 5
	b := bulkhead.New(limit)
	var wg sync.WaitGroup
	admitted := make(chan struct{}, limit*2)

	for i := 0; i < limit*2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := b.Acquire(context.Background()); err == nil {
				admitted <- struct{}{}
				time.Sleep(10 * time.Millisecond)
				b.Release()
			}
		}()
	}
	wg.Wait()
	close(admitted)
	if b.Active() != 0 {
		t.Fatalf("expected active=0 after all goroutines finished, got %d", b.Active())
	}
}

func TestGuard_RejectsWhenFull(t *testing.T) {
	b := bulkhead.New(1)
	_ = b.Acquire(context.Background()) // fill it

	next := func(_ context.Context, _ string, _ []byte) (time.Duration, error) {
		return time.Millisecond, nil
	}
	guarded := bulkhead.Guard(b, next)
	_, err := guarded(context.Background(), "/svc/Method", nil)
	if err != bulkhead.ErrFull {
		t.Fatalf("expected ErrFull, got %v", err)
	}
}

func TestGuard_CallsNext(t *testing.T) {
	b := bulkhead.New(2)
	called := false
	next := func(_ context.Context, _ string, _ []byte) (time.Duration, error) {
		called = true
		return 5 * time.Millisecond, nil
	}
	guarded := bulkhead.Guard(b, next)
	d, err := guarded(context.Background(), "/svc/Method", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected next to be called")
	}
	if d != 5*time.Millisecond {
		t.Fatalf("unexpected duration: %v", d)
	}
	if b.Active() != 0 {
		t.Fatalf("expected active=0 after call, got %d", b.Active())
	}
}
