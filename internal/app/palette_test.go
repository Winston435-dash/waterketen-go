package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"waterketen/internal/core"
	"waterketen/internal/ui/screens"
)

func paletteModel(t *testing.T, active screens.ScreenID) Model {
	t.Helper()
	m := New(t.TempDir())
	m.width, m.height, m.ready = 100, 30, true
	for i, screen := range m.screens {
		if screen.ID() == active {
			m.active = i
			break
		}
	}
	if k := m.kaart(); k != nil {
		k.SetScene(&core.Scene{})
	}
	return m.applySize()
}

func press(m Model, msg tea.KeyMsg) Model {
	next, _ := m.Update(msg)
	return next.(Model)
}

func pressText(m Model, text string) Model {
	for _, r := range text {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
		if r == ' ' {
			msg = tea.KeyMsg{Type: tea.KeySpace}
		}
		next, _ := m.Update(msg)
		m = next.(Model)
	}
	return m
}

func TestGlobalPaletteNavigationCommandChangesPanel(t *testing.T) {
	m := paletteModel(t, screens.ScreenOntdek)
	m = press(m, tea.KeyMsg{Type: tea.KeyCtrlP})
	m = pressText(m, "Ga naar Kaart")
	m = press(m, tea.KeyMsg{Type: tea.KeyEnter})
	if got := m.screens[m.active].ID(); got != screens.ScreenKaart {
		t.Fatalf("navigation command selected %v, want Kaart", got)
	}
}

func TestGlobalPaletteRefreshStartsProbeWithoutChangingPanel(t *testing.T) {
	m := paletteModel(t, screens.ScreenOntdek)
	m = press(m, tea.KeyMsg{Type: tea.KeyCtrlP})
	m = pressText(m, "Ververs bronnen")
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(Model)
	if cmd == nil {
		t.Fatal("refresh command did not dispatch source probes")
	}
	if got := m.screens[m.active].ID(); got != screens.ScreenOntdek {
		t.Fatalf("refresh changed the active panel to %v", got)
	}
	for _, screen := range m.screens {
		if sources, ok := screen.(*screens.Bronnen); ok {
			if !sources.Busy() {
				t.Fatal("refresh command did not start the source panel")
			}
			return
		}
	}
	t.Fatal("source panel was not registered")
}

func TestGlobalPaletteSearchRunsPanelCommand(t *testing.T) {
	m := paletteModel(t, screens.ScreenOntdek)
	m = press(m, tea.KeyMsg{Type: tea.KeyCtrlP})

	view := ansi.Strip(m.View())
	for _, want := range []string{"Opdrachten", "Ververs bronnen", "Ga naar Bronnen"} {
		if !strings.Contains(view, want) {
			t.Fatalf("global palette view does not contain %q:\n%s", want, view)
		}
	}

	m = pressText(m, "straat zoeken")
	view = ansi.Strip(m.View())
	if !strings.Contains(view, "straat zoeken") || strings.Contains(view, "Ga naar Bronnen") {
		t.Fatalf("global palette did not filter its commands:\n%s", view)
	}

	m = press(m, tea.KeyMsg{Type: tea.KeyEnter})
	if got := m.screens[m.active].ID(); got != screens.ScreenKaart {
		t.Fatalf("selected map command left active panel at %v", got)
	}
	if k := m.kaart(); k == nil || !k.OverlayOpen() {
		t.Fatal("selecting the map search command did not open map search")
	}
	if strings.Contains(ansi.Strip(m.View()), "Ververs bronnen") {
		t.Fatal("palette remained visible after selecting a command")
	}
}

func TestPanelPaletteIsScopedToActivePanel(t *testing.T) {
	m := paletteModel(t, screens.ScreenKaart)
	m = press(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})

	view := ansi.Strip(m.View())
	if !strings.Contains(view, "straat zoeken") {
		t.Fatalf("panel palette is missing map commands:\n%s", view)
	}
	if strings.Contains(view, "Ga naar Bronnen") || strings.Contains(view, "Ververs bronnen") {
		t.Fatalf("panel palette leaked global commands:\n%s", view)
	}

	m = pressText(m, "/")
	m = press(m, tea.KeyMsg{Type: tea.KeyEnter})
	if k := m.kaart(); k == nil || !k.OverlayOpen() {
		t.Fatal("selecting the panel search command did not open map search")
	}
}

func TestPaletteEscapeClosesWithoutDispatchingToPanel(t *testing.T) {
	m := paletteModel(t, screens.ScreenKaart)
	m = press(m, tea.KeyMsg{Type: tea.KeyCtrlP})
	m = press(m, tea.KeyMsg{Type: tea.KeyEsc})

	if strings.Contains(ansi.Strip(m.View()), "Ververs bronnen") {
		t.Fatal("palette remained visible after escape")
	}
	if got := m.screens[m.active].ID(); got != screens.ScreenKaart {
		t.Fatalf("escape changed the active panel to %v", got)
	}
	if k := m.kaart(); k == nil || k.OverlayOpen() {
		t.Fatal("escape leaked through to the map screen")
	}
}

func TestInventoryPanelPaletteCanSwitchToGlobal(t *testing.T) {
	m := paletteModel(t, screens.ScreenLagen)
	m = press(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	view := ansi.Strip(m.View())
	if !strings.Contains(view, "Lagen · lokaal/alle") || !strings.Contains(view, "Lagen · filter") {
		t.Fatalf("inventory panel palette omits its new actions:\n%s", view)
	}
	if strings.Contains(view, "Ververs bronnen") || !strings.Contains(view, "ctrl+p") {
		t.Fatalf("panel palette has the wrong scope or no global hint:\n%s", view)
	}

	m = press(m, tea.KeyMsg{Type: tea.KeyCtrlP})
	if view := ansi.Strip(m.View()); !strings.Contains(view, "Ververs bronnen") {
		t.Fatalf("ctrl+p did not expand to global commands:\n%s", view)
	}
}

func TestPanelPaletteIncludesZoomCommands(t *testing.T) {
	m := paletteModel(t, screens.ScreenKaart)
	m = press(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	m = pressText(m, "zoom in")
	view := ansi.Strip(m.View())
	if !strings.Contains(view, "Kaart · zoom in") {
		t.Fatalf("panel palette omitted an advertised map action:\n%s", view)
	}
}

func TestGlobalPaletteBindingsStayVisibleOnNarrowTerminals(t *testing.T) {
	m := paletteModel(t, screens.ScreenKaart)
	m.width = 30
	m = m.applySize()
	lines := strings.Split(ansi.Strip(m.View()), "\n")
	footer := lines[len(lines)-1]
	if !strings.Contains(footer, "ctrl+p") || !strings.Contains(footer, "?") {
		t.Fatalf("palette shortcuts disappeared from the narrow footer: %q", footer)
	}
}

func TestGlobalPaletteOwnsKeysAboveMapSearch(t *testing.T) {
	m := paletteModel(t, screens.ScreenKaart)
	m = press(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	if k := m.kaart(); k == nil || !k.OverlayOpen() {
		t.Fatal("test setup did not open map search")
	}

	m = press(m, tea.KeyMsg{Type: tea.KeyCtrlP})
	m = press(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if view := ansi.Strip(m.View()); !strings.Contains(view, "Zoek · q") {
		t.Fatalf("palette did not own the typed key:\n%s", view)
	}
	m = press(m, tea.KeyMsg{Type: tea.KeyEsc})
	if k := m.kaart(); k == nil || !k.OverlayOpen() {
		t.Fatal("closing the palette also closed the underlying map search")
	}
}
