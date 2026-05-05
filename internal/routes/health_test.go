package routes

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestDefaultHealthConfig(t *testing.T) {
	cfg := DefaultHealthConfig()
	if cfg.MaxErrors != 5 {
		t.Errorf("expected MaxErrors=5, got %d", cfg.MaxErrors)
	}
	if cfg.StalenessAge != 5*time.Minute {
		t.Errorf("unexpected StalenessAge: %v", cfg.StalenessAge)
	}
	if cfg.Output == nil {
		t.Error("expected non-nil Output")
	}
}

func TestHealthMonitor_InitialStatusOK(t *testing.T) {
	h := NewHealthMonitor(DefaultHealthConfig())
	s := h.Status()
	// No checks recorded yet — last_check is zero which is within staleness window
	// because zero time satisfies the IsZero branch.
	if !s.OK {
		t.Error("expected initial status to be OK")
	}
	if s.ErrorCount != 0 {
		t.Errorf("expected 0 errors, got %d", s.ErrorCount)
	}
}

func TestHealthMonitor_RecordError(t *testing.T) {
	cfg := DefaultHealthConfig()
	cfg.MaxErrors = 2
	h := NewHealthMonitor(cfg)

	h.RecordError(errors.New("route capture failed"))
	if h.Status().ErrorCount != 1 {
		t.Error("expected error count 1")
	}
	if h.Status().LastError != "route capture failed" {
		t.Error("expected last error message to be stored")
	}

	h.RecordError(errors.New("second error"))
	if h.Status().OK {
		t.Error("expected status DEGRADED after reaching MaxErrors")
	}
}

func TestHealthMonitor_RecordCheckUpdatesTimestamp(t *testing.T) {
	h := NewHealthMonitor(DefaultHealthConfig())
	before := time.Now()
	h.RecordCheck()
	after := time.Now()

	s := h.Status()
	if s.LastCheck.Before(before) || s.LastCheck.After(after) {
		t.Errorf("LastCheck %v not between %v and %v", s.LastCheck, before, after)
	}
}

func TestHealthMonitor_StaleCheckDegraded(t *testing.T) {
	cfg := DefaultHealthConfig()
	cfg.StalenessAge = 1 * time.Millisecond
	h := NewHealthMonitor(cfg)
	h.RecordCheck()
	time.Sleep(5 * time.Millisecond)

	if h.Status().OK {
		t.Error("expected DEGRADED status after stale check")
	}
}

func TestHealthMonitor_RecordChange(t *testing.T) {
	h := NewHealthMonitor(DefaultHealthConfig())
	h.RecordChange()
	s := h.Status()
	if s.LastChange.IsZero() {
		t.Error("expected LastChange to be set")
	}
}

func TestHealthMonitor_Print(t *testing.T) {
	var buf bytes.Buffer
	cfg := DefaultHealthConfig()
	cfg.Output = &buf
	h := NewHealthMonitor(cfg)
	h.RecordCheck()
	h.Print()

	out := buf.String()
	if !strings.Contains(out, "health:") {
		t.Errorf("expected 'health:' in output, got: %s", out)
	}
	if !strings.Contains(out, "OK") {
		t.Errorf("expected 'OK' in output, got: %s", out)
	}
}
