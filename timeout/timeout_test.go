package timeout_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/grpcannon/timeout"
)

func TestNew_ZeroDuration_NoDeadline(t *testing.T) {
	e := timeout.New(0)
	ctx, cancel := e.Apply(context.Background())
	defer cancel()
	if _, ok := ctx.Deadline(); ok {
		t.Fatal("expected no deadline for zero duration")
	}
}

func TestNew_PositiveDuration_HasDeadline(t *testing.T) {
	e := timeout.New(5 * time.Second)
	ctx, cancel := e.Apply(context.Background())
	defer cancel()
	if _, ok := ctx.Deadline(); !ok {
		t.Fatal("expected a deadline to be set")
	}
}

func TestWrap_FnSucceeds(t *testing.T) {
	e := timeout.New(time.Second)
	err := e.Wrap(context.Background(), func(_ context.Context) error {
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestWrap_FnReturnsError(t *testing.T) {
	sentinel := errors.New("fn error")
	e := timeout.New(time.Second)
	err := e.Wrap(context.Background(), func(_ context.Context) error {
		return sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}

func TestWrap_Timeout(t *testing.T) {
	e := timeout.New(20 * time.Millisecond)
	err := e.Wrap(context.Background(), func(ctx context.Context) error {
		time.Sleep(200 * time.Millisecond)
		return nil
	})
	if !errors.Is(err, timeout.ErrDeadlineExceeded) {
		t.Fatalf("expected ErrDeadlineExceeded, got %v", err)
	}
}

func TestWrap_NoDuration_DoesNotTimeout(t *testing.T) {
	e := timeout.New(0)
	err := e.Wrap(context.Background(), func(_ context.Context) error {
		time.Sleep(10 * time.Millisecond)
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}
