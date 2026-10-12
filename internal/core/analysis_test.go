package core

import (
	"math"
	"path/filepath"
	"reflect"
	"testing"

	"waterketen/internal/testutil"
)

func TestLoadLayerReturnsReadErrors(t *testing.T) {
	path := filepath.Join(t.TempDir(), "broken.gpkg")
	testutil.GeoPackage(t, path, `INSERT INTO gpkg_geometry_columns VALUES('missing','geom',28992);`)
	data, err := LoadLayer(AssetMeta{Area: "test", Layer: "riool", ProcessedPath: path})
	if err == nil || data != nil {
		t.Fatalf("unreadable table returned success: %+v / %v", data, err)
	}
	layers, failures := LoadManagedLayers([]AssetMeta{{Area: "test", Layer: "riool", ProcessedPath: path}})
	if len(layers) != 0 || len(failures) != 1 {
		t.Fatalf("load diagnostics lost: %v / %v", layers, failures)
	}
}

func TestLoadLayerAndSummaryUseBothBOBEnds(t *testing.T) {
	path := filepath.Join(t.TempDir(), "layer.gpkg")
	db := testutil.GeoPackage(t, path, `INSERT INTO gpkg_geometry_columns VALUES('features','geom',28992);
		CREATE TABLE features(fid INTEGER PRIMARY KEY, geom BLOB, STRAATNAAM TEXT, BOB_ACTUEE REAL, BOB_ACTU_1 REAL);`)
	for i, ends := range [][2]float64{{-2, -4}, {0, -2}, {0, 0}} {
		if _, err := db.Exec(`INSERT INTO features VALUES(?,?,?,?,?)`, i+1, testutil.Point(float64(i), 1), "Teststraat", ends[0], ends[1]); err != nil {
			t.Fatal(err)
		}
	}
	data, err := LoadLayer(AssetMeta{Area: "test", Layer: "riool", ProcessedPath: path})
	if err != nil {
		t.Fatal(err)
	}
	if len(data.Features) != 3 || data.Features[0].BOB != -3 || data.Features[1].BOB != -2 || !math.IsNaN(data.Features[2].BOB) {
		t.Fatalf("wrong BOBs: %+v", data.Features)
	}
	s := Summarize(data.Features)
	if s.Features != 3 || s.Points != 3 || s.BOBMeasured != 2 || s.BOBMedian != -2.5 || s.Streets != 1 {
		t.Fatalf("wrong summary: %+v", s)
	}
}

func TestSummaryIsPureAndHandlesEmptyAndNonFiniteBOB(t *testing.T) {
	if s := Summarize(nil); s != (LayerSummary{}) {
		t.Fatalf("empty summary: %+v", s)
	}
	features := []Feature{
		{Geometry: Geometry{Kind: "linestring", Line: LineString{{X: 0, Y: 0}, {X: 3, Y: 4}}}, BOB: -4, Street: "A"},
		{Geometry: Geometry{Kind: "multipoint"}, BOB: math.Inf(1), Street: "A"},
		{Geometry: Geometry{Kind: "collection"}, BOB: -2, Street: "B"},
	}
	before := append([]Feature(nil), features...)
	s := Summarize(features)
	if s.Lines != 1 || s.Points != 1 || s.Other != 1 || s.TotalM != 5 || s.BOBMeasured != 2 || s.BOBMin != -4 || s.BOBMax != -2 || s.BOBMedian != -3 || s.Streets != 2 {
		t.Fatalf("wrong summary: %+v", s)
	}
	if !reflect.DeepEqual(before, features) {
		t.Fatal("summary mutated input")
	}
}
