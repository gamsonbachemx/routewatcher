package routes

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewExporter_InvalidFormat(t *testing.T) {
	_, err := NewExporter(ExportConfig{FilePath: "/tmp/x", Format: "csv"})
	if err == nil {
		t.Fatal("expected error for unsupported format")
	}
}

func TestNewExporter_EmptyPath(t *testing.T) {
	_, err := NewExporter(ExportConfig{FilePath: "", Format: "text"})
	if err == nil {
		t.Fatal("expected error for empty file path")
	}
}

func sampleDiff() Diff {
	return Diff{
		Added: []Route{
			{Destination: "10.0.0.0/8", Gateway: "192.168.1.1", Iface: "eth0", Protocol: "kernel"},
		},
		Removed: []Route{},
	}
}

func TestExporter_WriteText(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "routes.log")

	ex, err := NewExporter(ExportConfig{FilePath: path, Format: "text", Append: false})
	if err != nil {
		t.Fatalf("NewExporter: %v", err)
	}

	if err := ex.Write(sampleDiff()); err != nil {
		t.Fatalf("Write: %v", err)
	}

	data, _ := os.ReadFile(path)
	if !strings.Contains(string(data), "10.0.0.0/8") {
		t.Errorf("expected route in text output, got:\n%s", data)
	}
}

func TestExporter_WriteJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "routes.json")

	ex, err := NewExporter(ExportConfig{FilePath: path, Format: "json", Append: false})
	if err != nil {
		t.Fatalf("NewExporter: %v", err)
	}

	if err := ex.Write(sampleDiff()); err != nil {
		t.Fatalf("Write: %v", err)
	}

	data, _ := os.ReadFile(path)
	var rec ExportRecord
	if err := json.Unmarshal(data, &rec); err != nil {
		t.Fatalf("invalid JSON output: %v\nraw: %s", err, data)
	}
	if len(rec.Diff.Added) != 1 {
		t.Errorf("expected 1 added route, got %d", len(rec.Diff.Added))
	}
	if rec.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}

func TestExporter_AppendMode(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "routes.log")

	ex, err := NewExporter(ExportConfig{FilePath: path, Format: "text", Append: true})
	if err != nil {
		t.Fatalf("NewExporter: %v", err)
	}

	for i := 0; i < 3; i++ {
		if err := ex.Write(sampleDiff()); err != nil {
			t.Fatalf("Write iteration %d: %v", i, err)
		}
	}

	data, _ := os.ReadFile(path)
	count := strings.Count(string(data), "10.0.0.0/8")
	if count != 3 {
		t.Errorf("expected 3 occurrences in appended file, got %d", count)
	}
}
