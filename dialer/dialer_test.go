package dialer_test

import (
	"context"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"

	"github.com/user/grpcannon/dialer"
)

// startFakeServer starts a bare gRPC server on a random local port and returns its address.
func startFakeServer(t *testing.T) string {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	srv := grpc.NewServer()
	go srv.Serve(lis) //nolint:errcheck
	t.Cleanup(srv.Stop)
	return lis.Addr().String()
}

func TestConnect_Success(t *testing.T) {
	addr := startFakeServer(t)
	conn, err := dialer.Connect(context.Background(), dialer.Options{
		Address: addr,
		Timeout: 3 * time.Second,
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	defer conn.Close()
}

func TestConnect_EmptyAddress(t *testing.T) {
	_, err := dialer.Connect(context.Background(), dialer.Options{})
	if err == nil {
		t.Fatal("expected error for empty address")
	}
}

func TestConnect_Timeout(t *testing.T) {
	// Nothing listening on this port — should time out quickly.
	_, err := dialer.Connect(context.Background(), dialer.Options{
		Address: "127.0.0.1:19999",
		Timeout: 200 * time.Millisecond,
	})
	if err == nil {
		t.Fatal("expected connection timeout error")
	}
}
