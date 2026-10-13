package screens

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"

	"waterketen/internal/core"
)

// hostSizes are the terminal geometries the overlays must survive inside the panel
// content area. The last few are the "someone resized to a sliver" cases.
var hostSizes = []struct {
	name string
	w, h int
}{
	{"desktop", 116, 34},
	{"landscape", 70, 40},
	{"portrait", 46, 24},
	{"tiny", 38, 18},
	{"sliver", 28, 12},
	{"micro", 22, 8},
}

func sampleAreas() []core.Area {
	return []core.Area{
		{Code: "WK059903", Name: "Delfshaven", Kind: core.KindWijk},
		{Code: "BU05990320", Name: "Delfshaven", Kind: core.KindBuurt},
		{Code: "BU05992459", Name: "Bedrijvenpark Noord-West", Kind: core.KindBuurt},
		{Code: "BU05999998", Name: "Buitenwater", Kind: core.KindBuurt},
	}
}

// fitsHost asserts an overlay occupies at most its host box: never more rows than
// the host has, never a line wider than the host. It returns the lines so a caller
// can make further assertions.
func fitsHost(t *testing.T, what string, view string, w, h int) []string {
	t.Helper()
	lines := strings.Split(view, "\n")
	if len(lines) > h {
		t.Errorf("%s op %dx%d: %d regels, past niet", what, w, h, len(lines))
	}
	for i, line := range lines {
		if lw := lipgloss.Width(line); lw > w {
			t.Errorf("%s op %dx%d: regel %d is %d cellen: %q", what, w, h, i, lw, line)
		}
	}
	return lines
}

// TestAreaPickerFitsItsHost is the mobile contract for the area dialog. It renders
// INSIDE the panel's content area, so a dialog wider than its host made lipgloss
// wrap the PANEL BORDER instead — the map's frame broke on a 48-column phone, and
// only a real capture showed it.
func TestAreaPickerFitsItsHost(t *testing.T) {
	for _, host := range hostSizes {
		p := NewAreaPicker(sampleAreas())
		p.SetSize(host.w, host.h)
		fitsHost(t, "gebiedkiezer", p.View(), host.w, host.h)
	}
}

// TestLayerPickerFitsItsHost does the same for the layer dialog, where the checkbox
// prefix makes every entry two cells wider.
func TestLayerPickerFitsItsHost(t *testing.T) {
	for _, host := range hostSizes {
		p := NewLayerPicker(core.DrawOrder, []string{"straten", "riool"}, core.DrawOrder)
		p.SetSize(host.w, host.h)
		fitsHost(t, "lagenkiezer", p.View(), host.w, host.h)
	}
}

// TestStreetSearchFitsItsHost mirrors the picker contract for '/'.
func TestStreetSearchFitsItsHost(t *testing.T) {
	for _, host := range hostSizes {
		s := &StreetSearch{
			query:   "west",
			matches: []string{"Westzeedijk", "West-Kruiskade", "West-Pietermanstraat"},
		}
		s.SetSize(host.w, host.h)
		fitsHost(t, "straatzoeker", s.View(), host.w, host.h)
	}
}

// TestPickerEntriesAreOneRow: a long area label must be CLIPPED, not wrapped. A
// wrapped entry costs two rows, so half as many entries fit and the list stops
// looking like a list — measured with the CBS-code list on a 48-column phone.
func TestPickerEntriesAreOneRow(t *testing.T) {
	areas := []core.Area{
		{Code: "BU05992459", Name: "Bedrijvenpark Noord-West Met Een Zeer Lange Naam", Kind: core.KindBuurt},
	}
	p := NewAreaPicker(areas)
	p.SetSize(48, 26)
	rows := 0
	for _, line := range strings.Split(p.View(), "\n") {
		if strings.Contains(line, "BU05992459") {
			rows++
		}
	}
	if rows != 1 {
		t.Errorf("een lange labelinvoer besloeg %d regels, verwacht 1", rows)
	}
}

// TestPickerWindowShrinksWithHost: the number of visible list rows must follow the
// host height, or on a short terminal the entries push the hint (and at the extreme,
// the dialog's own bottom border) off the screen.
func TestPickerWindowShrinksWithHost(t *testing.T) {
	p := NewAreaPicker(sampleAreas())
	tall, short := 40, 14
	p.SetSize(60, tall)
	grew := p.windowRows()
	p.SetSize(60, short)
	shrunk := p.windowRows()
	if !(shrunk < grew) {
		t.Errorf("windowRows op hoogte %d = %d, op %d = %d — hoort te krimpen",
			short, shrunk, tall, grew)
	}
	if shrunk < 1 {
		t.Errorf("windowRows zakte onder 1 (%d)", shrunk)
	}
}

// TestDialogNeverExceedsItsPanel is the arithmetic guard behind the border-wrap
// bug: whatever the host width, the dialog must leave room for the panel chrome.
func TestDialogNeverExceedsItsPanel(t *testing.T) {
	p := NewAreaPicker(sampleAreas())
	for _, host := range hostSizes {
		p.SetSize(host.w, host.h)
		if d := p.dialogWidth(); d > host.w {
			t.Errorf("dialoog %d breder dan host %d (host %dx%d)", d, host.w, host.w, host.h)
		}
		if d := p.dialogWidth(); d < 20 && host.w >= 28 {
			t.Errorf("dialoog %d te smal voor host %d", d, host.w)
		}
	}
}
