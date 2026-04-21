package admit_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/grpcannon/admit"
)

func TestGuard_AdmitsAndCallsNext(t *testing.T) {
	c := admit.New(5)
	called := false
	next := func(_ context.Context, _ string) (time.Duration, error) {
		called = true
		return 10 * time.Millisecond, nil
	}
	guarded := admit.Guard(c, next)

	d, err := guarded(context.Background(), "/svc/Method")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected next to be called")
	}
	if d != 10*time.Millisecond {
		t.Fatalf("unexpected duration: %v", d)
	}
	if got := c.InFlight(); got != 0 {
		t.Fatalf("expected 0 in-flight after call, got %d", got)
	}
}

func TestGuard_RejectsWhenFull(t *testing.T) {
	c := admit.New(1)
	// Hold the single slot.
	rel, _ := c.Admit(context.Background())
	defer rel()

	next := func(_ context.Context, _ string) (time.Duration, error) {
		return 0, nil
	}
	guarded := admit.Guard(c, next)

	_, err := guarded(context.Background(), "/svc/Method")
	if !errors.Is(err, admit.ErrRejected) {
		t.Fatalf("expected ErrRejected, got %v", err)
	}
}

func TestGuard_ReleasesOnNextError(t *testing.T) {
	c := admit.New(2)
	nextErr := errors.New("downstream error")
	next := func(_ context.Context, _ string) (time.Duration, error) {
		return 0, nextErr
	}
	guarded := admit.Guard(c, next)

	_, err := guarded(context.Background(), "/svc/Method")
	if !errors.Is(err, nextErr) {
		t.Fatalf("expected downstream error, got %v", err)
	}
	if got := c.InFlight(); got != 0 {
		t.Fatalf("expected 0 in-flight after error, got %d", got)
	}
}
