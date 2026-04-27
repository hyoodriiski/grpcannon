package shed_test

import (
	"context"
	"errors"
	"testing"

	"grpcannon/shed"
)

func TestGuard_CallsNextWhenCapacityAvailable(t *testing.T) {
	s := shed.New(5)
	called := false
	invoker := shed.Guard(s, func(_ context.Context) error {
		called = true
		return nil
	})

	if err := invoker(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected next invoker to be called")
	}
}

func TestGuard_ShedsWhenFull(t *testing.T) {
	s := shed.New(1)

	// Saturate the shed manually.
	ctx := context.Background()
	if err := s.Acquire(ctx); err != nil {
		t.Fatalf("unexpected error on first acquire: %v", err)
	}
	defer s.Release()

	called := false
	invoker := shed.Guard(s, func(_ context.Context) error {
		called = true
		return nil
	})

	err := invoker(ctx)
	if !errors.Is(err, shed.ErrShed) {
		t.Fatalf("expected ErrShed, got %v", err)
	}
	if called {
		t.Fatal("next invoker should not have been called")
	}
}

func TestGuard_ReleasesOnNextError(t *testing.T) {
	s := shed.New(2)
	sentinel := errors.New("invoke failed")

	invoker := shed.Guard(s, func(_ context.Context) error {
		return sentinel
	})

	if err := invoker(context.Background()); !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}

	// Capacity should be restored; a second call should succeed.
	var secondCalled bool
	invoker2 := shed.Guard(s, func(_ context.Context) error {
		secondCalled = true
		return nil
	})
	if err := invoker2(context.Background()); err != nil {
		t.Fatalf("unexpected error on second call: %v", err)
	}
	if !secondCalled {
		t.Fatal("expected second invoker to be called")
	}
}
