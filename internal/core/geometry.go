package core

// Geometry helpers: iteration, bounding boxes, bbox intersection tests,
// distance from a point to a geometry, polygon-outlining, linear
// interpolation along a line. This replaces the shapely surface the Python
// atlas uses (sindex.query, boundary, representative_point, interpolate,
// distance). It does NOT grow beyond what the atlas needs.

import "math"

// BBox is a (minx, miny, maxx, maxy) rectangle in EPSG:28992.
type BBox struct{ MinX, MinY, MaxX, MaxY float64 }

func (b BBox) intersects(o BBox) bool {
	return b.MinX <= o.MaxX && b.MaxX >= o.MinX && b.MinY <= o.MaxY && b.MaxY >= o.MinY
}

// Grow expands the box by r on all sides.
func (b BBox) Grow(r float64) BBox {
	return BBox{b.MinX - r, b.MinY - r, b.MaxX + r, b.MaxY + r}
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
// The Python atlas uses a STRtree via `gdf.sindex.query(box)`. Our feature
// counts per load are 2-60k, which makes a grid-index over cell buckets the
// right call: cheap to build, one-shot per scene, fast enough per query.

const gridIndexCell = 64 // metres per bucket

// GridIndex buckets features by the cells their bbox touches.
type GridIndex struct {
	cell    float64
	buckets map[[2]int][]int
	bounds  []BBox
}

// NewGridIndex builds a bucket index over the features' bounding boxes.
func NewGridIndex(geoms []Geometry) GridIndex {
	idx := GridIndex{cell: gridIndexCell, buckets: map[[2]int][]int{}, bounds: make([]BBox, len(geoms))}
	for i, g := range geoms {
		b, ok := g.Bounds()
		if !ok {
			continue
		}
		idx.bounds[i] = b
		x0, y0 := int(math.Floor(b.MinX/idx.cell)), int(math.Floor(b.MinY/idx.cell))
		x1, y1 := int(math.Floor(b.MaxX/idx.cell)), int(math.Floor(b.MaxY/idx.cell))
		for x := x0; x <= x1; x++ {
			for y := y0; y <= y1; y++ {
				key := [2]int{x, y}
				idx.buckets[key] = append(idx.buckets[key], i)
			}
		}
	}
	return idx
}

// Query returns the indices of features whose bbox intersects `b`, deduped
// and ascending (the Python version returns sindex order; we guarantee a
// stable order which the renderers already rely on).
func (idx GridIndex) Query(b BBox) []int {
	seen := map[int]struct{}{}
	var out []int
	x0, y0 := int(math.Floor(b.MinX/idx.cell)), int(math.Floor(b.MinY/idx.cell))
	x1, y1 := int(math.Floor(b.MaxX/idx.cell)), int(math.Floor(b.MaxY/idx.cell))
	for x := x0; x <= x1; x++ {
		for y := y0; y <= y1; y++ {
			for _, i := range idx.buckets[[2]int{x, y}] {
				if _, dup := seen[i]; dup {
					continue
				}
				seen[i] = struct{}{}
				if idx.bounds[i].intersects(b) {
					out = append(out, i)
				}
			}
		}
	}
	// insertion order is grid-scan order; sort for determinism
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j-1] > out[j]; j-- {
			out[j-1], out[j] = out[j], out[j-1]
		}
	}
	return out
}
