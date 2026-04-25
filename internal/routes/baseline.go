package routes

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// BaselineEntry holds a saved routing snapshot used as a reference point.
type BaselineEntry struct {
	CapturedAt time.Time `json:"captured_at"`
	Routes     Snapshot  `json:"routes"`
}

// BaselineStore manages loading and saving a routing baseline to disk.
type BaselineStore struct {
	mu   sync.RWMutex
	path string
}

// NewBaselineStore creates a BaselineStore backed by the given file path.
func NewBaselineStore(path string) (*BaselineStore, error) {
	if path == "" {
		return nil, fmt.Errorf("baseline path must not be empty")
	}
	return &BaselineStore{path: path}, nil
}

// Save writes the given snapshot as the current baseline.
func (b *BaselineStore) Save(s Snapshot) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	entry := BaselineEntry{
		CapturedAt: time.Now().UTC(),
		Routes:     s,
	}
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal baseline: %w", err)
	}
	if err := os.WriteFile(b.path, data, 0644); err != nil {
		return fmt.Errorf("write baseline: %w", err)
	}
	return nil
}

// Load reads the baseline from disk and returns the entry.
func (b *BaselineStore) Load() (*BaselineEntry, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	data, err := os.ReadFile(b.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no baseline found at %s", b.path)
		}
		return nil, fmt.Errorf("read baseline: %w", err)
	}
	var entry BaselineEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, fmt.Errorf("parse baseline: %w", err)
	}
	return &entry, nil
}

// CompareToBaseline captures the current routing table and diffs it against
// the stored baseline, returning the diff and baseline metadata.
func (b *BaselineStore) CompareToBaseline() (*Diff, *BaselineEntry, error) {
	entry, err := b.Load()
	if err != nil {
		return nil, nil, err
	}
	current, err := Capture()
	if err != nil {
		return nil, nil, fmt.Errorf("capture routes: %w", err)
	}
	d := Compare(entry.Routes, current)
	return &d, entry, nil
}
