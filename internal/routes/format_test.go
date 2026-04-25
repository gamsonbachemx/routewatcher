package routes

import (
	"strings"
	"testing"
	"time"
)

var testTime = time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)

func TestDiff_HasChanges(t *testing.T) {
	tests := []struct {
		name     string
		diff     Diff
		expected bool
	}{
		{"no changes", Diff{At: testTime}, false},
		{"added only", Diff{Added: []string{"192.168.1.0/24"}, At: testTime}, true},
		{"removed only", Diff{Removed: []string{"10.0.0.0/8"}, At: testTime}, true},
		{"both", Diff{Added: []string{"1.2.3.0/24"}, Removed: []string{"10.0.0.0/8"}, At: testTime}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.diff.HasChanges(); got != tt.expected {
				t.Errorf("HasChanges() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestDiff_Events(t *testing.T) {
	d := Diff{
		Added:   []string{"192.168.1.0/24"},
		Removed: []string{"10.0.0.0/8"},
		At:      testTime,
	}
	events := d.Events()
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	if events[0].Type != Added || events[0].Route != "192.168.1.0/24" {
		t.Errorf("unexpected first event: %+v", events[0])
	}
	if events[1].Type != Removed || events[1].Route != "10.0.0.0/8" {
		t.Errorf("unexpected second event: %+v", events[1])
	}
}

func TestFormatText_WithChanges(t *testing.T) {
	d := &Diff{
		Added:   []string{"192.168.1.0/24"},
		Removed: []string{"10.0.0.0/8"},
		At:      testTime,
	}
	out := FormatText(d)
	if !strings.Contains(out, "+ 192.168.1.0/24") {
		t.Errorf("expected added route in output, got: %s", out)
	}
	if !strings.Contains(out, "- 10.0.0.0/8") {
		t.Errorf("expected removed route in output, got: %s", out)
	}
}

func TestFormatText_NoChanges(t *testing.T) {
	d := &Diff{At: testTime}
	out := FormatText(d)
	if !strings.Contains(out, "No route changes") {
		t.Errorf("expected no-changes message, got: %s", out)
	}
}

func TestFormatJSON_WithChanges(t *testing.T) {
	d := &Diff{
		Added:   []string{"192.168.1.0/24"},
		Removed: []string{"10.0.0.0/8"},
		At:      testTime,
	}
	out := FormatJSON(d)
	if !strings.Contains(out, `"added"`) || !strings.Contains(out, `"removed"`) {
		t.Errorf("malformed JSON output: %s", out)
	}
	if !strings.Contains(out, "192.168.1.0/24") {
		t.Errorf("missing added route in JSON: %s", out)
	}
	if !strings.Contains(out, "10.0.0.0/8") {
		t.Errorf("missing removed route in JSON: %s", out)
	}
}
