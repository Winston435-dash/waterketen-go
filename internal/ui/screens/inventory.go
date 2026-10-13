package screens

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"waterketen/internal/core"
	"waterketen/internal/ui"
	"waterketen/internal/ui/components"
)

// inventoryFilter owns typed keys only while its inline search is being edited.
// Enter applies, Escape cancels; outside editing Escape clears the active filter.
type inventoryFilter struct {
	query, previous string
	editing         bool
}

func (f *inventoryFilter) update(msg tea.KeyMsg) bool {
	if f.editing {
		switch {
		case key.Matches(msg, ui.Keys.Enter):
			f.editing = false
		case key.Matches(msg, ui.Keys.Back):
			f.query, f.editing = f.previous, false
		case msg.Type == tea.KeyBackspace:
			runes := []rune(f.query)
			if len(runes) > 0 {
				f.query = string(runes[:len(runes)-1])
			}
		default:
			f.query += TypedText(msg)
		}
		return true
	}
	if key.Matches(msg, ui.Keys.Filter) {
		f.previous, f.editing = f.query, true
		return true
	}
	if key.Matches(msg, ui.Keys.Back) && f.query != "" {
		f.query = ""
		return true
	}
	return false
}

func (f inventoryFilter) matches(text string) bool {
	return strings.Contains(core.NormaliseName(text), core.NormaliseName(f.query))
}

func (f inventoryFilter) view(width int) string {
	text := "/: filter op naam of code"
	if f.query != "" || f.editing {
		text = "Filter: " + f.query
	}
	if f.editing {
		text += " · enter toepassen · esc annuleren"
	}
	return components.SingleLine(components.Strong(text), width)
}

func inventoryKeys(editing bool) []key.Binding {
	if editing {
		return []key.Binding{ui.Keys.Enter, ui.Keys.Back}
	}
	return []key.Binding{ui.Keys.Filter, ui.Keys.Up, ui.Keys.Down, ui.Keys.PageUp, ui.Keys.PageDown, ui.Keys.Home, ui.Keys.End, ui.Keys.Back, ui.Keys.Quit}
}

func scrollInventory(body *viewport.Model, msg tea.Msg) tea.Cmd {
	if msg, ok := msg.(tea.KeyMsg); ok {
		if key.Matches(msg, ui.Keys.Home) {
			body.GotoTop()
			return nil
		}
		if key.Matches(msg, ui.Keys.End) {
			body.GotoBottom()
			return nil
		}
	}
	var cmd tea.Cmd
	*body, cmd = body.Update(msg)
	return cmd
}

// Lagen presents the same catalog as Ontdek, enriched with all managed/snapshot
// deliveries. Described remote layers stay visible but never look downloaded.
type Lagen struct {
	catalog   core.Catalog
	err       error
	filter    inventoryFilter
	localOnly bool
	body      viewport.Model
}

func NewLagen(catalog core.Catalog, err error) *Lagen {
	return &Lagen{catalog: catalog, err: err, body: viewport.New(1, 1)}
}
func (s *Lagen) ID() ScreenID        { return ScreenLagen }
func (s *Lagen) EditingFilter() bool { return s.filter.editing }
func (s *Lagen) KeyMap() []key.Binding {
	keys := inventoryKeys(s.filter.editing)
	if !s.filter.editing {
		keys = append([]key.Binding{ui.Keys.LocalOnly}, keys...)
	}
	return keys
}
func (s *Lagen) SetSize(w, h int) {
	s.body.Width, s.body.Height = max(1, w-4), max(1, h-3)
	s.refresh()
}
func (s *Lagen) Info() string {
	local := 0
	for _, layer := range s.catalog.Layers {
		if layer.Available() {
			local++
		}
	}
	return fmt.Sprintf("%d bekend · %d met lokaal bestand", len(s.catalog.Layers), local)
}
func (s *Lagen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		if s.filter.update(msg) {
			s.refresh()
			s.body.GotoTop()
			return s, nil
		}
		if key.Matches(msg, ui.Keys.LocalOnly) {
			s.localOnly = !s.localOnly
			s.refresh()
			s.body.GotoTop()
			return s, nil
		}
	}
	return s, scrollInventory(&s.body, msg)
}
func (s *Lagen) refresh() {
	layers := append([]core.CatalogLayer(nil), s.catalog.Layers...)
	sort.SliceStable(layers, func(i, j int) bool {
		if layers[i].Available() != layers[j].Available() {
			return layers[i].Available()
		}
		return layers[i].Label < layers[j].Label
	})
	var lines []string
	mode := "alle beschreven lagen"
	if s.localOnly {
		mode = "alleen lagen met lokaal bestand"
	}
	lines = append(lines, components.Muted("Weergave: "+mode+" · l wisselt"), components.Muted("Groen = bestand aanwezig; inhoud nog niet getoetst. Aantallen zijn metadata."), "")
	if s.err != nil {
		lines = append(lines, components.Error(queryClean(s.err.Error())), "")
	}
	shown := 0
	for _, layer := range layers {
		if s.localOnly && !layer.Available() {
			continue
		}
		if !s.filter.matches(layer.Label + " " + layer.ID + " " + layer.SourceID + " " + layer.Official) {
			continue
		}
		shown++
		style := components.StatusIdle
		if layer.Available() {
			style = components.StatusOK
		} else if len(layer.Local) > 0 {
			style = components.StatusDown
		}
		geometry, source := layer.Geometry, layer.SourceID
		if geometry == "" {
			geometry = "geometrie onbekend"
		}
		if source == "" {
			source = "herkomst bij lokale levering"
		}
		lines = append(lines, components.Badge(style, queryClean(layer.Label+" · "+layer.ID)), components.Muted(queryClean(geometry+" · bron: "+source)))
		if len(layer.Local) == 0 {
			lines = append(lines, components.Muted("  Alleen beschreven — geen lokale levering"))
			for _, source := range s.catalog.Sources {
				if source.ID == layer.SourceID {
					lines = append(lines, components.Muted("  Licentie bron: "+queryClean(source.License)))
					break
				}
			}
		}
		for _, local := range layer.Local {
			state := "bestand ontbreekt"
			if local.OnDisk {
				state = "bestand aanwezig"
			}
			unit := "objecten"
			if local.Storage == "gebiedenregister" {
				unit = "gebieden"
			}
			lines = append(lines, components.Muted(queryClean(fmt.Sprintf("  %s · %s: %s %s (metadata) · %s", local.Storage, local.Area, Thousands(local.Count), unit, state))),
				components.Muted("  Licentie levering: "+queryClean(licenseOrUnknown(local.License))))
		}
		lines = append(lines, "")
	}
	if shown == 0 {
		lines = append(lines, components.Muted("Geen lagen voor dit filter."))
	}
	s.body.SetContent(components.WrappedText(strings.Join(lines, "\n"), s.body.Width))
}
func (s *Lagen) View() string { return s.filter.view(s.body.Width) + "\n" + s.body.View() }

// Gebieden reads the actual CBS registry, lazily and off the event loop. Its
// count is the number read, not a manifest promise; all names/codes are searchable.
type Gebieden struct {
	lib             *core.Library
	assets          []core.AssetMeta
	areas           []core.Area
	loaded, opening bool
	err             error
	filter          inventoryFilter
	body            viewport.Model
	spinner         string
}

type AreasMsg struct {
	Areas []core.Area
	Err   error
}

func NewGebieden(lib *core.Library, assets []core.AssetMeta) *Gebieden {
	return &Gebieden{lib: lib, assets: assets, loaded: lib == nil, body: viewport.New(1, 1)}
}
func (s *Gebieden) ID() ScreenID            { return ScreenGebieden }
func (s *Gebieden) EditingFilter() bool     { return s.filter.editing }
func (s *Gebieden) Busy() bool              { return s.opening }
func (s *Gebieden) SetSpinner(frame string) { s.spinner = frame }
func (s *Gebieden) KeyMap() []key.Binding   { return inventoryKeys(s.filter.editing) }
func (s *Gebieden) SetSize(w, h int) {
	s.body.Width, s.body.Height = max(1, w-4), max(1, h-3)
	s.refresh()
}
func (s *Gebieden) Info() string {
	if s.opening {
		return "CBS-gebieden lezen…"
	}
	if s.lib != nil && !s.loaded {
		return "CBS-gebiedenregister"
	}
	if len(s.areas) > 0 {
		if s.lib == nil {
			return fmt.Sprintf("%d CBS-gebieden", len(s.areas))
		}
		return fmt.Sprintf("%d CBS-gebieden · snapshot %s", len(s.areas), s.lib.Snapshot)
	}
	return "geen CBS-register · managed gebiedsnamen"
}
func (s *Gebieden) LoadCmd() tea.Cmd {
	if s.loaded || s.opening || s.lib == nil {
		return nil
	}
	s.opening = true
	lib := s.lib
	return func() tea.Msg {
		areas, err := core.LoadAreas(lib)
		if err == nil && len(areas) == 0 {
			err = fmt.Errorf("Geen CBS-gebiedenregister in deze snapshot")
		}
		return AreasMsg{Areas: areas, Err: err}
	}
}
func (s *Gebieden) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case AreasMsg:
		s.loaded, s.opening, s.areas, s.err = true, false, msg.Areas, msg.Err
		s.refresh()
		return s, nil
	case tea.KeyMsg:
		if s.filter.update(msg) {
			s.refresh()
			s.body.GotoTop()
			return s, nil
		}
	}
	return s, scrollInventory(&s.body, msg)
}
func areaRank(kind string) int {
	switch kind {
	case core.KindGemeente:
		return 0
	case core.KindWijk:
		return 1
	case core.KindBuurt:
		return 2
	default:
		return 3
	}
}
func (s *Gebieden) refresh() {
	var lines []string
	if s.err != nil {
		lines = append(lines, components.Error(queryClean(s.err.Error())), "")
	}
	if len(s.areas) > 0 {
		counts := map[string]int{}
		areas := append([]core.Area(nil), s.areas...)
		for _, a := range areas {
			counts[a.Kind]++
		}
		sort.Slice(areas, func(i, j int) bool {
			if areaRank(areas[i].Kind) != areaRank(areas[j].Kind) {
				return areaRank(areas[i].Kind) < areaRank(areas[j].Kind)
			}
			return areas[i].Code < areas[j].Code
		})
		lines = append(lines, components.Muted(fmt.Sprintf("%d gemeente · %d wijken · %d buurten", counts[core.KindGemeente], counts[core.KindWijk], counts[core.KindBuurt])),
			components.Muted("Officiële CBS-grenzen · kies een gebied met g in Kaart of Query."), "")
		if s.lib != nil && s.lib.Manifest.AreaCount != len(areas) {
			lines = append(lines, components.Error(fmt.Sprintf("Manifest noemt %d gebieden; daadwerkelijk gelezen: %d", s.lib.Manifest.AreaCount, len(areas))), "")
		}
		shown := 0
		for _, a := range areas {
			if s.filter.matches(a.Label()) {
				lines = append(lines, queryClean(a.Label()))
				shown++
			}
		}
		if shown == 0 {
			lines = append(lines, components.Muted("Geen CBS-gebieden voor dit filter."))
		}
	} else if !s.opening {
		lines = append(lines, components.Muted("Geen officiële CBS-grenzen geladen. Hieronder alleen managed gebiedsnamen."), "")
		names := map[string]string{}
		for _, a := range s.assets {
			names[a.Area] = a.AreaLabel
		}
		var ids []string
		for id := range names {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		for _, id := range ids {
			label := id + " · " + names[id]
			if s.filter.matches(label) {
				lines = append(lines, queryClean(label))
			}
		}
		if len(ids) == 0 {
			lines = append(lines, components.Muted("Geen managed gebieden beschikbaar."))
		}
	}
	s.body.SetContent(components.WrappedText(strings.Join(lines, "\n"), s.body.Width))
}
func (s *Gebieden) View() string {
	if s.opening {
		return components.Busy(s.spinner, "CBS-gebieden worden gelezen…")
	}
	return s.filter.view(s.body.Width) + "\n" + s.body.View()
}
