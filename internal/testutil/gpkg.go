// Package testutil builds small, deterministic GeoPackage fixtures. Tests do not
// need a user's managed store or any external service.
package testutil

import (
	"bytes"
	"database/sql"
	"encoding/binary"
	"testing"

	_ "modernc.org/sqlite"
)

func GeoPackage(t testing.TB, path, schema string) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	_, err = db.Exec(`CREATE TABLE gpkg_geometry_columns (table_name TEXT, column_name TEXT, srs_id INTEGER);` + schema)
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func binaryGeometry(kind uint32, write func(*bytes.Buffer)) []byte {
	var b bytes.Buffer
	b.Write([]byte{'G', 'P', 0, 1, 0x40, 0x71, 0, 0, 1})
	binary.Write(&b, binary.LittleEndian, kind)
	write(&b)
	return b.Bytes()
}

func Point(x, y float64) []byte {
	return binaryGeometry(1, func(b *bytes.Buffer) {
		binary.Write(b, binary.LittleEndian, x)
		binary.Write(b, binary.LittleEndian, y)
	})
}

func Polygon(ring [][2]float64) []byte {
	return binaryGeometry(3, func(b *bytes.Buffer) {
		binary.Write(b, binary.LittleEndian, uint32(1))
		binary.Write(b, binary.LittleEndian, uint32(len(ring)))
		for _, p := range ring {
			binary.Write(b, binary.LittleEndian, p[0])
			binary.Write(b, binary.LittleEndian, p[1])
		}
	})
}
