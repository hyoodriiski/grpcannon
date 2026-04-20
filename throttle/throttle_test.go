package throttle

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestNew_InvalidCap(t *testing.T) {
	_, err := New(0)
	if err == nil {
		t.Fatal("expected error for zero cap")
	}
}

func TestNew_ValidCap(t *testing.T) {
	th, err := New(5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if th.Cap() != 5 {
		t.Fatalf("expected cap 5, got %d", th.Cap())
	}
}

func TestAcquireRelease(t *testing.T) {
	th, _ := New(2)
	ctx := context.Background()

	if err := th.Acquire(ctx); err != nil {
		t.Fatalf("acquire 1: %v", err)
	}
	if err := th.Acquire(ctx); err != nil {
		t.Fatalf("acquire 2: %v", err)
	}
	if th.InFlight() != 2 {
		t.Fatalf("expected 2 in-flight, got %d", th.InFlight())
	}
	th.Release()
	if th.InFlight() != 1 {
		t.Fatalf("expected 1 in-flight after release")
	}
}

func TestAcquire_ContextCancelled(t *testing.T) {
	th, _ := New(1)
	ctx := context.Background()
	_ = th.Acquire(ctx) // fill the slot

	cancel, cancelFn := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancelFn()

	err := th.Acquire(cancel)
	if err == nil {
		t.Fatal("expected context error when throttle full")
	}
}

func TestThrottle_ConcurrentSafe(t *testing.T) {
	th, _ := New(4)
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = th.Acquire(context.Background())
			time.Sleep(5 * time.Millisecond)
			th.Release()
		}()
	}
	wg.Wait()
	if th.InFlight() != 0 {
		t.Fatalf("expected 0 in-flight after all goroutines done")
	}
}
