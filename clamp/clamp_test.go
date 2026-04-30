package clamp_test

import (
	"context"
	"sync"
	"testing"

	"github.com/grpcannon/clamp"
)

func TestNew_ClampsNegativeMin(t *testing.T) {
	c := clamp.New(-5, 10)
	if c.InFlight() != 0 {
		t.Fatalf("expected 0 in-flight, got %d", c.InFlight())
	}
}

func TestNew_MaxBelowMin_EqualisesMax(t *testing.T) {
	c := clamp.New(5, 2)
	// Should not panic and max should have been raised to min.
	for i := 0; i < 5; i++ {
		if err := c.Acquire(context.Background()); err != nil {
			t.Fatalf("unexpected error on acquire %d: %v", i, err)
		}
	}
	// 6th acquire should exceed the ceiling (which equals min=5).
	if err := c.Acquire(context.Background()); err != clamp.ErrAbove {
		t.Fatalf("expected ErrAbove, got %v", err)
	}
}

func TestAcquire_WithinRange_Succeeds(t *testing.T) {
	c := clamp.New(1, 3)
	if err := c.Acquire(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.InFlight() != 1 {
		t.Fatalf("expected 1 in-flight, got %d", c.InFlight())
	}
}

func TestAcquire_ExceedsCeiling_ReturnsErrAbove(t *testing.T) {
	c := clamp.New(0, 2)
	for i := 0; i < 2; i++ {
		_ = c.Acquire(context.Background())
	}
	if err := c.Acquire(context.Background()); err != clamp.ErrAbove {
		t.Fatalf("expected ErrAbove, got %v", err)
	}
}

func TestRelease_DecrementsCounter(t *testing.T) {
	c := clamp.New(0, 5)
	_ = c.Acquire(context.Background())
	_ = c.Acquire(context.Background())
	c.Release()
	if c.InFlight() != 1 {
		t.Fatalf("expected 1 in-flight after release, got %d", c.InFlight())
	}
}

func TestRelease_NoOp_WhenZero(t *testing.T) {
	c := clamp.New(0, 5)
	c.Release() // should not panic or go negative
	if c.InFlight() != 0 {
		t.Fatalf("expected 0 in-flight, got %d", c.InFlight())
	}
}

func TestAcquire_CancelledContext_ReturnsErr(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	c := clamp.New(0, 10)
	if err := c.Acquire(ctx); err == nil {
		t.Fatal("expected error for cancelled context, got nil")
	}
}

func TestAcquire_ConcurrentSafe(t *testing.T) {
	c := clamp.New(0, 50)
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := c.Acquire(context.Background()); err == nil {
				defer c.Release()
			}
		}()
	}
	wg.Wait()
	if c.InFlight() != 0 {
		t.Fatalf("expected 0 in-flight after all releases, got %d", c.InFlight())
	}
}
