package app

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
)

// mobile is the set of terminal geometries this app must survive. It is not a
// guess: 118x36 is the desktop reference, ~72x42 is a phone in landscape (wide
// but few rows once the soft keyboard is up), ~48x26 is Termux in portrait, and
// the last two are the "someone resized to a sliver" cases that used to produce
// negative panel widths.
var mobile = []struct {
	name string
	w, h int
}{
	{"desktop", 118, 36},
	{"laptop", 100, 30},
	{"tablet", 84, 24},
	{"landscape", 72, 42},
	{"portrait", 48, 26},
	{"tiny", 40, 20},
	{"sliver", 30, 12},
	{"micro", 24, 8},
}

// renderSize draws the whole app at one terminal size and returns its lines.
func renderSize(t *testing.T, w, h int) []string {
	t.Helper()
	m := New(t.TempDir())
	m.width, m.height, m.ready = w, h, true
	m = m.applySize()
	return strings.Split(m.viewContent(), "\n")
}

// TestViewNeverExceedsTerminal is the mobile contract: lipgloss renders each
// element to its own width, so a box one cell too wide makes the terminal WRAP —
// and a wrap on a phone is not a cosmetic glitch, it pushes the command bar off
// the bottom of the screen and scrolls the header away. Every line must fit the
// width, and there must not be more lines than rows.
func TestViewNeverExceedsTerminal(t *testing.T) {
	for _, s := range mobile {
		lines := renderSize(t, s.w, s.h)
		if len(lines) > s.h {
			t.Errorf("%s %dx%d: view heeft %d regels, past niet in %d rijen",
				s.name, s.w, s.h, len(lines), s.h)
		}
		for i, line := range lines {
			if w := lipgloss.Width(line); w > s.w {
				t.Errorf("%s %dx%d: regel %d is %d cellen breed (max %d)\n  %q",
					s.name, s.w, s.h, i, w, s.w, line)
			}
		}
	}
}

// TestPanelBoxMatchesRenderedPanel keeps SetSize and rendering on the same
// full-width box at every supported terminal geometry.
func TestPanelBoxMatchesRenderedPanel(t *testing.T) {
	for _, s := range mobile {
		m := New(t.TempDir())
		m.width, m.height, m.ready = s.w, s.h, true
		l := m.layout()
		wantW, wantH := l.panelW, l.panelH
		if gotW, gotH := m.panelBox(); gotW != wantW || gotH != wantH {
			t.Errorf("%s %dx%d: panelBox=(%d,%d), layout=(%d,%d)",
				s.name, s.w, s.h, gotW, gotH, wantW, wantH)
		}
		// Navigation lives in the header, so content owns every terminal column.
		if l.panelW != s.w {
			t.Errorf("%s: paneel=%d, verwacht volledige breedte %d", s.name, l.panelW, s.w)
		}
		if l.panelW < 1 || l.panelH < 1 {
			t.Errorf("%s: paneelbox is niet-positief: %dx%d", s.name, l.panelW, l.panelH)
		}
	}
}
