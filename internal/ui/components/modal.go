package components

import (
	"github.com/charmbracelet/lipgloss"
)

// Modal renders `content` as a floating dialog over `background` (both full
// screen-sized views). The dialog is a Panel — the same token borders —
// centered. Nothing is painted behind it: the terminal background shows
// through, so `background` is accepted only to keep the call shape stable.
func Modal(background, title, content string, screenW, screenH, width int) string {
	_ = background
	dialog := Panel(title, content, width, lipgloss.Height(content)+4, true)
	return lipgloss.Place(screenW, screenH, lipgloss.Center, lipgloss.Center,
		dialog,
		lipgloss.WithWhitespaceChars(" "),
	)
}
