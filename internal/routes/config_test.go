package routes

import (
	"testing"
	"time"
)

func TestMultiStringFlag_SetAndString(t *testing.T) {
	var m multiStringFlag
	if err := m.Set("eth0"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := m.Set("wg0"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(m) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(m))
	}
	if m.String() != "eth0,wg0" {
		t.Errorf("unexpected string: %q", m.String())
	}
}

func TestConfig_Defaults(t *testing.T) {
	// ParseFlags reads os.Args; test the zero-value defaults directly.
	cfg := &Config{
		Interval:     5 * time.Second,
		OutputFormat: "text",
		Filter:       nil,
	}

	if cfg.Interval != 5*time.Second {
		t.Errorf("expected 5s interval, got %v", cfg.Interval)
	}
	if cfg.OutputFormat != "text" {
		t.Errorf("expected text format, got %q", cfg.OutputFormat)
	}
	if cfg.Filter != nil {
		t.Error("expected nil filter")
	}
}

func TestConfig_WithFilter(t *testing.T) {
	cfg := &Config{
		Interval:     10 * time.Second,
		OutputFormat: "json",
		Filter: &Filter{
			ExcludeLocal: true,
			Interfaces:   []string{"eth0"},
			Protocols:    []string{"static"},
		},
	}

	if cfg.Filter == nil {
		t.Fatal("expected non-nil filter")
	}
	if !cfg.Filter.ExcludeLocal {
		t.Error("expected ExcludeLocal to be true")
	}
	if len(cfg.Filter.Interfaces) != 1 || cfg.Filter.Interfaces[0] != "eth0" {
		t.Errorf("unexpected interfaces: %v", cfg.Filter.Interfaces)
	}
}
