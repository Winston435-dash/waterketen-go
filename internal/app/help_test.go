package app

import (
	"strings"
	"testing"
	"github.com/charmbracelet/x/ansi"
	"waterketen/internal/ui/screens"
)

func TestPickerHelpAdvertisesOnlyOwnedActions(t *testing.T) {
	m := paletteModel(t, screens.ScreenKaart)
	m = press(m, appKey("/"))
	footer := strings.Split(ansi.Strip(m.viewContent()), "\n")
	line := footer[len(footer)-1]
	if strings.Contains(line, "ctrl+p") || strings.Contains(line, "? hulp") { t.Fatalf("overlay advertises inaccessible globals: %s",line) }
	if !strings.Contains(line, "esc") { t.Fatal("overlay cancellation not visible") }
}
