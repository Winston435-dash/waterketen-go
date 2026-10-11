package core

// Braille rendering primitives. One terminal cell holds 2x4 dots, so a
// 110x38 pane shows 220x152 samples. Two non-negotiable rules carried over
// from the Python atlas:
//
//   - Interpolate segments (Bresenham). Real riool features average ~2.6
//     vertices: vertex-only plotting draws a 40 m pipe as two dots and the
//     network looks empty.
//   - Polygons draw outline-only (the Outlines() helper): a filled district
//     would hide the network.

const brailleBlank = 0x2800

// dotBits[y][x] is the bit of dot (x, y) inside one braille cell.
var dotBits = [4][2]int{{0x01, 0x08}, {0x02, 0x10}, {0x04, 0x20}, {0x40, 0x80}}

// DotGrid is a bit-array of dots addressed in dot coordinates.
type DotGrid struct {
	Width, Height       int // in terminal cells
	DotWidth, DotHeight int // in dots
	bits                [][]int
}

// NewDotGrid creates an empty grid of w x h cells.
func NewDotGrid(w, h int) DotGrid {
	bits := make([][]int, h)
	for i := range bits {
		bits[i] = make([]int, w)
	}
	return DotGrid{Width: w, Height: h, DotWidth: w * 2, DotHeight: h * 4, bits: bits}
}

// SetDot marks dot (x, y) in dot coordinates; out-of-range is a no-op.
func (g *DotGrid) SetDot(x, y int) {
	if x < 0 || y < 0 || x >= g.DotWidth || y >= g.DotHeight {
		return
	}
	cx, dx := x/2, x%2
	cy, dy := y/4, y%4
	g.bits[cy][cx] |= dotBits[dy][dx]
}

// lineInt walks a Bresenham segment between integer dot coordinates and calls
// plot for each dot. It is the single rasterising rule, shared by the colour
// grid (which also records colour per dot), so segment drawing cannot drift.
func lineInt(x0, y0, x1, y1 int, plot func(x, y int)) {
	dx := absi(x1 - x0)
	dy := -absi(y1 - y0)
	sx, sy := 1, 1
	if x0 > x1 {
		sx = -1
	}
	if y0 > y1 {
		sy = -1
	}
	err := dx + dy
	for {
		plot(x0, y0)
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x0 += sx
		}
		if e2 <= dx {
			err += dx
			y0 += sy
		}
	}
}

// Text renders the grid as `Height` strings of `Width` braille runes.
func (g DotGrid) Text() []string {
	out := make([]string, g.Height)
	for y, row := range g.bits {
		rs := make([]rune, g.Width)
		for x, bits := range row {
			if bits == 0 {
				rs[x] = ' '
			} else {
				rs[x] = rune(brailleBlank + bits)
			}
		}
		out[y] = string(rs)
	}
	return out
}

func absi(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
