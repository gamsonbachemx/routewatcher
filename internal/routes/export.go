package routes

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// ExportConfig holds configuration for exporting route diffs to a file.
type ExportConfig struct {
	FilePath string
	Format   string // "text" or "json"
	Append   bool
}

// Exporter writes route diffs to a file.
type Exporter struct {
	cfg ExportConfig
}

// NewExporter creates a new Exporter with the given config.
func NewExporter(cfg ExportConfig) (*Exporter, error) {
	if cfg.FilePath == "" {
		return nil, fmt.Errorf("export file path must not be empty")
	}
	if cfg.Format != "text" && cfg.Format != "json" {
		return nil, fmt.Errorf("unsupported export format %q: must be \"text\" or \"json\"", cfg.Format)
	}
	return &Exporter{cfg: cfg}, nil
}

// ExportRecord wraps a Diff with a timestamp for serialisation.
type ExportRecord struct {
	Timestamp time.Time `json:"timestamp"`
	Diff      Diff      `json:"diff"`
}

// Write appends or overwrites the export file with the given diff.
func (e *Exporter) Write(d Diff) error {
	flag := os.O_CREATE | os.O_WRONLY
	if e.cfg.Append {
		flag |= os.O_APPEND
	} else {
		flag |= os.O_TRUNC
	}

	f, err := os.OpenFile(e.cfg.FilePath, flag, 0o644)
	if err != nil {
		return fmt.Errorf("opening export file: %w", err)
	}
	defer f.Close()

	switch e.cfg.Format {
	case "json":
		rec := ExportRecord{Timestamp: time.Now().UTC(), Diff: d}
		enc := json.NewEncoder(f)
		enc.SetIndent("", "  ")
		if err := enc.Encode(rec); err != nil {
			return fmt.Errorf("encoding json export: %w", err)
		}
	case "text":
		line := fmt.Sprintf("[%s]\n%s\n", time.Now().UTC().Format(time.RFC3339), FormatText(d))
		if _, err := fmt.Fprint(f, line); err != nil {
			return fmt.Errorf("writing text export: %w", err)
		}
	}
	return nil
}
