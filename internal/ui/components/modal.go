package components

import (
	"charm.land/lipgloss/v2"
	"waterketen/internal/ui"
)

func DialogWidth(preferred,host int) int {
	if host<ui.CompactDialogWidth { return max(1,host) }
	return max(1,min(preferred,host-2*ui.SpacingLG))
}

// Modal renders `content` as a floating dialog over `background` (both full
// screen-sized views). The dialog is a Panel — the same token borders — centered.
// Nothing is painted behind it: the terminal background shows through, so
// `background` is accepted only to keep the call shape stable.
//
// The requested width is CLAMPED to the screen. A modal sized for a desktop (72
// columns) on a 40-column phone would be wider than the terminal, so lipgloss
// would wrap it into a ragged column that no longer looks like a dialog. The
// caller asks for a comfortable width; the terminal decides what fits.
func Modal(background, title, content string, screenW, screenH, width int) string {
	_ = background
	if screenW < 1 {
		screenW = 1
	}
	if screenH < 1 {
		screenH = 1
	}
	width = clampModalWidth(width, screenW)
	// Height is the content plus the panel chrome, capped to the screen so a long
	// picker on a short terminal cannot run past the bottom.
	height := lipgloss.Height(content) + 4
	if height > screenH {
		height = screenH
	}
	dialog := Panel(title, content, width, height, true)
	return lipgloss.Place(screenW, screenH, lipgloss.Center, lipgloss.Center,
		dialog,
		lipgloss.WithWhitespaceChars(" "),
	)
}

// clampModalWidth fits a dialog inside the screen: at least a readable 20 cells
// when the terminal allows it, otherwise the full width.
func clampModalWidth(width, screenW int) int {
	if width > screenW {
		width = screenW
	}
	if width < 20 && screenW >= 20 {
		width = 20
	}
	if width < 1 {
		width = 1
	}
	return width
}
