package core

import (
	"strings"
	"testing"
)

func testLibrary(t *testing.T) *Library {
	t.Helper()
	lib, err := OpenLibrary(dataRoot)
	if err != nil {
		t.Fatal(err)
	}
	if lib == nil {
		t.Skip("geen kaartbibliotheek — sla bibliotheek-tests over")
	}
	return lib
}

// TestAreasRealSnapshot pins the published area registry: 115 CBS areas, the
// three kinds, and the whole-of-Rotterdam bounds. It is the regression net for
// panel 5's "world" — if areas.gpkg reading or the BBox extraction drifts, the
// map silently frames the wrong city.
func TestAreasRealSnapshot(t *testing.T) {
	areas, err := LoadAreas(testLibrary(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(areas) != 115 {
		t.Errorf("gebieden=%d, verwacht 115", len(areas))
	}
	kinds := map[string]int{}
	for _, a := range areas {
		kinds[a.Kind]++
	}
	for kind, want := range map[string]int{KindGemeente: 1, KindWijk: 22, KindBuurt: 92} {
		if kinds[kind] != want {
			t.Errorf("%s=%d, verwacht %d", kind, kinds[kind], want)
		}
	}

	rot, ok := DefaultArea(areas)
	if !ok {
		t.Fatal("geen standaardgebied")
	}
	if rot.Code != DefaultAreaCode || rot.Name != "Rotterdam" || rot.Kind != KindGemeente {
		t.Fatalf("standaardgebied = %+v, verwacht GM0599 Rotterdam gemeente", rot)
	}
	// Python measured GM0599 as 55.500..101.033 × 428.647..447.000.
	got := rot.Bounds
	if diff(got.MinX, 55500) > 1 || diff(got.MinY, 428647.44) > 1 ||
		diff(got.MaxX, 101032.6) > 1 || diff(got.MaxY, 447000) > 1 {
		t.Errorf("Rotterdam-bounds afwijkend: %.1f,%.1f → %.1f,%.1f",
			got.MinX, got.MinY, got.MaxX, got.MaxY)
	}
	// 45.5 km wide: the whole municipality, not one district. Pinning the
	// number here is the point — panel 5's default is Rotterdam.
	if w := (got.MaxX - got.MinX) / 1000; w < 44 || w > 47 {
		t.Errorf("Rotterdam is %.1f km breed, verwacht ~45,5", w)
	}
}

// TestSelectAreaAmbiguity pins the rule the Python select_area enforces:
// "Delfshaven" is BOTH a wijk (WK059903) and a buurt (BU05990320), so resolving
// it by name must fail loudly with both codes rather than silently pick one.
func TestSelectAreaAmbiguity(t *testing.T) {
	areas, err := LoadAreas(testLibrary(t))
	if err != nil {
		t.Fatal(err)
	}
	_, err = SelectArea(areas, "Delfshaven")
	if err == nil {
		t.Fatal("Delfshaven is dubbelzinnig; verwachtte een fout")
	}
	for _, code := range []string{"WK059903", "BU05990320"} {
		if !strings.Contains(err.Error(), code) {
			t.Errorf("fout moet %s noemen, kreeg: %v", code, err)
		}
	}

	// An explicit code resolves without ambiguity, and name matching is
	// case/space insensitive (the Python normalise_name contract).
	byCode, err := SelectArea(areas, "WK059903")
	if err != nil {
		t.Fatalf("WK059903: %v", err)
	}
	if byCode.Name != "Delfshaven" || byCode.Kind != KindWijk {
		t.Errorf("WK059903 = %+v", byCode)
	}
	if byName, err := SelectArea(areas, "  rotterdam "); err != nil || byName.Code != DefaultAreaCode {
		t.Errorf("'  rotterdam ' hoort GM0599 te geven, kreeg %+v / %v", byName, err)
	}
	if _, err := SelectArea(areas, "Atlantis"); err == nil {
		t.Error("onbekend gebied hoort een fout te geven")
	}
}

// TestAreaLabelLeadsWithCode: the label starts with the CBS code, which is the
// unique identifier and the disambiguator ("Delfshaven" is both WK059903 and
// BU05990320). A phone truncates the label's tail, so the tail must be the part
// that is merely descriptive.
func TestAreaLabelLeadsWithCode(t *testing.T) {
	a := Area{Code: "WK059903", Name: "Delfshaven", Kind: KindWijk}
	got := a.Label()
	if !strings.HasPrefix(got, "WK059903") {
		t.Errorf("label %q begint niet met de CBS-code", got)
	}
	if !strings.Contains(got, "Delfshaven") {
		t.Errorf("label %q mist de naam", got)
	}
}
func TestFilterAreas(t *testing.T) {
	areas, err := LoadAreas(testLibrary(t))
	if err != nil {
		t.Fatal(err)
	}
	if got := len(FilterAreas(areas, "")); got != len(areas) {
		t.Errorf("lege query gaf %d, verwacht %d", got, len(areas))
	}
	hits := FilterAreas(areas, "delf")
	if len(hits) == 0 || hits[0].Name != "Delfshaven" {
		t.Errorf("zoeken op 'delf' gaf %+v", hits)
	}
	if byCode := FilterAreas(areas, "GM0599"); len(byCode) != 1 || byCode[0].Code != DefaultAreaCode {
		t.Errorf("zoeken op GM0599 = %+v", byCode)
	}
	if got := FilterAreas(areas, "zzz-niets"); len(got) != 0 {
		t.Errorf("onzin-query gaf %d treffers", len(got))
	}
}
