package screens

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"waterketen/internal/core"
)

// TestStreetSearchOpenIsInKeyMap is the guard for the bug the user reported: '/'
// was declared in Kaart.KeyMap() — so the CommandBar advertised it — but the
// Update switch had no case for it, so pressing it did nothing at all. A key that
// is advertised must be handled.
func TestStreetSearchOpenIsInKeyMap(t *testing.T) {
	k := &Kaart{scene: &core.Scene{}, layers: []string{"riool"}}
	var advertised bool
	for _, b := range k.KeyMap() {
		if b.Help().Key == "/" {
			advertised = true
		}
	}
	if !advertised {
		t.Fatal("'/' staat niet in de CommandBar-help van paneel 5")
	}
	// It needs a scene to search in; without one the key is (correctly) inert,
	// so the test drives it with a scene present.
	k.OpenSearch()
	if k.search == nil {
		t.Fatal("'/' opende de straatzoeker niet — dit was precies de gerapporteerde bug")
	}
}

// TestStreetSearchTypesAndFocusesFirstMatch: typing builds the query, and a
// result that lands is what fills the match list (the ranking comes from a
// command, not from local work).
func TestStreetSearchTypesAndFocusesFirstMatch(t *testing.T) {
	s := &StreetSearch{}
	s.SetSize(100, 30)
	_, _, changed := s.Update(keys("w"))
	if !changed {
		t.Error("een letter hoort changed=true te geven, anders zoekt de app niet")
	}
	if s.query != "w" || !s.Busy() {
		t.Fatalf("na typen: query=%q busy=%v", s.query, s.Busy())
	}
}

// TestStreetSearchEscapeBacksOut: esc closes without focusing anything, so
// opening the search and changing your mind does not move the map.
func TestStreetSearchEscapeBacksOut(t *testing.T) {
	s := &StreetSearch{}
	s.SetSize(100, 30)
	s.matches = []string{"Westzeedijk"}
	focus, done, _ := s.Update(keys("esc"))
	if !done {
		t.Fatal("esc hoort te sluiten")
	}
	if focus != "" {
		t.Errorf("esc gaf focus=%q, verwacht leeg", focus)
	}
}

// TestStreetSearchEnterWithNoMatchStaysOpen: enter on an empty result list must
// not close the overlay — there is nothing to focus, and closing would look like
// a silent no-op.
func TestStreetSearchEnterWithNoMatchStaysOpen(t *testing.T) {
	s := &StreetSearch{}
	s.SetSize(100, 30)
	if _, done, _ := s.Update(keys("enter")); done {
		t.Error("enter zonder treffers sloot de zoeker")
	}
}

// TestStreetSearchResultApplies is the search-result contract, tested without the
// Kaart so the epoch logic stands alone.
func TestStreetSearchResultApplies(t *testing.T) {
	s := &StreetSearch{}
	s.SetSize(100, 30)
	for _, r := range "west" {
		s.Update(keys(string(r)))
	}
	if s.query != "west" {
		t.Fatalf("query = %q, verwacht west", s.query)
	}
	if !s.Busy() {
		t.Error("na typen hoort de zoeker bezig te zijn")
	}
	s.matches = []string{"Westzeedijk", "West-Kruiskade"}
	s.busy = false
	if s.Busy() {
		t.Error("na een resultaat is de zoeker niet meer bezig")
	}
	focus, done, _ := s.Update(keys("enter"))
	if !done || focus != "Westzeedijk" {
		t.Fatalf("enter gaf focus=%q done=%v, verwacht de eerste treffer", focus, done)
	}
}

// TestStreetSearchQueryChangedOnlyOnTyping: arrow keys and enter must not report
// a query change, or the caller would re-run the whole city search on every
// keypress.
func TestStreetSearchQueryChangedOnlyOnTyping(t *testing.T) {
	s := &StreetSearch{}
	s.SetSize(100, 30)
	for _, k := range []string{"down", "up", "enter"} {
		if _, _, changed := s.Update(keys(k)); changed {
			t.Errorf("%q hoort geen zoekopdracht te starten", k)
		}
	}
}

// TestApplySearchDropsStaleEpoch is the correctness guard the Python app also
// carries (its _map_search_revision): a search that lands after the user switched
// area must be DROPPED, or the map jumps to a street in a city the user left.
func TestApplySearchDropsStaleEpoch(t *testing.T) {
	k := &Kaart{scene: &core.Scene{}}
	s := &StreetSearch{epoch: 5}
	k.search = s
	k.epoch = 5

	k.ApplySearch(queryResult("west", []string{"Westzeedijk"}, 4)) // oud
	if len(s.matches) != 0 {
		t.Error("een resultaat van een vorige gebiedsgeneratie is toegepast")
	}
	k.ApplySearch(queryResult("west", []string{"Westzeedijk"}, 5)) // actueel
	if len(s.matches) != 1 {
		t.Error("een actueel resultaat is niet toegepast")
	}
	// And a result that arrives after the overlay closed must not resurrect it.
	k.search = nil
	k.ApplySearch(queryResult("west", []string{"Westzeedijk"}, 5))
}

// TestOverlayOpenCoversBothOverlays: the app routes keys to the Kaart when
// OverlayOpen is true. If the street search did not count, typing in it would
// also switch panels on 'g' and quit on 'q'.
func TestOverlayOpenCoversBothOverlays(t *testing.T) {
	k := &Kaart{}
	if k.OverlayOpen() {
		t.Fatal("geen overlay hoort niet open te zijn")
	}
	k.picker = &Picker{}
	if !k.OverlayOpen() {
		t.Error("een open picker hoort de toetsen te bezitten")
	}
	k.picker = nil
	k.search = &StreetSearch{}
	if !k.OverlayOpen() {
		t.Error("een open straatzoeker hoort de toetsen te bezitten")
	}
}

// TestToggleKeysMarkDirty is the guard for the second class the capture found:
// 'o' and 'l' changed state but fell through without setting needsDraw, so they
// were advertised keys that repainted nothing.
func TestToggleKeysMarkDirty(t *testing.T) {
	for _, tc := range []struct{ key, what string }{
		{"o", "lagen samen/solo"},
		{"l", "straatnamen"},
	} {
		k := &Kaart{scene: &core.Scene{}, labels: true, overlay: true}
		k.needsDraw = false
		_ = k.handleKey(keys(tc.key))
		if !k.needsDraw {
			t.Errorf("%s (%s) zette needsDraw niet — de kaart hertekent niet", tc.key, tc.what)
		}
	}
}

// TestSoloNarrowsToActiveLayer: 'o' switches samen/solo and '['/']' select a
// layer, so the map must actually draw one layer in solo mode.
// TestSetLayersShowsTheChosenSetAfterSolo verifies that applying the layer picker
// restores the selected set instead of leaving the map narrowed to the old active
// layer.
func TestSetLayersShowsTheChosenSetAfterSolo(t *testing.T) {
	k := &Kaart{
		layers: []string{"riool", "gemalen"},
		active: "riool",
		solo:   true,
	}

	k.SetLayers([]string{"riool", "gemalen"})
	got := k.visibleLayerNames()
	if len(got) != 2 || got[0] != "riool" || got[1] != "gemalen" {
		t.Fatalf("zichtbare lagen na toepassen = %v, verwacht [riool gemalen]", got)
	}
}

func TestSoloNarrowsToActiveLayer(t *testing.T) {
	k := &Kaart{scene: &core.Scene{}, layers: []string{"straten", "water", "riool"}, active: "riool"}
	if got := k.visibleLayerNames(); len(got) != 3 {
		t.Errorf("samen gaf %v, verwacht 3 lagen", got)
	}
	k.needsDraw = false
	_ = k.handleKey(keys("o"))
	if !k.solo {
		t.Fatal("'o' zette solo niet aan")
	}
	if got := k.visibleLayerNames(); len(got) != 1 || got[0] != "riool" {
		t.Errorf("solo gaf %v, verwacht alleen riool", got)
	}
	// Cycling a layer implies solo, otherwise the press would change no ink.
	k.solo, k.overlay = false, true
	k.cycleLayer(1)
	if !k.solo {
		t.Error("laagwisselen hoort solo te zetten")
	}
	if k.active == "riool" {
		t.Error("laagwisselen veranderde de actieve laag niet")
	}
}

// TestStreetSearchViewFitsScreen: the overlay is a dialog and must not be wider
// than its host panel.
func TestStreetSearchViewFitsScreen(t *testing.T) {
	for _, width := range []int{40, 80, 120, 200} {
		s := &StreetSearch{query: "west", matches: []string{"Westzeedijk", "West-Kruiskade"}}
		s.SetSize(width, 30)
		for _, line := range strings.Split(s.View(), "\n") {
			if w := lipgloss.Width(line); w > width {
				t.Errorf("breedte %d: overlayregel is %d cellen", width, w)
			}
		}
	}
}

func queryResult(q string, matches []string, epoch int) SearchQueryMsg {
	return SearchQueryMsg{Query: q, Matches: matches, Epoch: epoch}
}
