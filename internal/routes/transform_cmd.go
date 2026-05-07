package routes

import (
	"fmt"
	"io"
	"os"
)

// TransformCmdConfig holds CLI-level config for transform commands.
type TransformCmdConfig struct {
	Transform TransformConfig
	Output    io.Writer
}

// DefaultTransformCmdConfig returns defaults for CLI use.
func DefaultTransformCmdConfig() TransformCmdConfig {
	return TransformCmdConfig{
		Transform: DefaultTransformConfig(),
		Output:    os.Stdout,
	}
}

// RunTransformShow captures the current routing table, applies transforms,
// and prints the result to cfg.Output.
func RunTransformShow(cfg TransformCmdConfig) error {
	snap, err := Capture()
	if err != nil {
		return fmt.Errorf("transform show: capture failed: %w", err)
	}
	tr := NewTransformer(cfg.Transform)
	out := tr.Apply(snap)
	for _, r := range out {
		fmt.Fprintf(cfg.Output, "%-20s %-16s %-10s %s\n",
			r.Destination, r.Gateway, r.Iface, r.Protocol)
	}
	return nil
}

// RunTransformDiff captures the current routing table, compares against a
// baseline snapshot, applies transforms to the diff, and prints changes.
func RunTransformDiff(baseline Snapshot, cfg TransformCmdConfig) error {
	snap, err := Capture()
	if err != nil {
		return fmt.Errorf("transform diff: capture failed: %w", err)
	}
	tr := NewTransformer(cfg.Transform)
	d := Compare(baseline, snap)
	td := tr.ApplyDiff(d)
	if !td.HasChanges() {
		fmt.Fprintln(cfg.Output, "no changes")
		return nil
	}
	for _, r := range td.Added {
		fmt.Fprintf(cfg.Output, "+ %-20s %-16s %-10s %s\n",
			r.Destination, r.Gateway, r.Iface, r.Protocol)
	}
	for _, r := range td.Removed {
		fmt.Fprintf(cfg.Output, "- %-20s %-16s %-10s %s\n",
			r.Destination, r.Gateway, r.Iface, r.Protocol)
	}
	return nil
}
