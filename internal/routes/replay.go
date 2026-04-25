package routes

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

// ReplayOptions controls how a history file is replayed.
type ReplayOptions struct {
	// Since filters entries recorded after this time. Zero means no filter.
	Since time.Time
	// Limit caps the number of entries returned. Zero means no limit.
	Limit int
}

// LoadHistory reads a JSONL history file and returns matching entries.
func LoadHistory(path string, opts ReplayOptions) ([]HistoryEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("replay: open %q: %w", path, err)
	}
	defer f.Close()
	return parseHistoryReader(f, opts)
}

func parseHistoryReader(r io.Reader, opts ReplayOptions) ([]HistoryEntry, error) {
	var entries []HistoryEntry
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var entry HistoryEntry
		if err := json.Unmarshal(line, &entry); err != nil {
			return nil, fmt.Errorf("replay: parse line: %w", err)
		}
		if !opts.Since.IsZero() && !entry.Timestamp.After(opts.Since) {
			continue
		}
		entries = append(entries, entry)
		if opts.Limit > 0 && len(entries) >= opts.Limit {
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("replay: scan: %w", err)
	}
	return entries, nil
}
