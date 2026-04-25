package routes

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func sampleDiffForHistory() Diff {
	return Diff{
		Added:   []Route{{Destination: "10.0.0.0/8", Gateway: "192.168.1.1"}},
		Removed: []Route{},
	}
}

func TestHistory_RecordAndRetrieve(t *testing.T) {
	h := NewHistory(10, "")
	d := sampleDiffForHistory()

	if err := h.Record(d); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entries := h.Entries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if len(entries[0].Diff.Added) != 1 {
		t.Errorf("expected 1 added route, got %d", len(entries[0].Diff.Added))
	}
	if entries[0].Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}

func TestHistory_RingBuffer(t *testing.T) {
	h := NewHistory(3, "")
	d := sampleDiffForHistory()

	for i := 0; i < 5; i++ {
		_ = h.Record(d)
	}

	entries := h.Entries()
	if len(entries) != 3 {
		t.Errorf("expected ring buffer capped at 3, got %d", len(entries))
	}
}

func TestHistory_DefaultMaxSize(t *testing.T) {
	h := NewHistory(0, "")
	if h.maxSize != 100 {
		t.Errorf("expected default maxSize 100, got %d", h.maxSize)
	}
}

func TestHistory_PersistToFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history.jsonl")

	h := NewHistory(10, path)
	d := sampleDiffForHistory()

	if err := h.Record(d); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := h.Record(d); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("could not read history file: %v", err)
	}

	lines := splitLines(data)
	if len(lines) != 2 {
		t.Errorf("expected 2 JSON lines, got %d", len(lines))
	}

	var entry HistoryEntry
	if err := json.Unmarshal([]byte(lines[0]), &entry); err != nil {
		t.Errorf("failed to parse first line as JSON: %v", err)
	}
}

func splitLines(data []byte) []string {
	var lines []string
	start := 0
	for i, b := range data {
		if b == '\n' && i > start {
			lines = append(lines, string(data[start:i]))
			start = i + 1
		}
	}
	return lines
}
