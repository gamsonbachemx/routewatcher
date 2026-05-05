package routes

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// DefaultCheckpointConfig returns a CheckpointConfig with sensible defaults.
func DefaultCheckpointConfig() CheckpointConfig {
	return CheckpointConfig{
		Path:     "/tmp/routewatcher_checkpoint.json",
		Interval: 5 * time.Minute,
	}
}

// CheckpointConfig holds configuration for the checkpoint manager.
type CheckpointConfig struct {
	Path     string
	Interval time.Duration
}

// CheckpointEntry represents a single saved checkpoint.
type CheckpointEntry struct {
	Timestamp time.Time        `json:"timestamp"`
	Snapshot  []RouteEntry     `json:"snapshot"`
}

// CheckpointManager periodically saves routing snapshots to disk.
type CheckpointManager struct {
	cfg  CheckpointConfig
	stop chan struct{}
}

// NewCheckpointManager creates a new CheckpointManager.
func NewCheckpointManager(cfg CheckpointConfig) *CheckpointManager {
	return &CheckpointManager{
		cfg:  cfg,
		stop: make(chan struct{}),
	}
}

// Save writes the current snapshot to the checkpoint file.
func (c *CheckpointManager) Save(snapshot []RouteEntry) error {
	entry := CheckpointEntry{
		Timestamp: time.Now(),
		Snapshot:  snapshot,
	}
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return fmt.Errorf("checkpoint marshal: %w", err)
	}
	return os.WriteFile(c.cfg.Path, data, 0644)
}

// Load reads the last saved checkpoint from disk.
func (c *CheckpointManager) Load() (*CheckpointEntry, error) {
	data, err := os.ReadFile(c.cfg.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("checkpoint read: %w", err)
	}
	var entry CheckpointEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, fmt.Errorf("checkpoint unmarshal: %w", err)
	}
	return &entry, nil
}

// Start begins periodic checkpointing using the provided snapshot function.
func (c *CheckpointManager) Start(snapshotFn func() ([]RouteEntry, error)) {
	go func() {
		ticker := time.NewTicker(c.cfg.Interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if snap, err := snapshotFn(); err == nil {
					_ = c.Save(snap)
				}
			case <-c.stop:
				return
			}
		}
	}()
}

// Stop halts the background checkpointing goroutine.
func (c *CheckpointManager) Stop() {
	close(c.stop)
}
