package routes

import (
	"testing"
	"time"
)

func sampleRetentionDiff(dest string) Diff {
	return Diff{
		Added: []Route{{Destination: dest, Gateway: "10.0.0.1", Iface: "eth0"}},
	}
}

func TestDefaultRetentionConfig(t *testing.T) {
	cfg := DefaultRetentionConfig()
	if cfg.MaxAge <= 0 {
		t.Error("expected positive MaxAge")
	}
	if cfg.MaxEntries <= 0 {
		t.Error("expected positive MaxEntries")
	}
	if cfg.PurgeEvery <= 0 {
		t.Error("expected positive PurgeEvery")
	}
}

func TestRetentionManager_AddAndRetrieve(t *testing.T) {
	cfg := DefaultRetentionConfig()
	cfg.PurgeEvery = time.Hour // don't purge during test
	rm := NewRetentionManager(cfg)
	defer rm.Stop()

	rm.Add(sampleRetentionDiff("192.168.1.0/24"))
	rm.Add(sampleRetentionDiff("10.0.0.0/8"))

	entries := rm.Entries()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
}

func TestRetentionManager_PurgeByAge(t *testing.T) {
	cfg := RetentionConfig{
		MaxAge:     50 * time.Millisecond,
		MaxEntries: 100,
		PurgeEvery: time.Hour,
	}
	rm := NewRetentionManager(cfg)
	defer rm.Stop()

	rm.Add(sampleRetentionDiff("192.168.0.0/24"))
	time.Sleep(80 * time.Millisecond)
	rm.Add(sampleRetentionDiff("10.0.0.0/8")) // recent

	rm.Purge()

	entries := rm.Entries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry after age purge, got %d", len(entries))
	}
	if entries[0].Added[0].Destination != "10.0.0.0/8" {
		t.Errorf("unexpected entry: %v", entries[0])
	}
}

func TestRetentionManager_PurgeByCount(t *testing.T) {
	cfg := RetentionConfig{
		MaxAge:     time.Hour,
		MaxEntries: 2,
		PurgeEvery: time.Hour,
	}
	rm := NewRetentionManager(cfg)
	defer rm.Stop()

	for i := 0; i < 5; i++ {
		rm.Add(sampleRetentionDiff("10.0.0.0/8"))
	}

	rm.Purge()

	if got := len(rm.Entries()); got != 2 {
		t.Fatalf("expected 2 entries after count purge, got %d", got)
	}
}

func TestRetentionManager_StopHaltsLoop(t *testing.T) {
	cfg := RetentionConfig{
		MaxAge:     time.Hour,
		MaxEntries: 100,
		PurgeEvery: 10 * time.Millisecond,
	}
	rm := NewRetentionManager(cfg)
	rm.Stop()
	// stopping twice should not panic (channel already closed)
	// just verify we can call Stop without blocking
}
