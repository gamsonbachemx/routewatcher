package routes

import (
	"fmt"
	"time"
)

// WatchOptions configures the watch behavior.
type WatchOptions struct {
	Interval time.Duration
	OnChange func(diff Diff)
	OnError  func(err error)
}

// Diff holds the result of comparing two route snapshots.
type Diff struct {
	Added   []Route
	Removed []Route
}

// HasChanges returns true if there are any added or removed routes.
func (d Diff) HasChanges() bool {
	return len(d.Added) > 0 || len(d.Removed) > 0
}

// Watch polls the routing table at the given interval and calls
// opts.OnChange whenever a difference is detected. It blocks until
// the provided stop channel is closed.
func Watch(stop <-chan struct{}, opts WatchOptions) error {
	if opts.Interval <= 0 {
		opts.Interval = 5 * time.Second
	}
	if opts.OnChange == nil {
		return fmt.Errorf("watch: OnChange callback must not be nil")
	}
	if opts.OnError == nil {
		opts.OnError = func(err error) {}
	}

	prev, err := Capture()
	if err != nil {
		return fmt.Errorf("watch: initial capture failed: %w", err)
	}

	ticker := time.NewTicker(opts.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return nil
		case <-ticker.C:
			curr, err := Capture()
			if err != nil {
				opts.OnError(fmt.Errorf("watch: capture failed: %w", err))
				continue
			}

			added, removed := Compare(prev, curr)
			d := Diff{Added: added, Removed: removed}
			if d.HasChanges() {
				opts.OnChange(d)
			}
			prev = curr
		}
	}
}
