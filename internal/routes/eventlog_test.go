package routes

import (
	"bytes"
	"strings"
	"testing"
)

func TestDefaultEventLogConfig(t *testing.T) {
	cfg := DefaultEventLogConfig()
	if cfg.MaxSize != 500 {
		t.Errorf("expected MaxSize 500, got %d", cfg.MaxSize)
	}
	if cfg.MinLevel != EventInfo {
		t.Errorf("expected MinLevel INFO, got %s", cfg.MinLevel)
	}
	if cfg.Output == nil {
		t.Error("expected non-nil Output")
	}
}

func TestEventLog_LogAndRetrieve(t *testing.T) {
	var buf bytes.Buffer
	cfg := EventLogConfig{Output: &buf, MaxSize: 100, MinLevel: EventInfo}
	el := NewEventLog(cfg)

	el.Log(EventInfo, "route added", "10.0.0.0/8")
	el.Log(EventWarn, "route flap detected", "192.168.1.0/24")

	entries := el.Entries()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Level != EventInfo {
		t.Errorf("expected INFO, got %s", entries[0].Level)
	}
	if entries[1].Route != "192.168.1.0/24" {
		t.Errorf("unexpected route: %s", entries[1].Route)
	}
	if !strings.Contains(buf.String(), "route added") {
		t.Error("expected output to contain 'route added'")
	}
}

func TestEventLog_RingBuffer(t *testing.T) {
	var buf bytes.Buffer
	cfg := EventLogConfig{Output: &buf, MaxSize: 3, MinLevel: EventInfo}
	el := NewEventLog(cfg)

	for i := 0; i < 5; i++ {
		el.Log(EventInfo, "msg", "")
	}
	if len(el.Entries()) != 3 {
		t.Errorf("expected ring buffer size 3, got %d", len(el.Entries()))
	}
}

func TestEventLog_MinLevelFilters(t *testing.T) {
	var buf bytes.Buffer
	cfg := EventLogConfig{Output: &buf, MaxSize: 100, MinLevel: EventWarn}
	el := NewEventLog(cfg)

	el.Log(EventInfo, "this should be filtered", "")
	el.Log(EventWarn, "this should pass", "")
	el.Log(EventError, "this should also pass", "")

	entries := el.Entries()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries after filtering, got %d", len(entries))
	}
	if entries[0].Level != EventWarn {
		t.Errorf("expected WARN, got %s", entries[0].Level)
	}
}

func TestEventLog_Clear(t *testing.T) {
	var buf bytes.Buffer
	cfg := EventLogConfig{Output: &buf, MaxSize: 100, MinLevel: EventInfo}
	el := NewEventLog(cfg)

	el.Log(EventInfo, "test", "")
	el.Clear()

	if len(el.Entries()) != 0 {
		t.Error("expected entries to be cleared")
	}
}

func TestEventLog_NilOutputUsesDefault(t *testing.T) {
	cfg := EventLogConfig{MaxSize: 10}
	el := NewEventLog(cfg)
	if el.cfg.Output == nil {
		t.Error("expected fallback output to be set")
	}
}
