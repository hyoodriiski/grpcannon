// Package ceiling implements a lightweight hard-cap concurrency limiter.
// Unlike a semaphore it does not block callers — it immediately returns
// ErrAbove when the ceiling is reached, making it suitable for load-shedding
// scenarios where queuing is undesirable.
package ceiling
