package routes

import "strings"

// TransformConfig holds configuration for the route transformer.
type TransformConfig struct {
	Enabled        bool
	StripProtocol  bool
	NormalizeIface bool
	UppercaseProto bool
}

// DefaultTransformConfig returns sensible defaults.
func DefaultTransformConfig() TransformConfig {
	return TransformConfig{
		Enabled:        true,
		StripProtocol:  false,
		NormalizeIface: true,
		UppercaseProto: false,
	}
}

// Transformer applies normalization transforms to routes in a snapshot.
type Transformer struct {
	cfg TransformConfig
}

// NewTransformer constructs a Transformer with the given config.
func NewTransformer(cfg TransformConfig) *Transformer {
	return &Transformer{cfg: cfg}
}

// Apply returns a new snapshot with transforms applied to each route.
func (t *Transformer) Apply(snap Snapshot) Snapshot {
	if !t.cfg.Enabled {
		return snap
	}
	out := make(Snapshot, len(snap))
	for i, r := range snap {
		out[i] = t.transformRoute(r)
	}
	return out
}

// ApplyDiff returns a new Diff with transforms applied to added and removed routes.
func (t *Transformer) ApplyDiff(d Diff) Diff {
	if !t.cfg.Enabled {
		return d
	}
	return Diff{
		Added:   t.Apply(d.Added),
		Removed: t.Apply(d.Removed),
	}
}

func (t *Transformer) transformRoute(r Route) Route {
	if t.cfg.NormalizeIface {
		r.Iface = strings.TrimSpace(strings.ToLower(r.Iface))
	}
	if t.cfg.UppercaseProto {
		r.Protocol = strings.ToUpper(r.Protocol)
	}
	if t.cfg.StripProtocol {
		r.Protocol = ""
	}
	return r
}
