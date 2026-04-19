package metrics

import (
	"fmt"
	"io"
	"os"
	"sort"
	"time"
)

// Summary holds aggregate statistics derived from a Snapshot.
type Summary struct {
	Total     int
	Errors    int
	ErrorRate float64
	Min       time.Duration
	Max       time.Duration
	Mean      time.Duration
	P50       time.Duration
	P95       time.Duration
	P99       time.Duration
}

// Summarise computes a Summary from a Snapshot.
func Summarise(s Snapshot) Summary {
	sum := Summary{
		Total:     s.Total,
		Errors:    s.Errors,
		ErrorRate: s.ErrorRate(),
	}
	if len(s.Latencies) == 0 {
		return sum
	}
	sorted := make([]time.Duration, len(s.Latencies))
	copy(sorted, s.Latencies)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })

	var total time.Duration
	for _, d := range sorted {
		total += d
	}
	sum.Min = sorted[0]
	sum.Max = sorted[len(sorted)-1]
	sum.Mean = total / time.Duration(len(sorted))
	sum.P50 = percentile(sorted, 0.50)
	sum.P95 = percentile(sorted, 0.95)
	sum.P99 = percentile(sorted, 0.99)
	return sum
}

func percentile(sorted []time.Duration, p float64) time.Duration {
	idx := int(float64(len(sorted)-1) * p)
	return sorted[idx]
}

// Print writes a human-readable summary to w (defaults to os.Stdout).
func (s Summary) Print(w io.Writer) {
	if w == nil {
		w = os.Stdout
	}
	fmt.Fprintf(w, "Total: %d  Errors: %d  ErrorRate: %.2f%%\n", s.Total, s.Errors, s.ErrorRate*100)
	fmt.Fprintf(w, "Min: %v  Mean: %v  Max: %v\n", s.Min, s.Mean, s.Max)
	fmt.Fprintf(w, "P50: %v  P95: %v  P99: %v\n", s.P50, s.P95, s.P99)
}
