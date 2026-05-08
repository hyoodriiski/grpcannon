package drift

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestGuard_RecordsLatencyAndPassesThrough(t *testing.T) {
	d := New(100*time.Millisecond, 2.0, 10)
	called := false
	next := func(_ context.Context) error {
		called = true
		return nil
	}
	inv := Guard(d, next)
	if err := inv(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected next to be called")
	}
}

func TestGuard_NextError_ReturnedDirectly(t *testing.T) {
	d := New(100*time.Millisecond, 2.0, 10)
	sentinel := errors.New("rpc error")
	next := func(_ context.Context) error { return sentinel }
	inv := Guard(d, next)
	if err := inv(context.Background()); !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}

func TestGuard_DriftDetected_ReturnsErrDrifted(t *testing.T) {
	d := New(1*time.Millisecond, 1.5, 5)
	// Invoker that sleeps long enough to trigger drift.
	next := func(_ context.Context) error {
		time.Sleep(20 * time.Millisecond)
		return nil
	}
	inv := Guard(d, next)
	var last error
	for i := 0; i < 5; i++ {
		last = inv(context.Background())
	}
	if !errors.Is(last, ErrDrifted) {
		t.Fatalf("expected ErrDrifted, got %v", last)
	}
}

func TestGuard_NextErrorTakesPrecedenceOverDrift(t *testing.T) {
	d := New(1*time.Millisecond, 1.5, 5)
	sentinel := errors.New("downstream")
	next := func(_ context.Context) error {
		time.Sleep(20 * time.Millisecond)
		return sentinel
	}
	inv := Guard(d, next)
	var last error
	for i := 0; i < 5; i++ {
		last = inv(context.Background())
	}
	if !errors.Is(last, sentinel) {
		t.Fatalf("expected sentinel error to take precedence, got %v", last)
	}
}
