package routes

import "strings"

// AnnotateConfig holds configuration for the route annotator.
type AnnotateConfig struct {
	// Rules maps a substring match on Destination to a human-readable annotation.
	Rules map[string]string
}

// DefaultAnnotateConfig returns an AnnotateConfig with sensible defaults.
func DefaultAnnotateConfig() AnnotateConfig {
	return AnnotateConfig{
		Rules: map[string]string{
			"default": "default gateway",
			"169.254":  "link-local",
			"127.":     "loopback",
			"::1":      "loopback (IPv6)",
			"fe80":     "link-local (IPv6)",
		},
	}
}

// Annotator attaches human-readable labels to routes based on configurable rules.
type Annotator struct {
	cfg AnnotateConfig
}

// NewAnnotator creates a new Annotator with the given config.
func NewAnnotator(cfg AnnotateConfig) *Annotator {
	return &Annotator{cfg: cfg}
}

// Annotate returns a copy of the snapshot with Annotation fields populated.
func (a *Annotator) Annotate(snap Snapshot) Snapshot {
	annotated := make(Snapshot, len(snap))
	for i, r := range snap {
		r.Annotation = a.matchAnnotation(r.Destination)
		annotated[i] = r
	}
	return annotated
}

// AnnotateDiff annotates all routes in the Added and Removed slices of a Diff.
func (a *Annotator) AnnotateDiff(d Diff) Diff {
	d.Added = a.Annotate(d.Added)
	d.Removed = a.Annotate(d.Removed)
	return d
}

func (a *Annotator) matchAnnotation(destination string) string {
	for pattern, label := range a.cfg.Rules {
		if strings.Contains(destination, pattern) {
			return label
		}
	}
	return ""
}
