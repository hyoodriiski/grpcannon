// Package shed implements load shedding: requests are dropped when
// the number of in-flight calls exceeds a configured ceiling.
package shed

import (
	"context"
	"errors"
	"sync/atomic"
)

// ErrShed is returned when a request is shed due to overload.
var ErrShed = errors.New("shed: load too high, request dropped")

// Shed tracks in-flight requests and rejects new ones above the limit.
type Shed struct {
	max     int64
	inflight atomic.Int64
}

// New creates a Shed that allows at most max concurrent calls.
// A max of zero or less disables shedding (all requests pass through).
func New(max int) *Shed {
	return &Shed{max: int64(max)}
}

// Acquire attempts to register a new in-flight request.
// It returns a release func and nil on success, or ErrShed when the
// ceiling has been reached.
func (s *Shed) Acquire(ctx context.Context) (func(), error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if s.max > 0 {
		current := s.inflight.Add(1)
		if current > s.max {
			s.inflight.Add(-1)
			return nil, ErrShed
		}
	}
	release := func() {
		if s.max > 0 {
			s.inflight.Add(-1)
		}
	}
	return release, nil
}

// Inflight returns the current number of in-flight requests.
func (s *Shed) Inflight() int64 {
	return s.inflight.Load()
}
