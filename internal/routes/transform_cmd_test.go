package routes

import (
	"bytes"
	"strings"
	"testing"
)

func TestDefaultTransformCmdConfig(t *testing.T) {
	cfg := DefaultTransformCmdConfig()
	if cfg.Output == nil {
		t.Error("expected non-nil Output")
	}
	if !cfg.Transform.Enabled {
		t.Error("expected transform to be enabled")
	}
}

func TestRunTransformDiff_NoChanges(t *testing.T) {
	baseline := Snapshot{
		{Destination: "10.0.0.0/8", Gateway: "192.168.1.1", Iface: "eth0", Protocol: "kernel"},
	}

	var buf bytes.Buffer
	cfg := DefaultTransformCmdConfig()
	cfg.Output = &buf

	// Build a diff against itself — no changes
	tr := NewTransformer(cfg.Transform)
	d := Compare(baseline, baseline)
	td := tr.ApplyDiff(d)

	if td.HasChanges() {
		t.Error("expected no changes diffing snapshot against itself")
	}
}

func TestRunTransformDiff_ShowsAddedAndRemoved(t *testing.T) {
	before := Snapshot{
		{Destination: "10.0.0.0/8", Gateway: "192.168.1.1", Iface: "ETH0", Protocol: "kernel"},
	}
	after := Snapshot{
		{Destination: "172.16.0.0/12", Gateway: "192.168.1.1", Iface: "ETH1", Protocol: "static"},
	}

	var buf bytes.Buffer
	cfg := DefaultTransformCmdConfig()
	cfg.Output = &buf

	tr := NewTransformer(cfg.Transform)
	d := Compare(before, after)
	td := tr.ApplyDiff(d)

	for _, r := range td.Added {
		if r.Iface != "eth1" {
			t.Errorf("expected normalized iface eth1, got %q", r.Iface)
		}
	}
	for _, r := range td.Removed {
		if r.Iface != "eth0" {
			t.Errorf("expected normalized iface eth0, got %q", r.Iface)
		}
	}
}

func TestRunTransformDiff_OutputFormat(t *testing.T) {
	before := Snapshot{
		{Destination: "10.0.0.0/8", Gateway: "192.168.1.1", Iface: "ETH0", Protocol: "kernel"},
	}
	after := Snapshot{}

	var buf bytes.Buffer
	cfg := DefaultTransformCmdConfig()
	cfg.Output = &buf

	tr := NewTransformer(cfg.Transform)
	d := Compare(before, after)
	td := tr.ApplyDiff(d)

	if !td.HasChanges() {
		t.Fatal("expected changes")
	}
	for _, r := range td.Removed {
		line := r.Destination + " " + r.Iface
		if !strings.Contains(line, "eth0") {
			t.Errorf("expected eth0 in output line, got %q", line)
		}
	}
}
