package routes

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// EventLevel represents the severity of a log event.
type EventLevel string

const (
	EventInfo  EventLevel = "INFO"
	EventWarn  EventLevel = "WARN"
	EventError EventLevel = "ERROR"
)

// EventEntry is a single structured log entry.
type EventEntry struct {
	Timestamp time.Time  `json:"timestamp"`
	Level     EventLevel `json:"level"`
	Message   string     `json:"message"`
	Route     string     `json:"route,omitempty"`
}

// DefaultEventLogConfig returns a sensible default configuration.
func DefaultEventLogConfig() EventLogConfig {
	return EventLogConfig{
		Output:   os.Stderr,
		MaxSize:  500,
		MinLevel: EventInfo,
	}
}

// EventLogConfig controls EventLog behaviour.
type EventLogConfig struct {
	Output   io.Writer
	MaxSize  int
	MinLevel EventLevel
}

// EventLog records structured events emitted during route watching.
type EventLog struct {
	mu      sync.Mutex
	entries []EventEntry
	cfg     EventLogConfig
}

// NewEventLog creates an EventLog with the given config.
func NewEventLog(cfg EventLogConfig) *EventLog {
	if cfg.Output == nil {
		cfg.Output = os.Stderr
	}
	if cfg.MaxSize <= 0 {
		cfg.MaxSize = 500
	}
	return &EventLog{cfg: cfg}
}

// Log records an event if its level meets the minimum threshold.
func (el *EventLog) Log(level EventLevel, msg, route string) {
	if !el.meetsLevel(level) {
		return
	}
	entry := EventEntry{
		Timestamp: time.Now().UTC(),
		Level:     level,
		Message:   msg,
		Route:     route,
	}
	el.mu.Lock()
	defer el.mu.Unlock()
	if len(el.entries) >= el.cfg.MaxSize {
		el.entries = el.entries[1:]
	}
	el.entries = append(el.entries, entry)
	fmt.Fprintf(el.cfg.Output, "[%s] %s %s\n", entry.Level, entry.Timestamp.Format(time.RFC3339), entry.Message)
}

// Entries returns a copy of all recorded entries.
func (el *EventLog) Entries() []EventEntry {
	el.mu.Lock()
	defer el.mu.Unlock()
	out := make([]EventEntry, len(el.entries))
	copy(out, el.entries)
	return out
}

// Clear removes all recorded entries.
func (el *EventLog) Clear() {
	el.mu.Lock()
	defer el.mu.Unlock()
	el.entries = nil
}

func (el *EventLog) meetsLevel(level EventLevel) bool {
	order := map[EventLevel]int{EventInfo: 0, EventWarn: 1, EventError: 2}
	return order[level] >= order[el.cfg.MinLevel]
}
