package screens

import (
	"github.com/charmbracelet/x/ansi"
	"strings"
	"testing"
	"waterketen/internal/core"
	"waterketen/internal/ui/components"
)

func TestMapOverlaysUseActualContentBoxOnOpenAndResize(t *testing.T) {
	for _, size := range []struct{ w, h int }{{120, 34}, {48, 18}, {30, 10}, {24, 6}} {
		s := NewKaart("", nil, nil)
		s.SetSize(size.w, size.h)
		s.SetScene(&core.Scene{})
		s.OpenSearch()
		fitsHost(t, "composed map search", s.View(), max(1, size.w-4), max(1, size.h-2))
		if size.h >= 10 {
			panel := ansi.Strip(components.Panel("Kaart", s.View(), size.w, size.h, true))
			// A dialog must retain its lower border above the outer panel border.
			if strings.Count(panel, "╰") < 2 {
				t.Fatalf("%dx%d dialog lower border clipped: %s", size.w, size.h, panel)
			}
		}
		s.SetSize(30, 10)
		fitsHost(t, "resized map search", s.View(), 26, 8)
	}
}
