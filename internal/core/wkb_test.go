package core

import (
	"encoding/binary"
	"math"
	"testing"
)

// wkbPoint builds a little-endian Point with the given WKB type code and
// ordinates, so a test can feed the parser a 2D, Z, M-only or ZM geometry.
func wkbPoint(typ uint32, coords ...float64) []byte {
	b := []byte{1, 0, 0, 0, 0}
	binary.LittleEndian.PutUint32(b[1:5], typ)
	for _, c := range coords {
		var buf [8]byte
		binary.LittleEndian.PutUint64(buf[:], math.Float64bits(c))
		b = append(b, buf[:]...)
	}
	return b
}

// wkbLine builds a little-endian LineString header plus its vertex ordinates:
// the WKB type code, the vertex count, then the coordinates.
func wkbLine(typ uint32, count uint32, coords ...float64) []byte {
	b := []byte{1, 0, 0, 0, 0}
	binary.LittleEndian.PutUint32(b[1:5], typ)
	var cnt [4]byte
	binary.LittleEndian.PutUint32(cnt[:], count)
	b = append(b, cnt[:]...)
	for _, c := range coords {
		var buf [8]byte
		binary.LittleEndian.PutUint64(buf[:], math.Float64bits(c))
		b = append(b, buf[:]...)
	}
	return b
}

// TestWKBDimensions pins the Z/M decoding. An M-only geometry carries a measure
// in its third ordinate; reading it as Z would silently turn a measure into an
// elevation. The parser must keep Z = NaN there, and read Z only when the type
// actually declares it (1001/3001).
func TestWKBDimensions(t *testing.T) {
	cases := []struct {
		name    string
		typ     uint32
		coords  []float64
		wantZ   float64
		wantNaN bool
	}{
		{"2D", 1, []float64{1, 2}, 0, true},
		{"Z", 1001, []float64{1, 2, 3}, 3, false},
		{"M-only", 2001, []float64{1, 2, 99}, 0, true},
		{"ZM", 3001, []float64{1, 2, 3, 99}, 3, false},
	}
	for _, c := range cases {
		g, err := parseWKB(wkbPoint(c.typ, c.coords...))
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		if g.Kind != "point" {
			t.Errorf("%s: kind=%q, wil point", c.name, g.Kind)
		}
		if g.Point.X != c.coords[0] || g.Point.Y != c.coords[1] {
			t.Errorf("%s: XY=%.1f,%.1f, wil %.1f,%.1f", c.name, g.Point.X, g.Point.Y, c.coords[0], c.coords[1])
		}
		if c.wantNaN {
			if !math.IsNaN(g.Point.Z) {
				t.Errorf("%s: Z=%.1f, wil NaN (geen hoogte in deze geometrie)", c.name, g.Point.Z)
			}
			continue
		}
		if g.Point.Z != c.wantZ {
			t.Errorf("%s: Z=%.1f, wil %.1f", c.name, g.Point.Z, c.wantZ)
		}
	}
}

// TestWKBLineStringZOffsets checks the ordinate stride on a multi-point
// geometry: an M-only line must advance 3 ordinates per vertex and still keep
// Z = NaN.
func TestWKBLineStringZOffsets(t *testing.T) {
	// LineString (1002, Z) with two 3D vertices: type + count + ordinates.
	g, err := parseWKB(wkbLine(1002, 2, 10, 20, 5, 11, 21, 6))
	if err != nil {
		t.Fatal(err)
	}
	if g.Kind != "linestring" || len(g.Line) != 2 {
		t.Fatalf("kind=%q punten=%d, wil linestring met 2 punten", g.Kind, len(g.Line))
	}
	if g.Line[1].X != 11 || g.Line[1].Y != 21 || g.Line[1].Z != 6 {
		t.Errorf("tweede punt = %.1f,%.1f,%.1f, wil 11,21,6", g.Line[1].X, g.Line[1].Y, g.Line[1].Z)
	}

	// LineString M-only (2002): third ordinate is a measure, Z stays NaN.
	g, err = parseWKB(wkbLine(2002, 2, 0, 0, 0, 10, 0, 42))
	if err != nil {
		t.Fatal(err)
	}
	if !math.IsNaN(g.Line[1].Z) {
		t.Errorf("M-only lijn: Z=%.1f, wil NaN", g.Line[1].Z)
	}
	if g.Line[1].X != 10 {
		t.Errorf("M-only lijn: X=%.1f, wil 10", g.Line[1].X)
	}
}
