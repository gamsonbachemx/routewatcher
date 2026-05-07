package routes

import (
	"encoding/json"
	"testing"
)

// TestEnveloper_IntegratesWithDiff verifies that a real Compare result can be
// wrapped and round-tripped through JSON without data loss.
func TestEnveloper_IntegratesWithDiff(t *testing.T) {
	old := Snapshot{
		Routes: []Route{
			{Destination: "10.0.0.0/8", Gateway: "192.168.1.1", Iface: "eth0", Protocol: "kernel"},
		},
	}
	new_ := Snapshot{
		Routes: []Route{
			{Destination: "10.0.0.0/8", Gateway: "192.168.1.1", Iface: "eth0", Protocol: "kernel"},
			{Destination: "172.16.0.0/12", Gateway: "10.0.0.1", Iface: "eth1", Protocol: "static"},
		},
	}

	diff := Compare(old, new_)
	e := NewEnveloper(DefaultEnvelopeConfig())
	env := e.Wrap(diff)

	b, err := env.Marshal()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded Envelope
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(decoded.Diff.Added) != 1 {
		t.Errorf("expected 1 added route after round-trip, got %d", len(decoded.Diff.Added))
	}
	if decoded.Diff.Added[0].Destination != "172.16.0.0/12" {
		t.Errorf("unexpected destination: %q", decoded.Diff.Added[0].Destination)
	}
}

// TestEnveloper_SequenceMonotonicallyIncreases verifies ordering across many wraps.
func TestEnveloper_SequenceMonotonicallyIncreases(t *testing.T) {
	e := NewEnveloper(DefaultEnvelopeConfig())
	d := Diff{Added: []Route{}, Removed: []Route{}}

	const n = 20
	var last uint64
	for i := 0; i < n; i++ {
		env := e.Wrap(d)
		if env.Sequence <= last {
			t.Fatalf("sequence did not increase at iteration %d: got %d, prev %d", i, env.Sequence, last)
		}
		last = env.Sequence
	}
	if last != n {
		t.Errorf("expected final sequence %d, got %d", n, last)
	}
}
