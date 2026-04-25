package routes

import (
	"strings"
	"testing"
	"time"
)

func makeDiff(added, removed []Route) Diff {
	return Diff{Added: added, Removed: removed}
}

func TestSummarize_Empty(t *testing.T) {
	now := time.Now()
	s := Summarize(nil, now, now)
	if s.TotalDiffs != 0 || s.TotalAdded != 0 || s.TotalRemoved != 0 {
		t.Errorf("expected zero counts, got %+v", s)
	}
}

func TestSummarize_Counts(t *testing.T) {
	now := time.Now()
	diffs := []Diff{
		makeDiff(
			[]Route{{Destination: "10.0.0.0/8", Iface: "eth0", Protocol: "kernel"}},
			[]Route{{Destination: "192.168.1.0/24", Iface: "eth1", Protocol: "static"}},
		),
		makeDiff(
			[]Route{{Destination: "172.16.0.0/12", Iface: "eth0", Protocol: "kernel"}},
			nil,
		),
	}

	s := Summarize(diffs, now, now.Add(time.Minute))

	if s.TotalDiffs != 2 {
		t.Errorf("expected 2 diffs, got %d", s.TotalDiffs)
	}
	if s.TotalAdded != 2 {
		t.Errorf("expected 2 added, got %d", s.TotalAdded)
	}
	if s.TotalRemoved != 1 {
		t.Errorf("expected 1 removed, got %d", s.TotalRemoved)
	}
}

func TestSummarize_InterfaceCounts(t *testing.T) {
	now := time.Now()
	diffs := []Diff{
		makeDiff(
			[]Route{{Iface: "eth0"}, {Iface: "eth0"}, {Iface: "eth1"}},
			nil,
		),
	}
	s := Summarize(diffs, now, now)

	ifaceMap := map[string]int{}
	for _, ic := range s.TopInterfaces {
		ifaceMap[ic.Iface] = ic.Count
	}
	if ifaceMap["eth0"] != 2 {
		t.Errorf("expected eth0 count=2, got %d", ifaceMap["eth0"])
	}
	if ifaceMap["eth1"] != 1 {
		t.Errorf("expected eth1 count=1, got %d", ifaceMap["eth1"])
	}
}

func TestFormatSummary_ContainsFields(t *testing.T) {
	now := time.Now()
	s := Summarize([]Diff{
		makeDiff([]Route{{Iface: "eth0", Protocol: "kernel"}}, nil),
	}, now, now.Add(time.Minute))

	out := FormatSummary(s)
	for _, want := range []string{"Summary", "Diffs", "Added", "Removed", "eth0", "kernel"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in output:\n%s", want, out)
		}
	}
}
