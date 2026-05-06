package ceiling

import (
	"context"
	"errors"
	"testing"
)

func TestGuard_CallsNextWhenCapacityAvailable(t *testing.T) {
	c := New(2)
	called := false
	next := InvokerFn(func(ctx context.Context) error {
		called = true
		return nil
	})
	guarded := Guard(c, next)
	if err := guarded(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected next to be called")
	}
	if c.Active() != 0 {
		t.Fatalf("expected active=0 after invocation, got %d", c.Active())
	}
}

func TestGuard_RejectsWhenFull(t *testing.T) {
	c := New(1)
	_ = c.Acquire(context.Background()) // fill the slot
	next := InvokerFn(func(ctx context.Context) error {
		t.Fatal("next should not be called")
		return nil
	})
	guarded := Guard(c, next)
	if err := guarded(context.Background()); !errors.Is(err, ErrAbove) {
		t.Fatalf("expected ErrAbove, got %v", err)
	}
}

func TestGuard_ReleasesOnNextError(t *testing.T) {
	c := New(5)
	sentinel := errors.New("invoke failed")
	next := InvokerFn(func(ctx context.Context) error {
		return sentinel
	})
	guarded := Guard(c, next)
	if err := guarded(context.Background()); !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
	if c.Active() != 0 {
		t.Fatalf("expected active=0 after error release, got %d", c.Active())
	}
}
