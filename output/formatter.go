package output

import (
	"fmt"
	"io"
	"text/tabwriter"
	"time"
)

// Format controls the output style.
type Format string

const (
	FormatText Format = "text"
	FormatJSON Format = "json"
)

// Summary holds the data passed to the formatter.
type Summary struct {
	Total       int
	Errors      int
	Duration    time.Duration
	Throughput  float64
	P50         time.Duration
	P95         time.Duration
	P99         time.Duration
	MaxLatency  time.Duration
	MeanLatency time.Duration
}

// WriteText writes a human-readable table to w.
func WriteText(w io.Writer, s Summary) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "--- grpcannon results ---")
	fmt.Fprintf(tw, "Total requests:\t%d\n", s.Total)
	fmt.Fprintf(tw, "Errors:\t%d\n", s.Errors)
	fmt.Fprintf(tw, "Duration:\t%s\n", s.Duration.Round(time.Millisecond))
	fmt.Fprintf(tw, "Throughput:\t%.2f req/s\n", s.Throughput)
	fmt.Fprintln(tw, "--- latency ---")
	fmt.Fprintf(tw, "Mean:\t%s\n", s.MeanLatency.Round(time.Microsecond))
	fmt.Fprintf(tw, "P50:\t%s\n", s.P50.Round(time.Microsecond))
	fmt.Fprintf(tw, "P95:\t%s\n", s.P95.Round(time.Microsecond))
	fmt.Fprintf(tw, "P99:\t%s\n", s.P99.Round(time.Microsecond))
	fmt.Fprintf(tw, "Max:\t%s\n", s.MaxLatency.Round(time.Microsecond))
	return tw.Flush()
}
