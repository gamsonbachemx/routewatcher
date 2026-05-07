package routes

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// DefaultLabelCmdConfig returns a LabelConfig suitable for CLI use.
func DefaultLabelCmdConfig() LabelConfig {
	cfg := DefaultLabelConfig()
	cfg.StaticLabels = map[string]string{
		"host": hostname(),
	}
	return cfg
}

// hostname returns the system hostname or "unknown" on error.
func hostname() string {
	h, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return h
}

// RunLabelShow prints all labels that would be applied to routes in a Diff
// using the provided LabelConfig, writing output to w.
func RunLabelShow(cfg LabelConfig, d Diff, w io.Writer) error {
	l := NewLabeler(cfg)
	labeled := l.LabelDiff(d)

	fmt.Fprintln(w, "=== Label Report ===")
	fmt.Fprintf(w, "Static Labels: %s\n", formatLabels(cfg.StaticLabels))
	fmt.Fprintf(w, "Prefix Rules:  %d rule(s)\n", len(cfg.PrefixRules))
	fmt.Fprintln(w, "")

	printLabeledRoutes(w, "Added", labeled.Added)
	printLabeledRoutes(w, "Removed", labeled.Removed)
	return nil
}

// RunLabelApply applies labels to a Diff and returns the labeled result.
func RunLabelApply(cfg LabelConfig, d Diff) Diff {
	return NewLabeler(cfg).LabelDiff(d)
}

func printLabeledRoutes(w io.Writer, section string, routes []Route) {
	if len(routes) == 0 {
		return
	}
	fmt.Fprintf(w, "[%s]\n", section)
	for _, r := range routes {
		fmt.Fprintf(w, "  %s via %s dev %s labels=[%s]\n",
			r.Destination, r.Gateway, r.Iface, formatLabels(r.Labels))
	}
}

func formatLabels(m map[string]string) string {
	if len(m) == 0 {
		return "(none)"
	}
	parts := make([]string, 0, len(m))
	for k, v := range m {
		parts = append(parts, k+"="+v)
	}
	return strings.Join(parts, ", ")
}
