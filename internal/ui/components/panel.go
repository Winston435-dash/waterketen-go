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
	// No background: the terminal's own background shows through, so a panel is
	// a floating outline rather than a filled card. Never paint a surface here.
	style := lipgloss.NewStyle().
		Border(border, true).
		BorderForeground(borderColor).
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
	// Cheap overlay: inject the title after the top-left corner of the first line.
	// (Lip Gloss v1 lacks native border titles; this keeps it semantic: panels
	// own their title, screens never draw one.)
	first, rest, found := cut(rendered)
	if !found {
		return rendered
	}
	label := lipgloss.NewStyle().Foreground(color).
		Render(" " + title + " ")
	width := lipgloss.Width(first)
	labelWidth := lipgloss.Width(label)
	const corner = 1 // after the rounded corner rune
	if labelWidth+corner+1 > width {
		// The title does not fit this terminal. On a phone that must not eat the
		// top border: clip the title to what is left rather than skipping it and
		// leaving the user with an untitled panel.
		room := width - corner - 1
		if room < 2 {
			return rendered
		}
		label = sliceRunes(label, 0, room)
		labelWidth = lipgloss.Width(label)
	}
	head := sliceRunes(first, 0, corner)
	tail := sliceRunes(first, corner+labelWidth, width)
	return head + label + tail + "\n" + rest
}

// cut splits s at the first newline.
func cut(s string) (first, rest string, found bool) {
	if i := indexByte(s, '\n'); i >= 0 {
		return s[:i], s[i+1:], true
	}
	return s, "", false
}

// SingleLine guarantees a string occupies EXACTLY ONE row of `width` cells: it
// takes the first line, truncates it ANSI-aware, and pads to the width. This is
// the primitive for the header and the command bar, which are each one row of the
// app's fixed anatomy.
//
// It exists because lipgloss Width() PADS but does not truncate, and Padding is
// added ON TOP of Width — so `Width(n).Padding(0,1)` renders n+2 cells and the
// terminal wraps it onto a second row. On a phone that extra row pushed the
// command bar off the bottom of the screen. Measured: the header and bar were
// both 2 rows wide at 30x12.
//
// A truncated string ends in "…" (when there is room for one). A silently chopped
// CBS code — "(BU0599108" — reads as a typo rather than a clipped line.
func SingleLine(s string, width int) string {
	if width < 1 {
		return ""
	}
	if i := indexByte(s, '\n'); i >= 0 {
		s = s[:i]
	}
	out := make([]rune, 0, width)
	col, inEsc, truncated := 0, false, false
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
		w := runeWidth(r)
		if col+w > width {
			truncated = true
			break
		}
		out = append(out, r)
		col += w
	}
	if truncated && width >= 2 {
		// Drop cells to make room for the ellipsis, counting only printable ones.
		out = trimCells(out, width-1)
		col = 0
		for _, r := range out {
			if r != 0x1b {
				col += runeWidth(r)
			}
		}
		out = append(out, '…')
		col++
	}
	for ; col < width; col++ {
		out = append(out, ' ')
	}
	return string(out)
}

// trimCells returns the prefix of out that occupies at most `max` printable cells,
// keeping every escape sequence intact.
func trimCells(out []rune, max int) []rune {
	col, inEsc := 0, false
	for i, r := range out {
		if r == 0x1b {
			inEsc = true
		}
		if inEsc {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEsc = false
			}
			continue
		}
		if col+runeWidth(r) > max {
			return out[:i]
		}
		col += runeWidth(r)
	}
	return out
}

// runeWidth is the terminal cell width of one rune: 0 for combining marks, 2 for
// the East-Asian wide and braille-adjacent blocks the map uses, else 1.
func runeWidth(r rune) int {
	switch {
	case r == 0:
		return 0
	case r < 0x20:
		return 0
	case r >= 0x0300 && r <= 0x036f: // combining diacritics
		return 0
	case r >= 0x1100 && r <= 0x115f, // Hangul Jamo
		r >= 0x2e80 && r <= 0xa4cf, // CJK
		r >= 0xac00 && r <= 0xd7a3, // Hangul
		r >= 0xf900 && r <= 0xfaff,
		r >= 0xfe30 && r <= 0xfe6f,
		r >= 0xff00 && r <= 0xff60,
		r >= 0xffe0 && r <= 0xffe6:
		return 2
	}
	return 1
}

func indexByte(s string, b byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			return i
		}
	}
	return -1
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
