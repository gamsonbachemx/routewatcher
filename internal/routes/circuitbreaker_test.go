package routes

import (
	"bytes"
	"testing"
	"time"
)

func TestDefaultCircuitBreakerConfig(t *testing.T) {
	cfg := DefaultCircuitBreakerConfig()
	if cfg.MaxFailures != 5 {
		t.Errorf("expected MaxFailures=5, got %d", cfg.MaxFailures)
	}
	if cfg.ResetTimeout != 30*time.Second {
		t.Errorf("expected ResetTimeout=30s, got %v", cfg.ResetTimeout)
	}
	if cfg.Output == nil {
		t.Error("expected non-nil Output")
	}
}

func TestCircuitBreaker_InitialStateClosed(t *testing.T) {
	cb := NewCircuitBreaker(DefaultCircuitBreakerConfig())
	if cb.State() != CircuitClosed {
		t.Errorf("expected closed, got %s", cb.State())
	}
	if !cb.Allow() {
		t.Error("expected Allow()=true in closed state")
	}
}

func TestCircuitBreaker_OpensAfterMaxFailures(t *testing.T) {
	var buf bytes.Buffer
	cfg := CircuitBreakerConfig{MaxFailures: 3, ResetTimeout: time.Minute, Output: &buf}
	cb := NewCircuitBreaker(cfg)

	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}

	if cb.State() != CircuitOpen {
		t.Errorf("expected open, got %s", cb.State())
	}
	if cb.Allow() {
		t.Error("expected Allow()=false in open state")
	}
	if cb.Failures() != 3 {
		t.Errorf("expected 3 failures, got %d", cb.Failures())
	}
}

func TestCircuitBreaker_TransitionsToHalfOpenAfterTimeout(t *testing.T) {
	var buf bytes.Buffer
	cfg := CircuitBreakerConfig{MaxFailures: 1, ResetTimeout: 10 * time.Millisecond, Output: &buf}
	cb := NewCircuitBreaker(cfg)

	cb.RecordFailure()
	if cb.State() != CircuitOpen {
		t.Fatalf("expected open state")
	}

	time.Sleep(20 * time.Millisecond)

	if !cb.Allow() {
		t.Error("expected Allow()=true after reset timeout")
	}
	if cb.State() != CircuitHalfOpen {
		t.Errorf("expected half-open, got %s", cb.State())
	}
}

func TestCircuitBreaker_RecordSuccessCloses(t *testing.T) {
	var buf bytes.Buffer
	cfg := CircuitBreakerConfig{MaxFailures: 1, ResetTimeout: 10 * time.Millisecond, Output: &buf}
	cb := NewCircuitBreaker(cfg)

	cb.RecordFailure()
	time.Sleep(20 * time.Millisecond)
	cb.Allow() // triggers half-open
	cb.RecordSuccess()

	if cb.State() != CircuitClosed {
		t.Errorf("expected closed after success, got %s", cb.State())
	}
	if cb.Failures() != 0 {
		t.Errorf("expected 0 failures after success, got %d", cb.Failures())
	}
}

func TestCircuitBreaker_HalfOpenFailureReopens(t *testing.T) {
	var buf bytes.Buffer
	cfg := CircuitBreakerConfig{MaxFailures: 1, ResetTimeout: 10 * time.Millisecond, Output: &buf}
	cb := NewCircuitBreaker(cfg)

	cb.RecordFailure()
	time.Sleep(20 * time.Millisecond)
	cb.Allow() // half-open
	cb.RecordFailure()

	if cb.State() != CircuitOpen {
		t.Errorf("expected open after half-open failure, got %s", cb.State())
	}
}

func TestCircuitState_String(t *testing.T) {
	cases := []struct {
		state CircuitState
		want  string
	}{
		{CircuitClosed, "closed"},
		{CircuitOpen, "open"},
		{CircuitHalfOpen, "half-open"},
		{CircuitState(99), "unknown"},
	}
	for _, tc := range cases {
		if got := tc.state.String(); got != tc.want {
			t.Errorf("state %d: expected %q, got %q", tc.state, tc.want, got)
		}
	}
}
