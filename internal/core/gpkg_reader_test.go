package core

import (
	"path/filepath"
	"testing"

	"waterketen/internal/testutil"
)

func TestGeoPackageReadsPreserveIDsAndQuotedNames(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ids.gpkg")
	db := testutil.GeoPackage(t, path, `INSERT INTO gpkg_geometry_columns VALUES('feature"table','geom',28992);
		CREATE TABLE "feature""table"(fid INTEGER PRIMARY KEY, geom BLOB, "quoted""property" TEXT);`)
	if _, err := db.Exec(`INSERT INTO "feature""table" VALUES(42,?,'value'),(99,?,'other')`, testutil.Point(1, 2), testutil.Point(3, 4)); err != nil {
		t.Fatal(err)
	}
	table, err := FeatureTable(db)
	if err != nil {
		t.Fatal(err)
	}
	if count, err := table.Count(); err != nil || count != 2 {
		t.Fatalf("count=%d / %v", count, err)
	}
	var rows []Row
	if err := table.Each(func(row Row) { rows = append(rows, row) }); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 || rows[0].ID != 42 || rows[1].ID != 99 || rows[0].Properties[`quoted"property`] != "value" {
		t.Fatalf("bad full read: %+v", rows)
	}
	rows = nil
	if err := table.Select([]int64{99}, func(row Row) { rows = append(rows, row) }); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].ID != 99 {
		t.Fatalf("bad viewport read: %+v", rows)
	}
}

func TestGeoPackageRejectsMissingGeometryColumn(t *testing.T) {
	db := testutil.GeoPackage(t, filepath.Join(t.TempDir(), "missing.gpkg"), `INSERT INTO gpkg_geometry_columns VALUES('features','geom',28992); CREATE TABLE features(fid INTEGER PRIMARY KEY);`)
	table, err := FeatureTable(db)
	if err != nil {
		t.Fatal(err)
	}
	if err := table.Each(func(Row) {}); err == nil {
		t.Fatal("missing geometry column reported an empty layer")
	}
}
