package core

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"testing"
)

func integrationDataRoot(t *testing.T) string {
	t.Helper()
	if testing.Short() || os.Getenv("WATERKETEN_INTEGRATION") != "1" {
		t.Skip("managed-data integration test; set WATERKETEN_INTEGRATION=1")
	}
	return DataRoot()
}

// TestGpkgRealData reads the real managed Delfshaven files and pins the
// numbers the Python pipeline established. This is the regression net for the
// whole rewrite: if the reader, the WKB decoder, or BOB handling drift, this
// test fails — matching the same file the Python pipeline wrote.
func TestGpkgRealData(t *testing.T) {
	dataRoot := integrationDataRoot(t)
	if !fileExists(filepath.Join(dataRoot, "processed", "delfshaven-riool.gpkg")) {
		t.Fatal("geen managed data voor integratieverificatie")
	}

	db, err := Open(filepath.Join(dataRoot, "processed", "delfshaven-riool.gpkg"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	tbl, err := FeatureTable(db)
	if err != nil {
		t.Fatal(err)
	}
	count, err := tbl.Count()
	if err != nil {
		t.Fatal(err)
	}
	if count != 4610 {
		t.Errorf("riool-rijen=%d, verwacht 4610", count)
	}

	var (
		lines, polys, other int
		minX, minY          = math.MaxFloat64, math.MaxFloat64
		maxX, maxY          = -math.MaxFloat64, -math.MaxFloat64
		bobMeasured         int
		streets             = map[string]struct{}{}
		materials           = map[string]struct{}{}
	)
	err = tbl.Each(func(r Row) {
		switch r.Geometry.Kind {
		case "linestring", "multilinestring":
			lines++
		case "polygon", "multipolygon":
			polys++
		default:
			other++
		}
		b, ok := r.Geometry.Bounds()
		if !ok {
			t.Error("geometrie zonder bounds")
			return
		}
		minX = math.Min(minX, b.MinX)
		minY = math.Min(minY, b.MinY)
		maxX = math.Max(maxX, b.MaxX)
		maxY = math.Max(maxY, b.MaxY)
		for _, col := range []string{"BOB_ACTUEE", "BOB_ACTU_1"} {
			if v, ok := r.Numbers[col]; ok && v != 0 {
				bobMeasured++
				break
			}
		}
		if s := r.Properties["STRAATNAAM"]; s != "" {
			streets[s] = struct{}{}
		}
		if m := r.Properties["MATERIAAL1"]; m != "" {
			materials[m] = struct{}{}
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	if other != 0 || polys != 0 {
		t.Errorf("riool moet alleen lijnen bevatten; vond %d polygonen, %d overig", polys, other)
	}
	if lines != 4610 {
		t.Errorf("lijnen=%d, verwacht 4610", lines)
	}
	// Python: 88008..91584 (x) × 435145..437596 (y)
	if diff(minX, 88008) > 1 || diff(minY, 435145) > 1 || diff(maxX, 91584) > 1 || diff(maxY, 437596) > 1 {
		t.Errorf("bounds afwijkend: %.3f,%.3f → %.3f,%.3f", minX, minY, maxX, maxY)
	}
	if bobMeasured != 4461 {
		t.Errorf("gemeten BOB=%d, verwacht 4461 (minstens één uiteinde ≠ 0)", bobMeasured)
	}
	if len(streets) != 342 {
		t.Errorf("straten=%d, verwacht 342", len(streets))
	}
	if len(materials) != 12 {
		t.Errorf("materialen=%d, verwacht 12", len(materials))
	}
}

// TestAtlasRendersRealScene renders the full Delfshaven scene at the TUI's
// typical atlas size and asserts the picture is recognisably the network:
// a healthy number of drawn cells, the expected counts per layer, and a
// non-empty legend. This is the pin that catches regressions in Bresenham,
// clipping, cell mapping, and the overlay policy.
func TestAtlasRendersRealScene(t *testing.T) {
	dataRoot := integrationDataRoot(t)
	if !fileExists(filepath.Join(dataRoot, "processed", "delfshaven-riool.gpkg")) {
		t.Fatal("geen managed data voor integratieverificatie")
	}
	scene := mustScene(t, "delfshaven", []string{"riool", "drinkwaterleidingen"})
	view := FitView(scene.Bounds)
	r := RenderScene(scene, view, 110, 38, "riool", true, "netwerk", true, true, nil, "", nil, true)
	if r.Counts["riool"] != 4610 {
		t.Errorf("riool-count=%d, verwacht 4610", r.Counts["riool"])
	}
	if r.Counts["drinkwaterleidingen"] == 0 {
		t.Error("drinkwaterleidingen wordt niet meegetekend in overlay")
	}
	if len(r.Legend) == 0 {
		t.Error("lege legend")
	}
	drawn := 0
	for _, row := range r.Cells {
		for _, c := range row {
			if c.Rune != ' ' && c.Rune != 0x2800 {
				drawn++
			}
		}
	}
	if drawn < 500 {
		t.Errorf("slechts %d cellen getekend — netwerk zou er voller uit moeten zien", drawn)
	}
}

func fileExists(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && !fi.IsDir()
}

func diff(a, b float64) float64 { return math.Abs(a - b) }

func mustScene(t *testing.T, area string, layers []string) *Scene {
	t.Helper()
	dataRoot := integrationDataRoot(t)
	var sources []SceneSource
	for _, name := range layers {
		path := filepath.Join(dataRoot, "processed", fmt.Sprintf("%s-%s.gpkg", area, name))
		if !fileExists(path) {
			t.Fatalf("laag ontbreekt: %s", path)
		}
		db, err := Open(path)
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		tbl, err := FeatureTable(db)
		if err != nil {
			t.Fatal(err)
		}
		var feats []Feature
		tbl.Each(func(r Row) {
			f := Feature{
				Layer:      name,
				Position:   len(feats),
				Geometry:   r.Geometry,
				Properties: r.Properties,
			}
			if s := r.Properties["STRAATNAAM"]; s != "" {
				f.Street = s
			}
			if f.Street == "" {
				f.Street = r.Properties["name"]
			}
			a, okA := r.Numbers["BOB_ACTUEE"]
			bb, okB := r.Numbers["BOB_ACTU_1"]
			useA, useB := okA && a != 0, okB && bb != 0
			switch {
			case useA && useB:
				f.BOB = (a + bb) / 2
			case useA:
				f.BOB = a
			case useB:
				f.BOB = bb
			default:
				f.BOB = math.NaN()
			}
			feats = append(feats, f)
		})
		sources = append(sources, SceneSource{Name: name, Features: feats})
	}
	scene, err := NewScene(area, sources, nil)
	if err != nil {
		t.Fatal(err)
	}
	return scene
}
