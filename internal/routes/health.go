package routes

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// HealthStatus represents the current health state of the watcher.
type HealthStatus struct {
	OK           bool      `json:"ok"`
	LastCheck    time.Time `json:"last_check"`
	LastChange   time.Time `json:"last_change,omitempty"`
	ErrorCount   int       `json:"error_count"`
	LastError    string    `json:"last_error,omitempty"`
	UptimeSeconds float64  `json:"uptime_seconds"`
}

// DefaultHealthConfig returns a HealthConfig with sensible defaults.
func DefaultHealthConfig() HealthConfig {
	return HealthConfig{
		Output:       os.Stdout,
		MaxErrors:    5,
		StalenessAge: 5 * time.Minute,
	}
}

// HealthConfig controls the behaviour of the HealthMonitor.
type HealthConfig struct {
	Output       io.Writer
	MaxErrors    int
	StalenessAge time.Duration
}

// HealthMonitor tracks watcher liveness and error counts.
type HealthMonitor struct {
	mu        sync.Mutex
	cfg       HealthConfig
	start     time.Time
	lastCheck time.Time
	lastChange time.Time
	errorCount int
	lastError  string
}

// NewHealthMonitor creates a new HealthMonitor using the given config.
func NewHealthMonitor(cfg HealthConfig) *HealthMonitor {
	if cfg.Output == nil {
		cfg.Output = os.Stdout
	}
	if cfg.MaxErrors == 0 {
		cfg.MaxErrors = 5
	}
	if cfg.StalenessAge == 0 {
		cfg.StalenessAge = 5 * time.Minute
	}
	return &HealthMonitor{cfg: cfg, start: time.Now()}
}

// RecordCheck marks that a poll cycle completed successfully.
func (h *HealthMonitor) RecordCheck() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.lastCheck = time.Now()
}

// RecordChange marks that a routing table change was observed.
func (h *HealthMonitor) RecordChange() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.lastChange = time.Now()
}

// RecordError increments the error counter and stores the message.
func (h *HealthMonitor) RecordError(err error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.errorCount++
	if err != nil {
		h.lastError = err.Error()
	}
}

// Status returns a snapshot of the current health state.
func (h *HealthMonitor) Status() HealthStatus {
	h.mu.Lock()
	defer h.mu.Unlock()

	ok := h.errorCount < h.cfg.MaxErrors &&
		(h.lastCheck.IsZero() || time.Since(h.lastCheck) < h.cfg.StalenessAge)

	return HealthStatus{
		OK:            ok,
		LastCheck:     h.lastCheck,
		LastChange:    h.lastChange,
		ErrorCount:    h.errorCount,
		LastError:     h.lastError,
		UptimeSeconds: time.Since(h.start).Seconds(),
	}
}

// Print writes a human-readable health summary to the configured output.
func (h *HealthMonitor) Print() {
	s := h.Status()
	status := "OK"
	if !s.OK {
		status = "DEGRADED"
	}
	fmt.Fprintf(h.cfg.Output, "health: %s | uptime: %.0fs | errors: %d | last_check: %s\n",
		status, s.UptimeSeconds, s.ErrorCount, s.LastCheck.Format(time.RFC3339))
}
