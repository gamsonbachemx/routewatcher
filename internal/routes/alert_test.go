package routes

import (
	"bytes"
	"strings"
	"testing"
)

func TestAlerter_NoAlertBelowThreshold(t *testing.T) {
	var buf bytes.Buffer
	cfg := AlertConfig{
		MinChanges: 3,
		Level:      AlertWarning,
		Output:     &buf,
	}
	a := NewAlerter(cfg)
	d := Diff{
		Added:   []Route{{Destination: "10.0.0.0/8"}},
		Removed: []Route{},
	}
	a.Notify(d)
	if buf.Len() != 0 {
		t.Errorf("expected no output for 1 change with threshold 3, got: %s", buf.String())
	}
}

func TestAlerter_EmitsAlertAtThreshold(t *testing.T) {
	var buf bytes.Buffer
	cfg := AlertConfig{
		MinChanges: 1,
		Level:      AlertCritical,
		Output:     &buf,
	}
	a := NewAlerter(cfg)
	d := Diff{
		Added:   []Route{{Destination: "192.168.1.0/24"}},
		Removed: []Route{{Destination: "10.0.0.0/8"}},
	}
	a.Notify(d)
	out := buf.String()
	if !strings.Contains(out, "CRITICAL") {
		t.Errorf("expected CRITICAL in output, got: %s", out)
	}
	if !strings.Contains(out, "+1 added") {
		t.Errorf("expected '+1 added' in output, got: %s", out)
	}
	if !strings.Contains(out, "-1 removed") {
		t.Errorf("expected '-1 removed' in output, got: %s", out)
	}
}

func TestAlerter_DefaultOutputNotNil(t *testing.T) {
	cfg := AlertConfig{
		MinChanges: 1,
		Level:      AlertInfo,
		Output:     nil, // should fall back to os.Stderr
	}
	a := NewAlerter(cfg)
	if a.cfg.Output == nil {
		t.Error("expected Output to be non-nil after NewAlerter")
	}
}

func TestDefaultAlertConfig(t *testing.T) {
	cfg := DefaultAlertConfig()
	if cfg.MinChanges != 1 {
		t.Errorf("expected MinChanges=1, got %d", cfg.MinChanges)
	}
	if cfg.Level != AlertWarning {
		t.Errorf("expected Level=WARNING, got %s", cfg.Level)
	}
	if cfg.Output == nil {
		t.Error("expected non-nil Output in DefaultAlertConfig")
	}
}

func TestAlerter_NoAlertOnEmptyDiff(t *testing.T) {
	var buf bytes.Buffer
	cfg := AlertConfig{
		MinChanges: 1,
		Level:      AlertWarning,
		Output:     &buf,
	}
	a := NewAlerter(cfg)
	d := Diff{
		Added:   []Route{},
		Removed: []Route{},
	}
	a.Notify(d)
	if buf.Len() != 0 {
		t.Errorf("expected no output for empty diff, got: %s", buf.String())
	}
}
