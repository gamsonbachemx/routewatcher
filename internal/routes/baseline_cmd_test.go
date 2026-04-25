package routes

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func writeSampleBaseline(t *testing.T, path string, snap Snapshot) {
	t.Helper()
	entry := BaselineEntry{
		CapturedAt: time.Now().UTC(),
		Routes:     snap,
	}
	data, _ := json.MarshalIndent(entry, "", "  ")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("writeSampleBaseline: %v", err)
	}
}

func TestDefaultBaselineConfig(t *testing.T) {
	cfg := DefaultBaselineConfig()
	if cfg.Path == "" {
		t.Error("expected non-empty default path")
	}
	if cfg.Output == nil {
		t.Error("expected non-nil default output")
	}
}

func TestRunBaselineShow_Output(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "baseline.json")

	snap := Snapshot{
		{Destination: "10.0.0.0/8", Gateway: "192.168.1.1", Iface: "eth0"},
	}
	writeSampleBaseline(t, path, snap)

	var buf bytes.Buffer
	cfg := BaselineCommandConfig{Path: path, Output: &buf}

	if err := RunBaselineShow(cfg); err != nil {
		t.Fatalf("RunBaselineShow: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "10.0.0.0/8") {
		t.Errorf("expected destination in output, got: %s", out)
	}
	if !strings.Contains(out, "captured:") {
		t.Errorf("expected 'captured:' in output, got: %s", out)
	}
}

func TestRunBaselineShow_MissingFile(t *testing.T) {
	cfg := BaselineCommandConfig{
		Path:   "/tmp/routewatcher_no_such_baseline.json",
		Output: &bytes.Buffer{},
	}
	if err := RunBaselineShow(cfg); err == nil {
		t.Fatal("expected error for missing baseline")
	}
}

func TestRunBaselineDiff_NoChanges(t *testing.T) {
	// Write a baseline with a known snapshot; diff output should note no changes
	// when current routes match (we mock by saving current capture as baseline).
	dir := t.TempDir()
	path := filepath.Join(dir, "baseline.json")

	current, err := Capture()
	if err != nil {
		t.Skip("cannot capture routes in this environment")
	}

	store, _ := NewBaselineStore(path)
	_ = store.Save(current)

	var buf bytes.Buffer
	cfg := BaselineCommandConfig{Path: path, Output: &buf}
	if err := RunBaselineDiff(cfg); err != nil {
		t.Fatalf("RunBaselineDiff: %v", err)
	}
	if !strings.Contains(buf.String(), "baseline captured:") {
		t.Errorf("expected header in output: %s", buf.String())
	}
}
