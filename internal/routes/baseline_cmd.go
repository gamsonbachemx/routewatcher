package routes

import (
	"fmt"
	"io"
	"os"
	"text/tabwriter"
	"time"
)

// BaselineCommandConfig holds options for baseline subcommands.
type BaselineCommandConfig struct {
	Path   string
	Output io.Writer
}

// DefaultBaselineConfig returns a BaselineCommandConfig with sensible defaults.
func DefaultBaselineConfig() BaselineCommandConfig {
	return BaselineCommandConfig{
		Path:   "/var/lib/routewatcher/baseline.json",
		Output: os.Stdout,
	}
}

// RunBaselineSave captures the current routing table and saves it as the baseline.
func RunBaselineSave(cfg BaselineCommandConfig) error {
	store, err := NewBaselineStore(cfg.Path)
	if err != nil {
		return err
	}
	snap, err := Capture()
	if err != nil {
		return fmt.Errorf("capture routes: %w", err)
	}
	if err := store.Save(snap); err != nil {
		return err
	}
	fmt.Fprintf(cfg.Output, "baseline saved: %d routes written to %s\n", len(snap), cfg.Path)
	return nil
}

// RunBaselineDiff loads the stored baseline and prints a diff against current routes.
func RunBaselineDiff(cfg BaselineCommandConfig) error {
	store, err := NewBaselineStore(cfg.Path)
	if err != nil {
		return err
	}
	d, entry, err := store.CompareToBaseline()
	if err != nil {
		return err
	}
	w := tabwriter.NewWriter(cfg.Output, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "baseline captured:\t%s\n", entry.CapturedAt.Format(time.RFC3339))
	fmt.Fprintf(w, "added routes:\t%d\n", len(d.Added))
	fmt.Fprintf(w, "removed routes:\t%d\n", len(d.Removed))
	w.Flush()

	if !d.HasChanges() {
		fmt.Fprintln(cfg.Output, "no changes from baseline")
		return nil
	}
	fmt.Fprintln(cfg.Output, FormatText(*d))
	return nil
}

// RunBaselineShow prints the stored baseline routes.
func RunBaselineShow(cfg BaselineCommandConfig) error {
	store, err := NewBaselineStore(cfg.Path)
	if err != nil {
		return err
	}
	entry, err := store.Load()
	if err != nil {
		return err
	}
	w := tabwriter.NewWriter(cfg.Output, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "captured:\t%s\n", entry.CapturedAt.Format(time.RFC3339))
	fmt.Fprintf(w, "routes:\t%d\n", len(entry.Routes))
	w.Flush()
	for _, r := range entry.Routes {
		fmt.Fprintf(cfg.Output, "  %-20s via %-16s dev %s\n", r.Destination, r.Gateway, r.Iface)
	}
	return nil
}
