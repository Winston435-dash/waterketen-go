package screens

import (
	tea "charm.land/bubbletea/v2"
	"strings"
)

// TypedText accepts printable keys and v2's separate bracketed-paste messages.
// Control shortcuts never become text. Flatten pasted newlines in single-line fields.
func TypedText(msg tea.Msg) string {
	var text string
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if msg.Mod.Contains(tea.ModCtrl) || msg.Mod.Contains(tea.ModAlt) {
			return ""
		}
		text = msg.Text
		if msg.Code == tea.KeySpace {
			text = " "
		}
	case tea.PasteMsg:
		text = msg.Content
	}
	return strings.Map(func(r rune) rune {
		if r == '\n' || r == '\r' || r == '\t' {
			return ' '
		}
		if r < 0x20 || r == 0x7f {
			return -1
		}
		return r
	}, text)
}

func isSpace(msg tea.KeyPressMsg) bool {
	return msg.Code == tea.KeySpace && msg.Mod == 0
}
