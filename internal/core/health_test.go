package core

import "testing"

// TestAssembleHealth pins the contract the TUI relies on when it collects the
// wired probes one message at a time: the wired records keep their order, and
// the unwired (display-only) records are appended with Count -1 and a note.
func TestAssembleHealth(t *testing.T) {
	wired := []SourceHealth{
		{SourceID: "a", Status: HealthUp, Count: 5},
		{SourceID: "b", Status: HealthDown, Count: -1, Error: "timeout"},
	}
	got := AssembleHealth(wired)
	if len(got) != len(wired)+len(UnwiredSources) {
		t.Fatalf("len=%d, verwacht %d", len(got), len(wired)+len(UnwiredSources))
	}
	for i, w := range wired {
		if got[i] != w {
			t.Errorf("positie %d = %+v, verwacht %+v", i, got[i], w)
		}
	}
	for _, h := range got[len(wired):] {
		if h.Status != HealthNotWired || h.Count != -1 || h.Note == "" {
			t.Errorf("ongeldige niet-bedrade bron: %+v", h)
		}
	}
}

// TestProbeAllIncludesUnwired guards the CLI path: a full probe still reports
// the detected-but-unwired sources after a live run.
func TestProbeAllIncludesUnwired(t *testing.T) {
	if testing.Short() {
		t.Skip("raakt het netwerk")
	}
	got := ProbeAll()
	if len(got) != len(WiredSources)+len(UnwiredSources) {
		t.Fatalf("len=%d, verwacht %d", len(got), len(WiredSources)+len(UnwiredSources))
	}
	for i, src := range WiredSources {
		if got[i].SourceID != src.ID {
			t.Errorf("positie %d = %s, verwacht %s", i, got[i].SourceID, src.ID)
		}
	}
}
