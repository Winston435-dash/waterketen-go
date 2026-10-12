package core

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

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
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, `{"count":0}`) }))
	defer server.Close()
	previous := WiredSources
	WiredSources = []Source{{ID: "a", URL: server.URL, ProbeKind: "arcgis"}, {ID: "b", URL: server.URL, ProbeKind: "arcgis"}}
	t.Cleanup(func() { WiredSources = previous })
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

func TestProbeSourceValidatesHTTPAndPayload(t *testing.T) {
	for _, tc := range []struct {
		name, body, kind string
		code             int
		status           HealthStatus
		count            int
	}{
		{"http-error", `{"count":10}`, "arcgis", 500, HealthDown, -1},
		{"arcgis-error", `{"error":{"code":500}}`, "arcgis", 200, HealthDown, -1},
		{"count-with-error", `{"count":10,"error":{"code":500}}`, "arcgis", 200, HealthDown, -1},
		{"missing-count", `{}`, "arcgis", 200, HealthDown, -1},
		{"null-count", `{"count":null}`, "arcgis", 200, HealthDown, -1},
		{"negative-count", `{"count":-1}`, "arcgis", 200, HealthDown, -1},
		{"invalid-json", `no`, "arcgis", 200, HealthDown, -1},
		{"empty", `{"count":0}`, "arcgis", 200, HealthUp, 0},
		{"count", `{"count":42}`, "arcgis", 200, HealthUp, 42},
		{"wfs", `<FeatureCollection numberMatched="42"/>`, "wfs", 200, HealthUp, 42},
		{"ogc", `{"features":[]}`, "ogc", 200, HealthUp, -1},
		{"ogc-error", `{"error":"broken"}`, "ogc", 200, HealthDown, -1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(tc.code); fmt.Fprint(w, tc.body) }))
			defer server.Close()
			got := ProbeSource(Source{ID: "test", URL: server.URL, ProbeKind: tc.kind})
			if got.Status != tc.status || got.Count != tc.count {
				t.Fatalf("got %+v, want %s / %d", got, tc.status, tc.count)
			}
			if tc.status == HealthDown && got.Error == "" {
				t.Fatal("missing failure reason")
			}
		})
	}
}

func TestLiveSources(t *testing.T) {
	if testing.Short() || os.Getenv("WATERKETEN_LIVE_TESTS") != "1" {
		t.Skip("live probes; set WATERKETEN_LIVE_TESTS=1")
	}
	for _, health := range ProbeAll() {
		if health.Status == HealthDown {
			t.Errorf("%s: %s", health.SourceID, health.Error)
		}
	}
}

func TestProbeHTTPErrorIncludesStatus(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	defer server.Close()
	if got := ProbeSource(Source{URL: server.URL, ProbeKind: "arcgis"}); !strings.Contains(got.Error, "404") {
		t.Fatalf("status lost: %+v", got)
	}
}
