package components

import (
	"github.com/charmbracelet/lipgloss"

	"waterketen/internal/ui"
)

// Status is the health of one thing: a probe, a layer, an export.
type Status int

const (
	StatusOK   Status = iota // up / complete
	StatusWarn               // partial, degraded, truncated
	StatusDown               // unreachable / failed
	StatusIdle               // not wired, not probed, "gedetecteerd"
)

// Badge renders a one-cell status mark with the token color, followed by an
// optional label. This is the ONLY way status is shown; screens never color
// their own ●/✓.
func Badge(status Status, label string) string {
	var mark string
	var fg lipgloss.TerminalColor
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

// Attention renders emphasised, non-text ink in a width — the sidebar's active
// panel, an alert count. The token lives here, not in a screen, so a screen is
// structurally unable to reach for lipgloss on its own.
func Attention(text string, width int) string {
	return lipgloss.NewStyle().Foreground(ui.ColorPrimary).Bold(true).Width(width).Render(text)
}

// MutedWidth renders secondary text padded to a fixed width.
func MutedWidth(text string, width int) string {
	return lipgloss.NewStyle().Foreground(ui.ColorMuted).Width(width).Render(text)
}

// Strong renders emphasized text (titles inside content, numbers that matter).
func Strong(text string) string {
	return lipgloss.NewStyle().Foreground(ui.ColorText).Bold(true).Render(text)
}

// Error renders a line in the error token. Prefer EmptyState for whole panels.
func Error(text string) string {
	return lipgloss.NewStyle().Foreground(ui.ColorError).Render(text)
}
