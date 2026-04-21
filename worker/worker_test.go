package worker_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"google.golang.org/grpc"

	"github.com/example/grpcannon/worker"
)

func noopInvoker(_ context.Context, _ *grpc.ClientConn, _ worker.Task) worker.Result {
	return worker.Result{Latency: time.Millisecond}
}

func errorInvoker(_ context.Context, _ *grpc.ClientConn, _ worker.Task) worker.Result {
	return worker.Result{Err: errors.New("fail")}
}

func TestPool_Run_CollectsResults(t *testing.T) {
	pool := worker.NewPool(nil, 4, noopInvoker)
	ctx := context.Background()
	tasks := worker.Dispatch(ctx, 20, func(i int) worker.Task {
		return worker.Task{Method: "Test", Payload: nil}
	})
	results := pool.Run(ctx, tasks)
	count := 0
	for r := range results {
		if r.Err != nil {
			t.Errorf("unexpected error: %v", r.Err)
		}
		count++
	}
	if count != 20 {
		t.Errorf("expected 20 results, got %d", count)
	}
}

func TestPool_Run_ErrorResults(t *testing.T) {
	pool := worker.NewPool(nil, 2, errorInvoker)
	ctx := context.Background()
	tasks := worker.Dispatch(ctx, 5, func(i int) worker.Task {
		return worker.Task{Method: "Fail"}
	})
	results := pool.Run(ctx, tasks)
	errCount := 0
	for r := range results {
		if r.Err != nil {
			errCount++
		}
	}
	if errCount != 5 {
		t.Errorf("expected 5 errors, got %d", errCount)
	}
}

func TestPool_Run_CancelContext(t *testing.T) {
	var calls atomic.Int32
	invoker := func(ctx context.Context, _ *grpc.ClientConn, _ worker.Task) worker.Result {
		calls.Add(1)
		time.Sleep(10 * time.Millisecond)
		return worker.Result{}
	}
	pool := worker.NewPool(nil, 2, invoker)
	ctx, cancel := context.WithCancel(context.Background())
	tasks := worker.Dispatch(ctx, 100, func(i int) worker.Task { return worker.Task{} })
	results := pool.Run(ctx, tasks)
	time.AfterFunc(25*time.Millisecond, cancel)
	for range results {
	}
	if calls.Load() >= 100 {
		t.Error("expected cancellation to stop work early")
	}
}

func TestDispatch_SendsAllTasks(t *testing.T) {
	ctx := context.Background()
	const n = 10
	ch := worker.Dispatch(ctx, n, func(i int) worker.Task {
		return worker.Task{Method: "m"}
	})
	count := 0
	for range ch {
		count++
	}
	if count != n {
		t.Errorf("expected %d tasks, got %d", n, count)
	}
}

// TestDispatch_IndexPassedToFactory verifies that the factory function receives
// the correct sequential index for each task, starting from 0.
func TestDispatch_IndexPassedToFactory(t *testing.T) {
	ctx := context.Background()
	const n = 5
	var indices []int
	ch := worker.Dispatch(ctx, n, func(i int) worker.Task {
		indices = append(indices, i)
		return worker.Task{Method: "m"}
	})
	for range ch {
	}
	if len(indices) != n {
		t.Fatalf("expected %d index calls, got %d", n, len(indices))
	}
	for want, got := range indices {
		if got != want {
			t.Errorf("index[%d] = %d, want %d", want, got, want)
		}
	}
}
