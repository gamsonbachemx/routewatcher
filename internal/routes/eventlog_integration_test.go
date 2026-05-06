package routes

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestEventLog_IntegratesWithDiff(t *testing.T) {
	var buf bytes.Buffer
	cfg := EventLogConfig{Output: &buf, MaxSize: 100, MinLevel: EventInfo}
	el := NewEventLog(cfg)

	diff := Diff{
		Added: []Route{
			{Destination: "10.1.0.0/16", Gateway: "192.168.1.1", Iface: "eth0"},
		},
		Removed: []Route{
			{Destination: "10.2.0.0/16", Gateway: "192.168.1.1", Iface: "eth0"},
		},
	}

	for _, r := range diff.Added {
		el.Log(EventInfo, "route added: "+r.Destination, r.Destination)
	}
	for _, r := range diff.Removed {
		el.Log(EventWarn, "route removed: "+r.Destination, r.Destination)
	}

	entries := el.Entries()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Level != EventInfo || entries[1].Level != EventWarn {
		t.Error("unexpected event levels")
	}
	if !strings.Contains(buf.String(), "10.1.0.0/16") {
		t.Error("expected output to mention added route")
	}
}

func TestEventLog_TimestampsAreSet(t *testing.T) {
	var buf bytes.Buffer
	before := time.Now().UTC()
	cfg := EventLogConfig{Output: &buf, MaxSize: 10, MinLevel: EventInfo}
	el := NewEventLog(cfg)
	el.Log(EventInfo, "check timestamp", "")
	after := time.Now().UTC()

	entries := el.Entries()
	if len(entries) == 0 {
		t.Fatal("expected at least one entry")
	}
	ts := entries[0].Timestamp
	if ts.Before(before) || ts.After(after) {
		t.Errorf("timestamp %v not in expected range [%v, %v]", ts, before, after)
	}
}
