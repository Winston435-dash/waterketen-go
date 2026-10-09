package components

import (
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"

	"waterketen/internal/ui"
)

// compactWidth is the column count below which the command bar drops its
// descriptions and shows only the keys.
const compactWidth = 72

// CommandBar is the footer. The app's fixed anatomy is Header → Content →
// CommandBar; this is the third element and every screen renders exactly one.
// It renders ONLY the bindings the active screen handed it — the sidebar
// already lists the panel keys, so duplicating them here would just overflow
// (the classic footer-mangling bug from the Textual app).
//
// ShortHelpView takes []key.Binding; screens pass their own subset.
type CommandBar struct {
	help help.Model
}

func NewCommandBar() CommandBar {
	h := help.New()
	h.Styles.ShortKey = lipgloss.NewStyle().Foreground(ui.ColorPrimary)
	h.Styles.ShortDesc = lipgloss.NewStyle().Foreground(ui.ColorMuted)
	h.Styles.ShortSeparator = lipgloss.NewStyle().Foreground(ui.ColorBorder)
	return CommandBar{help: h}
}

// View renders ONE row of the screen's active bindings, clipped to width.
//
// The one-row guarantee matters on a phone: lipgloss help wraps to as many rows
// as it needs, and the app builds the footer as a fixed one-row slot, so a wrap
// would push the bar off the bottom of the screen. SingleLine clips instead.
//
// Below compactWidth the descriptions are dropped and only the keys are shown:
// "← west • → oost • …" is worthless if the terminal cuts it mid-word after three
// bindings, and on Termux the keys are what the user needs to see.
func (c CommandBar) View(width int, bindings []key.Binding) string {
	if width < 1 {
		return ""
	}
	if width >= compactWidth {
		c.help.Width = width
		c.help.ShowAll = false
		return SingleLine(c.help.ShortHelpView(bindings), width)
	}
	sep := lipgloss.NewStyle().Foreground(ui.ColorBorder).Render(" · ")
	ink := lipgloss.NewStyle().Foreground(ui.ColorPrimary).Render
	parts := make([]string, 0, len(bindings))
	for _, b := range bindings {
		if k := b.Help().Key; k != "" {
			parts = append(parts, ink(k))
		}
	}
	// Clip by WHOLE bindings, not mid-token: a footer ending in a dangling "·"
	// reads as a rendering bug. SingleLine is the last-resort clip.
	return SingleLine(clipParts(parts, lipgloss.Width(sep), width), width)
}

// clipParts joins parts with a separator, dropping trailing parts that would not
// fit so the row never ends in a separator.
func clipParts(parts []string, sepWidth, width int) string {
	sep := lipgloss.NewStyle().Foreground(ui.ColorBorder).Render(" · ")
	total, end := 0, 0
	for i, p := range parts {
		add := lipgloss.Width(p)
		if i > 0 {
			add += sepWidth
		}
		if total+add > width {
			break
		}
		total, end = total+add, i+1
	}
	if end == 0 && len(parts) > 0 {
		return parts[0]
	}
	return strings.Join(parts[:end], sep)
}
