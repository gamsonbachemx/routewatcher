package routes

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// AuditEntry records a single routing table change event for audit purposes.
type AuditEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Added     int       `json:"added"`
	Removed   int       `json:"removed"`
	Details   []string  `json:"details,omitempty"`
}

// AuditConfig holds configuration for the Auditor.
type AuditConfig struct {
	Output   io.Writer
	FilePath string
	IncludeDetails bool
}

// DefaultAuditConfig returns an AuditConfig with sensible defaults.
func DefaultAuditConfig() AuditConfig {
	return AuditConfig{
		Output:         os.Stdout,
		IncludeDetails: false,
	}
}

// Auditor writes structured audit log entries for each observed diff.
type Auditor struct {
	cfg AuditConfig
	mu  sync.Mutex
	f   *os.File
}

// NewAuditor creates a new Auditor. If cfg.FilePath is set, entries are
// appended to that file in addition to cfg.Output.
func NewAuditor(cfg AuditConfig) (*Auditor, error) {
	a := &Auditor{cfg: cfg}
	if cfg.FilePath != "" {
		f, err := os.OpenFile(cfg.FilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("audit: open file: %w", err)
		}
		a.f = f
	}
	return a, nil
}

// Record logs a diff as an audit entry.
func (a *Auditor) Record(d Diff) error {
	entry := AuditEntry{
		Timestamp: time.Now().UTC(),
		Added:     len(d.Added),
		Removed:   len(d.Removed),
	}
	if a.cfg.IncludeDetails {
		for _, r := range d.Added {
			entry.Details = append(entry.Details, fmt.Sprintf("+%s via %s", r.Destination, r.Gateway))
		}
		for _, r := range d.Removed {
			entry.Details = append(entry.Details, fmt.Sprintf("-%s via %s", r.Destination, r.Gateway))
		}
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("audit: marshal: %w", err)
	}
	line := string(data) + "\n"

	a.mu.Lock()
	defer a.mu.Unlock()

	if a.cfg.Output != nil {
		fmt.Fprint(a.cfg.Output, line)
	}
	if a.f != nil {
		fmt.Fprint(a.f, line)
	}
	return nil
}

// Close releases any open file handle held by the Auditor.
func (a *Auditor) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.f != nil {
		return a.f.Close()
	}
	return nil
}
