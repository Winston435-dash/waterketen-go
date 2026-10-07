// Package components holds the canonical waterketen UI kit: Panel, DataTable,
// CommandBar, StatusBadge, EmptyState, Modal. Screens assemble these; they may
// not draw their own borders, padding, or colors.
package components

import (
	"github.com/charmbracelet/lipgloss"

	"waterketen/internal/ui"
)

// Panel is THE content container: a rounded border in the token border color,
// a title in the top border, and token padding inside. width/height are the
// OUTER size; the returned string always occupies exactly that box.
func Panel(title, body string, width, height int, active bool) string {
	border := ui.BorderPanel
	borderColor := ui.ColorBorder
	if active {
		border = ui.BorderActive
		borderColor = ui.ColorPrimary
	}
	style := lipgloss.NewStyle().
		Border(border, true).
		BorderForeground(borderColor).
		BorderBackground(ui.ColorSurface).
		Background(ui.ColorSurface).
		Foreground(ui.ColorText).
		Padding(0, ui.SpacingXS).
		Width(width - border.GetLeftSize() - border.GetRightSize()).
		Height(height - border.GetTopSize() - border.GetBottomSize()).
		MaxWidth(width).
		MaxHeight(height)
	rendered := style.Render(body)
	return lipgloss.NewStyle().MaxWidth(width).MaxHeight(height).Render(
		lipgloss.PlaceHorizontal(width, lipgloss.Left, withBorderTitle(rendered, title, borderColor)),
	)
}

// withBorderTitle paints `title` over the panel's top border line.
func withBorderTitle(rendered, title string, color lipgloss.TerminalColor) string {
	if title == "" {
		return rendered
	}
	lines := lipgloss.NewStyle().Render(rendered)
	_ = lines
	// Cheap overlay: inject the title after the top-left corner of the first line.
	// (Lip Gloss v1 lacks native border titles; this keeps it semantic: panels
	// own their title, screens never draw one.)
	first, rest, found := cut(rendered)
	if !found {
		return rendered
	}
	label := lipgloss.NewStyle().Foreground(color).Background(ui.ColorSurface).
		Render(" " + title + " ")
	width := lipgloss.Width(first)
	labelWidth := lipgloss.Width(label)
	const corner = 1 // after the rounded corner rune
	if labelWidth+corner+1 > width {
		return rendered
	}
	head := sliceRunes(first, 0, corner)
	tail := sliceRunes(first, corner+labelWidth, width)
	return head + label + tail + "\n" + rest
}

// cut splits s at the first newline.
func cut(s string) (first, rest string, found bool) {
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			return s[:i], s[i+1:], true
		}
	}
	return s, "", false
}

// sliceRunes returns the printable cell range [from, to) of a single-line
// rendered string. Lip Gloss widths are cell widths; we walk ANSI-aware.
func sliceRunes(s string, from, to int) string {
	return lipgloss.NewStyle().MaxWidth(to).Render(truncateLeft(s, from))
}

// truncateLeft drops the first n printable cells of a single line.
func truncateLeft(s string, n int) string {
	out := make([]rune, 0, len(s))
	col := 0
	inEsc := false
	for _, r := range s {
		if r == 0x1b {
			inEsc = true
		}
		if inEsc {
			out = append(out, r)
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEsc = false
			}
			continue
		}
		if col >= n {
			out = append(out, r)
		}
		col++
	}
	return string(out)
}
