package app

import (
	tea "charm.land/bubbletea/v2"
	"strings"
	"testing"
	"waterketen/internal/ui/screens"
)

func TestV2ViewOwnsTerminalModesBeforeFirstResize(t *testing.T) {
	v := New(t.TempDir()).View()
	if !v.AltScreen || v.MouseMode != tea.MouseModeCellMotion {
		t.Fatal("terminal modes missing during initialization")
	}
}

func TestV2PasteOwnedByPaletteFilterAndMapOverlay(t *testing.T) {
	for _, id := range []screens.ScreenID{screens.ScreenKaart, screens.ScreenLagen, screens.ScreenGebieden} {
		m := paletteModel(t, id)
		m = press(m, appKey("/"))
		m = press(m, tea.PasteMsg{Content: "q3j café west zeedijk"})
		if m.screens[m.active].ID() != id || !strings.Contains(m.viewContent(), "q3j café west zeedijk") {
			t.Fatal("paste leaked into navigation or lost Unicode/spaces")
		}
	}
	m := paletteModel(t, screens.ScreenOntdek)
	m = press(m, appKey("ctrl+p"))
	m = press(m, tea.PasteMsg{Content: "Ga naar Kaart"})
	m = press(m, appKey("enter"))
	if m.screens[m.active].ID() != screens.ScreenKaart {
		t.Fatal("v2 palette paste did not select command")
	}
}
