package runner

import (
	"fmt"
	"io"
	"sort"
	"time"
)

// Stats summarises a slice of Results.
type Stats struct {
	Total    int
	Errors   int
	Min      time.Duration
	Max      time.Duration
	Mean     time.Duration
	P50      time.Duration
	P95      time.Duration
	P99      time.Duration
}

// Compute derives Stats from results.
func Compute(results []Result) Stats {
	if len(results) == 0 {
		return Stats{}
	}
	latencies := make([]time.Duration, 0, len(results))
	var sum time.Duration
	var errs int
	for _, r := range results {
		if r.Err != nil {
			errs++
		}
		latencies = append(latencies, r.Latency)
		sum += r.Latency
	}
	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
	n := len(latencies)
	return Stats{
		Total:  n,
		Errors: errs,
		Min:    latencies[0],
		Max:    latencies[n-1],
		Mean:   sum / time.Duration(n),
		P50:    latencies[int(float64(n)*0.50)],
		P95:    latencies[int(float64(n)*0.95)],
		P99:    latencies[int(float64(n)*0.99)],
	}
}

// Print writes a human-readable histogram summary to w.
func Print(w io.Writer, s Stats) {
	fmt.Fprintf(w, "Total:  %d\n", s.Total)
	fmt.Fprintf(w, "Errors: %d\n", s.Errors)
	fmt.Fprintf(w, "Min:    %v\n", s.Min)
	fmt.Fprintf(w, "Max:    %v\n", s.Max)
	fmt.Fprintf(w, "Mean:   %v\n", s.Mean)
	fmt.Fprintf(w, "p50:    %v\n", s.P50)
	fmt.Fprintf(w, "p95:    %v\n", s.P95)
	fmt.Fprintf(w, "p99:    %v\n", s.P99)
}
