// Package admit provides a simple admission controller that caps the number
// of concurrent requests allowed into the system. Requests that exceed the
// cap are rejected immediately rather than queued.
package admit

import (
	"context"
	"errors"
	"sync/atomic"
)

// ErrRejected is returned when the admission limit has been reached.
var ErrRejected = errors.New("admit: request rejected — limit reached")

// Controller tracks in-flight requests and enforces a concurrency cap.
type Controller struct {
	max     int64
	inflight atomic.Int64
}

// New creates a Controller with the given concurrency cap.
// A max of zero means unlimited admissions.
func New(max int) *Controller {
	if max < 0 {
		max = 0
	}
	return &Controller{max: int64(max)}
}

// Admit attempts to admit a request. It returns a release function and nil
// on success. On failure it returns ErrRejected. The caller must invoke the
// returned release function exactly once when the request completes.
func (c *Controller) Admit(_ context.Context) (func(), error) {
	if c.max == 0 {
		// unlimited
		c.inflight.Add(1)
		return func() { c.inflight.Add(-1) }, nil
	}
	// Optimistic increment then check.
	v := c.inflight.Add(1)
	if v > c.max {
		c.inflight.Add(-1)
		return nil, ErrRejected
	}
	return func() { c.inflight.Add(-1) }, nil
}

// InFlight returns the current number of admitted (in-flight) requests.
func (c *Controller) InFlight() int64 {
	return c.inflight.Load()
}
