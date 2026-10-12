package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"waterketen/internal/core"
	"waterketen/internal/ui/screens"
)

func TestAreaRegistryCompletesWhileAnotherPanelIsActive(t *testing.T) {
	m := paletteModel(t, screens.ScreenGebieden)
	regions := screens.NewGebieden(&core.Library{Root: t.TempDir(), Snapshot: "fixture", Manifest: core.LibraryManifest{AreaCount: 1}}, nil)
	m.screens[m.active] = regions
	regions.SetSize(100, 28)
	if regions.LoadCmd() == nil {
		t.Fatal("area registry did not dispatch")
	}
	m = pressText(m, "7")
	next, _ := m.Update(screens.AreasMsg{Areas: []core.Area{{Code: "GM0599", Name: "Rotterdam", Kind: core.KindGemeente}}})
	m = next.(Model)
	if regions.Busy() || !strings.Contains(ansi.Strip(regions.View()), "GM0599") {
		t.Fatal("inactive region panel dropped completion")
	}
}

func TestInventoryFilterOwnsTypedGlobalKeysAndAllowsCtrlC(t *testing.T) {
	for _, id := range []screens.ScreenID{screens.ScreenLagen, screens.ScreenGebieden} {
		m := paletteModel(t, id)
		m = pressText(m, "/")
		m = pressText(m, "q3j")
		if m.screens[m.active].ID() != id || !strings.Contains(ansi.Strip(m.View()), "q3j") {
			t.Fatal("typing inventory filter quit or navigated")
		}
		_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
		if cmd == nil {
			t.Fatal("ctrl+c cannot exit during filter editing")
		}
		if _, ok := cmd().(tea.QuitMsg); !ok {
			t.Fatal("ctrl+c did not quit")
		}
	}
}

func TestAreaPanelStartsLazyLoadOnEntry(t *testing.T) {
	m := paletteModel(t, screens.ScreenOntdek)
	index := m.screenIndex(screens.ScreenGebieden)
	regions := screens.NewGebieden(&core.Library{Root: t.TempDir()}, nil)
	m.screens[index] = regions
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	m = next.(Model)
	if cmd == nil || !regions.Busy() {
		t.Fatal("area entry did not dispatch lazy read and spinner")
	}
}
