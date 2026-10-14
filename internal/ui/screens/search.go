package screens

import (
	"strconv"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

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
func (s *StreetSearch) Update(event tea.Msg) (focus string, done, changed bool) {
	msg, ok := event.(tea.KeyPressMsg)
	if !ok {
		if text := TypedText(event); text != "" {
			s.query += text
			s.busy = true
			return "", false, true
		}
		return "", false, false
	}
	switch {
	case key.Matches(msg, ui.Keys.Back):
		return "", true, false
	case key.Matches(msg, ui.Keys.Enter):
		if len(s.matches) == 0 {
			return "", false, false
		}
		return s.matches[0], true, false
	case msg.Code == tea.KeyBackspace:
		if s.query == "" {
			return "", false, false
		}
		// Backspace removes one RUNE, not one byte: a street name may end in a
		// multi-byte character, and s.query[:len-1] would split it and corrupt the
		// query (and the search that runs on it).
		runes := []rune(s.query)
		s.query = string(runes[:len(runes)-1])
		s.busy = true
		return "", false, true
	default:
		// TypedText, not `case tea.KeyRunes`: a SPACE arrives as KeySpace, so a
		// KeyRunes-only handler silently dropped every space — you could not type
		// "west zeedijk" in '/' at all.
		if text := TypedText(msg); text != "" {
			s.query += text
			s.busy = true
			return "", false, true
		}
	}
	return "", false, false
}

// View renders the search overlay: the query, the ranked matches, and the hint.
// Like the picker, every line is clipped to the dialog width so a long street name
// cannot wrap an entry onto two rows on a phone.
func (s *StreetSearch) View() string {
	innerW := s.innerWidth()
	rows := s.windowRows()
	var b strings.Builder
	b.WriteString(components.SingleLine(components.Strong("straat: "+s.query), innerW))
	b.WriteByte('\n')
	b.WriteByte('\n')
	if s.busy {
		b.WriteString(components.SingleLine(components.Muted("  zoeken…"), innerW))
		b.WriteByte('\n')
	} else if len(s.matches) == 0 {
		msg := "  typ een straatnaam"
		if s.query != "" {
			msg = "  geen straat gevonden"
		}
		b.WriteString(components.SingleLine(components.Muted(msg), innerW))
		b.WriteByte('\n')
	}
	for i, name := range s.matches {
		if i >= rows {
			break
		}
		line := "  " + name
		if i == 0 {
			b.WriteString(components.Selected(components.SingleLine(line, innerW), innerW))
		} else {
			b.WriteString(components.SingleLine(components.Muted(line), innerW))
		}
		b.WriteByte('\n')
	}
	b.WriteByte('\n')
	b.WriteString(components.SingleLine(
		components.Muted("typ om te zoeken · enter focust · esc sluit"), innerW))
	title := "Straat zoeken"
	if len(s.matches) > 0 {
		title = "Straat zoeken · " + strconv.Itoa(len(s.matches)) + " treffers"
	}
	return components.Modal("", title, b.String(), s.w, s.h, s.dialogWidth())
}

// dialogWidth is the dialog's outer width. The cap is p.w-8 (see Picker.dialogWidth):
// the dialog lives inside the panel's content area, so a dialog wider than its host
// makes lipgloss wrap the panel border instead of the dialog.
func (s *StreetSearch) dialogWidth() int {
	return components.DialogWidth(72, s.w)
}

func (s *StreetSearch) innerWidth() int {
	w := s.dialogWidth() - 4
	if w < 8 {
		w = 8
	}
	return w
}

// windowRows is how many match rows fit without pushing the hint off the bottom.
func (s *StreetSearch) windowRows() int {
	rows := s.h - 6
	if rows > 8 {
		rows = 8
	}
	if rows < 1 {
		rows = 1
	}
	return rows
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
	w, h := s.overlaySize()
	s.search.SetSize(w, h)
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

// ApplySearch stores only a result for the current query and atlas generation.
func (s *Kaart) ApplySearch(msg SearchQueryMsg) {
	if s.search == nil || msg.Epoch != s.epoch || msg.Epoch != s.search.epoch || msg.Query != s.search.query {
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
