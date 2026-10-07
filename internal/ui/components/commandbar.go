package components

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"

	"waterketen/internal/ui"
)

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

// View renders one line (truncated to width) of the screen's active bindings.
func (c CommandBar) View(width int, bindings []key.Binding) string {
	c.help.Width = width
	return lipgloss.NewStyle().
		Background(ui.ColorSurfaceHi).
		Foreground(ui.ColorText).
		Width(width).
		MaxWidth(width).
		Render(c.help.ShortHelpView(bindings))
}
