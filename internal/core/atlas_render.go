package core

// Atlas rendering: the actual draw. Lives in its own file next to atlas.go
// (which holds the data model and helpers) so the render function stays
// readable — it is the longest single function in the core.

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

// lowPriorityLayers draw first and dimmest: their ink should never outshine the
// network above it. Built once, not per frame.
var lowPriority = map[string]bool{
	"drinkwater": true, "water": true, "straten": true,
	"groen": true, "gebouwen": true, "spoor": true,
}

// RenderScene draws the scene into a width×height cell raster. This is the
// full port of atlas.render_atlas: RD background grid, candidate selection,
// per-layer colouring, highlight priority, the overview minimap with the
// viewport rectangle, street labels with collision avoidance, and the legend.
//
// displayLayers == nil mirrors the Python `display_layers=None` (derive from
// overlay+active); overview controls whether the minimap is drawn.
func RenderScene(
	scene *Scene,
	view View,
	width, height int,
	active string,
	overlay bool,
	mode string,
	labels bool,
	overview bool,
	selected *Feature,
	street string,
	displayLayers []string,
	decorations bool,
) Render {
	width = max(1, width)
	height = max(1, height)
	frame := view.Frame(scene.Bounds, width, height)
	a, b, c, d := frame.Bounds.MinX, frame.Bounds.MinY, frame.Bounds.MaxX, frame.Bounds.MaxY

	grid := newColourGrid(width, height)

	// Background RD reference grid. Restrained: a navigational reference, not
	// fictitious streets.
	step := niceDistance((c - a) / 5)
	grid.style = gridColour
	grid.priority = 0
	for x := math.Floor(a / step); x <= math.Floor(c/step); x++ {
		px, _ := frame.Project(x*step, d)
		for py := 0; py < grid.grid.DotHeight; py += 8 {
			grid.setDot(int(math.Round(px)), py)
		}
	}
	for y := math.Floor(b / step); y <= math.Floor(d/step); y++ {
		_, py := frame.Project(a, y*step)
		for px := 0; px < grid.grid.DotWidth; px += 6 {
			grid.setDot(px, int(math.Round(py)))
		}
	}

	// Per-layer draw — ordinary features first, highlights last.
	counts := map[string]int{}
	labelFeatures := map[string][]Geometry{}
	isSelected := func(f Feature) bool {
		return selected != nil && f.Layer == selected.Layer && f.Position == selected.Position
	}
	isHighlighted := func(f Feature) bool {
		return isSelected(f) || (street != "" && f.Street == street)
	}
	for _, layer := range scene.VisibleLayers(active, overlay, displayLayers) {
		idx, ok := scene.Index[layer]
		if !ok {
			continue
		}
		ids := idx.Query(frame.Bounds)
		features := scene.Layers[layer]
		counts[layer] = len(ids)

		drawFeature := func(f Feature) {
			highlight := isHighlighted(f)
			if highlight {
				grid.style = "#ffffff:#284a63" // "fg:bg" — resolved by the caller to CellStyle
				grid.priority = 5
			} else {
				grid.style = featureColour(f, mode, scene)
				if lowPriority[layer] {
					grid.priority = 1
				} else {
					grid.priority = 2
				}
			}
			geom := f.Geometry
			if f.Layer == "gemalen" {
				// Pump constructions are map markers, not building footprints.
				p := geom.RepresentativePoint()
				if !math.IsNaN(p.X) && !math.IsNaN(p.Y) && !math.IsInf(p.X, 0) && !math.IsInf(p.Y, 0) {
					px, py := frame.Project(p.X, p.Y)
					grid.setDot(int(math.Round(px)), int(math.Round(py)))
				}
			} else if pts := geom.Dots(); len(pts) > 0 {
				for _, p := range pts {
					px, py := frame.Project(p.X, p.Y)
					grid.setDot(int(math.Round(px)), int(math.Round(py)))
				}
			} else {
				// Clip lazily per segment: the grid drops out-of-range dots, so
				// a geometry intersection would only add work for ordinary lines.
				for _, line := range geom.Outlines() {
					if len(line) < 2 {
						continue
					}
					ax, ay := frame.Project(line[0].X, line[0].Y)
					for _, point := range line[1:] {
						bx, by := frame.Project(point.X, point.Y)
						// Fast-reject segments fully outside one axis.
						outside := (ax < -8 && bx < -8) ||
							(ax > float64(grid.grid.DotWidth+8) && bx > float64(grid.grid.DotWidth+8)) ||
							(ay < -8 && by < -8) ||
							(ay > float64(grid.grid.DotHeight+8) && by > float64(grid.grid.DotHeight+8))
						if !outside {
							grid.line(ax, ay, bx, by)
						}
						ax, ay = bx, by
					}
				}
			}
			if f.Street != "" && (layer == "riool" || layer == "straten") {
				labelFeatures[f.Street] = append(labelFeatures[f.Street], geom)
			}
		}

		// Query returns ascending feature IDs. Two passes reproduce the stable
		// normal-then-highlight ordering without copying and sorting []Feature.
		for _, id := range ids {
			f := features[id]
			if !isHighlighted(f) {
				drawFeature(f)
			}
		}
		for _, id := range ids {
			f := features[id]
			if isHighlighted(f) {
				drawFeature(f)
			}
		}
	}

	// Materialise the grid to cells.
	rows := make([][]Cell, height)
	for y := 0; y < height; y++ {
		rows[y] = make([]Cell, width)
	}
	texts := grid.grid.Text()
	for y := 0; y < height && y < len(texts); y++ {
		runes := []rune(texts[y])
		for x := 0; x < width && x < len(runes); x++ {
			r := runes[x]
			style := gridStyleToCell(grid.styles[y][x])
			rows[y][x] = Cell{Rune: r, Style: style}
		}
	}

	// Overview minimap + viewport rectangle.
	if overview && width >= 75 && height >= 22 {
		mw, mh := 20, 9
		mini := RenderScene(scene, FitView(scene.Bounds), mw, mh,
			active, overlay, mode, false, false, nil, "", displayLayers, false)
		mx, my := width-mw-3, 2
		for ry := 0; ry < mh; ry++ {
			for rx := 0; rx < mw; rx++ {
				if my+ry < height && mx+rx < width && my+ry >= 0 && mx+rx >= 0 {
					rows[my+ry][mx+rx] = mini.Cells[ry][rx]
				}
			}
		}
		// Viewport rectangle on the real network overview.
		p1x, p1y := mini.Frame.Project(a, d)
		p2x, p2y := mini.Frame.Project(c, b)
		l, r := int(math.Round(p1x/2)), int(math.Round(p2x/2))
		t, bt := int(math.Round(p1y/4)), int(math.Round(p2y/4))
		if l > r {
			l, r = r, l
		}
		if t > bt {
			t, bt = bt, t
		}
		l = max(0, min(mw-1, l))
		r = max(0, min(mw-1, r))
		t = max(0, min(mh-1, t))
		bt = max(0, min(mh-1, bt))
		if l <= r && t <= bt {
			for x := l; x <= r; x++ {
				putText(rows, mx+x, my+t, "─", string(StyleScale))
				putText(rows, mx+x, my+bt, "─", string(StyleScale))
			}
			for y := t; y <= bt; y++ {
				putText(rows, mx+l, my+y, "│", string(StyleScale))
				putText(rows, mx+r, my+y, "│", string(StyleScale))
			}
		}
		putText(rows, mx, my-1, " OVERZICHT ", string(StyleLabel))
	}

	// Street labels, collision-avoided. Reserve HUD, scale bar and minimap first.
	reserved := [][4]int{{0, 0, 12, 2}, {0, height - 2, min(width, 36), height}}
	if overview && width >= 75 && height >= 22 {
		mw, mh := 20, 9
		mx, my := width-mw-3, 2
		reserved = append(reserved, [4]int{mx - 1, my - 1, width, my + mh + 2})
	}
	if labels {
		type label struct {
			name  string
			total float64
		}
		var order []label
		for name, geoms := range labelFeatures {
			sum := 0.0
			for _, g := range geoms {
				sum += g.Length()
			}
			order = append(order, label{name, sum})
		}
		sort.SliceStable(order, func(i, j int) bool {
			if (order[i].name == street) != (order[j].name == street) {
				return order[i].name == street
			}
			if order[i].total != order[j].total {
				return order[i].total > order[j].total
			}
			return order[i].name < order[j].name
		})
		for _, lbl := range order {
			geoms := labelFeatures[lbl.name]
			longest := geoms[0]
			best := longest.Length()
			for _, g := range geoms[1:] {
				if n := g.Length(); n > best {
					longest, best = g, n
				}
			}
			var p Point
			if best > 0 {
				p = longest.Interpolate(0.5)
			} else {
				p = longest.RepresentativePoint()
			}
			px, py := frame.Project(p.X, p.Y)
			text := fmt.Sprintf(" %s ", lbl.name)
			if len([]rune(text)) > 32 {
				text = string([]rune(text)[:32])
			}
			col := int(math.Round(px/2)) - len([]rune(text))/2
			col = max(0, min(width-len([]rune(text)), col))
			row := int(math.Round(py / 4))
			row = max(2, min(height-3, row))
			rect := [4]int{col - 1, row - 1, col + len([]rune(text)) + 1, row + 2}
			if intersectsAny(rect, reserved) {
				continue
			}
			putText(rows, col, row, text, string(StyleLabel))
			reserved = append(reserved, rect)
			if len(reserved) > 20 {
				break
			}
		}
	}

	if decorations {
		putText(rows, 1, 0, " ↑ N · RD ", string(StyleNorth))
		metres := niceDistance(math.Max(frame.MetresPerDot*float64(min(24, width/3))*2, .001))
		cells := max(1, int(math.Round(metres/frame.MetresPerDot/2)))
		bar := "├" + strings.Repeat("─", max(1, cells-2)) + "┤ " + distanceLabel(metres)
		putText(rows, 1, height-1, bar, string(StyleScale))
	}

	// Legend: same branch structure as the Python atlas.
	var legend []LegendEntry
	if mode == "bob" && counts["riool"] > 0 {
		legend = []LegendEntry{
			valueLegendEntry("BOB < −4", "bob:very-deep"),
			valueLegendEntry("−4..−2", "bob:deep"),
			valueLegendEntry("−2..−1", "bob:mid"),
			valueLegendEntry("−1..0", "bob:shallow"),
			valueLegendEntry("≥ 0", "bob:above"),
			valueLegendEntry("onbekend", "bob:unknown"),
		}
		if counts["drinkwaterleidingen"] > 0 {
			legend = append(legend, LegendEntry{Label: "drinkwater", Style: StyleForLayer("drinkwaterleidingen"), Symbol: "━━"})
		}
	} else if mode == "materiaal" && counts["riool"] > 0 {
		names := make([]string, 0, len(scene.Materials))
		for m := range scene.Materials {
			names = append(names, m)
		}
		sort.Strings(names)
		for _, m := range names {
			legend = append(legend, valueLegendEntry(m, CellStyle("material:"+m)))
		}
		legend = append(legend, valueLegendEntry("onbekend", "material:unknown"))
		if counts["drinkwaterleidingen"] > 0 {
			legend = append(legend, LegendEntry{Label: "drinkwater", Style: StyleForLayer("drinkwaterleidingen"), Symbol: "━━"})
		}
	} else {
		names := make([]string, 0, len(counts))
		for n := range counts {
			names = append(names, n)
		}
		sort.Strings(names)
		for _, n := range names {
			if n == "drinkwater" {
				continue
			}
			legend = append(legend, layerLegendEntry(n))
		}
		if counts["drinkwater"] > 0 {
			legend = append(legend, LegendEntry{
				Label: "verzorgingsgrens (geen buizen)", Style: StyleForLayer("drinkwater"), Symbol: "■",
			})
		}
	}

	return Render{Cells: rows, Frame: frame, Counts: counts, Legend: legend}
}

func layerLegendEntry(name string) LegendEntry {
	symbol := "━━"
	switch name {
	case "gemalen", "putten":
		symbol = "•"
	case "drinkwater", "water", "groen", "gebouwen":
		symbol = "■"
	}
	return LegendEntry{Label: name, Style: StyleForLayer(name), Symbol: symbol}
}

func valueLegendEntry(label string, style CellStyle) LegendEntry {
	return LegendEntry{Label: label, Style: style, Symbol: "━━"}
}

func gridStyleToCell(style string) CellStyle {
	if style == "" {
		return StyleBlank
	}
	if style == gridColour {
		return StyleGrid
	}
	// highlight sentinel from render loop
	if strings.HasPrefix(style, "#ffffff:") {
		return StyleHighlight
	}
	// otherwise the style is a hex colour: reverse-map through the tables
	return ColourToStyle(style)
}

// ColourToStyle reverse-resolves a hex colour to its CellStyle. The renderer
// deals in colours because the colour decision is geometry logic; this helper
// exists once so the screen never maps hexes itself.
func ColourToStyle(hex string) CellStyle {
	switch hex {
	case bobBands["very-deep"]:
		return "bob:very-deep"
	case bobBands["deep"]:
		return "bob:deep"
	case bobBands["mid"]:
		// riool AND bob mid AND putten share this colour — resolve by priority:
		// this hex as a CELL colour means "layer ink".
		return StyleForLayer("riool")
	case bobBands["shallow"]:
		return "bob:shallow"
	case bobBands["above"]:
		return "bob:above"
	case bobBands["unknown"]:
		return "bob:unknown"
	}
	for name, c := range layerColours {
		if c == hex {
			return StyleForLayer(name)
		}
	}
	// Material palette colours are assigned deterministically; the screen can't
	// reverse them without the scene's map. Keep the raw hex as the style key —
	// the screen's style table accepts "material:#hex" and "hex:#hex" forms.
	return CellStyle("hex:" + hex)
}

func featureColour(f Feature, mode string, scene *Scene) string {
	if mode == "bob" && f.Layer == "riool" {
		if math.IsNaN(f.BOB) {
			return bobBands["unknown"]
		}
		for _, branch := range bobBandOrder {
			if f.BOB < branch.below {
				return bobBands[branch.band]
			}
		}
		return bobBands["above"]
	}
	if mode == "materiaal" && f.Layer == "riool" {
		if c, ok := scene.Materials[f.Properties["MATERIAAL1"]]; ok {
			return c
		}
		return bobBands["unknown"]
	}
	if c, ok := layerColours[f.Layer]; ok {
		return c
	}
	return "#8ee6a8"
}

func intersectsAny(rect [4]int, rects [][4]int) bool {
	for _, r := range rects {
		if rect[0] < r[2] && rect[2] > r[0] && rect[1] < r[3] && rect[3] > r[1] {
			return true
		}
	}
	return false
}
