package dialer

import (
	"errors"
	"sync/atomic"
	"time"

	"google.golang.org/grpc"
)

// Pool holds a fixed set of gRPC connections and distributes them round-robin.
type Pool struct {
	conns []*grpc.ClientConn
	next  uint64
}

// NewPool creates a pool of size connections to addr.
func NewPool(addr string, size int, timeout time.Duration) (*Pool, error) {
	if addr == "" {
		return nil, errors.New("address must not be empty")
	}
	if size <= 0 {
		return nil, errors.New("pool size must be greater than zero")
	}

	conns := make([]*grpc.ClientConn, 0, size)
	for i := 0; i < size; i++ {
		conn, err := Connect(addr, timeout)
		if err != nil {
			// close already-opened connections before returning
			for _, c := range conns {
				_ = c.Close()
			}
			return nil, err
		}
		conns = append(conns, conn)
	}

	return &Pool{conns: conns}, nil
}

// Get returns the next connection in round-robin order.
func (p *Pool) Get() *grpc.ClientConn {
	if len(p.conns) == 0 {
		return nil
	}
	idx := atomic.AddUint64(&p.next, 1) - 1
	return p.conns[idx%uint64(len(p.conns))]
}

// Size returns the number of connections in the pool.
func (p *Pool) Size() int {
	return len(p.conns)
}

// Close closes all connections in the pool.
func (p *Pool) Close() error {
	var last error
	for _, c := range p.conns {
		if err := c.Close(); err != nil {
			last = err
		}
	}
	return last
}
