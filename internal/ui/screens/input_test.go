package screens

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"waterketen/internal/core"
)

// space is the message Bubble Tea actually produces for the space bar: KeySpace,
// NOT KeyRunes. Getting this wrong is the whole bug, so the tests build it exactly
// as bubbletea/key.go does (len(runes)==1 && runes[0]==' ' → KeySpace).
func space() tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: tea.KeySpace, Text: " "}
}

// pasted is what a bracketed paste produces: KeyRunes with Paste set and the whole
// string in Runes.
func pasted(s string) tea.PasteMsg {
	return tea.PasteMsg{Content: s}
}

// TestTypedTextAcceptsSpaceAndRunes is the guard for the reported bug: '/' dropped
// every space, so "west zeedijk" was typed as "westzeedijk" and a leading space did
// nothing. A space is a KeySpace event, not a KeyRunes event.
func TestTypedTextAcceptsSpaceAndRunes(t *testing.T) {
	if got := TypedText(space()); got != " " {
		t.Errorf("TypedText(spatie) = %q, verwacht \" \"", got)
	}
	if got := TypedText(tea.KeyPressMsg{Code: 'a', Text: "a"}); got != "a" {
		t.Errorf("TypedText('a') = %q", got)
	}
	if got := TypedText(pasted("west zeedijk")); got != "west zeedijk" {
		t.Errorf("TypedText(plak) = %q", got)
	}
	// Non-text keys contribute nothing.
	for _, m := range []tea.KeyPressMsg{
		{Code: tea.KeyEnter}, {Code: tea.KeyEscape}, {Code: tea.KeyUp},
		{Code: tea.KeyBackspace}, {Code: tea.KeyTab},
	} {
		if got := TypedText(m); got != "" {
			t.Errorf("TypedText(%v) = %q, verwacht leeg", m.Code, got)
		}
	}
}

// TestStreetSearchAcceptsSpaces is the end-to-end version of the bug report: type a
// street name with a space in it and the query must contain the space.
func TestStreetSearchAcceptsSpaces(t *testing.T) {
	s := &StreetSearch{}
	s.SetSize(100, 30)
	for _, r := range "west" {
		s.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
	}
	if _, _, changed := s.Update(space()); !changed {
		t.Fatal("een spatie startte geen zoekopdracht — dit was de gerapporteerde bug")
	}
	for _, r := range "zeedijk" {
		s.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
	}
	if s.query != "west zeedijk" {
		t.Errorf("query = %q, verwacht \"west zeedijk\"", s.query)
	}
}

// TestStreetSearchAcceptsALeadingSpace: pressing space first must not be a no-op.
func TestStreetSearchAcceptsALeadingSpace(t *testing.T) {
	s := &StreetSearch{}
	s.SetSize(100, 30)
	_, _, changed := s.Update(space())
	if !changed || s.query != " " {
		t.Errorf("voorloopspatie: query=%q changed=%v", s.query, changed)
	}
}

// TestStreetSearchBackspaceRemovesARune: backspace must not split a multi-byte
// character, or the query (and the search on it) is corrupted.
func TestStreetSearchBackspaceRemovesARune(t *testing.T) {
	s := &StreetSearch{}
	s.SetSize(100, 30)
	for _, r := range "café" {
		s.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
	}
	s.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})
	if s.query != "caf" {
		t.Errorf("na backspace: query=%q, verwacht \"caf\" (geen halve é)", s.query)
	}
}

// TestStreetSearchAcceptsPaste: a pasted street name arrives as one KeyRunes message.
func TestStreetSearchAcceptsPaste(t *testing.T) {
	s := &StreetSearch{}
	s.SetSize(100, 30)
	s.Update(pasted("west zeedijk"))
	if s.query != "west zeedijk" {
		t.Errorf("na plakken: query=%q", s.query)
	}
}

// TestAreaPickerAcceptsSpaces: the area search must accept a space too (a name can
// contain one, and typing "s gravendijkwal" is a normal thing to do).
func TestAreaPickerAcceptsSpaces(t *testing.T) {
	p := NewAreaPicker([]core.Area{
		{Code: "WK059903", Name: "Delfshaven", Kind: core.KindWijk},
		{Code: "BU05990001", Name: "s Gravendijkwal", Kind: core.KindBuurt},
	})
	p.SetSize(100, 30)
	for _, r := range "s " {
		p.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
	}
	if p.query != "s " {
		t.Errorf("query = %q, verwacht \"s \"", p.query)
	}
	// And the toggling layer picker still treats space as a COMMAND, not text.
	lp := NewLayerPicker([]string{"riool", "water"}, nil, core.DrawOrder)
	lp.SetSize(100, 30)
	lp.Update(space())
	if lp.query != "" {
		t.Errorf("lagenkiezer kreeg tekst %q van een spatie", lp.query)
	}
	if !lp.filtered[0].mark {
		t.Error("spatie schakelde de laag niet aan")
	}
}
