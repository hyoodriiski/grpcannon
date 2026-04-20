// Package circuit implements a simple circuit breaker for gRPC call protection.
package circuit

import (
	"errors"
	"sync"
	"time"
)

// ErrOpen is returned when the circuit breaker is in the open state.
var ErrOpen = errors.New("circuit breaker is open")

// State represents the current state of the circuit breaker.
type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

// Breaker is a circuit breaker that trips after a threshold of consecutive failures.
type Breaker struct {
	mu           sync.Mutex
	state        State
	failures     int
	threshold    int
	resetAfter   time.Duration
	openedAt     time.Time
}

// New creates a Breaker that opens after threshold consecutive failures
// and attempts reset after resetAfter duration.
func New(threshold int, resetAfter time.Duration) *Breaker {
	if threshold <= 0 {
		threshold = 1
	}
	return &Breaker{
		threshold:  threshold,
		resetAfter: resetAfter,
	}
}

// Allow returns nil if the call should proceed, or ErrOpen if the circuit is open.
func (b *Breaker) Allow() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	switch b.state {
	case StateOpen:
		if time.Since(b.openedAt) >= b.resetAfter {
			b.state = StateHalfOpen
			return nil
		}
		return ErrOpen
	default:
		return nil
	}
}

// RecordSuccess resets the failure count and closes the circuit.
func (b *Breaker) RecordSuccess() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.failures = 0
	b.state = StateClosed
}

// RecordFailure increments the failure count and opens the circuit if threshold is reached.
func (b *Breaker) RecordFailure() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.failures++
	if b.failures >= b.threshold {
		b.state = StateOpen
		b.openedAt = time.Now()
	}
}

// CurrentState returns the current state of the breaker.
func (b *Breaker) CurrentState() State {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.state
}
