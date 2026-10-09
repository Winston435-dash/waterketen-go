package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"waterketen/internal/core"
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
	return b.String()
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
