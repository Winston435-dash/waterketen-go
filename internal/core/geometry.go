package core

// Geometry helpers: iteration, bounding boxes, bbox intersection tests,
// distance from a point to a geometry, polygon-outlining, linear
// interpolation along a line. This replaces the shapely surface the Python
// atlas uses (sindex.query, boundary, representative_point, interpolate,
// distance). It does NOT grow beyond what the atlas needs.

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
	return nil
}

// Bounds returns the geometry's bounding box.
func (g Geometry) Bounds() (BBox, bool) {
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
		if g.Kind == "point" {
			return g.Point
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

// RepresentativePoint is the centroid of the geometry's vertices — cheap and
// good enough as a label fallback (shapely's representative_point()).
func (g Geometry) RepresentativePoint() Point {
	var sx, sy float64
	var n int
	for _, line := range g.Outlines() {
		for _, p := range line {
			sx, sy, n = sx+p.X, sy+p.Y, n+1
		}
	}
	if n == 0 {
		if g.Kind == "point" {
			return g.Point
		}
		return Point{math.NaN(), math.NaN(), math.NaN()}
	}
	return Point{sx / float64(n), sy / float64(n), math.NaN()}
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
	entries []giEntry // sorted by minX
}

// giEntry is one feature's bbox. float32 keeps the whole index 24 bytes per
// feature so the scan window fits a few cache lines; RD coordinates in
// metres need ~1 mm precision, far inside float32's ~7 digits.
type giEntry struct {
	minX, minY, maxX, maxY float32
	id                     int32
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
		entries = append(entries, giEntry{
			float32(b.MinX), float32(b.MinY), float32(b.MaxX), float32(b.MaxY), int32(i),
		})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].minX < entries[j].minX })
	return GridIndex{entries: entries}
}

// Query returns the ascending indices of features whose bbox intersects `b`.
// Binary-search the first entry that can intersect on x, then scan until an
// entry starts past the query. NOT safe for concurrent use — scenes own their
// index and the render/nearest paths are single-goroutine.
func (idx *GridIndex) Query(b BBox) []int {
	es := idx.entries
	minX, maxX := float32(b.MinX), float32(b.MaxX)
	lo := sort.Search(len(es), func(i int) bool { return es[i].maxX >= minX })
	var out []int
	minY, maxY := float32(b.MinY), float32(b.MaxY)
	for i := lo; i < len(es); i++ {
		e := &es[i]
		if e.minX > maxX {
			break // sorted by minX: nothing further can match
		}
		if e.minY <= maxY && e.maxY >= minY {
			out = append(out, int(e.id))
		}
	}
	sort.Ints(out)
	return out
}
