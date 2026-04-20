package backoff_test

import (
	"testing"

	"github.com/example/grpcannon/backoff"
)

// BenchmarkDelay measures the cost of computing a single backoff delay.
func BenchmarkDelay(b *testing.B) {
	s := backoff.Default()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.Delay(i % 10)
	}
}

// BenchmarkSteps measures the cost of generating a slice of delays.
func BenchmarkSteps(b *testing.B) {
	s := backoff.Default()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.Steps(10)
	}
}
