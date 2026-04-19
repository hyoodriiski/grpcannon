package dialer

import (
	"context"
	"fmt"
	"sync"

	"google.golang.org/grpc"
)

// Pool manages a fixed set of gRPC client connections.
type Pool struct {
	mu    sync.Mutex
	conns []*grpc.ClientConn
	next  int
}

// NewPool creates a pool of `size` connections to the same target.
func NewPool(ctx context.Context, size int, opts Options) (*Pool, error) {
	if size <= 0 {
		return nil, fmt.Errorf("pool: size must be > 0")
	}
	conns := make([]*grpc.ClientConn, 0, size)
	for i := 0; i < size; i++ {
		conn, err := Connect(ctx, opts)
		if err != nil {
			// Close already-opened connections before returning.
			for _, c := range conns {
				c.Close() //nolint:errcheck
			}
			return nil, fmt.Errorf("pool: connection %d failed: %w", i, err)
		}
		conns = append(conns, conn)
	}
	return &Pool{conns: conns}, nil
}

// Get returns the next connection in round-robin order.
func (p *Pool) Get() *grpc.ClientConn {
	p.mu.Lock()
	defer p.mu.Unlock()
	conn := p.conns[p.next%len(p.conns)]
	p.next++
	return conn
}

// Close closes all connections in the pool.
func (p *Pool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	var firstErr error
	for _, c := range p.conns {
		if err := c.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}
