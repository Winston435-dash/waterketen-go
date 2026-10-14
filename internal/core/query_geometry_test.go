package core

import (
	"encoding/binary"
	"math"
	"testing"
)

func queryRing(coords ...[2]float64) LineString {
	out := make(LineString, len(coords))
	for i, p := range coords {
		out[i] = Point{X: p[0], Y: p[1]}
	}
	return out
}

// Rotterdam's clipped groen contains real GeometryCollections. Their parts
// remain one source object, including an enclosing polygon with no inside vertices.
func TestQuerySupportsCollectedAndMultiPointGeometry(t *testing.T) {
	collect := func(kind uint32, parts ...[]byte) []byte {
		b := make([]byte, 9)
		b[0] = 1
		binary.LittleEndian.PutUint32(b[1:5], kind)
		binary.LittleEndian.PutUint32(b[5:9], uint32(len(parts)))
		for _, part := range parts {
			b = append(b, part...)
		}
		return b
	}
	polygon := wkbLine(3, 1)
	polygon = append(polygon, wkbLine(2, 5, -1, -1, 11, -1, 11, 11, -1, 11, -1, -1)[5:]...)
	boundary := Geometry{Kind: "polygon", Polygon: Polygon{queryRing([2]float64{0, 0}, [2]float64{10, 0}, [2]float64{10, 10}, [2]float64{0, 10}, [2]float64{0, 0})}}
	filter, err := newAreaFilter(boundary)
	if err != nil {
		t.Fatal(err)
	}
	for _, raw := range [][]byte{
		collect(7, wkbPoint(1, 20, 20), collect(7, polygon)),
		collect(4, wkbPoint(1, 20, 20), wkbPoint(1, 5, 5)),
	} {
		geometry, err := parseWKB(raw)
		if err != nil {
			t.Fatal(err)
		}
		if !filter.intersects(geometry) {
			t.Fatal("collected intersection lost")
		}
		if bounds, ok := geometry.Bounds(); !ok || bounds.MaxX != 20 || bounds.MaxY != 20 {
			t.Fatalf("collected bounds: %+v / %v", bounds, ok)
		}
		if len(geometry.Dots()) == 0 {
			t.Fatal("collected point lost")
		}
		if geometry.Distance(Point{X: 20, Y: 20}) != 0 {
			t.Fatal("collected point not inspectable by distance")
		}
		if anchor := geometry.RepresentativePoint(); math.IsNaN(anchor.X) {
			t.Fatal("collection has no label anchor")
		}
		if anchor := geometry.Interpolate(0.5); math.IsNaN(anchor.X) {
			t.Fatal("collection cannot be framed")
		}
	}
}

func TestAreaFilterUsesPolygonHolesAndBoundary(t *testing.T) {
	outer := queryRing([2]float64{0, 0}, [2]float64{10, 0}, [2]float64{10, 10}, [2]float64{0, 10}, [2]float64{0, 0})
	hole := queryRing([2]float64{3, 3}, [2]float64{7, 3}, [2]float64{7, 7}, [2]float64{3, 7}, [2]float64{3, 3})
	filter, err := newAreaFilter(Geometry{Kind: "polygon", Polygon: Polygon{outer, hole}})
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name string
		g    Geometry
		want bool
	}{
		{"inside", Geometry{Kind: "point", Point: Point{X: 1, Y: 1}}, true},
		{"hole is not area", Geometry{Kind: "point", Point: Point{X: 5, Y: 5}}, false},
		{"outer boundary", Geometry{Kind: "point", Point: Point{X: 0, Y: 5}}, true},
		{"hole boundary", Geometry{Kind: "point", Point: Point{X: 3, Y: 5}}, true},
		{"just outside", Geometry{Kind: "point", Point: Point{X: -0.000001, Y: 5}}, false},
		{"crossing without inside vertices", Geometry{Kind: "linestring", Line: queryRing([2]float64{-1, 5}, [2]float64{11, 5})}, true},
		{"line entirely in hole", Geometry{Kind: "linestring", Line: queryRing([2]float64{4, 4}, [2]float64{6, 6})}, false},
		{"polygon contains area", Geometry{Kind: "polygon", Polygon: Polygon{queryRing([2]float64{-1, -1}, [2]float64{11, -1}, [2]float64{11, 11}, [2]float64{-1, 11}, [2]float64{-1, -1})}}, true},
		{"polygon entirely in hole", Geometry{Kind: "polygon", Polygon: Polygon{queryRing([2]float64{4, 4}, [2]float64{6, 4}, [2]float64{6, 6}, [2]float64{4, 6}, [2]float64{4, 4})}}, false},
		{"multipart crossing", Geometry{Kind: "multilinestring", Lines: []LineString{queryRing([2]float64{20, 20}, [2]float64{21, 21}), queryRing([2]float64{-1, 2}, [2]float64{11, 2})}}, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := filter.intersects(c.g); got != c.want {
				t.Fatalf("intersects = %v, want %v", got, c.want)
			}
		})
	}
	// A concave boundary excludes objects that lie only in its bounding box.
	concave, err := newAreaFilter(Geometry{Kind: "polygon", Polygon: Polygon{queryRing([2]float64{0, 0}, [2]float64{10, 0}, [2]float64{0, 10}, [2]float64{0, 0})}})
	if err != nil {
		t.Fatal(err)
	}
	if concave.intersects(Geometry{Kind: "point", Point: Point{X: 9, Y: 9}}) {
		t.Fatal("bbox-only match counted")
	}
	if _, err := newAreaFilter(Geometry{Kind: "point"}); err == nil {
		t.Fatal("non-polygon area accepted")
	}
}

func TestRepresentativePointStaysInsideConcavePolygon(t *testing.T) {
	geometry := Geometry{Kind: "polygon", Polygon: Polygon{queryRing(
		[2]float64{0, 0}, [2]float64{4, 0}, [2]float64{4, 4},
		[2]float64{3, 4}, [2]float64{3, 1}, [2]float64{1, 1},
		[2]float64{1, 4}, [2]float64{0, 4}, [2]float64{0, 0},
	)}}
	filter, err := newAreaFilter(geometry)
	if err != nil {
		t.Fatal(err)
	}
	point := geometry.RepresentativePoint()
	if !filter.covers(point) {
		t.Fatalf("representative point %+v falls outside the concave polygon", point)
	}
}
