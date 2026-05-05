package routes

import (
	"testing"
	"time"
)

func TestRetentionManager_AutoPurgeViaLoop(t *testing.T) {
	cfg := RetentionConfig{
		MaxAge:     30 * time.Millisecond,
		MaxEntries: 100,
		PurgeEvery: 20 * time.Millisecond,
	}
	rm := NewRetentionManager(cfg)
	defer rm.Stop()

	rm.Add(sampleRetentionDiff("172.16.0.0/12"))

	// Wait for the purge loop to remove the stale entry.
	deadline := time.Now().Add(300 * time.Millisecond)
	for time.Now().Before(deadline) {
		if len(rm.Entries()) == 0 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("expected entry to be purged by retention loop")
}

func TestRetentionManager_IntegratesWithHistory(t *testing.T) {
	histCfg := DefaultHistoryConfig()
	histCfg.MaxSize = 10
	h := NewHistory(histCfg)

	retCfg := RetentionConfig{
		MaxAge:     time.Hour,
		MaxEntries: 3,
		PurgeEvery: time.Hour,
	}
	rm := NewRetentionManager(retCfg)
	defer rm.Stop()

	diffs := []Diff{
		{Added: []Route{{Destination: "10.0.0.0/8"}}},
		{Added: []Route{{Destination: "192.168.0.0/16"}}},
		{Added: []Route{{Destination: "172.16.0.0/12"}}},
		{Added: []Route{{Destination: "100.64.0.0/10"}}},
	}

	for _, d := range diffs {
		h.Record(d)
		rm.Add(d)
	}

	rm.Purge()

	retained := rm.Entries()
	if len(retained) != 3 {
		t.Fatalf("expected 3 retained entries, got %d", len(retained))
	}

	// History should still have all 4 (managed independently).
	all := h.All()
	if len(all) != 4 {
		t.Fatalf("expected 4 history entries, got %d", len(all))
	}
}
