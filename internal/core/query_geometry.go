package core

import (
	"fmt"
	"math"
)

// areaFilter prepares polygon rings once. Its edge indexes keep exact area
// selection affordable for whole-city queries, without clipping source objects.
type areaFilter struct {
	polygons []queryPolygon
	edges    []queryEdge
	index    GridIndex
}

type queryEdge struct{ a, b Point }
type queryRingIndex struct {
	bounds BBox
	edges  []queryEdge
	index  GridIndex
}
type queryPolygon struct{ rings []queryRingIndex }

func newAreaFilter(g Geometry) (*areaFilter, error) {
	polygons := g.Polygons
	if g.Kind == "polygon" {
		polygons = []Polygon{g.Polygon}
	} else if g.Kind != "multipolygon" {
		return nil, fmt.Errorf("gebiedsgrens is geen polygon")
	}
	filter := &areaFilter{}
	for _, polygon := range polygons {
		if len(polygon) == 0 {
			return nil, fmt.Errorf("lege gebiedsgrens")
		}
		p := queryPolygon{}
		for _, ring := range polygon {
			r, err := newQueryRing(ring)
			if err != nil {
				return nil, err
			}
			p.rings = append(p.rings, r)
			filter.edges = append(filter.edges, r.edges...)
		}
		filter.polygons = append(filter.polygons, p)
	}
	if len(filter.polygons) == 0 {
		return nil, fmt.Errorf("lege gebiedsgrens")
	}
	filter.index = queryEdgeIndex(filter.edges)
	return filter, nil
}

func newQueryRing(ring LineString) (queryRingIndex, error) {
	if len(ring) < 3 {
		return queryRingIndex{}, fmt.Errorf("ongeldige polygon-ring")
	}
	for _, p := range ring {
		if math.IsNaN(p.X) || math.IsNaN(p.Y) || math.IsInf(p.X, 0) || math.IsInf(p.Y, 0) {
			return queryRingIndex{}, fmt.Errorf("niet-eindige gebiedscoördinaat")
		}
	}
	r := queryRingIndex{}
	r.bounds, _ = (Geometry{Kind: "linestring", Line: ring}).Bounds()
	for i := range ring {
		a, b := ring[i], ring[(i+1)%len(ring)]
		if a.X != b.X || a.Y != b.Y {
			r.edges = append(r.edges, queryEdge{a, b})
		}
	}
	if len(r.edges) < 3 {
		return queryRingIndex{}, fmt.Errorf("ongeldige polygon-ring")
	}
	r.index = queryEdgeIndex(r.edges)
	return r, nil
}

func queryEdgeIndex(edges []queryEdge) GridIndex {
	geometries := make([]Geometry, len(edges))
	for i, e := range edges {
		geometries[i] = Geometry{Kind: "linestring", Line: LineString{e.a, e.b}}
	}
	return NewGridIndex(geometries)
}

// location returns outside (0), inside (1), or on the ring (2). A horizontal
// ray uses half-open y intervals so a shared vertex is never counted twice.
func (r *queryRingIndex) location(p Point) int {
	if !r.bounds.intersects(BBox{p.X, p.Y, p.X, p.Y}) {
		return 0
	}
	inside := false
	for _, i := range r.index.Query(BBox{p.X, p.Y, r.bounds.MaxX, p.Y}) {
		e := r.edges[i]
		cross := queryCross(e.a, e.b, p)
		if cross == 0 && queryOnSegment(p, e.a, e.b) {
			return 2
		}
		if (e.a.Y > p.Y) != (e.b.Y > p.Y) {
			x := e.a.X + (p.Y-e.a.Y)*(e.b.X-e.a.X)/(e.b.Y-e.a.Y)
			if p.X < x {
				inside = !inside
			}
		}
	}
	if inside {
		return 1
	}
	return 0
}

func (p *queryPolygon) covers(point Point) bool {
	outer := p.rings[0].location(point)
	if outer == 0 {
		return false
	}
	if outer == 2 {
		return true
	}
	for i := 1; i < len(p.rings); i++ {
		switch p.rings[i].location(point) {
		case 2:
			return true // hole edges belong to the polygon boundary
		case 1:
			return false
		}
	}
	return true
}

func (f *areaFilter) covers(point Point) bool {
	for i := range f.polygons {
		if f.polygons[i].covers(point) {
			return true
		}
	}
	return false
}

func (f *areaFilter) intersects(g Geometry) bool {
	if len(g.Parts) > 0 {
		for _, part := range g.Parts {
			if f.intersects(part) {
				return true
			}
		}
		return false
	}
	if g.Kind == "point" {
		return f.covers(g.Point)
	}
	for _, line := range g.Outlines() {
		for _, point := range line {
			if f.covers(point) {
				return true
			}
		}
		for i := 1; i < len(line); i++ {
			a, b := line[i-1], line[i]
			box := BBox{math.Min(a.X, b.X), math.Min(a.Y, b.Y), math.Max(a.X, b.X), math.Max(a.Y, b.Y)}
			for _, edge := range f.index.Query(box) {
				e := f.edges[edge]
				if querySegmentsIntersect(a, b, e.a, e.b) {
					return true
				}
			}
		}
	}
	// With no crossing edges, a polygon may still enclose an entire area.
	polygons := g.Polygons
	if g.Kind == "polygon" {
		polygons = []Polygon{g.Polygon}
	}
	for _, polygon := range polygons {
		p := queryPolygon{}
		for _, ring := range polygon {
			r, err := newQueryRing(ring)
			if err != nil {
				return false
			}
			p.rings = append(p.rings, r)
		}
		if len(p.rings) == 0 {
			continue
		}
		for _, area := range f.polygons {
			if p.covers(area.rings[0].edges[0].a) {
				return true
			}
		}
	}
	return false
}

func queryCross(a, b, p Point) float64 { return (b.X-a.X)*(p.Y-a.Y) - (b.Y-a.Y)*(p.X-a.X) }
func queryOnSegment(p, a, b Point) bool {
	return p.X >= math.Min(a.X, b.X) && p.X <= math.Max(a.X, b.X) && p.Y >= math.Min(a.Y, b.Y) && p.Y <= math.Max(a.Y, b.Y)
}
func querySegmentsIntersect(a, b, c, d Point) bool {
	a1, a2, b1, b2 := queryCross(a, b, c), queryCross(a, b, d), queryCross(c, d, a), queryCross(c, d, b)
	if a1 == 0 && queryOnSegment(c, a, b) || a2 == 0 && queryOnSegment(d, a, b) || b1 == 0 && queryOnSegment(a, c, d) || b2 == 0 && queryOnSegment(b, c, d) {
		return true
	}
	return (a1 > 0 && a2 < 0 || a1 < 0 && a2 > 0) && (b1 > 0 && b2 < 0 || b1 < 0 && b2 > 0)
}
