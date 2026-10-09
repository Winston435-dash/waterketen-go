package screens

import tea "github.com/charmbracelet/bubbletea"

// typedText returns the characters a key message contributes to a text field, or
// "" when the key is not text.
//
// This exists because a SPACE is not a KeyRunes event. Bubble Tea parses " " into
// Key{Type: KeySpace, Runes: []rune{' '}} (bubbletea/key.go: len(runes)==1 &&
// runes[0]==' ' → KeySpace). A handler that only tests `msg.Type == tea.KeyRunes`
// therefore silently DROPS every space: typing "west zeedijk" produced
// "westzeedijk" (and a leading space produced nothing at all). That was the bug on
// '/' in panel 5.
//
// Pasted text arrives as KeyRunes with Paste set and the whole run in Runes, so it
// flows through here too — pasting a street name works for free.
func typedText(msg tea.KeyMsg) string {
	switch msg.Type {
	case tea.KeySpace:
		return " "
	case tea.KeyRunes:
		return string(msg.Runes)
	}
	return ""
}

// isSpace reports whether the key is a space. Screens that treat space as a COMMAND
// (toggling a layer in the picker) must use this rather than comparing
// msg.String(): a pasted space stringifies to "[…]" instead of " ".
func isSpace(msg tea.KeyMsg) bool {
	return msg.Type == tea.KeySpace
}
