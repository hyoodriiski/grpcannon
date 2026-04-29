package policy_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/nickcorin/grpcannon/backoff"
	"github.com/nickcorin/grpcannon/circuit"
	"github.com/nickcorin/grpcannon/policy"
)

var errBoom = errors.New("boom")

func TestExecute_SuccessFirstAttempt(t *testing.T) {
	p := policy.New(policy.Options{MaxAttempts: 3})
	calls := 0
	err := p.Execute(context.Background(), func(_ context.Context) error {
		calls++
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestExecute_RetriesOnError(t *testing.T) {
	p := policy.New(policy.Options{
		MaxAttempts: 3,
		Backoff:     backoff.Backoff{Base: time.Millisecond, Max: time.Millisecond},
	})
	var calls int32
	err := p.Execute(context.Background(), func(_ context.Context) error {
		atomic.AddInt32(&calls, 1)
		return errBoom
	})
	if !errors.Is(err, errBoom) {
		t.Fatalf("expected errBoom, got %v", err)
	}
	if atomic.LoadInt32(&calls) != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestExecute_SucceedsOnRetry(t *testing.T) {
	p := policy.New(policy.Options{
		MaxAttempts: 3,
		Backoff:     backoff.Backoff{Base: time.Millisecond, Max: time.Millisecond},
	})
	var calls int32
	err := p.Execute(context.Background(), func(_ context.Context) error {
		n := atomic.AddInt32(&calls, 1)
		if n < 3 {
			return errBoom
		}
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExecute_ContextCancelled(t *testing.T) {
	p := policy.New(policy.Options{
		MaxAttempts: 10,
		Backoff:     backoff.Backoff{Base: 50 * time.Millisecond, Max: 50 * time.Millisecond},
	})
	ctx, cancel := context.WithCancel(context.Background())
	var calls int32
	err := p.Execute(ctx, func(_ context.Context) error {
		if atomic.AddInt32(&calls, 1) == 2 {
			cancel()
		}
		return errBoom
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected Canceled, got %v", err)
	}
}

func TestExecute_CircuitOpen(t *testing.T) {
	b := circuit.New(circuit.Options{Threshold: 1, HalfOpenAfter: time.Hour})
	// Trip the breaker.
	b.Record(errBoom)

	p := policy.New(policy.Options{MaxAttempts: 3, Breaker: b})
	err := p.Execute(context.Background(), func(_ context.Context) error {
		return nil
	})
	if !errors.Is(err, policy.ErrCircuitOpen) {
		t.Fatalf("expected ErrCircuitOpen, got %v", err)
	}
}

func TestNew_DefaultsMaxAttempts(t *testing.T) {
	p := policy.New(policy.Options{})
	calls := 0
	_ = p.Execute(context.Background(), func(_ context.Context) error {
		calls++
		return errBoom
	})
	if calls != 1 {
		t.Fatalf("expected default of 1 attempt, got %d", calls)
	}
}
