package routes

import (
	"path/filepath"
	"testing"
	"time"
)

func TestCheckpoint_IntegratesWithBaseline(t *testing.T) {
	dir := t.TempDir()
	cpPath := filepath.Join(dir, "checkpoint.json")
	cm := NewCheckpointManager(CheckpointConfig{
		Path:     cpPath,
		Interval: time.Minute,
	})

	snap := []RouteEntry{
		{Destination: "172.16.0.0/12", Gateway: "10.0.0.1", Iface: "eth1", Protocol: "ospf"},
	}
	if err := cm.Save(snap); err != nil {
		t.Fatalf("Save: %v", err)
	}

	entry, err := cm.Load()
	if err != nil || entry == nil {
		t.Fatalf("Load: %v", err)
	}

	// Feed loaded snapshot into a baseline store and verify round-trip.
	bsPath := filepath.Join(dir, "baseline.json")
	bs := NewBaselineStore(bsPath)
	if err := bs.Save(entry.Snapshot); err != nil {
		t.Fatalf("BaselineStore.Save: %v", err)
	}
	loaded, err := bs.Load()
	if err != nil {
		t.Fatalf("BaselineStore.Load: %v", err)
	}
	if len(loaded.Routes) != len(snap) {
		t.Errorf("expected %d routes in baseline, got %d", len(snap), len(loaded.Routes))
	}
}

func TestCheckpoint_RoundTripPreservesFields(t *testing.T) {
	dir := t.TempDir()
	cm := NewCheckpointManager(CheckpointConfig{
		Path:     filepath.Join(dir, "cp.json"),
		Interval: time.Minute,
	})

	orig := []RouteEntry{
		{Destination: "192.168.0.0/16", Gateway: "10.10.0.1", Iface: "bond0", Protocol: "bgp", Metric: 100},
	}
	_ = cm.Save(orig)

	entry, err := cm.Load()
	if err != nil || entry == nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r := entry.Snapshot[0]
	if r.Destination != orig[0].Destination {
		t.Errorf("Destination mismatch: %q vs %q", r.Destination, orig[0].Destination)
	}
	if r.Gateway != orig[0].Gateway {
		t.Errorf("Gateway mismatch: %q vs %q", r.Gateway, orig[0].Gateway)
	}
	if r.Protocol != orig[0].Protocol {
		t.Errorf("Protocol mismatch: %q vs %q", r.Protocol, orig[0].Protocol)
	}
	if r.Metric != orig[0].Metric {
		t.Errorf("Metric mismatch: %d vs %d", r.Metric, orig[0].Metric)
	}
}
