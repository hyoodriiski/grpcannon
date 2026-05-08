// Package drift detects when observed latency has shifted significantly
// from a baseline, signalling potential performance degradation.
package drift

import (
	"errors"
	"sync"
	"time"
)

// ErrDrifted is returned when latency has exceeded the allowed drift factor.
var ErrDrifted = errors.New("drift: latency has drifted beyond threshold")

// Detector tracks a rolling baseline and reports drift.
type Detector struct {
	mu        sync.Mutex
	baseline  time.Duration
	factor    float64
	samples   []time.Duration
	cap       int
}

// New creates a Detector with the given baseline, drift factor, and sample
// window capacity. factor must be > 1.0 (e.g. 2.0 means 2× baseline).
func New(baseline time.Duration, factor float64, windowCap int) *Detector {
	if factor <= 1.0 {
		factor = 2.0
	}
	if windowCap <= 0 {
		windowCap = 100
	}
	return &Detector{
		baseline: baseline,
		factor:   factor,
		cap:      windowCap,
	}
}

// Record adds a latency observation and returns ErrDrifted if the rolling
// average exceeds baseline × factor.
func (d *Detector) Record(lat time.Duration) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.samples = append(d.samples, lat)
	if len(d.samples) > d.cap {
		d.samples = d.samples[len(d.samples)-d.cap:]
	}

	if d.avg() > time.Duration(float64(d.baseline)*d.factor) {
		return ErrDrifted
	}
	return nil
}

// Reset clears the sample window.
func (d *Detector) Reset() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.samples = d.samples[:0]
}

// Avg returns the current rolling average latency.
func (d *Detector) Avg() time.Duration {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.avg()
}

func (d *Detector) avg() time.Duration {
	if len(d.samples) == 0 {
		return 0
	}
	var total time.Duration
	for _, s := range d.samples {
		total += s
	}
	return total / time.Duration(len(d.samples))
}
