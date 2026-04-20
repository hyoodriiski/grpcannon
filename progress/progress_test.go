package progress_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/your-org/grpcannon/progress"
)

func TestRecord_IncrementsTotals(t *testing.T) {
	var buf bytes.Buffer
	r := progress.New(&buf, time.Second)

	r.Record(true)
	r.Record(true)
	r.Record(false)

	r.Stop()

	out := buf.String()
	if !strings.Contains(out, "total=3") {
		t.Errorf("expected total=3 in output, got: %s", out)
	}
	if !strings.Contains(out, "success=2") {
		t.Errorf("expected success=2 in output, got: %s", out)
	}
	if !strings.Contains(out, "failures=1") {
		t.Errorf("expected failures=1 in output, got: %s", out)
	}
}

func TestStop_PrintsFinalLine(t *testing.T) {
	var buf bytes.Buffer
	r := progress.New(&buf, time.Second)
	r.Stop()

	if buf.Len() == 0 {
		t.Error("expected at least one output line after Stop")
	}
}

func TestStart_PeriodicOutput(t *testing.T) {
	var buf bytes.Buffer
	r := progress.New(&buf, 20*time.Millisecond)
	r.Start()

	r.Record(true)
	r.Record(false)

	time.Sleep(60 * time.Millisecond)
	r.Stop()

	lines := strings.Count(buf.String(), "[progress]")
	// at least 2 ticks + 1 final
	if lines < 2 {
		t.Errorf("expected at least 2 progress lines, got %d", lines)
	}
}

func TestNew_DefaultInterval(t *testing.T) {
	var buf bytes.Buffer
	// zero interval should default to 1 second (no panic)
	r := progress.New(&buf, 0)
	if r == nil {
		t.Fatal("expected non-nil reporter")
	}
	r.Stop()
}
