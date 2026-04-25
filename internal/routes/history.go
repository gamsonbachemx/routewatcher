package routes

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// HistoryEntry records a diff event with a timestamp.
type HistoryEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Diff      Diff      `json:"diff"`
}

// History maintains an in-memory ring buffer of recent diff events
// and optionally persists them to a JSON file.
type History struct {
	mu      sync.RWMutex
	entries []HistoryEntry
	maxSize int
	filePath string
}

// NewHistory creates a History with the given capacity and optional file path.
// If filePath is non-empty, entries are appended to that file.
func NewHistory(maxSize int, filePath string) *History {
	if maxSize <= 0 {
		maxSize = 100
	}
	return &History{
		maxSize:  maxSize,
		filePath: filePath,
	}
}

// Record adds a Diff to the history buffer and persists it if a file is configured.
func (h *History) Record(d Diff) error {
	entry := HistoryEntry{
		Timestamp: time.Now().UTC(),
		Diff:      d,
	}

	h.mu.Lock()
	h.entries = append(h.entries, entry)
	if len(h.entries) > h.maxSize {
		h.entries = h.entries[len(h.entries)-h.maxSize:]
	}
	h.mu.Unlock()

	if h.filePath != "" {
		return h.persist(entry)
	}
	return nil
}

// Entries returns a copy of all recorded history entries.
func (h *History) Entries() []HistoryEntry {
	h.mu.RLock()
	defer h.mu.RUnlock()
	out := make([]HistoryEntry, len(h.entries))
	copy(out, h.entries)
	return out
}

// persist appends a single entry as a JSON line to the history file.
func (h *History) persist(entry HistoryEntry) error {
	f, err := os.OpenFile(h.filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("history: open file: %w", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	if err := enc.Encode(entry); err != nil {
		return fmt.Errorf("history: encode entry: %w", err)
	}
	return nil
}
