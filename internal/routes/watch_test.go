package routes

import (
	"testing"
	"time"
)

func TestDiff_HasChanges(t *testing.T) {
	tests := []struct {
		name     string
		diff     Diff
		expected bool
	}{
		{
			name:     "no changes",
			diff:     Diff{},
			expected: false,
		},
		{
			name:     "added routes",
			diff:     Diff{Added: []Route{{Destination: "10.0.0.0/8"}}},
			expected: true,
		},
		{
			name:     "removed routes",
			diff:     Diff{Removed: []Route{{Destination: "192.168.1.0/24"}}},
			expected: true,
		},
		{
			name: "both added and removed",
			diff: Diff{
				Added:   []Route{{Destination: "10.0.0.0/8"}},
				Removed: []Route{{Destination: "192.168.1.0/24"}},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.diff.HasChanges(); got != tt.expected {
				t.Errorf("HasChanges() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestWatch_NilOnChange(t *testing.T) {
	stop := make(chan struct{})
	defer close(stop)

	err := Watch(stop, WatchOptions{
		Interval: 100 * time.Millisecond,
		OnChange: nil,
	})
	if err == nil {
		t.Error("expected error when OnChange is nil, got nil")
	}
}

func TestWatch_StopsOnClose(t *testing.T) {
	stop := make(chan struct{})

	done := make(chan error, 1)
	go func() {
		done <- Watch(stop, WatchOptions{
			Interval: 50 * time.Millisecond,
			OnChange: func(d Diff) {},
		})
	}()

	time.Sleep(120 * time.Millisecond)
	close(stop)

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Watch returned unexpected error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("Watch did not stop after channel close")
	}
}
