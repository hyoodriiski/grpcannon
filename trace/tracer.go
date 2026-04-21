// Package trace provides lightweight per-request tracing for grpcannon,
// recording span start/end times and attaching metadata labels.
package trace

import (
	"sync"
	"time"
)

// Span holds timing and metadata for a single traced operation.
type Span struct {
	Method    string
	Start     time.Time
	End       time.Time
	Latency   time.Duration
	Err       error
	Labels    map[string]string
}

// Tracer collects completed spans in a concurrency-safe manner.
type Tracer struct {
	mu    sync.Mutex
	spans []Span
}

// New returns an initialised Tracer.
func New() *Tracer {
	return &Tracer{}
}

// Start begins a new span for the given method and returns a finish function.
// Calling the returned function records the span with elapsed latency.
func (t *Tracer) Start(method string, labels map[string]string) func(err error) {
	begin := time.Now()
	return func(err error) {
		end := time.Now()
		s := Span{
			Method:  method,
			Start:   begin,
			End:     end,
			Latency: end.Sub(begin),
			Err:     err,
			Labels:  labels,
		}
		t.mu.Lock()
		t.spans = append(t.spans, s)
		t.mu.Unlock()
	}
}

// Spans returns a snapshot of all recorded spans.
func (t *Tracer) Spans() []Span {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := make([]Span, len(t.spans))
	copy(out, t.spans)
	return out
}

// Reset clears all recorded spans.
func (t *Tracer) Reset() {
	t.mu.Lock()
	t.spans = t.spans[:0]
	t.mu.Unlock()
}
