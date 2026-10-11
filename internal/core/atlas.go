package core

import (
	"fmt"
	"math"
	"slices"
	"sort"
	"strings"
)

// --- Colours, modes, the layer vocabulary -----------------------------------
//
// These are the DATA colours (what a riool line is), NOT the UI tokens. The
// ui package owns the chrome; this table owns the map. The mapping is the one
// the Python atlas established and the TUI screenshots already taught.

var layerColours = map[string]string{
	"riool":               "#38d9ff",
	"drinkwaterleidingen": "#ffa95c",
	"drinkwater":          "#476c91",
	"straten":             "#627d91",
	"water":               "#416e91",
	"spoor":               "#b19bbf",
	"putten":              "#38d9ff",
	"gemalen":             "#cc94ff",
	"groen":               "#5a8f6f",
	"gebouwen":            "#4a5a68",
}

// Networks are drawn even in overlay mode regardless of the active layer.
var networks = []string{"drinkwaterleidingen", "riool"}

// Modes is the colour-mode cycle for the atlas: netwerk → bob → materiaal.
var Modes = []string{"netwerk", "bob", "materiaal"}

// materialPalette cycles over distinct materials in MATERIAAL1 order.
var materialPalette = []string{
	"#38d9ff", "#ffa95c", "#cc94ff", "#8ee6a8", "#ff7096", "#e9d778",
	"#73a7ff", "#dca775", "#bde693", "#ed8be4", "#9ddeda", "#ff765e",
}

const (
	gridColour   = "#172b3e" // the reference RD grid
	highlightInk = "#ffffff"
	labelInk     = "#b9d5e7"
	northInk     = "#a8c4df"
	scaleInk     = "#b3d3e9"
	// materialInk is the fallback ink for a material legend swatch whose colour
	// core cannot resolve without the scene's material map.
	materialInk = "#9ddeda"
)

// bobBands are the sewer-invert depth bands the "bob" colour mode draws, in
// metres NAP. The keys double as the CellStyle suffix ("bob:deep"), so the
// renderer and CellColour share one table instead of repeating the palette.
var bobBands = map[string]string{
	"very-deep": "#b28cff",
	"deep":      "#589dff",
	"mid":       "#38d9ff",
	"shallow":   "#8ee6a8",
	"above":     "#ffcc73",
	"unknown":   "#73879d",
}

// bobBandOrder maps a BOB threshold to a band: a pipe shallower than `below`
// takes `band`. Ordered deepest-first, matching featureColour.
var bobBandOrder = []struct {
	below float64
	band  string
}{
	{-4, "very-deep"},
	{-2, "deep"},
	{-1, "mid"},
	{0, "shallow"},
}

// --- Scene: data the atlas draws --------------------------------------------

// Feature is one drawable element: a geometry plus display attributes.
type Feature struct {
	Layer      string
	Position   int
	Geometry   Geometry
	Properties map[string]string
	Street     string
	BOB        float64 // NaN = unknown
}

// DrawOrder is the painting order for the library map, matching the Python
// kaart.DRAW_ORDER: background-ish layers first, networks on top, assets last.
// It is also the order the layer picker lists them in, so the list and the map
// agree about what is "under" what.
var DrawOrder = []string{
	"groen", "water", "gebouwen", "spoor", "straten", "drinkwater",
	"riool", "drinkwaterleidingen", "putten", "gemalen",
}

// OrderLayers filters `names` to the known layers and sorts them into
// DrawOrder. Unknown names are dropped rather than drawn blind.
func OrderLayers(names []string) []string {
	in := map[string]bool{}
	for _, n := range names {
		in[n] = true
	}
	var out []string
	for _, n := range DrawOrder {
		if in[n] {
			out = append(out, n)
		}
	}
	return out
}

// StreetRef points at ONE feature that carries a street name. The layer is part
// of the reference because indices are per-layer: reusing a riool index against
// the straten slice would frame a random street, and no test would notice.
type StreetRef struct {
	Layer string
	Index int
}

// Scene is a fully loaded multi-layer picture: features per layer, a spatial
// index per layer, streets by name, the material→colour legend.
type Scene struct {
	Area      string
	Bounds    BBox
	Layers    map[string][]Feature
	Index     map[string]GridIndex
	Streets   map[string][]StreetRef // street name -> the features that carry it
	Materials map[string]string
	Warnings  []string
	// Display pins the visible layers explicitly (the library panel's chosen
	// set). nil means "derive from overlay+active", the managed-asset default.
	Display []string
}

// Frame is the current view rect at a given zoom, plus helpers to move
// between world (RD) and cell coordinates.
type Frame struct {
	Bounds        BBox
	Width, Height int
	MetresPerDot  float64
}

// Project maps a world point to float cell coordinates.
func (f Frame) Project(x, y float64) (float64, float64) {
	return (x - f.Bounds.MinX) / f.MetresPerDot,
		(f.Bounds.MaxY - y) / f.MetresPerDot
}

// Cell maps a world point to a terminal CELL (a braille cell, not a dot).
func (f Frame) Cell(x, y float64) (int, int) {
	px, py := f.Project(x, y)
	return int(math.Round(px / 2)), int(math.Round(py / 4))
}

// World maps a cell to the world point at the centre of its 2x4 dot block.
func (f Frame) World(col, row float64) (float64, float64) {
	return f.Bounds.MinX + (col*2+.5)*f.MetresPerDot,
		f.Bounds.MaxY - (row*4+1.5)*f.MetresPerDot
}

// View is the navigational state: centre + zoom. zoom=1 means "fit the scene".
type View struct {
	CX, CY float64
	Zoom   float64
}

// FitView returns the view that exactly fits the bounds.
func FitView(b BBox) View {
	return View{CX: (b.MinX + b.MaxX) / 2, CY: (b.MinY + b.MaxY) / 2, Zoom: 1}
}

// Frame materialises the view at a given pixel budget.
func (v View) Frame(scene BBox, width, height int) Frame {
	dw := math.Max(float64(width*2-1), 1)
	dh := math.Max(float64(height*4-1), 1)
	metres := math.Max(math.Max(
		(scene.MaxX-scene.MinX)/dw,
		(scene.MaxY-scene.MinY)/dh,
	), .01) * 1.08 / v.Zoom
	sx, sy := dw*metres/2, dh*metres/2
	return Frame{
		Bounds:       BBox{v.CX - sx, v.CY - sy, v.CX + sx, v.CY + sy},
		Width:        width,
		Height:       height,
		MetresPerDot: metres,
	}
}

// ZoomAt changes the zoom while keeping the world point under (col, row) fixed.
func (v *View) ZoomAt(factor, col, row float64, scene BBox, width, height int) {
	bx, by := v.Frame(scene, width, height).World(col, row)
	v.Zoom = math.Max(.5, math.Min(1024, v.Zoom*factor))
	ax, ay := v.Frame(scene, width, height).World(col, row)
	v.CX += bx - ax
	v.CY += by - ay
}

// Pan moves the view by a fraction of the visible span.
func (v *View) Pan(east, north float64, scene BBox, width, height int) {
	f := v.Frame(scene, width, height)
	v.CX += east * (f.Bounds.MaxX - f.Bounds.MinX) * .18
	v.CY += north * (f.Bounds.MaxY - f.Bounds.MinY) * .18
}

// --- Scene construction -----------------------------------------------------
//
// Load a scene from open GeoPackage handles. The Python side used geopandas;
// we use the gpkg.go reader plus the geometry.go index. One damaged layer is
// a warning, not a blank map.

// SceneSource is one layer's features.
type SceneSource struct {
	Name     string
	Features []Feature
}

// NewScene builds a Scene from pre-read layers.
func NewScene(area string, layers []SceneSource, warnings []string) (*Scene, error) {
	scene := &Scene{
		Area:      area,
		Layers:    map[string][]Feature{},
		Index:     map[string]GridIndex{},
		Streets:   map[string][]StreetRef{},
		Materials: map[string]string{},
		Warnings:  warnings,
	}
	var fitBBox BBox
	started := false

	materialNames := map[string]struct{}{}
	for _, layer := range layers {
		if len(layer.Features) == 0 {
			continue
		}
		geoms := make([]Geometry, len(layer.Features))
		for i, f := range layer.Features {
			geoms[i] = f.Geometry
			if f.Street != "" {
				scene.Streets[f.Street] = append(scene.Streets[f.Street], StreetRef{layer.Name, i})
			}
			if m := f.Properties["MATERIAAL1"]; m != "" {
				materialNames[m] = struct{}{}
			}
		}
		scene.Layers[layer.Name] = layer.Features
		scene.Index[layer.Name] = NewGridIndex(geoms)
		// Only networks set the fit bounds: a country-wide service-area polygon
		// must never determine the network zoom.
		if slices.Contains(networks, layer.Name) {
			for _, g := range geoms {
				if b, ok := g.Bounds(); ok {
					fitBBox, started = mergeBBox(fitBBox, started, b)
				}
			}
		}
	}
	// No network: fall back to the union of everything.
	if !started {
		for _, layer := range layers {
			for _, f := range layer.Features {
				if b, ok := f.Geometry.Bounds(); ok {
					fitBBox, started = mergeBBox(fitBBox, started, b)
				}
			}
		}
	}
	if !started {
		return nil, fmt.Errorf("geen geometrie beschikbaar voor de atlas")
	}
	scene.Bounds = fitBBox

	names := make([]string, 0, len(materialNames))
	for m := range materialNames {
		names = append(names, m)
	}
	sort.Strings(names)
	for i, m := range names {
		scene.Materials[m] = materialPalette[i%len(materialPalette)]
	}
	return scene, nil
}

// VisibleLayers answers the same question as the Python atlas: overlay draws
// the networks plus (if active is not a network) the active layer.
func (s *Scene) VisibleLayers(active string, overlay bool, display []string) []string {
	if display == nil {
		display = s.Display
	}
	if display != nil {
		var out []string
		for _, n := range display {
			if _, ok := s.Layers[n]; ok {
				out = append(out, n)
			}
		}
		return out
	}
	if !overlay {
		if _, ok := s.Layers[active]; ok {
			return []string{active}
		}
		return nil
	}
	nets := []string{}
	for _, n := range networks {
		if _, ok := s.Layers[n]; ok {
			nets = append(nets, n)
		}
	}
	if _, ok := s.Layers[active]; ok && !slices.Contains(nets, active) {
		return append([]string{active}, nets...)
	}
	if len(nets) > 0 {
		return nets
	}
	out := []string{}
	for n := range s.Layers {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

// Candidates is the sindex.query equivalent: features of `layer` whose bbox
// intersects `bounds`.
func (s *Scene) Candidates(layer string, bounds BBox) []Feature {
	idx, ok := s.Index[layer]
	if !ok {
		return nil
	}
	ids := idx.Query(bounds)
	feats := s.Layers[layer]
	out := make([]Feature, len(ids))
	for i, id := range ids {
		out[i] = feats[id]
	}
	return out
}

// MatchStreetQuery tests one street name against a search query. It reports whether
// the name matches, and whether the match is a PREFIX (which ranks first).
//
// A name matches on the raw substring OR on the space-collapsed form. People type
// "west zeedijk" for "Westzeedijk" and "s gravendijkwal" for "'s-Gravendijkwal", so a
// search that answers "geen straat gevonden" to the obvious spelling is broken — even
// once the space key itself works. Collapsing only ever ADDS matches, so it cannot
// hide a street that the raw query already found.
func MatchStreetQuery(query, name string) (matched, prefix bool) {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return false, false
	}
	lower := strings.ToLower(name)
	cq, cn := collapseSpaces(q), collapseSpaces(lower)
	switch {
	case strings.Contains(lower, q):
		return true, strings.HasPrefix(lower, q)
	case strings.Contains(cn, cq):
		return true, strings.HasPrefix(cn, cq)
	}
	return false, false
}

// collapseSpaces drops every space and separator-ish character so "west zeedijk"
// and "westzeedijk" compare equal.
func collapseSpaces(s string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case ' ', '	', '-', '\'':
			return -1
		}
		return r
	}, s)
}

// Search returns up to 8 street names matching the query (case-insensitive,
// prefix-first), or the first 8 alphabetically when the query is empty.
func (s *Scene) Search(query string) []string {
	names := make([]string, 0, len(s.Streets))
	for n := range s.Streets {
		names = append(names, n)
	}
	sort.Strings(names)
	if strings.TrimSpace(query) == "" {
		if len(names) > 8 {
			names = names[:8]
		}
		return names
	}
	type match struct {
		name   string
		prefix bool
	}
	var hits []match
	for _, n := range names {
		if ok, prefix := MatchStreetQuery(query, n); ok {
			hits = append(hits, match{n, prefix})
		}
	}
	sort.SliceStable(hits, func(i, j int) bool {
		if hits[i].prefix != hits[j].prefix {
			return hits[i].prefix
		}
		return strings.ToLower(hits[i].name) < strings.ToLower(hits[j].name)
	})
	out := make([]string, 0, min(8, len(hits)))
	for i, h := range hits {
		if i >= 8 {
			break
		}
		out = append(out, h.name)
	}
	return out
}

// Nearest returns the closest feature to world point (x, y) within `radius`,
// over the layers visible given (active, overlay). Polygons measure to their
// outline. nil when nothing is close.
func (s *Scene) Nearest(x, y, radius float64, active string, overlay bool, display []string) *Feature {
	p := Point{X: x, Y: y, Z: math.NaN()}
	var winner *Feature
	best := radius
	for _, layer := range s.VisibleLayers(active, overlay, display) {
		for _, f := range s.Candidates(layer, BBox{x - radius, y - radius, x + radius, y + radius}) {
			d := f.Geometry.Distance(p)
			if d <= best {
				cpy := f
				winner = &cpy
				best = d
			}
		}
	}
	return winner
}

// StreetView focuses a view on all features of one street, with a 40 m margin.
// References carry their own layer, so a street known from riool and from straten
// contributes both geometries instead of one layer's indices being read against
// the other's slice.
func (s *Scene) StreetView(name string, width, height int) (View, bool) {
	refs, ok := s.Streets[name]
	if !ok {
		return View{}, false
	}
	var b BBox
	started := false
	for _, ref := range refs {
		feats := s.Layers[ref.Layer]
		if ref.Index < 0 || ref.Index >= len(feats) {
			continue
		}
		gb, ok := feats[ref.Index].Geometry.Bounds()
		if !ok {
			continue
		}
		b, started = mergeBBox(b, started, gb)
	}
	if !started {
		return View{}, false
	}
	target := b.Grow(40)
	base := FitView(s.Bounds).Frame(s.Bounds, width, height).MetresPerDot
	needed := FitView(target).Frame(target, width, height).MetresPerDot
	return View{
		CX:   (b.MinX + b.MaxX) / 2,
		CY:   (b.MinY + b.MaxY) / 2,
		Zoom: math.Max(1, math.Min(1024, base/needed)),
	}, true
}

// --- Rendering --------------------------------------------------------------

// Render is the output of one draw: a grid of (rune, styleKey) cells plus the
// legend and per-layer counts. styleKey indexes the StyleTable the caller
// resolves to actual lipgloss styles — keeping core free of lipgloss.
type Render struct {
	Cells  [][]Cell
	Frame  Frame
	Counts map[string]int
	Legend []LegendEntry
}

// Cell is one terminal cell on the map.
type Cell struct {
	Rune  rune
	Style CellStyle
}

// CellStyle is a semantic style reference, not a colour. Renderers pick one;
// the screen translates via a fixed table (the only place that may import
// lipgloss for the atlas).
type CellStyle string

const (
	StyleBlank     CellStyle = "bg"
	StyleGrid      CellStyle = "grid"
	StyleNorth     CellStyle = "north"
	StyleScale     CellStyle = "scale"
	StyleLabel     CellStyle = "label"
	StyleHighlight CellStyle = "highlight"
	// StyleLayerPrefix prefixes a per-layer style: "layer:riool".
	StyleLayerPrefix CellStyle = "layer:"
)

// StyleForLayer returns the style key for a layer drawn in the default mode.
func StyleForLayer(name string) CellStyle { return StyleLayerPrefix + CellStyle(name) }

// CellColour resolves a CellStyle to the concrete hex ink the map draws. This
// is the single owner of data ink: the screen turns the hex into a terminal
// colour and keeps no colour table of its own. Chrome styles are resolved here
// too (grid, label, ...); the UI layer adds only bold/underline on top.
func CellColour(s CellStyle) string {
	str := string(s)
	switch s {
	case StyleBlank, StyleGrid:
		return gridColour
	case StyleNorth:
		return northInk
	case StyleScale:
		return scaleInk
	case StyleLabel:
		return labelInk
	case StyleHighlight:
		return highlightInk
	}
	if name, ok := strings.CutPrefix(str, string(StyleLayerPrefix)); ok {
		if c, ok := layerColours[name]; ok {
			return c
		}
	}
	if band, ok := strings.CutPrefix(str, "bob:"); ok {
		if c, ok := bobBands[band]; ok {
			return c
		}
	}
	if hex, ok := strings.CutPrefix(str, "hex:"); ok {
		return hex
	}
	if strings.HasPrefix(str, "material:") {
		// core cannot resolve a material swatch without the scene's map; the TUI
		// legend therefore draws every material in one ink (a known limitation).
		return materialInk
	}
	return layerColours["riool"]
}

// LegendEntry is one legend item: label, data ink, and the map mark it explains.
type LegendEntry struct {
	Label string
	Style CellStyle
	// Symbol is a short terminal sample, such as a line, area block, or point dot.
	Symbol string
}

// colourGrid keeps a colour per CELL (not per dot) — collisions resolve by
// priority.
type colourGrid struct {
	grid       DotGrid
	styles     [][]string
	priorities [][]int
	style      string
	priority   int
}

func newColourGrid(w, h int) colourGrid {
	styles := make([][]string, h)
	prios := make([][]int, h)
	for i := range styles {
		styles[i] = make([]string, w)
		prios[i] = make([]int, w)
		for x := range styles[i] {
			styles[i][x] = gridColour
			prios[i][x] = -1
		}
	}
	return colourGrid{grid: NewDotGrid(w, h), styles: styles, priorities: prios}
}

func (g *colourGrid) setDot(x, y int) {
	g.grid.SetDot(x, y)
	if x < 0 || y < 0 || x >= g.grid.DotWidth || y >= g.grid.DotHeight {
		return
	}
	cx, cy := x/2, y/4
	if g.priority >= g.priorities[cy][cx] {
		g.styles[cy][cx] = g.style
		g.priorities[cy][cx] = g.priority
	}
}

func (g *colourGrid) line(ax, ay, bx, by float64) {
	lineInt(int(math.Round(ax)), int(math.Round(ay)), int(math.Round(bx)), int(math.Round(by)), g.setDot)
}

// putText writes a string into a (rune, style) grid, clipped to the width.
func putText(rows [][]Cell, col, row int, text string, style string) {
	if row < 0 || row >= len(rows) {
		return
	}
	c := col
	for _, r := range text {
		if c >= 0 && c < len(rows[row]) {
			rows[row][c] = Cell{Rune: r, Style: CellStyle(style)}
		}
		c++
	}
}

// niceDistance picks a "round" scale distance ≤ v: 1/2/5 × 10^n.
func niceDistance(v float64) float64 {
	v = math.Max(v, 1e-12)
	base := math.Pow(10, math.Floor(math.Log10(v)))
	for _, n := range []float64{10, 5, 2, 1} {
		if n*base <= v {
			return n * base
		}
	}
	return base
}

func distanceLabel(m float64) string {
	if m >= 1000 {
		return fmt.Sprintf("%g km", m/1000)
	}
	return fmt.Sprintf("%g m", m)
}
