package ceiling

import (
	"context"
	"sync"
	"testing"
)

func TestNew_ZeroMax_Unlimited(t *testing.T) {
	c := New(0)
	if c.Max() != 0 {
		t.Fatalf("expected max 0, got %d", c.Max())
	}
	for i := 0; i < 100; i++ {
		if err := c.Acquire(context.Background()); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
}

func TestNew_NegativeMax_TreatedAsUnlimited(t *testing.T) {
	c := New(-5)
	if c.Max() != 0 {
		t.Fatalf("expected max 0, got %d", c.Max())
	}
}

func TestAcquire_WithinLimit_Succeeds(t *testing.T) {
	c := New(3)
	for i := 0; i < 3; i++ {
		if err := c.Acquire(context.Background()); err != nil {
			t.Fatalf("acquire %d failed: %v", i, err)
		}
	}
	if c.Active() != 3 {
		t.Fatalf("expected active=3, got %d", c.Active())
	}
}

func TestAcquire_ExceedsLimit_ReturnsErrAbove(t *testing.T) {
	c := New(2)
	_ = c.Acquire(context.Background())
	_ = c.Acquire(context.Background())
	if err := c.Acquire(context.Background()); err != ErrAbove {
		t.Fatalf("expected ErrAbove, got %v", err)
	}
}

func TestRelease_DecrementsCounter(t *testing.T) {
	c := New(1)
	_ = c.Acquire(context.Background())
	c.Release()
	if c.Active() != 0 {
		t.Fatalf("expected active=0, got %d", c.Active())
	}
	if err := c.Acquire(context.Background()); err != nil {
		t.Fatalf("expected acquire after release to succeed: %v", err)
	}
}

func TestAcquire_CancelledContext_ReturnsErr(t *testing.T) {
	c := New(10)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := c.Acquire(ctx); err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestAcquire_ConcurrentSafe(t *testing.T) {
	const limit = 10
	c := New(limit)
	var wg sync.WaitGroup
	success := make(chan struct{}, 200)
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := c.Acquire(context.Background()); err == nil {
				success <- struct{}{}
				c.Release()
			}
		}()
	}
	wg.Wait()
	close(success)
	if c.Active() != 0 {
		t.Fatalf("expected active=0 after all releases, got %d", c.Active())
	}
}
