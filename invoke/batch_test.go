package invoke_test

import (
	"context"
	"testing"
	"time"

	"google.golang.org/grpc"

	"github.com/example/grpcannon/invoke"
)

func TestRunBatch_NilConn(t *testing.T) {
	results := invoke.RunBatch(context.Background(), nil, "/svc/Method", 3)
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	for _, r := range results {
		if r.Err == nil {
			t.Error("expected error for nil conn")
		}
	}
}

func TestRunBatch_EmptyMethod(t *testing.T) {
	conn, _ := grpc.Dial("localhost:1", grpc.WithInsecure())
	defer conn.Close()
	results := invoke.RunBatch(context.Background(), conn, "", 2)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for _, r := range results {
		if r.Err == nil {
			t.Error("expected error for empty method")
		}
	}
}

func TestRunBatch_CancelledContext(t *testing.T) {
	conn, _ := grpc.Dial("localhost:1", grpc.WithInsecure())
	defer conn.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	results := invoke.RunBatch(ctx, conn, "/svc/Method", 5)
	if len(results) != 5 {
		t.Fatalf("expected 5 results, got %d", len(results))
	}
}

func TestRunBatch_LatencyPopulated(t *testing.T) {
	addr, stop := startStubServer(t)
	defer stop()
	conn := dial(t, addr)
	defer conn.Close()
	results := invoke.RunBatch(context.Background(), conn, "/grpc.health.v1.Health/Check", 4)
	if len(results) != 4 {
		t.Fatalf("expected 4 results")
	}
	for _, r := range results {
		if r.Err == nil && r.Latency <= 0 {
			t.Error("expected positive latency on success")
		}
		if r.Latency > 5*time.Second {
			t.Errorf("latency suspiciously high: %v", r.Latency)
		}
	}
}
