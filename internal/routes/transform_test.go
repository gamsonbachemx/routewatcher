package routes

import (
	"testing"
)

func sampleTransformSnapshot() Snapshot {
	return Snapshot{
		{Destination: "10.0.0.0/8", Gateway: "192.168.1.1", Iface: "ETH0", Protocol: "kernel"},
		{Destination: "0.0.0.0/0", Gateway: "192.168.1.254", Iface: " eth1 ", Protocol: "static"},
	}
}

func TestDefaultTransformConfig(t *testing.T) {
	cfg := DefaultTransformConfig()
	if !cfg.Enabled {
		t.Error("expected Enabled to be true")
	}
	if !cfg.NormalizeIface {
		t.Error("expected NormalizeIface to be true")
	}
	if cfg.UppercaseProto {
		t.Error("expected UppercaseProto to be false")
	}
}

func TestTransformer_NormalizeIface(t *testing.T) {
	cfg := DefaultTransformConfig()
	tr := NewTransformer(cfg)
	out := tr.Apply(sampleTransformSnapshot())
	if out[0].Iface != "eth0" {
		t.Errorf("expected eth0, got %q", out[0].Iface)
	}
	if out[1].Iface != "eth1" {
		t.Errorf("expected eth1, got %q", out[1].Iface)
	}
}

func TestTransformer_UppercaseProto(t *testing.T) {
	cfg := DefaultTransformConfig()
	cfg.UppercaseProto = true
	tr := NewTransformer(cfg)
	out := tr.Apply(sampleTransformSnapshot())
	if out[0].Protocol != "KERNEL" {
		t.Errorf("expected KERNEL, got %q", out[0].Protocol)
	}
}

func TestTransformer_StripProtocol(t *testing.T) {
	cfg := DefaultTransformConfig()
	cfg.StripProtocol = true
	tr := NewTransformer(cfg)
	out := tr.Apply(sampleTransformSnapshot())
	for _, r := range out {
		if r.Protocol != "" {
			t.Errorf("expected empty protocol, got %q", r.Protocol)
		}
	}
}

func TestTransformer_DisabledPassesThrough(t *testing.T) {
	cfg := DefaultTransformConfig()
	cfg.Enabled = false
	tr := NewTransformer(cfg)
	snap := sampleTransformSnapshot()
	out := tr.Apply(snap)
	if out[0].Iface != snap[0].Iface {
		t.Errorf("expected unchanged iface %q, got %q", snap[0].Iface, out[0].Iface)
	}
}

func TestTransformer_ApplyDiff(t *testing.T) {
	cfg := DefaultTransformConfig()
	cfg.UppercaseProto = true
	tr := NewTransformer(cfg)
	d := Diff{
		Added:   sampleTransformSnapshot()[:1],
		Removed: sampleTransformSnapshot()[1:],
	}
	out := tr.ApplyDiff(d)
	if out.Added[0].Protocol != "KERNEL" {
		t.Errorf("expected KERNEL in Added, got %q", out.Added[0].Protocol)
	}
	if out.Removed[0].Protocol != "STATIC" {
		t.Errorf("expected STATIC in Removed, got %q", out.Removed[0].Protocol)
	}
}
