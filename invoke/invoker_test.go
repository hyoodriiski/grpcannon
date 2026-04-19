package invoke_test

import (
	"context"
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/grpcannon/invoke"
)

func startStubServer(t *testing.T) string {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	srv := grpc.NewServer()
	go srv.Serve(lis) //nolint:errcheck
	t.Cleanup(srv.Stop)
	return lis.Addr().String()
}

func dial(t *testing.T, addr string) *grpc.ClientConn {
	t.Helper()
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	return conn
}

func TestCall_NilConn(t *testing.T) {
	inv := invoke.New(nil, "/pkg.Svc/Method", nil)
	res := inv.Call(context.Background())
	if res.Err == nil {
		t.Fatal("expected error for nil conn")
	}
}

func TestCall_EmptyMethod(t *testing.T) {
	addr := startStubServer(t)
	conn := dial(t, addr)
	inv := invoke.New(conn, "", nil)
	res := inv.Call(context.Background())
	if res.Err == nil {
		t.Fatal("expected error for empty method")
	}
}

func TestCall_UnknownMethod_ReturnsLatency(t *testing.T) {
	addr := startStubServer(t)
	conn := dial(t, addr)
	inv := invoke.New(conn, "/pkg.Svc/Missing", map[string]string{})
	res := inv.Call(context.Background())
	// The call will fail (unimplemented) but latency should be recorded.
	if res.Latency <= 0 {
		t.Errorf("expected positive latency, got %v", res.Latency)
	}
	if res.Err == nil {
		t.Error("expected RPC error for unknown method")
	}
}
