package components

import (
	"charm.land/lipgloss/v2"
	"image/color"

	"waterketen/internal/ui"
)

// Status is the health of one thing: a probe, a layer, an export.
type Status int

const (
	StatusOK   Status = iota // reachable / file present; not proof of physical completeness
	StatusWarn               // partial, degraded, truncated
	StatusDown               // unreachable / failed
	StatusIdle               // not wired, not probed, "gedetecteerd"
)

// Badge renders a one-cell status mark with the token color, followed by an
// optional label. This is the ONLY way status is shown; screens never color
// their own ●/✓.
func Badge(status Status, label string) string {
	var mark string
	var fg color.Color
	switch status {
	case StatusOK:
		mark, fg = "●", ui.ColorOK
	case StatusWarn:
		mark, fg = "◐", ui.ColorWarning
	case StatusDown:
		mark, fg = "●", ui.ColorError
	default:
		mark, fg = "○", ui.ColorMuted
	}
	b := lipgloss.NewStyle().Foreground(fg).Render(mark)
	if label == "" {
		return b
	}
	return b + " " + lipgloss.NewStyle().Foreground(ui.ColorText).Render(label)
}

// Muted renders secondary text in the muted token. Single place so "dim"
// never drifts between screens.
func Muted(text string) string {
	return lipgloss.NewStyle().Foreground(ui.ColorMuted).Render(text)
}

// Strong renders emphasized text (titles inside content, numbers that matter).
func Strong(text string) string {
	return lipgloss.NewStyle().Foreground(ui.ColorText).Bold(true).Render(text)
}

// Selected renders the one highlighted row of a list: pink + bold ink, never a
// filled bar, so the terminal background stays visible. Screens cannot reach
// for lipgloss themselves, so list selection lives here with the other ink.
func Selected(text string, width int) string {
	return lipgloss.NewStyle().Foreground(ui.ColorPrimary).Bold(true).Width(width).Render(text)
}

// Error renders a line in the error token. Prefer EmptyState for whole panels.
func Error(text string) string {
	return lipgloss.NewStyle().Foreground(ui.ColorError).Render(text)
}
