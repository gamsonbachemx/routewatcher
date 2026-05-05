package routes

import (
	"testing"
)

func TestDefaultAnnotateConfig(t *testing.T) {
	cfg := DefaultAnnotateConfig()
	if len(cfg.Rules) == 0 {
		t.Fatal("expected default rules to be non-empty")
	}
	if _, ok := cfg.Rules["default"]; !ok {
		t.Error("expected 'default' rule to exist")
	}
}

func TestAnnotator_KnownDestination(t *testing.T) {
	a := NewAnnotator(DefaultAnnotateConfig())
	snap := Snapshot{
		{Destination: "default", Iface: "eth0"},
		{Destination: "169.254.0.0/16", Iface: "eth0"},
	}
	result := a.Annotate(snap)
	if result[0].Annotation != "default gateway" {
		t.Errorf("expected 'default gateway', got %q", result[0].Annotation)
	}
	if result[1].Annotation != "link-local" {
		t.Errorf("expected 'link-local', got %q", result[1].Annotation)
	}
}

func TestAnnotator_UnknownDestination(t *testing.T) {
	a := NewAnnotator(DefaultAnnotateConfig())
	snap := Snapshot{
		{Destination: "10.0.0.0/8", Iface: "eth0"},
	}
	result := a.Annotate(snap)
	if result[0].Annotation != "" {
		t.Errorf("expected empty annotation, got %q", result[0].Annotation)
	}
}

func TestAnnotator_CustomRule(t *testing.T) {
	cfg := AnnotateConfig{
		Rules: map[string]string{
			"10.0": "private-10",
		},
	}
	a := NewAnnotator(cfg)
	snap := Snapshot{
		{Destination: "10.0.1.0/24", Iface: "eth1"},
	}
	result := a.Annotate(snap)
	if result[0].Annotation != "private-10" {
		t.Errorf("expected 'private-10', got %q", result[0].Annotation)
	}
}

func TestAnnotator_AnnotateDiff(t *testing.T) {
	a := NewAnnotator(DefaultAnnotateConfig())
	d := Diff{
		Added:   Snapshot{{Destination: "default", Iface: "eth0"}},
		Removed: Snapshot{{Destination: "127.0.0.1", Iface: "lo"}},
	}
	result := a.AnnotateDiff(d)
	if result.Added[0].Annotation != "default gateway" {
		t.Errorf("Added: expected 'default gateway', got %q", result.Added[0].Annotation)
	}
	if result.Removed[0].Annotation != "loopback" {
		t.Errorf("Removed: expected 'loopback', got %q", result.Removed[0].Annotation)
	}
}

func TestAnnotator_DoesNotMutateOriginal(t *testing.T) {
	a := NewAnnotator(DefaultAnnotateConfig())
	orig := Snapshot{{Destination: "default", Iface: "eth0"}}
	_ = a.Annotate(orig)
	if orig[0].Annotation != "" {
		t.Error("original snapshot should not be mutated")
	}
}
