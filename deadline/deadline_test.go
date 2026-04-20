package deadline_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/your-org/grpcannon/deadline"
)

func TestNew_ZeroTimeout_NoDeadline(t *testing.T) {
	e := deadline.New(0)
	called := false
	err := e.Run(context.Background(), func(ctx context.Context) error {
		_, hasDeadline := ctx.Deadline()
		if hasDeadline {
			t.Error("expected no deadline on context when timeout is zero")
		}
		called = true
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("fn was not called")
	}
}

func TestRun_FnSucceeds(t *testing.T) {
	e := deadline.New(500 * time.Millisecond)
	err := e.Run(context.Background(), func(ctx context.Context) error {
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestRun_FnReturnsError(t *testing.T) {
	sentinel := errors.New("invoke failed")
	e := deadline.New(500 * time.Millisecond)
	err := e.Run(context.Background(), func(ctx context.Context) error {
		return sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}

func TestRun_Timeout_ReturnsDeadlineExceeded(t *testing.T) {
	e := deadline.New(50 * time.Millisecond)
	err := e.Run(context.Background(), func(ctx context.Context) error {
		time.Sleep(200 * time.Millisecond)
		return nil
	})
	if !errors.Is(err, deadline.ErrDeadlineExceeded) {
		t.Fatalf("expected ErrDeadlineExceeded, got %v", err)
	}
}

func TestRun_ParentCancelled(t *testing.T) {
	e := deadline.New(500 * time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := e.Run(ctx, func(ctx context.Context) error {
		time.Sleep(200 * time.Millisecond)
		return nil
	})
	if err == nil {
		t.Fatal("expected an error when parent context is cancelled")
	}
}

func TestRun_NegativeTimeout_NoDeadline(t *testing.T) {
	e := deadline.New(-1 * time.Second)
	called := false
	err := e.Run(context.Background(), func(ctx context.Context) error {
		called = true
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("fn was not called")
	}
}
