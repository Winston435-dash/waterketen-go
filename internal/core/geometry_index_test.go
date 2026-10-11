package core

import (
	"reflect"
	"testing"
)

func TestGridIndexQueryIncludesLongIntervalsAndSortsIDs(t *testing.T) {
	geoms := []Geometry{
		{Kind: "linestring", Line: LineString{{X: 0, Y: 0}, {X: 100, Y: 0}}},
		{Kind: "linestring", Line: LineString{{X: 20, Y: 10}, {X: 21, Y: 11}}},
		{Kind: "linestring", Line: LineString{{X: 40, Y: 0}, {X: 41, Y: 1}}},
		{Kind: "linestring", Line: LineString{{X: 5, Y: 0}, {X: 6, Y: 1}}},
	}
	idx := NewGridIndex(geoms)
	got := idx.Query(BBox{MinX: 39, MinY: -1, MaxX: 42, MaxY: 2})
	want := []int{0, 2}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Query() = %v, want %v", got, want)
	}
}

func TestGridIndexQueryMatchesBruteForce(t *testing.T) {
	geoms := make([]Geometry, 250)
	for i := range geoms {
		x := float64((i * 37) % 113)
		y := float64((i * 19) % 71)
		// Some intervals deliberately span most of the x-range, ensuring an
		// interval starting far before the query cannot be skipped.
		if i%17 == 0 {
			x = -50
		}
		geoms[i] = Geometry{Kind: "linestring", Line: LineString{
			{X: x, Y: y}, {X: x + float64((i%23)+1), Y: y + 3},
		}}
	}
	idx := NewGridIndex(geoms)
	for minX := -40.0; minX < 100; minX += 7.25 {
		query := BBox{MinX: minX, MinY: 15, MaxX: minX + 12, MaxY: 43}
		var want []int
		for i, geom := range geoms {
			bounds, ok := geom.Bounds()
			if ok && bounds.intersects(query) {
				want = append(want, i)
			}
		}
		if got := idx.Query(query); !reflect.DeepEqual(got, want) {
			t.Fatalf("Query(%+v) = %v, brute force = %v", query, got, want)
		}
	}
}
