package core

import (
	"strings"
	"testing"
)

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

// imwaSources is the Waterschapshuis IMWA family added 12 okt '26. Licences
// were verified per NGR record (all records carry CC0 / "Geen beperkingen"),
// not inherited from the service that also hosts the INSPIRE hydrography
// dataset — that one stays unwired at CC BY-NC-ND 4.0, so the two families are
// deliberately not treated alike.
//
// The version suffix is not decoration: PDOK's "vernieuwing" article of
// 22 dec 2023 published the v1_0 URLs, and those answer HTTP 404 today. The
// renamed `waterschappen-<naam>` services must be probed on v2_0 — except
// keringen, which only answers on v3_0.
var imwaSources = map[string]struct{ url, typeName string }{
	"hwh_kunstwerken_imwa": {
		url:      "https://service.pdok.nl/hwh/waterschappen-kunstwerken-imwa/wfs/v2_0",
		typeName: "waterschappen-kunstwerken-imwa:gemaal",
	},
	"hwh_oppervlaktewateren_imwa": {
		url:      "https://service.pdok.nl/hwh/waterschappen-oppervlaktewateren-imwa/wfs/v2_0",
		typeName: "waterschappen-oppervlaktewateren-imwa:oppervlaktewaterlichaam",
	},
	"hwh_keringen_imwa": {
		url:      "https://service.pdok.nl/hwh/waterschappen-keringen-imwa/wfs/v3_0",
		typeName: "waterschappen-keringen-imwa:waterkering",
	},
	"hwh_waterbeheergebieden_imwa": {
		url:      "https://service.pdok.nl/hwh/waterschappen-waterbeheergebieden-imwa/wfs/v2_0",
		typeName: "waterschappen-waterbeheergebieden-imwa:peilgebied",
	},
}

// TestImwaServicesAreWiredAsCC0 pins the four IMWA services: a CC0 licence, a
// WFS probe on the version that answers, and the layer the probe counts.
func TestImwaServicesAreWiredAsCC0(t *testing.T) {
	byID := bySourceID(WiredSources)
	for id, want := range imwaSources {
		s, ok := byID[id]
		if !ok {
			t.Errorf("IMWA-bron %s ontbreekt in het register", id)
			continue
		}
		if s.License != "CC0" {
			t.Errorf("%s licentie = %q, verwacht CC0 (per NGR-record geverifieerd)", id, s.License)
		}
		if s.ProbeKind != "wfs" {
			t.Errorf("%s probe_kind = %q, verwacht wfs", id, s.ProbeKind)
		}
		if s.URL != want.url {
			t.Errorf("%s url = %q, verwacht %q", id, s.URL, want.url)
		}
		if s.TypeName != want.typeName {
			t.Errorf("%s type_name = %q, verwacht %q", id, s.TypeName, want.typeName)
		}
	}
}

// TestObjectTypeSourcesRegistered pins the DAMO and RIONED object types that
// were missing from the registry. Both services publish a whole family of
// object types; without a row per type the panel cannot tell a rioolgemaal
// from an RWZI, and the DAMO `pomp` type exists only in its download service —
// so a separate `pomp` source would report a false zero.
func TestObjectTypeSourcesRegistered(t *testing.T) {
	byID := bySourceID(WiredSources)
	damo := map[string]string{
		"damo_rioolgemaal":  "waterschappen-waterketen-damo:rioolgemaal",
		"damo_rwzi":         "waterschappen-waterketen-damo:rwzi",
		"damo_lozingspunt":  "waterschappen-waterketen-damo:lozingspunt",
		"damo_overnamepunt": "waterschappen-waterketen-damo:overnamepunt",
	}
	for id, typeName := range damo {
		s, ok := byID[id]
		if !ok {
			t.Errorf("DAMO-objectsoort %s ontbreekt", id)
			continue
		}
		if s.ProbeKind != "wfs_damo" || s.TypeName != typeName {
			t.Errorf("%s = kind %q type %q, verwacht wfs_damo + %q", id, s.ProbeKind, s.TypeName, typeName)
		}
		if s.License != "CC BY-SA 4.0" {
			t.Errorf("%s licentie = %q, verwacht CC BY-SA 4.0", id, s.License)
		}
	}
	rioned := map[string]string{
		"rioned_rioolgemaal": "waterschap_rioolgemaal",
		"rioned_rwzi":        "waterschap_rwzi",
		"rioned_iba":         "waterschap_iba",
	}
	for id, collection := range rioned {
		s, ok := byID[id]
		if !ok {
			t.Errorf("RIONED-collectie %s ontbreekt", id)
			continue
		}
		if s.ProbeKind != "ogc" || s.TypeName != collection {
			t.Errorf("%s = kind %q collectie %q, verwacht ogc + %q", id, s.ProbeKind, s.TypeName, collection)
		}
		if s.License != "CC0" {
			t.Errorf("%s licentie = %q, verwacht CC0", id, s.License)
		}
	}
}

// TestNoDeadPdokServiceURLIsRegistered: the HWH service names from PDOK's
// 22 dec 2023 article answer 404 today. Registering one would put a source in
// the panel that is permanently "down" while looking authoritative.
func TestNoDeadPdokServiceURLIsRegistered(t *testing.T) {
	dead := []string{
		"/hwh/waterschapsgrenzenimso/",
		"/hwh/zoneringenimwa/",
		"/hwh/kunstwerkenimwa/",
		"/hwh/oppervlaktewaterenimwa/",
		"/hwh/waterbeheergebiedenimwa/",
		"/hwh/keringenimwa/",
	}
	for _, s := range WiredSources {
		for _, d := range dead {
			if strings.Contains(s.URL, d) {
				t.Errorf("dode PDOK-URL geregistreerd bij %s: %s", s.ID, s.URL)
			}
		}
	}
	for _, u := range UnwiredSources {
		for _, d := range dead {
			if strings.Contains(u.URL, d) {
				t.Errorf("dode PDOK-URL geregistreerd bij %s: %s", u.ID, u.URL)
			}
		}
	}
}

func bySourceID(sources []Source) map[string]Source {
	out := make(map[string]Source, len(sources))
	for _, s := range sources {
		out[s.ID] = s
	}
	return out
}

func hasSuffix(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}
