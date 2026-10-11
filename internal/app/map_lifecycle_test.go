package app

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	tea "github.com/charmbracelet/bubbletea"

	"waterketen/internal/core"
	"waterketen/internal/ui/screens"
)

func mapLifecycleModel(t *testing.T) Model {
	t.Helper()
	m := New(t.TempDir())
	m.width, m.height, m.ready = 100, 30, true
	for i, screen := range m.screens {
		if screen.ID() == screens.ScreenKaart {
			m.active = i
			break
		}
	}
	return m.applySize()
}

func lifecycleKey(m Model, r rune) Model {
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	return next.(Model)
}

func lifecycleRender(r rune) *core.Render {
	return &core.Render{Cells: [][]core.Cell{{{Rune: r, Style: core.StyleForLayer("riool")}}}}
}

func TestLateMapDrawCannotReplaceNewerRenderAfterPanelSwitch(t *testing.T) {
	m := mapLifecycleModel(t)
	k := m.kaart()
	k.SetScene(&core.Scene{Bounds: core.BBox{MinX: 0, MinY: 0, MaxX: 10, MaxY: 10}})
	k.Commit(lifecycleRender('⣿'))
	m.drawGen = 2
	m.drawing = false // the newer draw has already completed

	m = lifecycleKey(m, '7')
	next, _ := m.Update(screens.DrawnMsg{Render: lifecycleRender('░')})
	m = next.(Model)
	m = lifecycleKey(m, '5')

	view := ansi.Strip(k.View())
	if !strings.Contains(view, "⣿") || strings.Contains(view, "░") {
		t.Fatalf("late draw replaced the current map after panel navigation: %q", view)
	}
}

func TestMapResizeDispatchesRedrawInsteadOfLeavingMapBusy(t *testing.T) {
	m := mapLifecycleModel(t)
	k := m.kaart()
	k.SetScene(&core.Scene{Bounds: core.BBox{MinX: 0, MinY: 0, MaxX: 10, MaxY: 10}})

	next, cmd := m.Update(tea.WindowSizeMsg{Width: 110, Height: 32})
	m = next.(Model)
	if cmd == nil {
		t.Fatal("resizing the map did not dispatch its redraw")
	}
	if view := ansi.Strip(k.View()); strings.Contains(view, "Kaart wordt geladen") {
		t.Fatalf("resize left the managed map waiting on a load that cannot complete: %q", view)
	}
}
