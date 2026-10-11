package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"waterketen/internal/ui/screens"
)

func TestQueryLoadsOnEntryAndCompletesWhileInactive(t *testing.T) {
	m := New(t.TempDir())
	if len(m.screens) <= int(screens.ScreenQuery) {
		t.Fatal("panel 8 not instantiated")
	}
	m.width, m.height, m.ready = 110, 26, true
	m = m.applySize()
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("8")})
	m = next.(Model)
	if cmd == nil {
		t.Fatal("query metadata was not dispatched on panel entry")
	}
	msg := cmd()
	if _, ok := msg.(screens.QueryOpenMsg); !ok {
		t.Fatalf("expected query metadata, got %T", msg)
	}
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("1")})
	m = next.(Model)
	next, _ = m.Update(msg)
	m = next.(Model)
	q := m.screens[screens.ScreenQuery].(*screens.Query)
	if q.Busy() {
		t.Fatal("inactive query screen dropped its completion")
	}
	if !strings.Contains(q.View(), "geen kaartbibliotheek") {
		t.Fatal("metadata error not delivered to owner")
	}
}
