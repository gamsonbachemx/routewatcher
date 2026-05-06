package routes

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// CircuitState represents the state of the circuit breaker.
type CircuitState int

const (
	CircuitClosed CircuitState = iota
	CircuitOpen
	CircuitHalfOpen
)

func (s CircuitState) String() string {
	switch s {
	case CircuitClosed:
		return "closed"
	case CircuitOpen:
		return "open"
	case CircuitHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// CircuitBreakerConfig holds configuration for the circuit breaker.
type CircuitBreakerConfig struct {
	// MaxFailures is the number of consecutive failures before opening the circuit.
	MaxFailures int
	// ResetTimeout is how long to wait in open state before transitioning to half-open.
	ResetTimeout time.Duration
	// Output is where state change messages are written.
	Output io.Writer
}

// DefaultCircuitBreakerConfig returns sensible defaults.
func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		MaxFailures:  5,
		ResetTimeout: 30 * time.Second,
		Output:       os.Stdout,
	}
}

// CircuitBreaker protects downstream calls by tracking failure rates.
type CircuitBreaker struct {
	cfg      CircuitBreakerConfig
	mu       sync.Mutex
	state    CircuitState
	failures int
	openedAt time.Time
}

// NewCircuitBreaker creates a new CircuitBreaker with the given config.
func NewCircuitBreaker(cfg CircuitBreakerConfig) *CircuitBreaker {
	if cfg.Output == nil {
		cfg.Output = os.Stdout
	}
	if cfg.MaxFailures <= 0 {
		cfg.MaxFailures = DefaultCircuitBreakerConfig().MaxFailures
	}
	if cfg.ResetTimeout <= 0 {
		cfg.ResetTimeout = DefaultCircuitBreakerConfig().ResetTimeout
	}
	return &CircuitBreaker{cfg: cfg, state: CircuitClosed}
}

// Allow returns true if the circuit permits the operation to proceed.
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	switch cb.state {
	case CircuitOpen:
		if time.Since(cb.openedAt) >= cb.cfg.ResetTimeout {
			cb.state = CircuitHalfOpen
			fmt.Fprintf(cb.cfg.Output, "[circuitbreaker] state -> half-open\n")
			return true
		}
		return false
	case CircuitHalfOpen:
		return true
	default:
		return true
	}
}

// RecordSuccess resets failure count and closes the circuit.
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	if cb.state != CircuitClosed {
		fmt.Fprintf(cb.cfg.Output, "[circuitbreaker] state -> closed\n")
	}
	cb.failures = 0
	cb.state = CircuitClosed
}

// RecordFailure increments the failure count and may open the circuit.
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures++
	if cb.state == CircuitHalfOpen || cb.failures >= cb.cfg.MaxFailures {
		cb.state = CircuitOpen
		cb.openedAt = time.Now()
		fmt.Fprintf(cb.cfg.Output, "[circuitbreaker] state -> open (failures=%d)\n", cb.failures)
	}
}

// State returns the current circuit state.
func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

// Failures returns the current consecutive failure count.
func (cb *CircuitBreaker) Failures() int {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.failures
}
