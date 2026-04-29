package routes

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestDefaultMetricsConfig(t *testing.T) {
	cfg := DefaultMetricsConfig()
	if cfg.Output == nil {
		t.Fatal("expected non-nil Output")
	}
	if cfg.ResetOnRead {
		t.Fatal("expected ResetOnRead to be false by default")
	}
}

func TestMetrics_RecordAndSnapshot(t *testing.T) {
	m := NewMetrics(DefaultMetricsConfig())

	m.RecordPoll()
	m.RecordPoll()
	m.RecordChange()
	m.RecordAlert()
	m.RecordError()

	snap := m.Snapshot()
	if snap.PollCount != 2 {
		t.Errorf("expected PollCount=2, got %d", snap.PollCount)
	}
	if snap.ChangeCount != 1 {
		t.Errorf("expected ChangeCount=1, got %d", snap.ChangeCount)
	}
	if snap.AlertCount != 1 {
		t.Errorf("expected AlertCount=1, got %d", snap.AlertCount)
	}
	if snap.ErrorCount != 1 {
		t.Errorf("expected ErrorCount=1, got %d", snap.ErrorCount)
	}
}

func TestMetrics_ResetOnRead(t *testing.T) {
	cfg := DefaultMetricsConfig()
	cfg.ResetOnRead = true
	m := NewMetrics(cfg)

	m.RecordPoll()
	m.RecordChange()

	first := m.Snapshot()
	if first.PollCount != 1 || first.ChangeCount != 1 {
		t.Fatalf("unexpected first snapshot: polls=%d changes=%d", first.PollCount, first.ChangeCount)
	}

	second := m.Snapshot()
	if second.PollCount != 0 || second.ChangeCount != 0 {
		t.Errorf("expected counters reset after read, got polls=%d changes=%d", second.PollCount, second.ChangeCount)
	}
}

func TestMetrics_StartTimeSet(t *testing.T) {
	before := time.Now()
	m := NewMetrics(DefaultMetricsConfig())
	after := time.Now()

	snap := m.Snapshot()
	if snap.StartTime.Before(before) || snap.StartTime.After(after) {
		t.Errorf("StartTime %v not in expected range [%v, %v]", snap.StartTime, before, after)
	}
}

func TestMetrics_Print(t *testing.T) {
	var buf bytes.Buffer
	cfg := DefaultMetricsConfig()
	cfg.Output = &buf
	m := NewMetrics(cfg)

	m.RecordPoll()
	m.RecordPoll()
	m.RecordChange()

	m.Print()

	out := buf.String()
	for _, want := range []string{"polls=2", "changes=1", "alerts=0", "errors=0", "uptime="} {
		if !strings.Contains(out, want) {
			t.Errorf("expected output to contain %q, got: %s", want, out)
		}
	}
}
