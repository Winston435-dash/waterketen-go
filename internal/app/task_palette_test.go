package app

import (
	"strings"
	"testing"
	"waterketen/internal/ui/screens"
)

func TestGlobalPaletteListsTasksNotScrollCommands(t *testing.T) {
	m := paletteModel(t, screens.ScreenOntdek)
	p := newCommandPalette(m, paletteGlobal)
	for _, command := range p.commands {
		for _, generic := range []string{"omhoog", "omlaag", "pagina ↑", "pagina ↓", "begin", "eind", "west", "oost", "terug"} {
			if command.label == generic {
				t.Fatalf("generic scroll command leaked globally: %+v", command)
			}
		}
	}
	found := false
	for _, c := range p.commands {
		found = found || c.label == "Zoek straat"
	}
	if !found {
		t.Fatal("global palette missing semantic map search")
	}
	for _, c := range p.commands {
		if c.label == "Voer gebiedstelling uit" {
			t.Fatal("unloaded query advertised as executable")
		}
	}
}

func TestPaletteTextInputEditsInMiddle(t *testing.T) {
	m := paletteModel(t, screens.ScreenOntdek)
	m = press(m, appKey("ctrl+p"))
	m = pressText(m, "Zoek xstraat")
	for i := 0; i < 6; i++ {
		m = press(m, appKey("left"))
	}
	m = press(m, appKey("backspace"))
	if m.palette.query != "Zoek straat" {
		t.Fatalf("editing cursor lost: %q", m.palette.query)
	}
	if !strings.Contains(m.viewContent(), "Zoek straat") {
		t.Fatal("edited command not found")
	}
}
