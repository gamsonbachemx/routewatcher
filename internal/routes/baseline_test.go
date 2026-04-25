package routes

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewBaselineStore_EmptyPath(t *testing.T) {
	_, err := NewBaselineStore("")
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestBaselineStore_SaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "baseline.json")

	store, err := NewBaselineStore(path)
	if err != nil {
		t.Fatalf("NewBaselineStore: %v", err)
	}

	snap := Snapshot{
		{Destination: "10.0.0.0/8", Gateway: "192.168.1.1", Iface: "eth0", Protocol: "static"},
		{Destination: "0.0.0.0/0", Gateway: "192.168.1.254", Iface: "eth0", Protocol: "dhcp"},
	}

	if err := store.Save(snap); err != nil {
		t.Fatalf("Save: %v", err)
	}

	entry, err := store.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if len(entry.Routes) != len(snap) {
		t.Errorf("expected %d routes, got %d", len(snap), len(entry.Routes))
	}
	if entry.CapturedAt.IsZero() {
		t.Error("expected non-zero CapturedAt")
	}
}

func TestBaselineStore_LoadMissing(t *testing.T) {
	store, _ := NewBaselineStore("/tmp/nonexistent_baseline_xyz.json")
	_, err := store.Load()
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestBaselineStore_LoadCorrupt(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	_ = os.WriteFile(path, []byte("not json{"), 0644)

	store, _ := NewBaselineStore(path)
	_, err := store.Load()
	if err == nil {
		t.Fatal("expected error for corrupt file")
	}
}

func TestBaselineEntry_JSONRoundtrip(t *testing.T) {
	entry := BaselineEntry{
		CapturedAt: time.Now().UTC().Truncate(time.Second),
		Routes: Snapshot{
			{Destination: "172.16.0.0/12", Gateway: "10.0.0.1", Iface: "eth1"},
		},
	}
	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got BaselineEntry
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Routes[0].Destination != entry.Routes[0].Destination {
		t.Errorf("destination mismatch: %s != %s", got.Routes[0].Destination, entry.Routes[0].Destination)
	}
}
