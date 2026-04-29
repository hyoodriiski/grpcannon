// Package burst provides a token-bucket style burst limiter that allows
// short bursts of traffic above a sustained rate.
package burst

import (
	"context"
	"errors"
	"sync"
	"time"
)

// ErrExceeded is returned when the burst capacity is exhausted.
var ErrExceeded = errors.New("burst: capacity exceeded")

// Limiter is a token-bucket burst limiter.
type Limiter struct {
	mu       sync.Mutex
	tokens   float64
	cap      float64
	rate     float64 // tokens per second
	lastTick time.Time
	now      func() time.Time
}

// New creates a Limiter with the given burst capacity and sustained rate
// (tokens per second). cap must be >= 1 and rate must be > 0.
func New(cap int, rate float64) *Limiter {
	if cap < 1 {
		cap = 1
	}
	if rate <= 0 {
		rate = 1
	}
	now := time.Now()
	return &Limiter{
		tokens:   float64(cap),
		cap:      float64(cap),
		rate:     rate,
		lastTick: now,
		now:      time.Now,
	}
}

// refill adds tokens based on elapsed time since the last call.
func (l *Limiter) refill() {
	now := l.now()
	elapsed := now.Sub(l.lastTick).Seconds()
	l.lastTick = now
	l.tokens += elapsed * l.rate
	if l.tokens > l.cap {
		l.tokens = l.cap
	}
}

// Acquire attempts to consume one token. Returns ErrExceeded if no tokens
// are available. Returns immediately without blocking.
func (l *Limiter) Acquire(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.refill()
	if l.tokens < 1 {
		return ErrExceeded
	}
	l.tokens--
	return nil
}

// Tokens returns the current number of available tokens.
func (l *Limiter) Tokens() float64 {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.refill()
	return l.tokens
}
