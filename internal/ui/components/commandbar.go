package components

import (
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"

	"waterketen/internal/ui"
)

// compactWidth is the column count below which the command bar drops its
// descriptions and shows only the keys.
const compactWidth = 72

// CommandBar is the footer. The app's fixed anatomy is Header → Content →
// CommandBar; this is the third element and every screen renders exactly one.
// It renders ONLY the bindings the active screen handed it — the
// header already lists the panel keys, so duplicating them would just overflow
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

// barInk/barSep are the footer's ink and separator, built once. The bar is
// re-rendered on every frame, so per-render style construction would be the
// hottest allocation in the app for no reason.
var (
	barInk = lipgloss.NewStyle().Foreground(ui.ColorPrimary)
	barSep = lipgloss.NewStyle().Foreground(ui.ColorBorder).Render(" · ")
)

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
	// Keep the route to full pane help at the right edge. Important content
	// actions go first; global palette/nav hints must not crowd them out.
	var suffix string
	if len(bindings) > 0 && bindings[len(bindings)-1].Help().Key == "?" {
		suffix = barInk.Render("?") + " " + Muted("hulp")
		bindings = bindings[:len(bindings)-1]
	}
	room := width
	if suffix != "" {
		room = max(0, width-lipgloss.Width(suffix)-lipgloss.Width(barSep))
	}
	if width >= compactWidth {
		c.help.SetWidth(room)
		c.help.ShowAll = false
		content := c.help.ShortHelpView(bindings)
		if suffix != "" {
			if content != "" {
				content += barSep
			}
			content += suffix
		}
		return SingleLine(content, width)
	}
	// Try concise descriptions for the primary three actions first. Fall back
	// to complete key tokens rather than printing a half-word on small screens.
	primary := bindings[:min(3, len(bindings))]
	var described []string
	for _, binding := range primary {
		desc := binding.Help().Desc
		if desc == "straat zoeken" {
			desc = "zoek"
		}
		described = append(described, barInk.Render(binding.Help().Key)+" "+Muted(desc))
	}
	if full := strings.Join(described, barSep); len(primary) > 0 && lipgloss.Width(full) <= room {
		if suffix != "" {
			full += barSep + suffix
		}
		return SingleLine(full, width)
	}
	parts := make([]string, 0, len(bindings))
	for _, b := range bindings {
		if k := b.Help().Key; k != "" {
			parts = append(parts, barInk.Render(k))
		}
	}
	// Clip by WHOLE bindings, not mid-token: a footer ending in a dangling "·"
	// reads as a rendering bug. SingleLine is the last-resort clip.
	content := clipParts(parts, lipgloss.Width(barSep), room)
	if suffix != "" {
		if content != "" {
			content += barSep
		}
		content += suffix
	}
	return SingleLine(content, width)
}

// clipParts joins parts with a separator, dropping trailing parts that would not
// fit so the row never ends in a separator.
func clipParts(parts []string, sepWidth, width int) string {
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
	return strings.Join(parts[:end], barSep)
}
