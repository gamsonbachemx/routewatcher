package routes

import (
	"fmt"
	"strings"
	"time"
)

// ChangeType represents the type of route change.
type ChangeType string

const (
	Added   ChangeType = "ADDED"
	Removed ChangeType = "REMOVED"
)

// ChangeEvent represents a single route change with metadata.
type ChangeEvent struct {
	Type      ChangeType
	Route     string
	Timestamp time.Time
}

// Diff holds the result of comparing two route snapshots.
type Diff struct {
	Added   []string
	Removed []string
	At      time.Time
}

// HasChanges returns true if there are any added or removed routes.
func (d *Diff) HasChanges() bool {
	return len(d.Added) > 0 || len(d.Removed) > 0
}

// Events converts a Diff into a slice of ChangeEvents.
func (d *Diff) Events() []ChangeEvent {
	events := make([]ChangeEvent, 0, len(d.Added)+len(d.Removed))
	for _, r := range d.Added {
		events = append(events, ChangeEvent{Type: Added, Route: r, Timestamp: d.At})
	}
	for _, r := range d.Removed {
		events = append(events, ChangeEvent{Type: Removed, Route: r, Timestamp: d.At})
	}
	return events
}

// FormatText returns a human-readable text representation of the Diff.
func FormatText(d *Diff) string {
	if !d.HasChanges() {
		return fmt.Sprintf("[%s] No route changes detected.\n", d.At.Format(time.RFC3339))
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("[%s] Route changes detected:\n", d.At.Format(time.RFC3339)))

	for _, r := range d.Added {
		sb.WriteString(fmt.Sprintf("  + %s\n", r))
	}
	for _, r := range d.Removed {
		sb.WriteString(fmt.Sprintf("  - %s\n", r))
	}

	return sb.String()
}

// FormatJSON returns a simple JSON representation of the Diff.
func FormatJSON(d *Diff) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(`{"timestamp":%q,"added":[`, d.At.Format(time.RFC3339)))
	for i, r := range d.Added {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(fmt.Sprintf("%q", r))
	}
	sb.WriteString(`],"removed":[`)
	for i, r := range d.Removed {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(fmt.Sprintf("%q", r))
	}
	sb.WriteString("]}")  
	return sb.String()
}
