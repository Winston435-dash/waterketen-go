package components

import (
	"charm.land/bubbles/v2/key"
	"github.com/charmbracelet/x/ansi"
	"strings"
	"testing"
	"waterketen/internal/ui"
)

func TestShortHelpPreservesPrimaryActionsAndHelpRoute(t *testing.T) {
	for _, width := range []int{30, 48, 80, 120} {
		bar := ansi.Strip(NewCommandBar().View(width, []key.Binding{ui.Keys.Search, ui.Keys.Area, ui.Keys.Layers, ui.Keys.ZoomIn, ui.Keys.ZoomOut, ui.Keys.Fit, ui.Keys.PanelPalette}))
		for _, want := range []string{"/", "g", "t", "?"} {
			if !strings.Contains(bar, want) {
				t.Fatalf("width %d lost %q: %s", width, want, bar)
			}
		}
		if ansi.StringWidth(bar) != width || strings.Contains(bar, "\n") {
			t.Fatal("footer geometry drifted")
		}
	}
}

func TestSingleLineUsesGraphemeCellWidths(t *testing.T) {
	for _, value := range []string{"cafe\u0301", "👩‍🔬 gebied Rotterdam", "漢字 CBS-code", Strong("lange wijknaam WK059903")} {
		for _, width := range []int{1, 2, 4, 10, 24} {
			got := SingleLine(value, width)
			if ansi.StringWidth(got) != width {
				t.Fatalf("%q width %d rendered %d", value, width, ansi.StringWidth(got))
			}
			if ansi.StringWidth(value) > width && !strings.Contains(ansi.Strip(got), "…") {
				t.Fatal("truncation not marked")
			}
		}
	}
}
