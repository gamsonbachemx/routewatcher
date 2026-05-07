package routes

import (
	"encoding/json"
	"testing"
	"time"
)

func sampleEnvelopeDiff() Diff {
	return Diff{
		Added: []Route{
			{Destination: "10.0.0.0/8", Gateway: "192.168.1.1", Iface: "eth0"},
		},
		Removed: []Route{},
	}
}

func TestDefaultEnvelopeConfig(t *testing.T) {
	cfg := DefaultEnvelopeConfig()
	if cfg.Source != "routewatcher" {
		t.Errorf("expected source 'routewatcher', got %q", cfg.Source)
	}
	if cfg.Version != EnvelopeVersion {
		t.Errorf("expected version %q, got %q", EnvelopeVersion, cfg.Version)
	}
}

func TestEnveloper_WrapSetsFields(t *testing.T) {
	e := NewEnveloper(DefaultEnvelopeConfig())
	d := sampleEnvelopeDiff()
	before := time.Now().UTC()
	env := e.Wrap(d)
	after := time.Now().UTC()

	if env.Version != EnvelopeVersion {
		t.Errorf("unexpected version: %q", env.Version)
	}
	if env.Source != "routewatcher" {
		t.Errorf("unexpected source: %q", env.Source)
	}
	if env.Timestamp.Before(before) || env.Timestamp.After(after) {
		t.Errorf("timestamp out of range: %v", env.Timestamp)
	}
	if len(env.Diff.Added) != 1 {
		t.Errorf("expected 1 added route, got %d", len(env.Diff.Added))
	}
}

func TestEnveloper_SequenceIncrements(t *testing.T) {
	e := NewEnveloper(DefaultEnvelopeConfig())
	d := sampleEnvelopeDiff()

	env1 := e.Wrap(d)
	env2 := e.Wrap(d)

	if env1.Sequence != 1 {
		t.Errorf("expected seq 1, got %d", env1.Sequence)
	}
	if env2.Sequence != 2 {
		t.Errorf("expected seq 2, got %d", env2.Sequence)
	}
}

func TestEnvelope_Marshal_ValidJSON(t *testing.T) {
	e := NewEnveloper(DefaultEnvelopeConfig())
	env := e.Wrap(sampleEnvelopeDiff())

	b, err := env.Marshal()
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var out map[string]interface{}
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	for _, key := range []string{"version", "source", "timestamp", "sequence", "diff"} {
		if _, ok := out[key]; !ok {
			t.Errorf("missing key %q in envelope JSON", key)
		}
	}
}
