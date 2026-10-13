// Package components holds the canonical waterketen UI kit: Panel, DataTable,
// CommandBar, StatusBadge, EmptyState, Modal. Screens assemble these; they may
// not draw their own borders, padding, or colors.
package components

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

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
	// v2 Width/Height include the frame, and its word wrapper can remove ink
	// from styled leading/trailing spaces. Fit the already laid-out body ourselves
	// so map cells (including blank cells) survive the stack migration unchanged.
	innerW, innerH := max(0, width-4), max(0, height-2)
	rows := strings.Split(body, "\n")
	if len(rows) > innerH {
		rows = rows[:innerH]
	}
	for len(rows) < innerH {
		rows = append(rows, "")
	}
	for i, row := range rows {
		row = ansi.Truncate(row, innerW, "")
		rows[i] = row + strings.Repeat(" ", max(0, innerW-ansi.StringWidth(row)))
	}
	style := lipgloss.NewStyle().
		Border(border, true).
		BorderForeground(borderColor).
		Foreground(ui.ColorText).
		Padding(0, ui.SpacingXS).
		MaxWidth(width).
		MaxHeight(height)
	rendered := style.Render(strings.Join(rows, "\n"))
	return lipgloss.NewStyle().MaxWidth(width).MaxHeight(height).Render(
		lipgloss.PlaceHorizontal(width, lipgloss.Left, withBorderTitle(rendered, title, borderColor)),
	)
}

// withBorderTitle paints `title` over the panel's top border line.
func withBorderTitle(rendered, title string, color color.Color) string {
	if title == "" {
		return rendered
	}
	// Cheap overlay: inject the title after the top-left corner of the first line.
	// (Lip Gloss v1 lacks native border titles; this keeps it semantic: panels
	// own their title, screens never draw one.)
	first, rest, found := strings.Cut(rendered, "\n")
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
		label = SingleLine(label, room)
		labelWidth = lipgloss.Width(label)
	}
	head := sliceRunes(first, 0, corner)
	tail := sliceRunes(first, corner+labelWidth, width)
	return head + label + tail + "\n" + rest
}

// SingleLine guarantees a string occupies EXACTLY ONE row of `width` cells: it
// takes the first line, truncates it ANSI-aware, and pads to the width. This is
// the primitive for the header and the command bar, which are each one row of the
// app's fixed anatomy.
//
// Width alone does not truncate. The original v1 padding bug made these slots
// wrap on 30x12; v2 still needs explicit ANSI/grapheme-aware truncation here.
//
// A truncated string ends in "…" (when there is room for one). A silently chopped
// CBS code — "(BU0599108" — reads as a typo rather than a clipped line.
func SingleLine(s string, width int) string {
	if width < 1 {
		return ""
	}
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[:i]
	}
	out := ansi.Truncate(s, width, "…")
	return out + strings.Repeat(" ", max(0, width-ansi.StringWidth(out)))
}

// sliceRunes returns the printable cell range [from, to) of a single-line
// rendered string. Lip Gloss widths are cell widths; we walk ANSI-aware.
func sliceRunes(s string, from, to int) string {
	return ansi.Cut(s, from, to)
}

// EmptyState is the single pattern for "there is nothing here": a muted glyph,
// a bold one-line headline, then the action the user should take (the fetch
// hint, typically). It exists so no screen invents its own sad-empty-box.
func EmptyState(headline, hint string, width int) string {
	body := lipgloss.JoinVertical(lipgloss.Left,
		Muted("∅"),
		Strong(headline),
		"",
		Muted(hint),
	)
	return lipgloss.Place(width, 5, lipgloss.Left, lipgloss.Top, body)
}
