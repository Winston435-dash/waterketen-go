package screens

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"waterketen/internal/core"
)

func keys(s string) tea.KeyMsg {
	switch s {
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEscape}
	case " ":
		return tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}}
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func areaPicker(t *testing.T) *Picker {
	t.Helper()
	p := NewAreaPicker([]core.Area{
		{Code: "GM0599", Name: "Rotterdam", Kind: core.KindGemeente},
		{Code: "WK059903", Name: "Delfshaven", Kind: core.KindWijk},
		{Code: "BU05990320", Name: "Delfshaven", Kind: core.KindBuurt},
	})
	p.SetSize(100, 30)
	return p
}

// indexOfKey is where a slug sits in the picker's list, so a test can navigate
// to it without hardcoding a position that DrawOrder owns.
func indexOfKey(p *Picker, key string) int {
	for i, it := range p.filtered {
		if it.key == key {
			return i
		}
	}
	return -1
}

// TestAreaPickerTypingAndEnter is the guard for the bug real capture found: an
// UNBOUND key.Binding matches nothing, so enter silently did nothing and the
// picker could not be confirmed at all.
func TestAreaPickerTypingAndEnter(t *testing.T) {
	p := areaPicker(t)
	for _, r := range "delf" {
		if _, done := p.Update(keys(string(r))); done {
			t.Fatal("typen mag de picker niet sluiten")
		}
	}
	if len(p.filtered) != 2 {
		t.Fatalf("'delf' gaf %d treffers, verwacht 2 (wijk + buurt)", len(p.filtered))
	}
	res, done := p.Update(keys("enter"))
	if !done || !res.Confirm {
		t.Fatalf("enter hoort te bevestigen, kreeg done=%v res=%+v", done, res)
	}
	if res.Toggled {
		t.Error("een gebiedskeuze is geen toggling")
	}
	if res.Key != "WK059903" {
		t.Errorf("enter koos %q, verwacht WK059903", res.Key)
	}
}

// TestAreaPickerCursorWraps: down/up walk the filtered list and wrap.
func TestAreaPickerCursorWraps(t *testing.T) {
	p := areaPicker(t)
	p.Update(keys("down"))
	p.Update(keys("down"))
	p.Update(keys("down")) // 3 items -> wraps to 0
	res, _ := p.Update(keys("enter"))
	if res.Key != "GM0599" {
		t.Errorf("na 3x down = %q, verwacht GM0599 (gewrapt)", res.Key)
	}
}

// TestAreaPickerEscapeCancels: esc must not report a selection, or opening the
// picker and backing out would silently move the map.
func TestAreaPickerEscapeCancels(t *testing.T) {
	p := areaPicker(t)
	res, done := p.Update(keys("esc"))
	if !done {
		t.Fatal("esc hoort te sluiten")
	}
	if res.Confirm || res.Key != "" {
		t.Errorf("esc gaf %+v, verwacht een leeg (niet-bevestigd) resultaat", res)
	}
}

// TestLayerPickerTogglesAndApplies is the guard for the second bug: the layer
// picker's only exit was esc, which CANCELS — so a toggled set could never be
// applied. enter must apply, esc must cancel, and space must toggle.
func TestLayerPickerTogglesAndApplies(t *testing.T) {
	avail := []string{"straten", "water", "riool", "drinkwaterleidingen"}
	p := NewLayerPicker(avail, []string{"straten", "riool"}, core.DrawOrder)
	p.SetSize(100, 30)

	if !p.Toggling() {
		t.Fatal("de lagenkiezer hoort een toggling-picker te zijn")
	}
	if got := p.result(true).Marked; len(got) != 2 {
		t.Fatalf("beginselectie = %v, verwacht 2 lagen", got)
	}

	// The list is in DrawOrder, NOT in the order the layers were passed in, so
	// navigate to the row by key instead of assuming where it sits.
	target := indexOfKey(p, "straten")
	if target < 0 {
		t.Fatal("straten staat niet in de lagenlijst")
	}
	for i := 0; i < target; i++ {
		p.Update(keys("down"))
	}
	p.Update(keys(" "))
	res, done := p.Update(keys("enter"))
	if !done || !res.Confirm || !res.Toggled {
		t.Fatalf("enter hoort de lagen toe te passen, kreeg done=%v res=%+v", done, res)
	}
	if len(res.Marked) != 1 || res.Marked[0] != "riool" {
		t.Errorf("na uitzetten van straten = %v, verwacht alleen riool", res.Marked)
	}

	// An EMPTY set is a legitimate answer and must stay distinguishable from an
	// area pick: Toggled, not "no marks so it must be an area".
	p2 := NewLayerPicker(avail, nil, core.DrawOrder)
	p2.SetSize(100, 30)
	res2, _ := p2.Update(keys("enter"))
	if !res2.Toggled || len(res2.Marked) != 0 {
		t.Errorf("lege lagenlijst = %+v, verwacht Toggled met 0 lagen", res2)
	}

	// esc cancels: never apply a half-finished edit.
	p3 := NewLayerPicker(avail, []string{"riool"}, core.DrawOrder)
	p3.SetSize(100, 30)
	p3.Update(keys(" "))
	res3, _ := p3.Update(keys("esc"))
	if res3.Confirm {
		t.Errorf("esc gaf %+v, verwacht niet-bevestigd", res3)
	}
}

// TestLayerPickerNeverTypes: in the layer picker, letters must not build a
// query — they are not a search there, and a query would silently hide layers.
func TestLayerPickerNeverTypes(t *testing.T) {
	p := NewLayerPicker([]string{"riool", "water"}, nil, core.DrawOrder)
	p.SetSize(100, 30)
	p.Update(keys("r"))
	p.Update(keys("i"))
	if p.query != "" {
		t.Errorf("lagenkiezer kreeg query %q", p.query)
	}
	if len(p.filtered) != 2 {
		t.Errorf("lagenlijst is gefilterd tot %d", len(p.filtered))
	}
}

// TestPickerViewFitsScreen: the overlay is a dialog, so it must never be wider
// than the panel it floats over.
func TestPickerViewFitsScreen(t *testing.T) {
	for _, width := range []int{40, 80, 120, 200} {
		p := areaPicker(t)
		p.SetSize(width, 30)
		out := p.View()
		for _, line := range strings.Split(out, "\n") {
			if w := lipgloss.Width(line); w > width {
				t.Errorf("breedte %d: overlayregel is %d cellen, past niet", width, w)
			}
		}
	}
}

func TestLayerPickerSeparatesIMWAFromExistingPumps(t *testing.T) {
	available := core.OrderLayers([]string{"gemalen", "oppervlaktewatergemalen", "hwh-kunstwerken-duiker", "hwh-kunstwerken-gemaal", "hwh-kunstwerken-sluis", "hwh-kunstwerken-stuw"})
	p := NewLayerPicker(available, []string{"gemalen"}, available)
	p.SetSize(68, 40)
	for _, item := range p.all {
		if item.key == "gemalen" && (!item.mark || item.label != "gemalen") {
			t.Fatal("existing pump selection changed")
		}
		if strings.HasPrefix(item.key, "hwh-kunstwerken-") && (item.mark || !strings.HasPrefix(item.label, "IMWA — ")) {
			t.Fatal("IMWA source unclear or enabled without selection", item)
		}
	}
	p.cursor = indexOfKey(p, "hwh-kunstwerken-gemaal")
	p.Update(keys(" "))
	result, done := p.Update(keys("enter"))
	if !done || !result.Confirm || len(result.Marked) != 2 || result.Marked[1] != "hwh-kunstwerken-gemaal" {
		t.Fatalf("IMWA display label lost source key: %+v", result)
	}
}

func TestLayerPickerDisclosesAndReachesLayersBelowTheWindow(t *testing.T) {
	available := core.OrderLayers([]string{"groen", "water", "gebouwen", "spoor", "straten", "riool", "drinkwaterleidingen", "putten", "gemalen", "oppervlaktewatergemalen", "hwh-kunstwerken-duiker", "hwh-kunstwerken-gemaal", "hwh-kunstwerken-sluis", "hwh-kunstwerken-stuw"})
	p := NewLayerPicker(available, nil, available)
	p.SetSize(68, 40)
	if !strings.Contains(p.View(), "↑/↓ meer lagen · 1–12 van 14") {
		t.Fatal("hidden layers have no scroll indication")
	}
	for range 13 {
		p.Update(keys("down"))
	}
	if !strings.Contains(p.View(), "IMWA — sluizen") || !strings.Contains(p.View(), "IMWA — stuwen") || !strings.Contains(p.View(), "3–14 van 14") {
		t.Fatal("last IMWA layers not reachable")
	}
	p.Update(keys(" "))
	result, _ := p.Update(keys("enter"))
	if len(result.Marked) != 1 || result.Marked[0] != "hwh-kunstwerken-stuw" {
		t.Fatalf("last layer not selectable: %+v", result)
	}
}
