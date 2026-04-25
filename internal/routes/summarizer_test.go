package routes

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestSummarizer_FlushOnStop(t *testing.T) {
	var buf bytes.Buffer
	cfg := SummaryConfig{
		Interval: 10 * time.Second, // long enough not to fire during test
		Output:   &buf,
	}
	s := NewSummarizer(cfg)
	s.Record(makeDiff(
		[]Route{{Destination: "10.0.0.0/8", Iface: "eth0", Protocol: "kernel"}},
		nil,
	))
	s.Stop()

	out := buf.String()
	if !strings.Contains(out, "Summary") {
		t.Errorf("expected Summary in output, got: %q", out)
	}
	if !strings.Contains(out, "Added:   1") {
		t.Errorf("expected Added: 1 in output, got: %q", out)
	}
}

func TestSummarizer_NoOutputWhenEmpty(t *testing.T) {
	var buf bytes.Buffer
	cfg := SummaryConfig{
		Interval: 10 * time.Second,
		Output:   &buf,
	}
	s := NewSummarizer(cfg)
	s.Stop()

	if buf.Len() != 0 {
		t.Errorf("expected no output for empty buffer, got: %q", buf.String())
	}
}

func TestSummarizer_TickerFlush(t *testing.T) {
	var buf bytes.Buffer
	cfg := SummaryConfig{
		Interval: 50 * time.Millisecond,
		Output:   &buf,
	}
	s := NewSummarizer(cfg)
	s.Record(makeDiff(
		[]Route{{Destination: "192.168.0.0/16", Iface: "eth1", Protocol: "static"}},
		nil,
	))

	time.Sleep(120 * time.Millisecond)
	s.Stop()

	out := buf.String()
	if !strings.Contains(out, "Summary") {
		t.Errorf("expected Summary in ticker-flushed output, got: %q", out)
	}
}

func TestDefaultSummaryConfig(t *testing.T) {
	cfg := DefaultSummaryConfig()
	if cfg.Interval != 5*time.Minute {
		t.Errorf("expected 5m interval, got %v", cfg.Interval)
	}
	if cfg.Output == nil {
		t.Error("expected non-nil Output")
	}
}
