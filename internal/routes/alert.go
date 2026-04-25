package routes

import (
	"fmt"
	"io"
	"os"
	"time"
)

// AlertLevel represents the severity of a routing change.
type AlertLevel string

const (
	AlertInfo    AlertLevel = "INFO"
	AlertWarning AlertLevel = "WARNING"
	AlertCritical AlertLevel = "CRITICAL"
)

// AlertConfig controls alerting behaviour.
type AlertConfig struct {
	// MinChanges is the minimum number of changes to trigger an alert.
	MinChanges int
	// Level is the severity level to emit.
	Level AlertLevel
	// Output is the writer used for alert messages (defaults to os.Stderr).
	Output io.Writer
}

// DefaultAlertConfig returns an AlertConfig with sensible defaults.
func DefaultAlertConfig() AlertConfig {
	return AlertConfig{
		MinChanges: 1,
		Level:      AlertWarning,
		Output:     os.Stderr,
	}
}

// Alerter emits alert messages when routing changes exceed a threshold.
type Alerter struct {
	cfg AlertConfig
}

// NewAlerter creates an Alerter using the provided AlertConfig.
func NewAlerter(cfg AlertConfig) *Alerter {
	if cfg.Output == nil {
		cfg.Output = os.Stderr
	}
	return &Alerter{cfg: cfg}
}

// Notify checks the diff and writes an alert if the change threshold is met.
func (a *Alerter) Notify(d Diff) {
	total := len(d.Added) + len(d.Removed)
	if total < a.cfg.MinChanges {
		return
	}
	timestamp := time.Now().UTC().Format(time.RFC3339)
	fmt.Fprintf(
		a.cfg.Output,
		"[%s] [%s] route change detected: +%d added, -%d removed\n",
		timestamp,
		a.cfg.Level,
		len(d.Added),
		len(d.Removed)),
}
