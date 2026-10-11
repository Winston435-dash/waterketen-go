package core

import "testing"

func queryRing(coords ...[2]float64) LineString {
	out := make(LineString, len(coords))
	for i, p := range coords {
		out[i] = Point{X: p[0], Y: p[1]}
	}
	return out
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
		g Geometry
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
	if err != nil { t.Fatal(err) }
	if concave.intersects(Geometry{Kind: "point", Point: Point{X: 9, Y: 9}}) { t.Fatal("bbox-only match counted") }
	if _, err := newAreaFilter(Geometry{Kind: "point"}); err == nil { t.Fatal("non-polygon area accepted") }
}
