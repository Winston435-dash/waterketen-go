// GeoPackage reading without cgo: SQLite via database/sql (pure-Go driver)
// plus a small hand-rolled GeoPackageBinary/WKB parser. Only what waterketen
// needs: points, (multi)linestrings and (multi)polygons, 2D or XYZ.
package core

import (
	"database/sql"
	"encoding/binary"
	"fmt"
	"math"
	"strings"

	_ "modernc.org/sqlite"
)

// Bounds is (minx, miny, maxx, maxy), in the file's CRS — always RD New for
// the waterketen snapshots (EPSG:28992).
type Bounds struct{ MinX, MinY, MaxX, MaxY float64 }

// Point / LineString carry XYZ; Z is NaN for 2D geometries.
type Point struct{ X, Y, Z float64 }
type LineString []Point
type Polygon []LineString // ring 0 = outer, rest = holes

// Geometry is a tagged union of the supported WKB shapes.
type Geometry struct {
	Kind     string // "point" | "linestring" | "polygon" | "multilinestring" | "multipolygon"
	Point    Point
	Line     LineString
	Lines    []LineString
	Polygon  Polygon
	Polygons []Polygon
}

// Row is one row of a GeoPackage table: parsed geometry plus raw attributes.
type Row struct {
	ID         int64
	Geometry   Geometry
	Properties map[string]string
	Numbers    map[string]float64 // numeric columns kept numeric for analysis
}

// Table is one feature table from a GeoPackage.
type Table struct {
	db    *sql.DB
	name  string
	geom  string // geometry column name
	Lower string // lower-case name for idempotent lookups
}

// Open opens a .gpkg for reading.
func Open(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path+"?mode=ro")
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	return db, nil
}

// FeatureTable returns the one feature table of the file (the waterketen
// snapshots are single-layer). Errors when zero or several tables exist.
func FeatureTable(db *sql.DB) (*Table, error) {
	rows, err := db.Query(`SELECT table_name, column_name FROM gpkg_geometry_columns`)
	if err != nil {
		return nil, fmt.Errorf("geen gpkg_geometry_columns: %w", err)
	}
	defer rows.Close()
	type pair struct{ table, geom string }
	var found []pair
	for rows.Next() {
		var p pair
		if err := rows.Scan(&p.table, &p.geom); err != nil {
			return nil, err
		}
		found = append(found, p)
	}
	if len(found) != 1 {
		return nil, fmt.Errorf("verwachtte 1 feature-tabel, vond %d", len(found))
	}
	return &Table{db: db, name: found[0].table, geom: found[0].geom}, nil
}

// Count returns the row count of the table.
func (t *Table) Count() (int, error) {
	var n int
	err := t.db.QueryRow(`SELECT COUNT(*) FROM "` + t.name + `"`).Scan(&n)
	return n, err
}

// Each streams every feature with a geometry through fn. A nil geometry row
// is skipped (matches the Python pipeline: geometry-less rows are registration
// noise, not features).
func (t *Table) Each(fn func(Row)) error {
	cols, err := t.columns()
	if err != nil {
		return err
	}
	query := `SELECT "` + strings.Join(cols.names, `","`) + `" FROM "` + t.name + `"`
	rows, err := t.db.Query(query)
	if err != nil {
		return fmt.Errorf("lees %s: %w", t.name, err)
	}
	defer rows.Close()

	for rows.Next() {
		vals := make([]any, len(cols.names))
		scan := make([]any, len(cols.names))
		for i := range vals {
			scan[i] = &vals[i]
		}
		if err := rows.Scan(scan...); err != nil {
			return fmt.Errorf("scan %s: %w", t.name, err)
		}
		f := Row{Properties: map[string]string{}, Numbers: map[string]float64{}}
		for i, name := range cols.names {
			if name == t.geom {
				blob, ok := vals[i].([]byte)
				if !ok || len(blob) == 0 {
					continue
				}
				geom, err := parseGeoPackageBinary(blob)
				if err != nil {
					continue // corrupt cell: skip the feature, never the table
				}
				f.Geometry = geom
				continue
			}
			switch v := vals[i].(type) {
			case int64:
				f.Numbers[name] = float64(v)
				f.Properties[name] = fmt.Sprintf("%d", v)
			case float64:
				f.Numbers[name] = v
				f.Properties[name] = fmt.Sprintf("%g", v)
			case string:
				f.Properties[name] = v
			case []byte:
				f.Properties[name] = string(v)
			}
		}
		if f.Geometry.Kind == "" {
			continue
		}
		fn(f)
	}
	return rows.Err()
}

type columnInfo struct{ names []string }

func (t *Table) columns() (columnInfo, error) {
	rows, err := t.db.Query(`PRAGMA table_info("` + t.name + `")`)
	if err != nil {
		return columnInfo{}, err
	}
	defer rows.Close()
	var info columnInfo
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dflt any
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			return columnInfo{}, err
		}
		info.names = append(info.names, name)
	}
	return info, rows.Err()
}

// --- GeoPackageBinaryHeader + WKB ------------------------------------------

var envelopeSize = [8]int{0, 32, 48, 48, 64, 64, 0, 0}

// parseGeoPackageBinary parses the 'GP' header, skips the envelope, then WKB.
func parseGeoPackageBinary(b []byte) (Geometry, error) {
	if len(b) < 9 || b[0] != 'G' || b[1] != 'P' {
		return Geometry{}, fmt.Errorf("geen GeoPackageBinaryHeader")
	}
	flags := b[3]
	var order binary.ByteOrder = binary.LittleEndian
	if flags&1 == 0 {
		order = binary.BigEndian
	}
	_ = order // srs_id word (bytes 4..8); not needed for geometry parsing
	envKind := (flags >> 1) & 0x7
	off := 8 + envelopeSize[envKind]
	if off >= len(b) {
		return Geometry{}, fmt.Errorf("envelope loopt buiten blob")
	}
	return parseWKB(b[off:])
}

// parseWKB decodes ISO WKB including the Z/M 1000/2000/3000 type offsets.
func parseWKB(b []byte) (Geometry, error) {
	g, _, err := readWKB(b)
	return g, err
}

func readWKB(b []byte) (Geometry, int, error) {
	if len(b) < 5 {
		return Geometry{}, 0, fmt.Errorf("WKB te kort")
	}
	var order binary.ByteOrder = binary.LittleEndian
	if b[0] == 0 {
		order = binary.BigEndian
	} else if b[0] != 1 {
		return Geometry{}, 0, fmt.Errorf("onbekende WKB byte order %d", b[0])
	}
	raw := order.Uint32(b[1:5])
	dims := 2
	hasZ := false
	base := raw
	switch {
	case raw >= 3000: // WKB ZM
		dims, base = 4, raw-3000
		hasZ = true
	case raw >= 2000: // WKB M-only: the 3rd ordinate is a measure, not Z.
		// The parser keeps Z = NaN for the whole geometry, so a measure can
		// never be mistaken for an elevation.
		dims, base = 3, raw-2000
	case raw >= 1000: // WKB Z
		dims, base = 3, raw-1000
		hasZ = true
	}
	off := 5

	read := func(p *Point) error {
		need := 8 * dims
		if off+need > len(b) {
			return fmt.Errorf("WKB buffer te kort voor punt")
		}
		p.X = math.Float64frombits(order.Uint64(b[off:]))
		p.Y = math.Float64frombits(order.Uint64(b[off+8:]))
		p.Z = math.NaN()
		if hasZ {
			p.Z = math.Float64frombits(order.Uint64(b[off+16:]))
		}
		off += need
		return nil
	}
	readCount := func() (int, error) {
		if off+4 > len(b) {
			return 0, fmt.Errorf("WKB buffer te kort voor telling")
		}
		n := int(order.Uint32(b[off:]))
		off += 4
		return n, nil
	}

	switch base {
	case 1: // Point
		var p Point
		if err := read(&p); err != nil {
			return Geometry{}, 0, err
		}
		return Geometry{Kind: "point", Point: p}, off, nil
	case 2: // LineString
		n, err := readCount()
		if err != nil {
			return Geometry{}, 0, err
		}
		line := make(LineString, n)
		for i := range line {
			if err := read(&line[i]); err != nil {
				return Geometry{}, 0, err
			}
		}
		return Geometry{Kind: "linestring", Line: line}, off, nil
	case 3: // Polygon
		n, err := readCount()
		if err != nil {
			return Geometry{}, 0, err
		}
		poly := make(Polygon, n)
		for ring := range poly {
			m, err := readCount()
			if err != nil {
				return Geometry{}, 0, err
			}
			poly[ring] = make(LineString, m)
			for i := range poly[ring] {
				if err := read(&poly[ring][i]); err != nil {
					return Geometry{}, 0, err
				}
			}
		}
		return Geometry{Kind: "polygon", Polygon: poly}, off, nil
	case 5, 6: // MultiLineString / MultiPolygon
		n, err := readCount()
		if err != nil {
			return Geometry{}, 0, err
		}
		out := Geometry{}
		if base == 5 {
			out.Kind = "multilinestring"
		} else {
			out.Kind = "multipolygon"
		}
		for i := 0; i < n; i++ {
			sub, used, err := readWKB(b[off:])
			if err != nil {
				return Geometry{}, 0, err
			}
			off += used
			if base == 5 && sub.Kind == "linestring" {
				out.Lines = append(out.Lines, sub.Line)
			}
			if base == 6 && sub.Kind == "polygon" {
				out.Polygons = append(out.Polygons, sub.Polygon)
			}
		}
		return out, off, nil
	default:
		return Geometry{}, 0, fmt.Errorf("WKB-type %d niet ondersteund", raw)
	}
}
