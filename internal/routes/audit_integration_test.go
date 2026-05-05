package routes

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// TestAuditor_IntegratesWithWatcher verifies that an Auditor can be wired into
// a Watch loop and receives diffs as they are emitted.
func TestAuditor_IntegratesWithWatcher(t *testing.T) {
	var buf bytes.Buffer
	cfg := DefaultAuditConfig()
	cfg.Output = &buf
	cfg.IncludeDetails = true

	auditor, err := NewAuditor(cfg)
	if err != nil {
		t.Fatalf("NewAuditor: %v", err)
	}
	defer auditor.Close()

	call := 0
	capture := func() ([]Route, error) {
		call++
		switch call {
		case 1:
			return []Route{{Destination: "10.0.0.0/8", Gateway: "192.168.1.1", Iface: "eth0"}}, nil
		case 2:
			return []Route{
				{Destination: "10.0.0.0/8", Gateway: "192.168.1.1", Iface: "eth0"},
				{Destination: "172.16.0.0/12", Gateway: "10.0.0.1", Iface: "eth1"},
			}, nil
		default:
			return nil, nil
		}
	}

	recorded := make(chan struct{}, 1)
	onChange := func(d Diff) {
		_ = auditor.Record(d)
		recorded <- struct{}{}
	}

	stop := make(chan struct{})
	go func() {
		_ = Watch(WatchConfig{
			Interval: 10 * time.Millisecond,
			CaptureFn: capture,
			OnChange:  onChange,
			Stop:      stop,
		})
	}()

	select {
	case <-recorded:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for audit record")
	}
	close(stop)

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) == 0 {
		t.Fatal("expected at least one audit line")
	}
	var entry AuditEntry
	if err := json.Unmarshal([]byte(lines[0]), &entry); err != nil {
		t.Fatalf("unmarshal audit entry: %v", err)
	}
	if entry.Added == 0 && entry.Removed == 0 {
		t.Error("expected non-zero added or removed in audit entry")
	}
}

// TestAuditor_EmptyDiffNotRecorded ensures that callers can guard against
// recording empty diffs (no-op behaviour expected at call site).
func TestAuditor_EmptyDiffNotRecorded(t *testing.T) {
	var buf bytes.Buffer
	cfg := DefaultAuditConfig()
	cfg.Output = &buf

	a, _ := NewAuditor(cfg)
	defer a.Close()

	empty := Diff{}
	if empty.HasChanges() {
		_ = a.Record(empty)
	}

	if buf.Len() != 0 {
		t.Errorf("expected no output for empty diff, got: %s", buf.String())
	}
}
