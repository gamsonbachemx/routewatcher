package routes

import (
	"testing"
	"time"
)

func sampleRollupDiff(added, removed int) Diff {
	d := Diff{}
	for i := 0; i < added; i++ {
		d.Added = append(d.Added, Route{Destination: "10.0.0." + string(rune('0'+i))})
	}
	for i := 0; i < removed; i++ {
		d.Removed = append(d.Removed, Route{Destination: "192.168.0." + string(rune('0'+i))})
	}
	return d
}

func TestDefaultRollupConfig(t *testing.T) {
	cfg := DefaultRollupConfig()
	if cfg.Window <= 0 {
		t.Errorf("expected positive Window, got %v", cfg.Window)
	}
	if cfg.MaxDiffs <= 0 {
		t.Errorf("expected positive MaxDiffs, got %d", cfg.MaxDiffs)
	}
}

func TestRollup_FlushEmitsBatch(t *testing.T) {
	var got []Diff
	r := NewRollup(DefaultRollupConfig(), func(batch []Diff) {
		got = append(got, batch...)
	})
	r.Add(sampleRollupDiff(1, 0))
	r.Add(sampleRollupDiff(0, 1))
	r.Flush()
	if len(got) != 2 {
		t.Errorf("expected 2 diffs, got %d", len(got))
	}
}

func TestRollup_FlushClearsBuffer(t *testing.T) {
	calls := 0
	r := NewRollup(DefaultRollupConfig(), func(batch []Diff) { calls++ })
	r.Add(sampleRollupDiff(1, 0))
	r.Flush()
	r.Flush() // second flush should be a no-op
	if calls != 1 {
		t.Errorf("expected 1 flush call, got %d", calls)
	}
}

func TestRollup_MaxDiffsTriggersFlush(t *testing.T) {
	var got [][]Diff
	cfg := RollupConfig{Window: time.Minute, MaxDiffs: 3}
	r := NewRollup(cfg, func(batch []Diff) {
		got = append(got, batch)
	})
	for i := 0; i < 3; i++ {
		r.Add(sampleRollupDiff(1, 0))
	}
	if len(got) != 1 {
		t.Errorf("expected 1 auto-flush, got %d", len(got))
	}
	if len(got[0]) != 3 {
		t.Errorf("expected batch of 3, got %d", len(got[0]))
	}
}

func TestRollup_NilCallbackSafe(t *testing.T) {
	r := NewRollup(DefaultRollupConfig(), nil)
	r.Add(sampleRollupDiff(1, 1))
	r.Flush() // should not panic
}

func TestRollup_StopFlushesRemaining(t *testing.T) {
	var got []Diff
	cfg := RollupConfig{Window: time.Minute, MaxDiffs: 100}
	r := NewRollup(cfg, func(batch []Diff) { got = append(got, batch...) })
	r.Start()
	r.Add(sampleRollupDiff(2, 1))
	r.Stop()
	if len(got) != 1 {
		t.Errorf("expected 1 diff after stop, got %d", len(got))
	}
}

func TestFormatRollup_Empty(t *testing.T) {
	s := FormatRollup(nil)
	if s != "rollup: no changes" {
		t.Errorf("unexpected output: %q", s)
	}
}

func TestFormatRollup_WithChanges(t *testing.T) {
	batch := []Diff{
		sampleRollupDiff(2, 1),
		sampleRollupDiff(0, 3),
	}
	s := FormatRollup(batch)
	if s == "" {
		t.Error("expected non-empty rollup format string")
	}
}
