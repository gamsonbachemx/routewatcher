package routes

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestSchemaValidator_IntegratesWithCapture verifies that a real snapshot
// produced by parseOutput satisfies the default schema requirements.
func TestSchemaValidator_IntegratesWithCapture(t *testing.T) {
	raw := `Kernel IP routing table
Destination     Gateway         Genmask         Flags Metric Ref    Use Iface
0.0.0.0         192.168.1.1     0.0.0.0         UG    100    0        0 eth0
10.0.0.0        0.0.0.0         255.0.0.0       U     0      0        0 lo
`
	snap := parseOutput(raw)
	v := NewSchemaValidator(DefaultSchemaConfig())
	errs := v.ValidateSnapshot(snap)
	if len(errs) != 0 {
		var msgs []string
		for _, e := range errs {
			msgs = append(msgs, e.Error())
		}
		t.Errorf("unexpected validation errors: %s", strings.Join(msgs, "; "))
	}
}

// TestMarshalVersioned_IntegratesWithDiff verifies the versioned envelope
// round-trips correctly when built from a real Compare result.
func TestMarshalVersioned_IntegratesWithDiff(t *testing.T) {
	before := Snapshot{
		Routes: []Route{
			{Destination: "10.0.0.0/8", Gateway: "192.168.1.1", Iface: "eth0"},
		},
	}
	after := Snapshot{
		Routes: []Route{
			{Destination: "10.0.0.0/8", Gateway: "192.168.1.1", Iface: "eth0"},
			{Destination: "172.16.0.0/12", Gateway: "192.168.1.254", Iface: "eth1"},
		},
	}
	d := Compare(before, after)
	b, err := MarshalVersioned(d)
	if err != nil {
		t.Fatalf("MarshalVersioned error: %v", err)
	}

	var envelope struct {
		Schema  string `json:"schema"`
		Payload struct {
			Added   []Route `json:"added"`
			Removed []Route `json:"removed"`
		} `json:"payload"`
	}
	if err := json.Unmarshal(b, &envelope); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if envelope.Schema != SchemaVersion {
		t.Errorf("schema mismatch: got %q", envelope.Schema)
	}
	if len(envelope.Payload.Added) != 1 {
		t.Errorf("expected 1 added route, got %d", len(envelope.Payload.Added))
	}
	if len(envelope.Payload.Removed) != 0 {
		t.Errorf("expected 0 removed routes, got %d", len(envelope.Payload.Removed))
	}
}
