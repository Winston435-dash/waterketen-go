package core

import "testing"

// knownProbeKinds is the closed set of probe implementations in health.go.
// A new source with a new kind needs a probe branch too — this test makes that
// a conscious act instead of a silently "down" source in the panel.
var knownProbeKinds = map[string]bool{
	"wfs":      true,
	"wfs_damo": true,
	"arcgis":   true,
	"ogc":      true,
}

// TestWiredSourceRegistry pins the invariants every wired source must satisfy:
// a unique id, a label, a URL, a licence string and a probe kind that exists.
// Licence is required because it is a decision here, not a detail: CC0, CC-BY-SA
// and "unknown" are materially different permissions.
func TestWiredSourceRegistry(t *testing.T) {
	seen := map[string]bool{}
	for _, s := range WiredSources {
		if s.ID == "" || s.Label == "" || s.URL == "" {
			t.Errorf("bron mist id/label/url: %+v", s)
		}
		if s.License == "" {
			t.Errorf("bron %s heeft geen licentie — dat is een beslissing, geen detail", s.ID)
		}
		if !knownProbeKinds[s.ProbeKind] {
			t.Errorf("bron %s heeft onbekende probe_kind %q", s.ID, s.ProbeKind)
		}
		if seen[s.ID] {
			t.Errorf("dubbele source-id %q", s.ID)
		}
		seen[s.ID] = true
		// A WFS/OGC probe must name its layer: the GWSW layer name sent to DAMO
		// returns HTTP 400, so the type name belongs to the source.
		if (s.ProbeKind == "wfs" || s.ProbeKind == "wfs_damo" || s.ProbeKind == "ogc") && s.TypeName == "" {
			t.Errorf("bron %s (%s) mist type_name", s.ID, s.ProbeKind)
		}
	}
}

// TestSourcesAddedFromRegister pins the two sources that were verified live
// against their real endpoints, plus their licence. Measured 8 okt '26:
// Zuid-Holland WFS answers numberMatched=10 for DRINKWATER_BEDRIJVEN (the ten
// utilities, Evides over Delfshaven), RWS regiogebieden/0 answers count=48.
func TestSourcesAddedFromRegister(t *testing.T) {
	byID := map[string]Source{}
	for _, s := range WiredSources {
		byID[s.ID] = s
	}
	zh, ok := byID["zh_drinkwater"]
	if !ok {
		t.Fatal("zh_drinkwater ontbreekt")
	}
	if zh.ProbeKind != "wfs" || zh.TypeName != "water:DRINKWATER_BEDRIJVEN" {
		t.Errorf("zh_drinkwater verwacht wfs + water:DRINKWATER_BEDRIJVEN, kreeg %+v", zh)
	}
	if zh.License != "Publiek domein (PDM 1.0)" {
		t.Errorf("zh_drinkwater licentie = %q", zh.License)
	}
	rws, ok := byID["rws_regiogebieden"]
	if !ok {
		t.Fatal("rws_regiogebieden ontbreekt")
	}
	// Measured: the FeatureServer form of this service answers HTTP 500; the
	// MapServer layer 0 is the working endpoint.
	if rws.ProbeKind != "arcgis" || !hasSuffix(rws.URL, "/MapServer/0") {
		t.Errorf("rws_regiogebieden verwacht arcgis + /MapServer/0, kreeg %+v", rws)
	}
}

// TestUnwiredSourcesDocumentReason: a detected-but-unwired source without a
// note would show up in the panel as an unexplained entry.
func TestUnwiredSourcesDocumentReason(t *testing.T) {
	for _, u := range UnwiredSources {
		if u.ID == "" || u.Label == "" || u.URL == "" || u.Note == "" {
			t.Errorf("niet-bedrade bron mist een veld: %+v", u)
		}
		if u.License == "" {
			t.Errorf("niet-bedrade bron %s mist licentie", u.ID)
		}
	}
}

func hasSuffix(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}
