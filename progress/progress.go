// Package progress provides a real-time progress reporter for ongoing load test runs.
package progress

import (
	"fmt"
	"io"
	"sync/atomic"
	"time"
)

// Reporter periodically prints progress stats to a writer.
type Reporter struct {
	total     int64
	success   int64
	failures  int64
	interval  time.Duration
	writer    io.Writer
	stopCh    chan struct{}
}

// New creates a new Reporter that writes to w every interval.
func New(w io.Writer, interval time.Duration) *Reporter {
	if interval <= 0 {
		interval = time.Second
	}
	return &Reporter{
		interval: interval,
		writer:   w,
		stopCh:   make(chan struct{}),
	}
}

// Record registers a single completed request outcome.
func (r *Reporter) Record(success bool) {
	atomic.AddInt64(&r.total, 1)
	if success {
		atomic.AddInt64(&r.success, 1)
	} else {
		atomic.AddInt64(&r.failures, 1)
	}
}

// Start begins periodic progress reporting in a background goroutine.
func (r *Reporter) Start() {
	go func() {
		ticker := time.NewTicker(r.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				r.print()
			case <-r.stopCh:
				return
			}
		}
	}()
}

// Stop halts the background reporter and prints a final summary line.
func (r *Reporter) Stop() {
	close(r.stopCh)
	r.print()
}

func (r *Reporter) print() {
	t := atomic.LoadInt64(&r.total)
	s := atomic.LoadInt64(&r.success)
	f := atomic.LoadInt64(&r.failures)
	fmt.Fprintf(r.writer, "[progress] total=%-6d success=%-6d failures=%-6d\n", t, s, f)
}
