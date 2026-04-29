package routes

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// MetricsConfig holds configuration for the metrics collector.
type MetricsConfig struct {
	Output    io.Writer
	ResetOnRead bool
}

// DefaultMetricsConfig returns a MetricsConfig with sensible defaults.
func DefaultMetricsConfig() MetricsConfig {
	return MetricsConfig{
		Output:    os.Stdout,
		ResetOnRead: false,
	}
}

// Metrics tracks runtime statistics for routewatcher.
type Metrics struct {
	mu           sync.Mutex
	cfg          MetricsConfig
	StartTime    time.Time
	PollCount    int64
	ChangeCount  int64
	AlertCount   int64
	ErrorCount   int64
}

// NewMetrics creates a new Metrics instance.
func NewMetrics(cfg MetricsConfig) *Metrics {
	return &Metrics{
		cfg:       cfg,
		StartTime: time.Now(),
	}
}

// RecordPoll increments the poll counter.
func (m *Metrics) RecordPoll() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.PollCount++
}

// RecordChange increments the change counter.
func (m *Metrics) RecordChange() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ChangeCount++
}

// RecordAlert increments the alert counter.
func (m *Metrics) RecordAlert() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.AlertCount++
}

// RecordError increments the error counter.
func (m *Metrics) RecordError() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ErrorCount++
}

// Snapshot returns a copy of current metric values.
func (m *Metrics) Snapshot() Metrics {
	m.mu.Lock()
	defer m.mu.Unlock()
	snap := *m
	if m.cfg.ResetOnRead {
		m.PollCount = 0
		m.ChangeCount = 0
		m.AlertCount = 0
		m.ErrorCount = 0
	}
	return snap
}

// Print writes a human-readable summary of metrics to the configured output.
func (m *Metrics) Print() {
	snap := m.Snapshot()
	uptime := time.Since(snap.StartTime).Round(time.Second)
	fmt.Fprintf(m.cfg.Output,
		"uptime=%s polls=%d changes=%d alerts=%d errors=%d\n",
		uptime, snap.PollCount, snap.ChangeCount, snap.AlertCount, snap.ErrorCount,
	)
}
