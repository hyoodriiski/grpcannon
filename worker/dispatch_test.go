package worker_test

import (
	"context"
	"testing"

	"github.com/example/grpcannon/worker"
)

func TestDispatch_CancelEarly(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately
	ch := worker.Dispatch(ctx, 1000, func(i int) worker.Task {
		return worker.Task{Method: "x"}
	})
	count := 0
	for range ch {
		count++
	}
	// With an already-cancelled context very few (possibly 0) tasks should be sent.
	if count >= 1000 {
		t.Errorf("expected early exit, got %d tasks", count)
	}
}

func TestDispatch_ZeroTasks(t *testing.T) {
	ctx := context.Background()
	ch := worker.Dispatch(ctx, 0, func(i int) worker.Task {
		return worker.Task{}
	})
	count := 0
	for range ch {
		count++
	}
	if count != 0 {
		t.Errorf("expected 0 tasks, got %d", count)
	}
}
