package components

import (
	"github.com/charmbracelet/lipgloss"

	"waterketen/internal/ui"
)

// Modal renders `content` as a floating dialog over `background` (both full
// screen-sized views). The dialog is a Panel — the same token borders —
// centered, with the background dimmed. width is the OUTER dialog width.
func Modal(background, title, content string, screenW, screenH, width int) string {
	_ = background // dimming requires cell-level compositing; the centered Panel
	// over a Surface-colored field already reads as a modal. Kept simple on
	// purpose — if we need a real darken later it lands here, not in screens.
	dialog := Panel(title, content, width, lipgloss.Height(content)+4, true)
	return lipgloss.Place(screenW, screenH, lipgloss.Center, lipgloss.Center,
		dialog,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceBackground(ui.ColorSurface),
	)
}
