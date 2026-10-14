package app

import (
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

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
	next, _ := m.Update(appKey(string(r)))
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
	next, _ := m.Update(screens.DrawnMsg{Render: lifecycleRender('░'), Generation: 1})
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

func TestRedrawAfterMapErrorDoesNotStartSpinnerChain(t *testing.T) {
	m := mapLifecycleModel(t)
	k := m.kaart()
	k.SetError(errors.New("load failed"))

	next, cmd := m.Update(appKey("r"))
	m = next.(Model)
	if cmd != nil {
		t.Fatal("redraw after an error scheduled a render or spinner tick")
	}
	if k.Busy() {
		t.Fatal("map remained busy after a redraw with no scene")
	}
}

func TestPanelReturnQueuesFreshDrawAfterStaleDrawCompletes(t *testing.T) {
	m := mapLifecycleModel(t)
	k := m.kaart()
	k.SetScene(&core.Scene{Bounds: core.BBox{MinX: 0, MinY: 0, MaxX: 10, MaxY: 10}})
	k.Commit(lifecycleRender('⣿'))
	k.MarkDirty()
	m.drawGen = 1
	m.drawing = true // an older render is still in flight

	m = lifecycleKey(m, '7')
	next, _ := m.Update(screens.SceneMsg{Scene: &core.Scene{Bounds: core.BBox{MinX: 0, MinY: 0, MaxX: 20, MaxY: 20}}})
	m = next.(Model)
	next, _ = m.Update(screens.DrawnMsg{Render: lifecycleRender('░'), Generation: 1})
	m = next.(Model)

	next, cmd := m.Update(appKey("5"))
	m = next.(Model)
	if cmd == nil {
		t.Fatal("returning to Kaart did not queue a draw for the latest scene")
	}
	if !k.NeedsDraw() {
		t.Fatal("stale render completion incorrectly marked the latest scene clean")
	}

	next, drawCmd := m.Update(drawFrameMsg{gen: m.drawGen})
	if drawCmd == nil {
		t.Fatal("queued map frame did not start the render")
	}
	done := drawCmd()
	next, _ = next.(Model).Update(done)
	m = next.(Model)
	if k.NeedsDraw() || strings.Contains(ansi.Strip(k.View()), "Kaart wordt getekend") {
		t.Fatal("latest map render did not complete after returning to the panel")
	}
}
