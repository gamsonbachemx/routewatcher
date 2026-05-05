package routes

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func sampleAuditDiff() Diff {
	return Diff{
		Added: []Route{
			{Destination: "10.0.0.0/8", Gateway: "192.168.1.1", Iface: "eth0"},
		},
		Removed: []Route{
			{Destination: "172.16.0.0/12", Gateway: "192.168.1.254", Iface: "eth0"},
		},
	}
}

func TestDefaultAuditConfig(t *testing.T) {
	cfg := DefaultAuditConfig()
	if cfg.Output == nil {
		t.Error("expected non-nil default output")
	}
	if cfg.IncludeDetails {
		t.Error("expected IncludeDetails to be false by default")
	}
}

func TestAuditor_RecordWritesToOutput(t *testing.T) {
	var buf bytes.Buffer
	cfg := DefaultAuditConfig()
	cfg.Output = &buf

	a, err := NewAuditor(cfg)
	if err != nil {
		t.Fatalf("NewAuditor: %v", err)
	}
	defer a.Close()

	if err := a.Record(sampleAuditDiff()); err != nil {
		t.Fatalf("Record: %v", err)
	}

	var entry AuditEntry
	if err := json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &entry); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if entry.Added != 1 || entry.Removed != 1 {
		t.Errorf("expected added=1 removed=1, got added=%d removed=%d", entry.Added, entry.Removed)
	}
	if entry.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}

func TestAuditor_IncludeDetails(t *testing.T) {
	var buf bytes.Buffer
	cfg := DefaultAuditConfig()
	cfg.Output = &buf
	cfg.IncludeDetails = true

	a, err := NewAuditor(cfg)
	if err != nil {
		t.Fatalf("NewAuditor: %v", err)
	}
	defer a.Close()

	_ = a.Record(sampleAuditDiff())

	var entry AuditEntry
	_ = json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &entry)
	if len(entry.Details) != 2 {
		t.Errorf("expected 2 detail lines, got %d", len(entry.Details))
	}
}

func TestAuditor_WritesToFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.log")

	var buf bytes.Buffer
	cfg := DefaultAuditConfig()
	cfg.Output = &buf
	cfg.FilePath = path

	a, err := NewAuditor(cfg)
	if err != nil {
		t.Fatalf("NewAuditor: %v", err)
	}
	_ = a.Record(sampleAuditDiff())
	_ = a.Close()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read audit file: %v", err)
	}
	if !strings.Contains(string(data), "\"added\":1") {
		t.Errorf("expected audit file to contain added count, got: %s", string(data))
	}
}

func TestNewAuditor_InvalidPath(t *testing.T) {
	cfg := DefaultAuditConfig()
	cfg.FilePath = "/nonexistent/path/audit.log"
	_, err := NewAuditor(cfg)
	if err == nil {
		t.Error("expected error for invalid file path")
	}
}
