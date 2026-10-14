package core

// Geometry helpers: iteration, bounding boxes, bbox intersection tests,
// distance from a point to a geometry, polygon-outlining, linear
// interpolation along a line. This replaces the shapely surface the Python
// atlas uses (sindex.query, boundary, representative_point, interpolate,
// distance). Exact area-query predicates live separately in query_geometry.go.

import (
	"math"
	"sort"
)

// BBox is a (minx, miny, maxx, maxy) rectangle in EPSG:28992.
type BBox struct{ MinX, MinY, MaxX, MaxY float64 }

func (b BBox) intersects(o BBox) bool {
	return b.MinX <= o.MaxX && b.MaxX >= o.MinX && b.MinY <= o.MaxY && b.MaxY >= o.MinY
}

// Grow expands the box by r on all sides.
func (b BBox) Grow(r float64) BBox {
	return BBox{b.MinX - r, b.MinY - r, b.MaxX + r, b.MaxY + r}
}

// mergeBBox folds o into b; when ok is false b is treated as empty and seeded
// with o. It is the one place a running bounding box is extended, so the fit
// bounds, street views and geometry bounds all grow the same way.
func mergeBBox(b BBox, ok bool, o BBox) (BBox, bool) {
	if !ok {
		return o, true
	}
	b.MinX = math.Min(b.MinX, o.MinX)
	b.MinY = math.Min(b.MinY, o.MinY)
	b.MaxX = math.Max(b.MaxX, o.MaxX)
	b.MaxY = math.Max(b.MaxY, o.MaxY)
	return b, true
}

// Lines normalises the geometry to its set of polylines: lines stay lines,
// polygons become their RINGS (outline only — a filled district hides the
// network), points get no line.
func (g Geometry) Outlines() []LineString {
	switch g.Kind {
	case "linestring":
		return []LineString{g.Line}
	case "multilinestring":
		return g.Lines
	case "polygon":
		return g.Polygon
	case "collection":
		var out []LineString
		for _, part := range g.Parts {
			out = append(out, part.Outlines()...)
		}
		return out
	case "multipolygon":
		var out []LineString
		for _, p := range g.Polygons {
			out = append(out, p...)
		}
		return out
	}
	return nil
}

// Points normalises the geometry to its individual dots (the point case).
func (g Geometry) Dots() []Point {
	if g.Kind == "point" {
		return []Point{g.Point}
	}
	var out []Point
	for _, part := range g.Parts {
		out = append(out, part.Dots()...)
	}
	return out
}

// Bounds returns the geometry's bounding box.
func (g Geometry) Bounds() (BBox, bool) {
	if len(g.Parts) > 0 {
		var bounds BBox
		valid := false
		for _, part := range g.Parts {
			if b, ok := part.Bounds(); ok {
				bounds, valid = mergeBBox(bounds, valid, b)
			}
		}
		return bounds, valid
	}
	var b BBox
	started := false
	visit := func(p Point) {
		if !started {
			b = BBox{p.X, p.Y, p.X, p.Y}
			started = true
			return
		}
		b.MinX = math.Min(b.MinX, p.X)
		b.MinY = math.Min(b.MinY, p.Y)
		b.MaxX = math.Max(b.MaxX, p.X)
		b.MaxY = math.Max(b.MaxY, p.Y)
	}
	if g.Kind == "point" {
		visit(g.Point)
		return b, true
	}
	for _, line := range g.Outlines() {
		for _, p := range line {
			visit(p)
		}
	}
	return b, started
}

// Length is the planimetric length of the geometry's outlines.
func (g Geometry) Length() float64 {
	total := 0.0
	for _, line := range g.Outlines() {
		for i := 1; i < len(line); i++ {
			total += math.Hypot(line[i].X-line[i-1].X, line[i].Y-line[i-1].Y)
		}
	}
	return total
}

// Distance is the minimum distance from p to the geometry's outlines (or to
// its point). Polygons measure to the OUTLINE, matching the Python
// `geom.boundary` behaviour before distance().
func (g Geometry) Distance(p Point) float64 {
	if g.Kind == "point" {
		return math.Hypot(g.Point.X-p.X, g.Point.Y-p.Y)
	}
	best := math.Inf(1)
	for _, dot := range g.Dots() {
		best = math.Min(best, math.Hypot(p.X-dot.X, p.Y-dot.Y))
	}
	for _, line := range g.Outlines() {
		for i := 1; i < len(line); i++ {
			d := segDistance(p, line[i-1], line[i])
			if d < best {
				best = d
			}
		}
	}
	return best
}

// segDistance is the point-to-segment distance.
func segDistance(p, a, b Point) float64 {
	dx, dy := b.X-a.X, b.Y-a.Y
	if dx == 0 && dy == 0 {
		return math.Hypot(p.X-a.X, p.Y-a.Y)
	}
	t := ((p.X-a.X)*dx + (p.Y-a.Y)*dy) / (dx*dx + dy*dy)
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	return math.Hypot(p.X-(a.X+t*dx), p.Y-(a.Y+t*dy))
}

// Interpolate returns the point at fraction `t` (0..1) of total line length
// along the geometry's longest polyline — shapely's
// `geom.interpolate(.5, normalized=True)` for the label anchor.
func (g Geometry) Interpolate(t float64) Point {
	lines := g.Outlines()
	if len(lines) == 0 {
		if points := g.Dots(); len(points) > 0 {
			return points[0]
		}
		return Point{math.NaN(), math.NaN(), math.NaN()}
	}
	longest := lines[0]
	best := lineLength(longest)
	for _, l := range lines[1:] {
		if n := lineLength(l); n > best {
			longest, best = l, n
		}
	}
	if best == 0 {
		return longest[0]
	}
	target := t * best
	acc := 0.0
	for i := 1; i < len(longest); i++ {
		seg := math.Hypot(longest[i].X-longest[i-1].X, longest[i].Y-longest[i-1].Y)
		if acc+seg >= target {
			f := 0.0
			if seg > 0 {
				f = (target - acc) / seg
			}
			return Point{
				longest[i-1].X + f*(longest[i].X-longest[i-1].X),
				longest[i-1].Y + f*(longest[i].Y-longest[i-1].Y),
				math.NaN(),
			}
		}
		acc += seg
	}
	return longest[len(longest)-1]
}

func lineLength(l LineString) float64 {
	n := 0.0
	for i := 1; i < len(l); i++ {
		n += math.Hypot(l[i].X-l[i-1].X, l[i].Y-l[i-1].Y)
	}
	return n
}

// RepresentativePoint chooses an anchor on the geometry. Polygon anchors are
// kept inside the polygon; line anchors use their midpoint and collections use
// the first part with a finite anchor.
func (g Geometry) RepresentativePoint() Point {
	switch g.Kind {
	case "point":
		return g.Point
	case "polygon":
		return polygonRepresentativePoint(g.Polygon)
	case "multipolygon":
		for _, polygon := range g.Polygons {
			if point := polygonRepresentativePoint(polygon); finiteXY(point) {
				return point
			}
		}
	case "collection":
		for _, part := range g.Parts {
			if point := part.RepresentativePoint(); finiteXY(point) {
				return point
			}
		}
	case "linestring", "multilinestring":
		if point := g.Interpolate(0.5); finiteXY(point) {
			return point
		}
	case "multipoint":
		if points := g.Dots(); len(points) > 0 {
			return points[0]
		}
	}
	var sx, sy float64
	var n int
	for _, line := range g.Outlines() {
		for _, p := range line {
			sx, sy, n = sx+p.X, sy+p.Y, n+1
		}
	}
	for _, p := range g.Dots() {
		sx, sy, n = sx+p.X, sy+p.Y, n+1
	}
	if n == 0 {
		return Point{math.NaN(), math.NaN(), math.NaN()}
	}
	return Point{sx / float64(n), sy / float64(n), math.NaN()}
}

func polygonRepresentativePoint(polygon Polygon) Point {
	var sx, sy float64
	var n int
	var ys []float64
	for _, ring := range polygon {
		for _, point := range ring {
			if !finiteXY(point) {
				continue
			}
			sx, sy, n = sx+point.X, sy+point.Y, n+1
			ys = append(ys, point.Y)
		}
	}
	if n == 0 {
		return Point{math.NaN(), math.NaN(), math.NaN()}
	}
	candidate := Point{sx / float64(n), sy / float64(n), math.NaN()}
	if polygonCovers(candidate, polygon) {
		return candidate
	}

	// Between consecutive vertex Y values, the polygon's horizontal crossings
	// are stable. Pick the midpoint of the widest interval that is truly inside,
	// including holes in the even/odd crossing sequence.
	sort.Float64s(ys)
	uniqueY := ys[:0]
	for _, y := range ys {
		if len(uniqueY) == 0 || uniqueY[len(uniqueY)-1] != y {
			uniqueY = append(uniqueY, y)
		}
	}
	bestWidth := 0.0
	var best Point
	for i := 1; i < len(uniqueY); i++ {
		if uniqueY[i] <= uniqueY[i-1] {
			continue
		}
		y := uniqueY[i-1] + (uniqueY[i]-uniqueY[i-1])/2
		var crossings []float64
		for _, ring := range polygon {
			for j, a := range ring {
				if len(ring) < 2 {
					break
				}
				b := ring[(j+1)%len(ring)]
				if !finiteXY(a) || !finiteXY(b) || (a.Y > y) == (b.Y > y) {
					continue
				}
				crossings = append(crossings, a.X+(y-a.Y)*(b.X-a.X)/(b.Y-a.Y))
			}
		}
		sort.Float64s(crossings)
		for j := 0; j+1 < len(crossings); j += 2 {
			width := crossings[j+1] - crossings[j]
			if width <= bestWidth {
				continue
			}
			point := Point{crossings[j] + width/2, y, math.NaN()}
			if polygonCovers(point, polygon) {
				best, bestWidth = point, width
			}
		}
	}
	if bestWidth > 0 {
		return best
	}
	// A degenerate ring has no interior scanline; its boundary is still part of
	// the geometry and is a safer anchor than an outside vertex average.
	for _, ring := range polygon {
		for _, point := range ring {
			if finiteXY(point) && polygonCovers(point, polygon) {
				point.Z = math.NaN()
				return point
			}
		}
	}
	return Point{math.NaN(), math.NaN(), math.NaN()}
}

func polygonCovers(point Point, polygon Polygon) bool {
	if len(polygon) == 0 {
		return false
	}
	outer := ringLocation(point, polygon[0])
	if outer == 0 {
		return false
	}
	if outer == 2 {
		return true
	}
	for _, hole := range polygon[1:] {
		switch ringLocation(point, hole) {
		case 2:
			return true
		case 1:
			return false
		}
	}
	return true
}

func ringLocation(point Point, ring LineString) int {
	if len(ring) < 3 {
		return 0
	}
	inside := false
	for i, a := range ring {
		b := ring[(i+1)%len(ring)]
		cross := (b.X-a.X)*(point.Y-a.Y) - (b.Y-a.Y)*(point.X-a.X)
		if cross == 0 && point.X >= math.Min(a.X, b.X) && point.X <= math.Max(a.X, b.X) &&
			point.Y >= math.Min(a.Y, b.Y) && point.Y <= math.Max(a.Y, b.Y) {
			return 2
		}
		if (a.Y > point.Y) != (b.Y > point.Y) {
			x := a.X + (point.Y-a.Y)*(b.X-a.X)/(b.Y-a.Y)
			if point.X < x {
				inside = !inside
			}
		}
	}
	if inside {
		return 1
	}
	return 0
}

func finiteXY(point Point) bool {
	return !math.IsNaN(point.X) && !math.IsNaN(point.Y) && !math.IsInf(point.X, 0) && !math.IsInf(point.Y, 0)
}

// --- Spatial index ----------------------------------------------------------
//
// The Python atlas uses a STRtree via `gdf.sindex.query(box)`. We bucketed by
// 64 m grid cells first, but the cell map dominated the render profile at 98%
// of frames (hashing [2]int keys per candidate). A sorted entry array with a
// binary search on minX plus a bbox filter answers the same query with no
// hashing and no map at all — a linear scan over the candidate window on
// cache-friendly float32 fields.

// GridIndex is a spatial index over the features' bounding boxes.
type GridIndex struct {
	entries    []giEntry // sorted by minX
	prefixMaxX []float64 // monotonic maxima, so old long intervals are not skipped
}

// giEntry is one feature's bbox. Keep coordinates at full precision so a
// query touching an edge cannot become a false miss through float rounding.
type giEntry struct {
	minX, minY, maxX, maxY float64
	id                     int
}

// NewGridIndex builds the index over the features' bounding boxes. The type
// keeps its historical name; the structure is the sorted array above.
func NewGridIndex(geoms []Geometry) GridIndex {
	entries := make([]giEntry, 0, len(geoms))
	for i, g := range geoms {
		b, ok := g.Bounds()
		if !ok {
			continue
		}
		entries = append(entries, giEntry{b.MinX, b.MinY, b.MaxX, b.MaxY, i})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].minX < entries[j].minX })
	prefixMaxX := make([]float64, len(entries))
	for i, e := range entries {
		prefixMaxX[i] = e.maxX
		if i > 0 && prefixMaxX[i-1] > prefixMaxX[i] {
			prefixMaxX[i] = prefixMaxX[i-1]
		}
	}
	return GridIndex{entries: entries, prefixMaxX: prefixMaxX}
}

// Query returns the ascending indices of features whose bbox intersects `b`.
// Search the monotonic prefix maximum for the first possible x-overlap, then
// scan entries by minX until the query's right edge. The immutable index is
// safe to query concurrently.
func (idx *GridIndex) Query(b BBox) []int {
	es := idx.entries
	lo := sort.Search(len(es), func(i int) bool { return idx.prefixMaxX[i] >= b.MinX })
	var out []int
	for i := lo; i < len(es); i++ {
		e := &es[i]
		if e.minX > b.MaxX {
			break
		}
		if e.maxX >= b.MinX && e.minY <= b.MaxY && e.maxY >= b.MinY {
			out = append(out, e.id)
		}
	}
	sort.Ints(out)
	return out
}
