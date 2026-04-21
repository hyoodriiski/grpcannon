package fence_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"grpcannon/fence"
)

func TestNew_InvalidMax(t *testing.T) {
	_, err := fence.New(0)
	if err == nil {
		t.Fatal("expected error for zero max")
	}
}

func TestNew_ValidMax(t *testing.T) {
	f, err := fence.New(3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.Active() != 0 {
		t.Fatalf("expected 0 active, got %d", f.Active())
	}
}

func TestAcquireRelease(t *testing.T) {
	f, _ := fence.New(2)
	ctx := context.Background()

	if err := f.Acquire(ctx); err != nil {
		t.Fatalf("first acquire failed: %v", err)
	}
	if err := f.Acquire(ctx); err != nil {
		t.Fatalf("second acquire failed: %v", err)
	}
	if f.Active() != 2 {
		t.Fatalf("expected 2 active, got %d", f.Active())
	}
	f.Release()
	if f.Active() != 1 {
		t.Fatalf("expected 1 active after release, got %d", f.Active())
	}
	f.Release()
}

func TestAcquire_ContextCancelled(t *testing.T) {
	f, _ := fence.New(1)
	ctx := context.Background()

	// Fill the single slot.
	_ = f.Acquire(ctx)

	ctx2, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	err := f.Acquire(ctx2)
	if err == nil {
		t.Fatal("expected context cancellation error")
	}
	f.Release()
}

func TestClose_RejectsNewAcquires(t *testing.T) {
	f, _ := fence.New(2)
	f.Close()

	err := f.Acquire(context.Background())
	if err != fence.ErrFenceClosed {
		t.Fatalf("expected ErrFenceClosed, got %v", err)
	}
}

func TestFence_ConcurrentSafe(t *testing.T) {
	const workers = 20
	const max = 5

	f, _ := fence.New(max)
	ctx := context.Background()

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := f.Acquire(ctx); err != nil {
				return
			}
			time.Sleep(2 * time.Millisecond)
			f.Release()
		}()
	}
	wg.Wait()

	if f.Active() != 0 {
		t.Fatalf("expected 0 active after all goroutines finished, got %d", f.Active())
	}
}
