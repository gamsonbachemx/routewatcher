package routes

import (
	"testing"
	"time"
)

func sampleDedupeDiff() Diff {
	return Diff{
		Added: []Route{
			{Destination: "10.0.0.0/8", Gateway: "192.168.1.1", Iface: "eth0"},
		},
		Removed: []Route{
			{Destination: "172.16.0.0/12", Gateway: "10.0.0.1", Iface: "eth1"},
		},
	}
}

func TestDefaultDedupeConfig(t *testing.T) {
	cfg := DefaultDedupeConfig()
	if cfg.TTL <= 0 {
		t.Errorf("expected positive TTL, got %v", cfg.TTL)
	}
}

func TestDeduplicator_FirstCallNotDuplicate(t *testing.T) {
	d := NewDeduplicator(DefaultDedupeConfig())
	diff := sampleDedupeDiff()
	if d.IsDuplicate(diff) {
		t.Error("expected first call to not be a duplicate")
	}
}

func TestDeduplicator_SecondCallIsDuplicate(t *testing.T) {
	d := NewDeduplicator(DefaultDedupeConfig())
	diff := sampleDedupeDiff()
	d.IsDuplicate(diff) // prime
	if !d.IsDuplicate(diff) {
		t.Error("expected second identical call to be a duplicate")
	}
}

func TestDeduplicator_DifferentDiffNotDuplicate(t *testing.T) {
	d := NewDeduplicator(DefaultDedupeConfig())
	diff1 := sampleDedupeDiff()
	diff2 := Diff{
		Added: []Route{{Destination: "192.168.0.0/16", Gateway: "10.1.1.1", Iface: "eth2"}},
	}
	d.IsDuplicate(diff1)
	if d.IsDuplicate(diff2) {
		t.Error("expected different diff to not be a duplicate")
	}
}

func TestDeduplicator_ExpiresAfterTTL(t *testing.T) {
	cfg := DedupeConfig{TTL: 50 * time.Millisecond}
	d := NewDeduplicator(cfg)
	diff := sampleDedupeDiff()
	d.IsDuplicate(diff)
	time.Sleep(80 * time.Millisecond)
	if d.IsDuplicate(diff) {
		t.Error("expected entry to have expired after TTL")
	}
}

func TestDeduplicator_ResetClearsState(t *testing.T) {
	d := NewDeduplicator(DefaultDedupeConfig())
	diff := sampleDedupeDiff()
	d.IsDuplicate(diff)
	d.Reset()
	if d.IsDuplicate(diff) {
		t.Error("expected Reset to clear remembered fingerprints")
	}
}

func TestDeduplicator_EmptyDiff(t *testing.T) {
	d := NewDeduplicator(DefaultDedupeConfig())
	empty := Diff{}
	d.IsDuplicate(empty)
	if !d.IsDuplicate(empty) {
		t.Error("expected repeated empty diff to be deduplicated")
	}
}
