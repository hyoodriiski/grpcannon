// Package slope computes the rate-of-change (slope) of a metric over a
// sliding window of samples. It is useful for detecting trends in latency
// or error-rate during a load test run.
package slope

import (
	"sync"
	"time"
)

// Point is a timestamped scalar observation.
type Point struct {
	At    time.Time
	Value float64
}

// Slope holds a fixed-size ring of Points and exposes a Rate method that
// returns the least-squares linear regression slope (units per second).
type Slope struct {
	mu      sync.Mutex
	window  int
	samples []Point
}

// New returns a Slope that retains at most window observations.
// window must be >= 2; values below 2 are clamped to 2.
func New(window int) *Slope {
	if window < 2 {
		window = 2
	}
	return &Slope{window: window}
}

// Record appends a new observation.
func (s *Slope) Record(v float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.samples = append(s.samples, Point{At: time.Now(), Value: v})
	if len(s.samples) > s.window {
		s.samples = s.samples[len(s.samples)-s.window:]
	}
}

// Rate returns the least-squares slope in value-units per second.
// Returns 0 if fewer than 2 samples have been recorded.
func (s *Slope) Rate() float64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.samples) < 2 {
		return 0
	}
	return leastSquares(s.samples)
}

// Reset discards all recorded samples.
func (s *Slope) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.samples = s.samples[:0]
}

// leastSquares computes dy/dx where x is elapsed seconds from the first point.
func leastSquares(pts []Point) float64 {
	n := float64(len(pts))
	t0 := pts[0].At
	var sumX, sumY, sumXY, sumX2 float64
	for _, p := range pts {
		x := p.At.Sub(t0).Seconds()
		y := p.Value
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}
	denom := n*sumX2 - sumX*sumX
	if denom == 0 {
		return 0
	}
	return (n*sumXY - sumX*sumY) / denom
}
