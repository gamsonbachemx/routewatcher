package routes

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func sampleCheckpointSnapshot() []RouteEntry {
	return []RouteEntry{
		{Destination: "10.0.0.0/8", Gateway: "192.168.1.1", Iface: "eth0", Protocol: "kernel"},
		{Destination: "0.0.0.0/0", Gateway: "192.168.1.254", Iface: "eth0", Protocol: "static"},
	}
}

func TestDefaultCheckpointConfig(t *testing.T) {
	cfg := DefaultCheckpointConfig()
	if cfg.Path == "" {
		t.Error("expected non-empty default path")
	}
	if cfg.Interval <= 0 {
		t.Error("expected positive default interval")
	}
}

func TestCheckpointManager_SaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	cfg := CheckpointConfig{Path: filepath.Join(dir, "cp.json"), Interval: time.Minute}
	cm := NewCheckpointManager(cfg)

	snap := sampleCheckpointSnapshot()
	if err := cm.Save(snap); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	entry, err := cm.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if entry == nil {
		t.Fatal("expected non-nil entry")
	}
	if len(entry.Snapshot) != len(snap) {
		t.Errorf("expected %d routes, got %d", len(snap), len(entry.Snapshot))
	}
	if entry.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}

func TestCheckpointManager_LoadMissing(t *testing.T) {
	cfg := CheckpointConfig{Path: "/tmp/routewatcher_no_such_checkpoint_xyz.json", Interval: time.Minute}
	cm := NewCheckpointManager(cfg)
	entry, err := cm.Load()
	if err != nil {
		t.Fatalf("expected nil error for missing file, got: %v", err)
	}
	if entry != nil {
		t.Error("expected nil entry for missing file")
	}
}

func TestCheckpointManager_LoadCorrupt(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "corrupt.json")
	_ = os.WriteFile(path, []byte("not json{"), 0644)

	cm := NewCheckpointManager(CheckpointConfig{Path: path, Interval: time.Minute})
	_, err := cm.Load()
	if err == nil {
		t.Error("expected error for corrupt checkpoint")
	}
}

func TestCheckpointManager_StartStop(t *testing.T) {
	dir := t.TempDir()
	cfg := CheckpointConfig{Path: filepath.Join(dir, "cp.json"), Interval: 50 * time.Millisecond}
	cm := NewCheckpointManager(cfg)

	called := make(chan struct{}, 5)
	snap := sampleCheckpointSnapshot()
	cm.Start(func() ([]RouteEntry, error) {
		called <- struct{}{}
		return snap, nil
	})

	select {
	case <-called:
		// at least one tick fired
	case <-time.After(500 * time.Millisecond):
		t.Error("expected snapshot function to be called")
	}
	cm.Stop()
}
