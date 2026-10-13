package screens

import (
	"context"
	"fmt"
	"net/url"
	"slices"
	"sort"
	"strings"
	"time"
	"unicode"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"waterketen/internal/core"
	"waterketen/internal/ui"
	"waterketen/internal/ui/components"
)

type queryMode uint8

const (
	querySummary queryMode = iota
	queryObjects
	queryDetails
)
const queryPageSize = 50

// Query owns one snapshot and one cancellable operation. Commands capture
// immutable requests; only matching generations can commit their results.
type Query struct {
	root                                   string
	library                                *core.Library
	areas                                  []core.Area
	area                                   core.Area
	layers                                 []core.QueryLayer
	selected                               []string
	loaded, opening, running, pageLoading  bool
	generation                             int
	cancel                                 context.CancelFunc
	result                                 *core.QueryResult
	err                                    error
	picker                                 *Picker
	mode                                   queryMode
	resultFocus                            bool
	resultCursor, objectCursor, pageOffset int
	objects                                []core.Row
	spinner                                string
	width                                  int
	body                                   viewport.Model
}

type QueryOpenMsg struct {
	Generation int
	Library    *core.Library
	Areas      []core.Area
	Err        error
}
type QueryDoneMsg struct {
	Generation int
	Result     *core.QueryResult
	Err        error
}
type QueryObjectsMsg struct {
	Generation, Offset int
	Rows               []core.Row
	Err                error
}

func NewQuery(root string) *Query {
	return &Query{root: root, selected: []string{"rioolgemalen", "putten", "riool"}, body: viewport.New(viewport.WithWidth(1), viewport.WithHeight(1))}
}
func (s *Query) ID() ScreenID            { return ScreenQuery }
func (s *Query) Busy() bool              { return s.opening || s.running || s.pageLoading }
func (s *Query) SetSpinner(frame string) { s.spinner = frame }
func (s *Query) OverlayOpen() bool       { return s.picker != nil }
func (s *Query) PaletteKeyMap() []key.Binding {
	if !s.loaded || s.library == nil || s.area.Code == "" || s.Busy() {
		return nil
	}
	return s.KeyMap()
}
func (s *Query) Info() string {
	if s.area.Code == "" {
		return "offline gebiedsquery"
	}
	return s.area.Label()
}
func (s *Query) SetSize(w, h int) {
	s.width = max(1, w-4)
	s.body.SetWidth(s.width)
	s.body.SetHeight(max(1, h-2))
	if s.picker != nil {
		s.picker.SetSize(s.width, s.body.Height())
	}
	s.refresh()
	if s.mode == queryObjects {
		s.revealObject()
	} else if s.resultFocus {
		s.revealResult()
	}
}
func (s *Query) KeyMap() []key.Binding {
	if s.picker != nil {
		return s.picker.KeyMap()
	}
	bindings := []key.Binding{ui.Keys.Area, ui.Keys.QueryLayers, ui.Keys.RunQuery}
	if s.resultFocus || s.mode != querySummary {
		bindings[2] = ui.Keys.Inspect
	}
	if s.mode==queryObjects { bindings[2]=ui.Keys.Details }
	if s.mode == queryDetails {
		bindings = nil
	}
	if s.result != nil && s.mode == querySummary {
		bindings = append(bindings, ui.Keys.Rerun)
	}
	bindings = append(bindings, ui.Keys.Up, ui.Keys.Down, ui.Keys.PageUp, ui.Keys.PageDown, ui.Keys.Home, ui.Keys.End, ui.Keys.Back, ui.Keys.Quit)
	return bindings
}

func (s *Query) ShortKeyMap() []key.Binding {
	if s.picker != nil {
		return s.picker.KeyMap()
	}
	keys := s.KeyMap()
	if s.mode == queryDetails {
		return []key.Binding{ui.Keys.Back, ui.Keys.Up, ui.Keys.Down}
	}
	if s.mode == queryObjects {
		return []key.Binding{ui.Keys.Details, ui.Keys.Back, ui.Keys.Up, ui.Keys.Down}
	}
	if s.result != nil && s.resultFocus {
		return []key.Binding{ui.Keys.Inspect, ui.Keys.Rerun, ui.Keys.Back, ui.Keys.Area, ui.Keys.QueryLayers}
	}
	if s.mode == querySummary {
		return []key.Binding{ui.Keys.RunQuery, ui.Keys.Area, ui.Keys.QueryLayers, ui.Keys.Back}
	}
	return append(keys[:min(3, len(keys))], ui.Keys.Back)
}

// LoadCmd lazily opens the published library when panel 8 is first entered.
func (s *Query) LoadCmd() tea.Cmd {
	if s.loaded || s.opening {
		return nil
	}
	ctx, generation := s.beginWork()
	s.opening = true
	root := s.root
	return func() tea.Msg {
		lib, err := core.OpenLibrary(root)
		var areas []core.Area
		if err == nil && lib == nil {
			err = fmt.Errorf("geen kaartbibliotheek; bouw deze met de Python waterketen-tool")
		}
		if err == nil {
			areas, err = core.LoadAreas(lib)
		}
		if err == nil && len(areas) == 0 {
			err = fmt.Errorf("deze snapshot heeft geen gebiedenregister")
		}
		if ctx.Err() != nil {
			err = ctx.Err()
		}
		return QueryOpenMsg{Generation: generation, Library: lib, Areas: areas, Err: err}
	}
}
func (s *Query) beginWork() (context.Context, int) {
	s.Cancel()
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	return ctx, s.generation
}

// Cancel invalidates work without allowing a late completion to revive it.
func (s *Query) Cancel() {
	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}
	s.generation++
	s.opening, s.running, s.pageLoading = false, false, false
}
func (s *Query) finishWork() {
	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}
	s.opening, s.running, s.pageLoading = false, false, false
}
func (s *Query) clearResult() {
	s.Cancel()
	s.result, s.objects, s.err = nil, nil, nil
	s.mode, s.resultFocus, s.resultCursor = querySummary, false, 0
	s.body.GotoTop()
}
func (s *Query) runCmd() tea.Cmd {
	if s.running {
		return nil
	}
	if s.library == nil || s.area.Code == "" {
		return nil
	}
	if len(s.selected) == 0 {
		s.err = fmt.Errorf("kies minstens één laag met l")
		s.refresh()
		return nil
	}
	s.clearResult()
	ctx, generation := s.beginWork()
	s.running = true
	lib, code, layers := s.library, s.area.Code, slices.Clone(s.selected)
	s.refresh()
	return func() tea.Msg {
		result, err := core.RunAreaQuery(ctx, lib, code, layers)
		return QueryDoneMsg{Generation: generation, Result: result, Err: err}
	}
}
func (s *Query) pageCmd() tea.Cmd {
	if s.result == nil || s.resultCursor >= len(s.result.Layers) {
		return nil
	}
	row := s.result.Layers[s.resultCursor]
	if row.Status != core.QueryOK || row.Count == 0 {
		return nil
	}
	offset := s.objectCursor / queryPageSize * queryPageSize
	if offset == s.pageOffset && (s.pageLoading || len(s.objects) > 0) {
		s.refresh()
		s.revealObject()
		return nil
	}
	ctx, generation := s.beginWork()
	s.pageOffset, s.pageLoading, s.objects, s.err = offset, true, nil, nil
	lib := s.library
	s.body.GotoTop()
	s.refresh()
	return func() tea.Msg {
		rows, err := core.ReadQueryObjects(ctx, lib, row, offset, queryPageSize)
		return QueryObjectsMsg{Generation: generation, Offset: offset, Rows: rows, Err: err}
	}
}

func (s *Query) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case QueryOpenMsg:
		if msg.Generation != s.generation || !s.opening {
			return s, nil
		}
		s.finishWork()
		s.loaded, s.library, s.areas, s.err = true, msg.Library, msg.Areas, msg.Err
		s.layers = core.QueryLayers(s.library)
		s.area, _ = core.DefaultArea(s.areas)
		s.refresh()
		return s, nil
	case QueryDoneMsg:
		if msg.Generation != s.generation || !s.running {
			return s, nil
		}
		s.finishWork()
		s.result, s.err, s.resultFocus = msg.Result, msg.Err, msg.Err == nil
		s.body.GotoTop()
		s.refresh()
		s.revealResult()
		return s, nil
	case QueryObjectsMsg:
		if msg.Generation != s.generation || !s.pageLoading || s.mode != queryObjects {
			return s, nil
		}
		s.finishWork()
		s.pageOffset, s.objects, s.err = msg.Offset, msg.Rows, msg.Err
		s.refresh()
		s.revealObject()
		return s, nil
	case tea.PasteMsg:
		if s.picker != nil {
			s.picker.Update(msg)
		}
		return s, nil
	case tea.KeyPressMsg:
		if s.picker != nil {
			res, done := s.picker.Update(msg)
			if done {
				s.picker = nil
				if res.Confirm {
					if res.Toggled && !slices.Equal(s.selected, res.Marked) {
						s.clearResult()
						s.selected = res.Marked
					}
					if !res.Toggled && res.Key != s.area.Code {
						if area, err := core.SelectArea(s.areas, res.Key); err == nil {
							s.clearResult()
							s.area = area
						}
					}
				}
				s.refresh()
			}
			return s, nil
		}
		switch {
		case key.Matches(msg, ui.Keys.Area):
			if len(s.areas) > 0 {
				s.picker = NewAreaPicker(s.areas)
				s.picker.SetSize(s.width, s.body.Height())
			}
			return s, nil
		case key.Matches(msg, ui.Keys.QueryLayers):
			if s.loaded && s.library != nil {
				p := &Picker{mode: pickToggle, title: "Querylagen", intro: "lagen voor de gebiedstelling", hint: "spatie schakelt · enter past toe · esc terug"}
				for _, l := range s.layers {
					label := l.Label
					if !l.Available {
						label += " · niet lokaal"
					}
					p.all = append(p.all, pickItem{key: l.ID, label: label, mark: slices.Contains(s.selected, l.ID)})
				}
				p.SetSize(s.width, s.body.Height())
				s.picker = p
			}
			return s, nil
		case key.Matches(msg, ui.Keys.Back):
			if s.Busy() {
				s.Cancel()
				s.loaded = true
				s.err = fmt.Errorf("Query afgebroken")
			}
			if s.mode == queryDetails {
				s.mode = queryObjects
			} else if s.mode == queryObjects {
				s.mode, s.resultFocus = querySummary, s.result != nil
			} else {
				s.mode, s.resultFocus = querySummary, false
			}
			s.body.GotoTop()
			s.refresh()
			if s.mode == queryObjects {
				s.revealObject()
			} else if s.resultFocus {
				s.revealResult()
			}
			return s, nil
		case key.Matches(msg, ui.Keys.Rerun):
			if s.mode == querySummary {
				return s, s.runCmd()
			}
			return s, nil
		case key.Matches(msg, ui.Keys.Enter):
			if s.mode == queryDetails {
				return s, nil
			}
			if s.mode == queryObjects {
				if i := s.objectCursor - s.pageOffset; i >= 0 && i < len(s.objects) {
					s.mode = queryDetails
					s.body.GotoTop()
					s.refresh()
				}
				return s, nil
			}
			if !s.resultFocus || s.result == nil {
				return s, s.runCmd()
			}
			if len(s.result.Layers) == 0 {
				return s, nil
			}
			row := s.result.Layers[s.resultCursor]
			if row.Status != core.QueryOK || row.Count == 0 {
				return s, nil
			}
			s.mode, s.objectCursor, s.pageOffset, s.objects = queryObjects, 0, -1, nil
			return s, s.pageCmd()
		}
		if s.mode == queryObjects {
			old := s.objectCursor
			s.objectCursor = queryMove(msg, s.objectCursor, s.result.Layers[s.resultCursor].Count, max(1, s.body.Height()-3))
			if old != s.objectCursor {
				return s, s.pageCmd()
			}
			return s, nil
		}
		if s.mode == querySummary && s.resultFocus && s.result != nil && len(s.result.Layers) > 0 && key.Matches(msg, ui.Keys.Up, ui.Keys.Down, ui.Keys.PageUp, ui.Keys.PageDown, ui.Keys.Home, ui.Keys.End) {
			s.resultCursor = queryMove(msg, s.resultCursor, len(s.result.Layers), max(1, s.body.Height()-4))
			s.refresh()
			s.revealResult()
			return s, nil
		}
		if key.Matches(msg, ui.Keys.Home) {
			s.body.GotoTop()
			return s, nil
		}
		if key.Matches(msg, ui.Keys.End) {
			s.body.GotoBottom()
			return s, nil
		}
	}
	var cmd tea.Cmd
	s.body, cmd = s.body.Update(msg)
	return s, cmd
}

func queryMove(msg tea.KeyPressMsg, cursor, total, page int) int {
	switch {
	case key.Matches(msg, ui.Keys.Up):
		cursor--
	case key.Matches(msg, ui.Keys.Down):
		cursor++
	case key.Matches(msg, ui.Keys.PageUp):
		cursor -= page
	case key.Matches(msg, ui.Keys.PageDown):
		cursor += page
	case key.Matches(msg, ui.Keys.Home):
		cursor = 0
	case key.Matches(msg, ui.Keys.End):
		cursor = total - 1
	}
	return min(max(0, cursor), max(0, total-1))
}

// Keep the page range reachable alongside the last object, even when the page
// is taller than the viewport. Do not push the selection off a one-row viewport.
func (s *Query) revealObject() {
	line := 2 + s.objectCursor - s.pageOffset
	if s.body.Height() > 1 && len(s.objects) > 0 && s.objectCursor-s.pageOffset == len(s.objects)-1 {
		line++
	}
	s.reveal(line)
}

func (s *Query) revealResult() {
	if s.result != nil && s.resultFocus && len(s.result.Layers) > 0 {
		s.resultCursor = min(s.resultCursor, len(s.result.Layers)-1)
		s.reveal(4 + s.resultCursor) // one context row, then the three-row result heading
	}
}

func (s *Query) reveal(line int) {
	if line < s.body.YOffset() {
		s.body.SetYOffset(line)
	}
	if line >= s.body.YOffset()+s.body.Height() {
		s.body.SetYOffset(line - s.body.Height() + 1)
	}
}
func (s *Query) formLines(width int) []string {
	if s.body.Height() < 18 {
		return []string{components.Strong("GEBIED · " + s.area.Code),
			components.SingleLine(s.area.Kind+" "+s.area.Name+" · g wijzigt", width),
			components.Strong(fmt.Sprintf("LAGEN · %d gekozen · l wijzigt", len(s.selected))),
			"", components.Strong("[ Enter: query uitvoeren ]")}
	}
	lines := []string{components.Strong("GEBIED"), components.Strong("  " + s.area.Kind + " " + s.area.Name + "  ▾  g"), components.Muted("  " + s.area.Code + " · CBS"), "", components.Strong("LAGEN")}
	for _, l := range s.layers {
		if !slices.Contains(s.selected, l.ID) {
			continue
		}
		label := l.Label
		if !l.Available {
			label += " · niet lokaal"
		}
		lines = append(lines, "  [x] "+label)
	}
	if len(s.selected) == 0 {
		lines = append(lines, components.Muted("  Geen lagen gekozen"))
	}
	lines = append(lines, components.Muted(fmt.Sprintf("  l: %d snapshotkeuzes · %d gekozen", len(s.layers), len(s.selected))), "")
	button := "  [ Query uitvoeren ]"
	if !s.resultFocus {
		button = components.Selected(button, width)
	} else {
		button = components.Strong(button)
	}
	return append(lines, button)
}
func (s *Query) countLines(width int) []string {
	lines := []string{components.Strong("RESULTAAT"), components.Muted("Objecten binnen / op de gebiedsgrens"), ""}
	if s.result == nil {
		if s.running {
			return append(lines, components.Busy(s.spinner, "Gebied en lagen tellen…"))
		}
		return append(lines, components.Muted("Enter: query uitvoeren"))
	}
	for i, row := range s.result.Layers {
		count := "—"
		if row.Status == core.QueryOK {
			count = Thousands(row.Count)
		}
		line := fmt.Sprintf("%7s  %s", count, row.Layer.Label)
		if row.Status == core.QueryFailed {
			line += " · leesfout"
		}
		if row.Status == core.QueryUnavailable {
			line += " · niet lokaal"
		}
		if s.resultFocus && i == s.resultCursor {
			line = components.Selected("›"+line[1:], width)
		} else {
			line = components.Strong(line)
		}
		lines = append(lines, line)
	}
	return lines
}
func (s *Query) refresh() {
	if s.width == 0 {
		return
	}
	var lines []string
	if s.mode == queryDetails {
		i := s.objectCursor - s.pageOffset
		if i >= 0 && i < len(s.objects) {
			row := s.objects[i]
			doc := []core.CatalogLine{{Text: fmt.Sprintf("Object %d · %s", row.ID, s.result.Layers[s.resultCursor].Layer.Label), Kind: "heading"}}
			var names []string
			for name := range row.Properties {
				names = append(names, name)
			}
			sort.Strings(names)
			for _, name := range names {
				doc = append(doc, core.CatalogLine{Text: queryClean(name + ": " + row.Properties[name]), Kind: "plain"})
			}
			layer := s.result.Layers[s.resultCursor]
			doc = append(doc, core.CatalogLine{Text: queryClean("Bron: " + layer.SourceURL), Kind: "meta"}, core.CatalogLine{Text: queryProvenance(layer), Kind: "meta"})
			s.body.SetContent(components.CatalogDocument(doc, s.width))
			return
		}
	}
	if s.mode == queryObjects {
		layer := s.result.Layers[s.resultCursor]
		lines = []string{components.Strong(fmt.Sprintf("%s · %s objecten", layer.Layer.Label, Thousands(layer.Count))), components.Muted("↑↓ / pgdn / end: objecten · enter: details")}
		if s.pageLoading {
			lines = append(lines, components.Busy(s.spinner, "Objecten lezen…"))
		}
		for i, row := range s.objects {
			name := row.Properties["naam"]
			if name == "" {
				name = row.Properties["name"]
			}
			if name == "" {
				name = row.Properties["STRAATNAAM"]
			}
			line := fmt.Sprintf("  %d · %s", row.ID, queryClean(name))
			if s.pageOffset+i == s.objectCursor {
				line = components.Selected("›"+line[1:], s.width)
			}
			lines = append(lines, line)
		}
		if len(s.objects) > 0 {
			lines = append(lines, components.Muted(fmt.Sprintf("Pagina %d–%d van %s · geen sample", s.pageOffset+1, s.pageOffset+len(s.objects), Thousands(layer.Count))))
		}
	} else if s.library != nil && s.area.Code != "" {
		if s.result != nil {
			lines = []string{components.Muted(fmt.Sprintf("%s · %d lagen · g/l wijzigen", s.area.Code, len(s.selected)))}
			lines = append(lines, s.countLines(s.width)...)
			if s.resultFocus {
				lines = append(lines, components.Muted("Enter: objecten · ctrl+r: opnieuw"))
			} else {
				lines = append(lines, components.Muted("Enter: opnieuw tellen · ↑↓: lezen"))
			}
		} else if s.width >= 90 {
			leftWidth := 36
			body := components.CatalogColumns(strings.Join(s.formLines(leftWidth), "\n"), strings.Join(s.countLines(s.width-leftWidth-3), "\n"), leftWidth)
			lines = strings.Split(body, "\n")
		} else {
			lines = append(s.formLines(s.width), "")
			lines = append(lines, s.countLines(s.width)...)
		}
		lines = append(lines, components.Muted("Snapshot · "+queryDate(s.library.Manifest.CreatedAt)))
		if s.result != nil && len(s.result.Layers) > 0 {
			row := s.result.Layers[s.resultCursor]
			lines = append(lines, components.Muted(queryProvenance(row)))
			if row.Error != "" {
				lines = append(lines, components.Error(queryClean(row.Error)))
			}
		}
		lines = append(lines, components.Muted("Registratie ≠ volledig fysiek netwerk"))
	} else if s.opening {
		lines = []string{components.Busy(s.spinner, "Gebiedenregister lezen…")}
	}
	if s.err != nil {
		lines = append(lines, components.Error(queryClean(s.err.Error())))
	}
	for i, line := range lines {
		lines[i] = components.SingleLine(line, s.width)
	}
	s.body.SetContent(strings.Join(lines, "\n"))
}
func (s *Query) View() string {
	if s.picker != nil {
		return s.picker.View()
	}
	// Spinner ink is supplied by the app, without a separate animation chain.
	if s.Busy() {
		s.refresh()
	}
	return s.body.View()
}
func queryClean(text string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return ' '
		}
		return r
	}, text)
}
func queryDate(value string) string {
	if value == "" {
		return "datum onbekend"
	}
	if date, err := time.Parse(time.RFC3339, value); err == nil {
		if zone, err := time.LoadLocation("Europe/Amsterdam"); err == nil {
			date = date.In(zone)
		}
		return date.Format("02-01-2006 15:04 MST")
	}
	return queryClean(value)
}
func queryProvenance(row core.QueryLayerResult) string {
	source := row.Layer.SourceID
	if parsed, err := url.Parse(row.SourceURL); err == nil && parsed.Hostname() != "" {
		source = parsed.Hostname()
	}
	if strings.Contains(source, "gwsw") {
		source = "GWSW"
	}
	if source == "" {
		source = "bron onbekend"
	}
	license, date := row.License, row.SourceDate
	if license == "" {
		license = "licentie onbekend"
	}
	if date == "" {
		date = "onbekend"
	}
	return queryClean(source + " · " + license + " · brondatum " + date)
}
