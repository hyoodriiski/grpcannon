package report

import (
	"fmt"
	"io"
	"os"
	"time"
)

// Summary holds aggregated results from a load test run.
type Summary struct {
	Total       int
	Successes   int
	Failures    int
	Duration    time.Duration
	Latencies   []time.Duration
	P50         time.Duration
	P95         time.Duration
	P99         time.Duration
	Throughput  float64 // requests per second
}

// New builds a Summary from raw results.
func New(latencies []time.Duration, failures int, total time.Duration) Summary {
	n := len(latencies)
	s := Summary{
		Total:     n,
		Successes: n - failures,
		Failures:  failures,
		Duration:  total,
		Latencies: latencies,
	}
	if total.Seconds() > 0 {
		s.Throughput = float64(n) / total.Seconds()
	}
	return s
}

// Print writes a human-readable summary to w (defaults to os.Stdout).
func Print(s Summary, w io.Writer) {
	if w == nil {
		w = os.Stdout
	}
	fmt.Fprintf(w, "\n=== grpcannon Report ===\n")
	fmt.Fprintf(w, "Total Requests : %d\n", s.Total)
	fmt.Fprintf(w, "Successes      : %d\n", s.Successes)
	fmt.Fprintf(w, "Failures       : %d\n", s.Failures)
	fmt.Fprintf(w, "Duration       : %s\n", s.Duration.Round(time.Millisecond))
	fmt.Fprintf(w, "Throughput     : %.2f req/s\n", s.Throughput)
	fmt.Fprintf(w, "Latency p50    : %s\n", s.P50)
	fmt.Fprintf(w, "Latency p95    : %s\n", s.P95)
	fmt.Fprintf(w, "Latency p99    : %s\n", s.P99)
}
