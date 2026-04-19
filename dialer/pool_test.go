package dialer

import (
	"testing"
	"time"

	"google.golang.org/grpc"
)

func TestNewPool_CreatesConnections(t *testing.T) {
	addr, stop := startFakeServer(t)
	defer stop()

	pool, err := NewPool(addr, 3, time.Second)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer pool.Close()

	if pool.Size() != 3 {
		t.Errorf("expected pool size 3, got %d", pool.Size())
	}
}

func TestNewPool_ZeroSize(t *testing.T) {
	_, err := NewPool("localhost:9999", 0, time.Second)
	if err == nil {
		t.Fatal("expected error for zero pool size")
	}
}

func TestNewPool_InvalidAddress(t *testing.T) {
	_, err := NewPool("", 2, time.Second)
	if err == nil {
		t.Fatal("expected error for empty address")
	}
}

func TestPool_Get_RoundRobin(t *testing.T) {
	addr, stop := startFakeServer(t)
	defer stop()

	pool, err := NewPool(addr, 3, time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer pool.Close()

	seen := make(map[*grpc.ClientConn]int)
	for i := 0; i < 9; i++ {
		conn := pool.Get()
		if conn == nil {
			t.Fatal("expected non-nil connection")
		}
		seen[conn]++
	}

	if len(seen) != 3 {
		t.Errorf("expected 3 distinct connections, got %d", len(seen))
	}
}

func TestPool_Close(t *testing.T) {
	addr, stop := startFakeServer(t)
	defer stop()

	pool, err := NewPool(addr, 2, time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := pool.Close(); err != nil {
		t.Errorf("expected clean close, got %v", err)
	}
}
