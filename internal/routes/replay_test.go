package routes

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func writeHistoryFile(t *testing.T, entries []HistoryEntry) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "history.jsonl")
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create file: %v", err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	for _, e := range entries {
		if err := enc.Encode(e); err != nil {
			t.Fatalf("encode: %v", err)
		}
	}
	return path
}

func TestLoadHistory_AllEntries(t *testing.T) {
	now := time.Now().UTC()
	entries := []HistoryEntry{
		{Timestamp: now.Add(-2 * time.Minute), Diff: Diff{Added: []Route{{Destination: "10.0.0.0/8"}}}},
		{Timestamp: now.Add(-1 * time.Minute), Diff: Diff{Removed: []Route{{Destination: "172.16.0.0/12"}}}},
	}
	path := writeHistoryFile(t, entries)

	got, err := LoadHistory(path, ReplayOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 entries, got %d", len(got))
	}
}

func TestLoadHistory_SinceFilter(t *testing.T) {
	now := time.Now().UTC()
	cutoff := now.Add(-90 * time.Second)
	entries := []HistoryEntry{
		{Timestamp: now.Add(-2 * time.Minute), Diff: Diff{}},
		{Timestamp: now.Add(-30 * time.Second), Diff: Diff{}},
	}
	path := writeHistoryFile(t, entries)

	got, err := LoadHistory(path, ReplayOptions{Since: cutoff})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("expected 1 entry after cutoff, got %d", len(got))
	}
}

func TestLoadHistory_Limit(t *testing.T) {
	now := time.Now().UTC()
	entries := []HistoryEntry{
		{Timestamp: now.Add(-3 * time.Minute)},
		{Timestamp: now.Add(-2 * time.Minute)},
		{Timestamp: now.Add(-1 * time.Minute)},
	}
	path := writeHistoryFile(t, entries)

	got, err := LoadHistory(path, ReplayOptions{Limit: 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 entries with limit, got %d", len(got))
	}
}

func TestLoadHistory_InvalidFile(t *testing.T) {
	_, err := LoadHistory("/nonexistent/path.jsonl", ReplayOptions{})
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestLoadHistory_MalformedLine(t *testing.T) {
	r := strings.NewReader("{bad json\n")
	_, err := parseHistoryReader(r, ReplayOptions{})
	if err == nil {
		t.Error("expected error for malformed JSON line")
	}
}

func TestLoadHistory_EmptyFile(t *testing.T) {
	path := writeHistoryFile(t, []HistoryEntry{})

	got, err := LoadHistory(path, ReplayOptions{})
	if err != nil {
		t.Fatalf("unexpected error for empty file: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected 0 entries for empty file, got %d", len(got))
	}
}
