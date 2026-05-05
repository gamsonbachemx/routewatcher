package routes

import (
	"errors"
	"testing"
	"time"
)

// TestHealthMonitor_IntegratesWithWatcher verifies that the HealthMonitor
// correctly reflects state when wired up alongside a Watch-style poll loop.
func TestHealthMonitor_IntegratesWithWatcher(t *testing.T) {
	cfg := DefaultHealthConfig()
	cfg.MaxErrors = 3
	cfg.StalenessAge = 500 * time.Millisecond
	h := NewHealthMonitor(cfg)

	// Simulate several successful poll cycles.
	for i := 0; i < 5; i++ {
		h.RecordCheck()
		time.Sleep(10 * time.Millisecond)
	}

	if !h.Status().OK {
		t.Fatal("expected OK after successful poll cycles")
	}

	// Simulate a change event.
	h.RecordChange()
	if h.Status().LastChange.IsZero() {
		t.Error("expected LastChange to be populated")
	}

	// Inject errors up to but not exceeding the threshold.
	h.RecordError(errors.New("transient error"))
	h.RecordError(errors.New("transient error"))
	if !h.Status().OK {
		t.Error("expected OK while below MaxErrors threshold")
	}

	// One more error should tip it over.
	h.RecordError(errors.New("fatal error"))
	if h.Status().OK {
		t.Error("expected DEGRADED after exceeding MaxErrors")
	}
}

// TestHealthMonitor_UptimeIncreases confirms that uptime grows over time.
func TestHealthMonitor_UptimeIncreases(t *testing.T) {
	h := NewHealthMonitor(DefaultHealthConfig())
	s1 := h.Status()
	time.Sleep(20 * time.Millisecond)
	s2 := h.Status()

	if s2.UptimeSeconds <= s1.UptimeSeconds {
		t.Errorf("expected uptime to increase: %.4f -> %.4f", s1.UptimeSeconds, s2.UptimeSeconds)
	}
}
