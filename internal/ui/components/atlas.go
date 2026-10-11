package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"waterketen/internal/core"
	"waterketen/internal/ui"
)

// RenderAtlas turns the core renderer's semantic grid into styled terminal
// output. It is the ONE core→lipgloss translation point for the atlas: core
// owns the data ink (core.CellColour), the kit owns only emphasis
// (bold/underline), and screens never touch a colour. Runs of one style are
// emitted in a single Render call so the escape output stays compact.
func RenderAtlas(r *core.Render) string {
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
	legend := renderAtlasLegend(r.Legend, len(r.Cells[0]))
	if legend == "" {
		return grid
	}
	return grid + "\n" + legend
}

func renderAtlasLegend(entries []core.LegendEntry, width int) string {
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

// atlasStyle maps a core CellStyle to lipgloss. There is deliberately no colour
// table here: the ink comes from core.CellColour (single owner), and this
// function adds only the emphasis the map needs — highlight is bright +
// underlined, the compass/scale/labels are bold, never a filled box.
func atlasStyle(s core.CellStyle) lipgloss.Style {
	st := lipgloss.NewStyle().Foreground(lipgloss.Color(core.CellColour(s)))
	switch s {
	case core.StyleHighlight:
		return st.Bold(true).Underline(true)
	case core.StyleNorth, core.StyleScale, core.StyleLabel:
		return st.Bold(true)
	}
	return st
}
