package core

import "testing"

// TestCellColourResolvesEveryStyle is the guard that keeps core the single
// owner of data ink: every CellStyle the renderer can emit must resolve to a
// non-empty hex, or the map would silently drop a colour.
func TestCellColourResolvesEveryStyle(t *testing.T) {
	styles := []CellStyle{
		StyleBlank, StyleGrid, StyleNorth, StyleScale, StyleLabel, StyleHighlight,
		"bob:very-deep", "bob:deep", "bob:mid", "bob:shallow", "bob:above", "bob:unknown",
		"hex:#abcdef", "material:Onbekend", "iets-onbekend",
	}
	for _, layer := range []string{"riool", "drinkwater", "drinkwaterleidingen", "straten", "water"} {
		styles = append(styles, StyleForLayer(layer))
	}
	for _, s := range styles {
		if got := CellColour(s); got == "" {
			t.Errorf("CellColour(%q) = leeg", s)
		}
	}
}

// TestCellColourMatchesCoreTables guards against a second, drifting colour
// table: the resolver must return exactly what the atlas tables define.
func TestCellColourMatchesCoreTables(t *testing.T) {
	if got, want := CellColour(StyleForLayer("riool")), layerColours["riool"]; got != want {
		t.Errorf("riool-ink=%q, wil %q", got, want)
	}
	if got, want := CellColour("bob:deep"), bobBands["deep"]; got != want {
		t.Errorf("bob-deep-ink=%q, wil %q", got, want)
	}
	if got, want := CellColour("hex:#123456"), "#123456"; got != want {
		t.Errorf("hex-ink=%q, wil %q", got, want)
	}
}
