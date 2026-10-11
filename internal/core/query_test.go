package core

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"errors"
	"math"
	"path/filepath"
	"reflect"
	"testing"
)

// A real SQLite/R-tree fixture includes two pumps, a non-pump, and a bbox-only
// match. Counts must be independent of the size of the requested object page.
func queryTestLibrary(t *testing.T) *Library {
	t.Helper()
	root := t.TempDir()
	create := func(file, schema string) *sql.DB {
		t.Helper()
		db, err := sql.Open("sqlite", filepath.Join(root, file))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { db.Close() })
		if _, err := db.Exec(schema); err != nil {
			t.Fatal(err)
		}
		return db
	}
	areas := create("areas.gpkg", `CREATE TABLE gpkg_geometry_columns (table_name TEXT, column_name TEXT, srs_id INTEGER);
		INSERT INTO gpkg_geometry_columns VALUES ('areas','geom',28992);
		CREATE TABLE areas (fid INTEGER PRIMARY KEY, code TEXT, name TEXT, kind TEXT, geom BLOB);`)
	var wkb bytes.Buffer
	wkb.WriteByte(1)
	binary.Write(&wkb, binary.LittleEndian, uint32(3))
	binary.Write(&wkb, binary.LittleEndian, uint32(1))
	binary.Write(&wkb, binary.LittleEndian, uint32(4))
	for _, p := range [][2]float64{{0, 0}, {10, 0}, {0, 10}, {0, 0}} {
		binary.Write(&wkb, binary.LittleEndian, p[0])
		binary.Write(&wkb, binary.LittleEndian, p[1])
	}
	blob := append([]byte{'G', 'P', 0, 1, 0x40, 0x71, 0, 0}, wkb.Bytes()...)
	if _, err := areas.Exec(`INSERT INTO areas VALUES (1,'WK000001','Testwijk','wijk',?)`, blob); err != nil {
		t.Fatal(err)
	}
	data := create("gemalen.gpkg", `CREATE TABLE gpkg_geometry_columns (table_name TEXT, column_name TEXT, srs_id INTEGER);
		INSERT INTO gpkg_geometry_columns VALUES ('pumps','geom',28992);
		CREATE TABLE pumps (fid INTEGER PRIMARY KEY, naam TEXT, type TEXT, geom BLOB);
		CREATE VIRTUAL TABLE rtree_pumps_geom USING rtree(id,minx,maxx,miny,maxy);`)
	for i, p := range [][2]float64{{1, 1}, {2, 2}, {3, 3}, {9, 9}} {
		var point bytes.Buffer
		point.WriteByte(1)
		binary.Write(&point, binary.LittleEndian, uint32(1))
		binary.Write(&point, binary.LittleEndian, p[0])
		binary.Write(&point, binary.LittleEndian, p[1])
		blob := append([]byte{'G', 'P', 0, 1, 0x40, 0x71, 0, 0}, point.Bytes()...)
		kind := "http://data.gwsw.nl/1.5/totaal/Rioolgemaal"
		if i == 2 {
			kind = "http://data.gwsw.nl/1.5/totaal/Inspectieput"
		}
		if _, err := data.Exec(`INSERT INTO pumps VALUES (?,?,?,?)`, i+1, "pump", kind, blob); err != nil {
			t.Fatal(err)
		}
		if _, err := data.Exec(`INSERT INTO rtree_pumps_geom VALUES (?,?,?,?,?)`, i+1, p[0], p[0], p[1], p[1]); err != nil {
			t.Fatal(err)
		}
	}
	lib := &Library{Root: root, Snapshot: "test"}
	if err := json.Unmarshal([]byte(`{"crs":"EPSG:28992","layers":{"gemalen":{"file":"gemalen.gpkg","source":{"license":"CC0"}}}}`), &lib.Manifest); err != nil {
		t.Fatal(err)
	}
	return lib
}

func TestAreaQueryRejectsNonFiniteCoordinates(t *testing.T) {
	lib := queryTestLibrary(t)
	db, err := sql.Open("sqlite", lib.LayerPath("gemalen"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var blob []byte
	if err := db.QueryRow(`SELECT geom FROM pumps WHERE fid=1`).Scan(&blob); err != nil {
		t.Fatal(err)
	}
	binary.LittleEndian.PutUint64(blob[13:21], math.Float64bits(math.NaN()))
	if _, err := db.Exec(`UPDATE pumps SET geom=? WHERE fid=1`, blob); err != nil {
		t.Fatal(err)
	}
	result, err := RunAreaQuery(context.Background(), lib, "WK000001", []string{"rioolgemalen"})
	if err != nil {
		t.Fatal(err)
	}
	if row := result.Layers[0]; row.Status != QueryFailed || row.Count != -1 || len(row.IDs) != 0 {
		t.Fatalf("corrupt coordinates reported as complete: %+v", row)
	}
}

func TestAreaQueryCountsExactTypedObjectsWithoutPageLimit(t *testing.T) {
	lib := queryTestLibrary(t)
	result, err := RunAreaQuery(context.Background(), lib, "WK000001", []string{"rioolgemalen", "putten"})
	if err != nil {
		t.Fatal(err)
	}
	pumps := result.Layers[0]
	if pumps.Status != QueryOK || pumps.Count != 2 || !reflect.DeepEqual(pumps.IDs, []int64{1, 2}) {
		t.Fatalf("expected two real pumps, got %+v", pumps)
	}
	if result.Layers[1].Status != QueryUnavailable || result.Layers[1].Count != -1 {
		t.Fatalf("missing layer reported as a count: %+v", result.Layers[1])
	}
	for offset, want := range []int64{1, 2} {
		rows, err := ReadQueryObjects(context.Background(), lib, pumps, offset, 1)
		if err != nil {
			t.Fatal(err)
		}
		if len(rows) != 1 || rows[0].ID != want {
			t.Fatalf("page %d: %+v", offset, rows)
		}
		if pumps.Count != 2 {
			t.Fatal("page size changed total")
		}
	}
	rows, err := ReadQueryObjects(context.Background(), lib, pumps, 2, 1)
	if err != nil || len(rows) != 0 {
		t.Fatalf("end page: %+v / %v", rows, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := RunAreaQuery(ctx, lib, "WK000001", []string{"rioolgemalen"}); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation lost: %v", err)
	}
}
