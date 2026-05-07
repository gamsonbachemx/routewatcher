package routes

import "strings"

// LabelConfig controls how routes are labeled.
type LabelConfig struct {
	// StaticLabels are applied to every route unconditionally.
	StaticLabels map[string]string
	// PrefixRules maps a destination prefix to a label value for the "prefix" key.
	PrefixRules map[string]string
	// Enabled controls whether labeling is active.
	Enabled bool
}

// DefaultLabelConfig returns a LabelConfig with sensible defaults.
func DefaultLabelConfig() LabelConfig {
	return LabelConfig{
		StaticLabels: map[string]string{},
		PrefixRules:  map[string]string{},
		Enabled:      true,
	}
}

// Labeler attaches metadata labels to routes in a Diff.
type Labeler struct {
	cfg LabelConfig
}

// NewLabeler creates a Labeler from the given config.
func NewLabeler(cfg LabelConfig) *Labeler {
	return &Labeler{cfg: cfg}
}

// LabelDiff returns a copy of the Diff with labels applied to each route.
func (l *Labeler) LabelDiff(d Diff) Diff {
	if !l.cfg.Enabled {
		return d
	}
	out := Diff{
		Added:   make([]Route, len(d.Added)),
		Removed: make([]Route, len(d.Removed)),
	}
	for i, r := range d.Added {
		out.Added[i] = l.labelRoute(r)
	}
	for i, r := range d.Removed {
		out.Removed[i] = l.labelRoute(r)
	}
	return out
}

// labelRoute applies static and prefix-based labels to a single Route.
func (l *Labeler) labelRoute(r Route) Route {
	if r.Labels == nil {
		r.Labels = map[string]string{}\n	}
	for k, v := range l.cfg.StaticLabels {
		r.Labels[k] = v
	}
	for prefix, label := range l.cfg.PrefixRules {
		if strings.HasPrefix(r.Destination, prefix) {
			r.Labels["prefix"] = label
			break
		}
	}
	return r
}
