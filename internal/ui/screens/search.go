package screens

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"waterketen/internal/ui"
	"waterketen/internal/ui/components"
)

// StreetSearch is the '/' overlay: type part of a street name, see ranked
// matches, enter to focus the map on the first one. Two data worlds again, and
// the split is honest: the library world searches the snapshot's straten layer
// over the WHOLE city (a wijk view must not hide streets outside it), the managed
// world searches only what is already loaded — which is all it has.
//
// The search itself is the only part that runs off the event loop; ranking and
// filtering are negligible once the name list is in.
type StreetSearch struct {
	query   string
	matches []string
	busy    bool
	// epoch is the area generation this search belongs to. A result that lands
	// after the user switched area must be DROPPED: focusing a street from the
	// previous area would move the map to a place the user already left.
	epoch int
	w, h  int
}

// SearchQueryMsg carries a street-name list back from a search run.
type SearchQueryMsg struct {
	Query   string
	Matches []string
	Epoch   int
}

// busy reports whether a search is in flight.
func (s *StreetSearch) Busy() bool { return s != nil && s.busy }

// SetSize records the overlay's host size.
func (s *StreetSearch) SetSize(w, h int) { s.w, s.h = w, h }

// Update handles one key and reports: the street to focus ("" = none), whether
// the overlay should close, and whether the QUERY changed so the caller knows to
// run a new search. Without that last flag the caller would search on every
// keystroke — including the arrows and enter.
func (s *StreetSearch) Update(msg tea.KeyMsg) (focus string, done, changed bool) {
	switch {
	case key.Matches(msg, ui.Keys.Back):
		return "", true, false
	case key.Matches(msg, ui.Keys.Enter):
		if len(s.matches) == 0 {
			return "", false, false
		}
		return s.matches[0], true, false
	case msg.Type == tea.KeyBackspace:
		if s.query == "" {
			return "", false, false
		}
		s.query = s.query[:len(s.query)-1]
		s.busy = true
		return "", false, true
	case msg.Type == tea.KeyRunes:
		s.query += string(msg.Runes)
		s.busy = true
		return "", false, true
	}
	return "", false, false
}

// View renders the search overlay: the query, the ranked matches, and the hint.
func (s *StreetSearch) View() string {
	var b strings.Builder
	b.WriteString(components.Strong("straat: " + s.query))
	b.WriteByte('\n')
	b.WriteByte('\n')
	if s.busy {
		b.WriteString(components.MutedText("  zoeken…"))
		b.WriteByte('\n')
	} else if len(s.matches) == 0 {
		if s.query == "" {
			b.WriteString(components.MutedText("  typ een straatnaam"))
		} else {
			b.WriteString(components.MutedText("  geen straat gevonden"))
		}
		b.WriteByte('\n')
	}
	for i, name := range s.matches {
		line := "  " + name
		if i == 0 {
			b.WriteString(components.Selected(line, max(20, s.w-14)))
		} else {
			b.WriteString(components.MutedText(line))
		}
		b.WriteByte('\n')
	}
	b.WriteByte('\n')
	b.WriteString(components.MutedText("typ om te zoeken · enter focust · esc sluit"))
	title := "Straat zoeken"
	if len(s.matches) > 0 {
		title = "Straat zoeken · " + itoa(len(s.matches)) + " treffers"
	}
	return components.Modal("", title, b.String(), s.w, s.h, min(72, max(30, s.w-8)))
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}

// OpenSearch opens the '/' overlay. It needs only a loaded scene: the library
// world searches the whole city's straten layer, but the MANAGED world can search
// too (core.Scene.Search over the loaded street names), and gating on the atlas
// would have made '/' dead for anyone without a library snapshot.
func (s *Kaart) OpenSearch() {
	if s.scene == nil {
		return
	}
	s.search = &StreetSearch{epoch: s.epoch}
	s.search.SetSize(s.w, s.h)
}

// CloseSearch closes the overlay.
func (s *Kaart) CloseSearch() { s.search = nil }

// StreetSearchOpen reports whether the '/' overlay owns the keyboard.
func (s *Kaart) StreetSearchOpen() bool { return s.search != nil }

// SearchCmd runs the street-name query off the event loop.
func (s *Kaart) SearchCmd(query string) tea.Cmd {
	atlas, epoch := s.atlas, s.epoch
	if atlas == nil {
		// Managed world: the names are already in the loaded scene.
		scene := s.scene
		return func() tea.Msg {
			return SearchQueryMsg{Query: query, Matches: scene.Search(query), Epoch: epoch}
		}
	}
	return func() tea.Msg {
		return SearchQueryMsg{Query: query, Matches: atlas.StreetNames(query), Epoch: epoch}
	}
}

// ApplySearch stores a finished search result. A result from an earlier area
// generation is dropped, not applied.
func (s *Kaart) ApplySearch(msg SearchQueryMsg) {
	if s.search == nil || msg.Epoch != s.epoch {
		return
	}
	s.search.busy = false
	s.search.matches = msg.Matches
}

// FocusStreetCmd focuses the map on one street name. In the library world this
// refits the view and reloads the viewport; in the managed world it only needs a
// redraw.
func (s *Kaart) FocusStreetCmd(name string) tea.Cmd {
	if name == "" {
		return nil
	}
	s.street = name
	s.selected = nil
	if s.atlas != nil {
		w, h := s.innerSize()
		view, ok := s.atlas.StreetView(name, w, h)
		if !ok {
			return nil
		}
		s.view = view
		return s.requestLoad()
	}
	if s.scene == nil {
		return nil
	}
	w, h := s.innerSize()
	if view, ok := s.scene.StreetView(name, w, h); ok {
		s.view = view
	}
	s.needsDraw = true
	return nil
}

// focusStreet is the street the map highlights; the empty string means none.
func (s *Kaart) focusStreet() string { return s.street }
