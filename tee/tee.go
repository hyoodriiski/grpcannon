// Package tee provides a writer that duplicates output to multiple io.Writer targets.
package tee

import (
	"fmt"
	"io"
	"sync"
)

// Writer duplicates writes to all registered writers.
type Writer struct {
	mu      sync.Mutex
	targets []io.Writer
}

// New returns a Writer that fans out to the provided targets.
// Nil targets are silently ignored.
func New(targets ...io.Writer) *Writer {
	filtered := make([]io.Writer, 0, len(targets))
	for _, t := range targets {
		if t != nil {
			filtered = append(filtered, t)
		}
	}
	return &Writer{targets: filtered}
}

// Add appends a new target writer. Nil is ignored.
func (w *Writer) Add(t io.Writer) {
	if t == nil {
		return
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	w.targets = append(w.targets, t)
}

// Write writes p to all registered targets.
// It returns the number of bytes written and the first error encountered.
func (w *Writer) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	for _, t := range w.targets {
		if _, err := t.Write(p); err != nil {
			return 0, fmt.Errorf("tee: write failed: %w", err)
		}
	}
	return len(p), nil
}

// Len returns the number of registered targets.
func (w *Writer) Len() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return len(w.targets)
}
