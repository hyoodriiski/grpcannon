package hook_test

import (
	"context"
	"errors"
	"testing"

	"github.com/example/grpcannon/hook"
)

func TestRegister_NilFnIgnored(t *testing.T) {
	r := hook.New()
	r.Register(hook.BeforeRun, nil)
	if err := r.RunBefore(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunBefore_ExecutesInOrder(t *testing.T) {
	r := hook.New()
	var order []int
	r.Register(hook.BeforeRun, func(_ context.Context, _ hook.Phase) error {
		order = append(order, 1)
		return nil
	})
	r.Register(hook.BeforeRun, func(_ context.Context, _ hook.Phase) error {
		order = append(order, 2)
		return nil
	})
	if err := r.RunBefore(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(order) != 2 || order[0] != 1 || order[1] != 2 {
		t.Fatalf("unexpected order: %v", order)
	}
}

func TestRunBefore_StopsOnError(t *testing.T) {
	r := hook.New()
	sentinel := errors.New("hook error")
	called := 0
	r.Register(hook.BeforeRun, func(_ context.Context, _ hook.Phase) error {
		called++
		return sentinel
	})
	r.Register(hook.BeforeRun, func(_ context.Context, _ hook.Phase) error {
		called++
		return nil
	})
	err := r.RunBefore(context.Background())
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel, got %v", err)
	}
	if called != 1 {
		t.Fatalf("expected 1 call, got %d", called)
	}
}

func TestRunAfter_ExecutesHooks(t *testing.T) {
	r := hook.New()
	executed := false
	r.Register(hook.AfterRun, func(_ context.Context, p hook.Phase) error {
		if p != hook.AfterRun {
			t.Errorf("expected AfterRun phase, got %v", p)
		}
		executed = true
		return nil
	})
	if err := r.RunAfter(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !executed {
		t.Fatal("after hook was not executed")
	}
}

func TestRun_ContextCancelled(t *testing.T) {
	r := hook.New()
	r.Register(hook.BeforeRun, func(_ context.Context, _ hook.Phase) error {
		return nil
	})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := r.RunBefore(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}
