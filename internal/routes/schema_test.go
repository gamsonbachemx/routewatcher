package routes

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestDefaultSchemaConfig(t *testing.T) {
	cfg := DefaultSchemaConfig()
	if cfg.StrictMode {
		t.Error("expected StrictMode to be false by default")
	}
	if len(cfg.RequiredFields) == 0 {
		t.Error("expected at least one required field by default")
	}
}

func TestSchemaValidator_ValidRoute(t *testing.T) {
	v := NewSchemaValidator(DefaultSchemaConfig())
	r := Route{Destination: "10.0.0.0/8", Gateway: "192.168.1.1", Iface: "eth0"}
	if err := v.ValidateRoute(r); err != nil {
		t.Errorf("expected valid route, got error: %v", err)
	}
}

func TestSchemaValidator_MissingDestination(t *testing.T) {
	v := NewSchemaValidator(DefaultSchemaConfig())
	r := Route{Gateway: "192.168.1.1", Iface: "eth0"}
	err := v.ValidateRoute(r)
	if err == nil {
		t.Fatal("expected error for missing destination")
	}
	if !strings.Contains(err.Error(), "destination") {
		t.Errorf("expected error to mention 'destination', got: %v", err)
	}
}

func TestSchemaValidator_MissingMultipleFields(t *testing.T) {
	v := NewSchemaValidator(DefaultSchemaConfig())
	r := Route{} // all empty
	err := v.ValidateRoute(r)
	if err == nil {
		t.Fatal("expected error for empty route")
	}
}

func TestSchemaValidator_ValidateSnapshot(t *testing.T) {
	v := NewSchemaValidator(DefaultSchemaConfig())
	s := Snapshot{
		Routes: []Route{
			{Destination: "0.0.0.0/0", Gateway: "10.0.0.1", Iface: "eth0"},
			{Gateway: "10.0.0.1", Iface: "eth0"}, // missing destination
		},
	}
	errs := v.ValidateSnapshot(s)
	if len(errs) != 1 {
		t.Errorf("expected 1 validation error, got %d", len(errs))
	}
}

func TestSchemaValidator_CustomRequiredFields(t *testing.T) {
	cfg := SchemaConfig{RequiredFields: []string{"protocol"}}
	v := NewSchemaValidator(cfg)
	r := Route{Destination: "10.0.0.0/8", Gateway: "192.168.1.1", Iface: "eth0"}
	if err := v.ValidateRoute(r); err == nil {
		t.Error("expected error when protocol is missing")
	}
	r.Protocol = "kernel"
	if err := v.ValidateRoute(r); err != nil {
		t.Errorf("expected no error when protocol is set, got: %v", err)
	}
}

func TestMarshalVersioned(t *testing.T) {
	d := Diff{
		Added:   []Route{{Destination: "1.2.3.0/24", Gateway: "10.0.0.1", Iface: "eth0"}},
		Removed: []Route{},
	}
	b, err := MarshalVersioned(d)
	if err != nil {
		t.Fatalf("MarshalVersioned failed: %v", err)
	}
	var out map[string]interface{}
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if out["schema"] != SchemaVersion {
		t.Errorf("expected schema %q, got %v", SchemaVersion, out["schema"])
	}
	if _, ok := out["payload"]; !ok {
		t.Error("expected 'payload' key in output")
	}
}
