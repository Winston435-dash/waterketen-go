package components

import (
	"fmt"
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/lucasb-eyer/go-colorful"

	"waterketen/internal/core"
	"waterketen/internal/ui"
)

// RenderAtlas turns the core renderer's semantic grid into styled terminal
// output. It is the ONE core→lipgloss translation point for the atlas: core
// owns the data ink (core.CellColour), the kit owns only emphasis
// (bold/underline), and screens never touch a colour. Runs of one style are
// emitted in a single Render call so the escape output stays compact.
func RenderAtlas(r *core.Render, compact ...bool) string {
	var b strings.Builder
	for y, row := range r.Cells {
		var cur string
		var curStyle core.CellStyle
		flush := func() {
			if cur == "" {
				return
			}
			b.WriteString(atlasStyle(curStyle).Render(cur))
			cur = ""
		}
		for _, cell := range row {
			if cell.Style != curStyle {
				flush()
				curStyle = cell.Style
			}
			cur += string(cell.Rune)
		}
		flush()
		if y < len(r.Cells)-1 {
			b.WriteByte('\n')
		}
	}
	grid := b.String()
	if len(r.Legend) == 0 || len(r.Cells) == 0 {
		return grid
	}
	width := len(r.Cells[0])
	legend := AtlasLegend(r.Legend, width)
	if len(compact) > 0 && compact[0] {
		legend = SingleLine(Muted(fmt.Sprintf("Legenda: %d lagen · L opent", len(r.Legend))), width)
	}
	if legend == "" {
		return grid
	}
	return grid + "\n" + legend
}

func AtlasLegend(entries []core.LegendEntry, width int) string {
	if width < 8 || len(entries) == 0 {
		return ""
	}

	heading := lipgloss.NewStyle().Foreground(ui.ColorMuted).Render("Legenda:")
	var lines []string
	line := heading
	lineWidth := lipgloss.Width(heading)
	for _, entry := range entries {
		symbol := entry.Symbol
		if symbol == "" {
			symbol = "•"
		}
		mark := atlasStyle(entry.Style).Render(symbol)
		labelWidth := width - lipgloss.Width(mark) - 1
		if labelWidth < 1 {
			continue
		}
		label := entry.Label
		if lipgloss.Width(label) > labelWidth {
			label = strings.TrimRight(SingleLine(label, labelWidth), " ")
		}
		item := mark + " " + label
		itemWidth := lipgloss.Width(item)
		if lineWidth+2+itemWidth > width {
			lines = append(lines, line)
			line, lineWidth = item, itemWidth
			continue
		}
		line += "  " + item
		lineWidth += 2 + itemWidth
	}
	if line != "" {
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

type AtlasGeometry struct {
	Width, GridHeight int
	Compact           bool
}

// AtlasLayout owns the height reserved for the actual legend. It is used both
// before drawing and when presenting the committed frame; no fixed six-row gap.
func AtlasLayout(width, height int, entries []core.LegendEntry) AtlasGeometry {
	compact := height <= ui.CompactMapRows
	rows := 0
	if len(entries) > 0 {
		rows = strings.Count(AtlasLegend(entries, width), "\n") + 1
		if compact {
			rows = 1
		}
	}
	if rows >= height-2 {
		compact = true
		rows = min(rows, 1)
	}
	return AtlasGeometry{Width: max(1, width), GridHeight: max(1, height-rows), Compact: compact}
}

// atlasStyle maps a core CellStyle to lipgloss. There is deliberately no colour
// table here: the ink comes from core.CellColour (single owner), and this
// function adds only the emphasis the map needs — highlight is bright +
// underlined, the compass/scale/labels are bold, never a filled box.
func atlasStyle(s core.CellStyle) lipgloss.Style {
	st := lipgloss.NewStyle().Foreground(atlasInk(core.CellColour(s)))
	switch s {
	case core.StyleHighlight:
		return st.Bold(true).Underline(true)
	case core.StyleNorth, core.StyleScale, core.StyleLabel:
		return st.Bold(true)
	}
	return st
}

// atlasInk preserves the RGB bytes emitted by v1's termenv codec. Its float
// conversion truncated some channels (e.g. #cc94ff -> 204;147;255). Keeping that
// quantization here makes the migration ink-identical without a second palette:
// core.CellColour remains the only owner of the colour values.
func atlasInk(hex string) color.Color {
	c, err := colorful.Hex(hex)
	if err != nil {
		return lipgloss.Color(hex)
	}
	return color.RGBA{R: uint8(c.R * 255), G: uint8(c.G * 255), B: uint8(c.B * 255), A: 255}
}
