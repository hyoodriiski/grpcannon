package retry_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/grpcannon/retry"
)

var errFail = errors.New("fail")

func TestDefault_Values(t *testing.T) {
	p := retry.Default()
	if p.MaxAttempts != 3 {
		t.Fatalf("expected 3, got %d", p.MaxAttempts)
	}
}

func TestDo_SuccessFirstAttempt(t *testing.T) {
	p := retry.Policy{MaxAttempts: 3, Delay: 0}
	calls := 0
	err := p.Do(context.Background(), func() error {
		calls++
		return nil
	})
	if err != nil || calls != 1 {
		t.Fatalf("unexpected: err=%v calls=%d", err, calls)
	}
}

func TestDo_RetriesOnError(t *testing.T) {
	p := retry.Policy{MaxAttempts: 3, Delay: 0}
	calls := 0
	err := p.Do(context.Background(), func() error {
		calls++
		return errFail
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestDo_SucceedsOnRetry(t *testing.T) {
	p := retry.Policy{MaxAttempts: 3, Delay: 0}
	calls := 0
	err := p.Do(context.Background(), func() error {
		calls++
		if calls < 3 {
			return errFail
		}
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestDo_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	p := retry.Policy{MaxAttempts: 5, Delay: 10 * time.Millisecond}
	err := p.Do(ctx, func() error { return errFail })
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestDo_ZeroMaxAttempts(t *testing.T) {
	p := retry.Policy{MaxAttempts: 0}
	called := false
	_ = p.Do(context.Background(), func() error {
		called = true
		return nil
	})
	if !called {
		t.Fatal("fn should be called once when MaxAttempts=0")
	}
}
