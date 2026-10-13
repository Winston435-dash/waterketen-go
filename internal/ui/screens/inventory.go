package screens

import (
	"fmt"
	"sort"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"waterketen/internal/core"
	"waterketen/internal/ui"
	"waterketen/internal/ui/components"
)

type inventoryFilter struct {
	query, previous      string
	editing, initialized bool
	input                textinput.Model
}

func (f *inventoryFilter) initInput() {
	if !f.initialized {
		f.input = components.SearchInput("Filter: ")
		f.input.SetValue(f.query)
		f.input.CursorEnd()
		f.initialized = true
	}
}
func (f *inventoryFilter) update(msg tea.Msg) bool {
	k, isKey := msg.(tea.KeyPressMsg)
	if f.editing {
		f.initInput()
		if isKey && key.Matches(k, ui.Keys.Enter) {
			f.editing = false
			return true
		}
		if isKey && key.Matches(k, ui.Keys.Back) {
			f.query, f.editing = f.previous, false
			f.input.SetValue(f.query)
			return true
		}
		if p, ok := msg.(tea.PasteMsg); ok {
			p.Content = TypedText(p)
			msg = p
		}
		f.input, _ = f.input.Update(msg)
		f.query = f.input.Value()
		return true
	}
	if !isKey {
		return false
	}
	if key.Matches(k, ui.Keys.Filter) {
		f.previous, f.editing = f.query, true
		f.initInput()
		f.input.SetValue(f.query)
		f.input.CursorEnd()
		return true
	}
	if key.Matches(k, ui.Keys.Back) && f.query != "" {
		f.query = ""
		return true
	}
	return false
}
func (f inventoryFilter) matches(text string) bool {
	return strings.Contains(core.NormaliseName(text), core.NormaliseName(f.query))
}
func (f *inventoryFilter) view(width int) string {
	if f.editing {
		f.initInput()
		f.input.SetWidth(max(1, width-8))
		return components.SingleLine(f.input.View(), width)
	}
	if f.query != "" {
		return components.SingleLine(components.Strong("Filter: "+f.query+" · esc wist"), width)
	}
	return components.SingleLine(components.Muted("/ filter · enter details"), width)
}
func inventoryKeys(editing bool) []key.Binding {
	if editing {
		return []key.Binding{ui.Keys.Enter, ui.Keys.Back}
	}
	return []key.Binding{ui.Keys.Filter, ui.Keys.Details, ui.Keys.Up, ui.Keys.Down, ui.Keys.PageUp, ui.Keys.PageDown, ui.Keys.Home, ui.Keys.End, ui.Keys.Back, ui.Keys.Quit}
}

func inventoryPaletteKeys(keys []key.Binding,list components.Inventory,details,editing bool) []key.Binding {
	if details||editing { return nil }
	_,selected:=list.Selected()
	var available []key.Binding
	for _,binding:=range keys { if binding.Help().Key=="enter"&&!selected { continue };available=append(available,binding) }
	return available
}
func scrollInventory(body *viewport.Model, msg tea.Msg) tea.Cmd {
	if msg, ok := msg.(tea.KeyPressMsg); ok {
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

type Lagen struct {
	catalog            core.Catalog
	err                error
	filter             inventoryFilter
	localOnly, details bool
	body               viewport.Model
	list               components.Inventory
}

func NewLagen(catalog core.Catalog, err error) *Lagen {
	return &Lagen{catalog: catalog, err: err, body: viewport.New(), list: components.NewInventory()}
}
func (s *Lagen) ID() ScreenID        { return ScreenLagen }
func (s *Lagen) EditingFilter() bool { return s.filter.editing }
func (s *Lagen) PaletteKeyMap() []key.Binding { return inventoryPaletteKeys(s.KeyMap(),s.list,s.details,s.filter.editing) }
func (s *Lagen) KeyMap() []key.Binding {
	if s.details {
		return []key.Binding{ui.Keys.Back, ui.Keys.Up, ui.Keys.Down, ui.Keys.PageUp, ui.Keys.PageDown, ui.Keys.Home, ui.Keys.End}
	}
	keys := inventoryKeys(s.filter.editing)
	if !s.filter.editing {
		keys = append([]key.Binding{ui.Keys.LocalOnly}, keys...)
	}
	return keys
}
func (s *Lagen) ShortKeyMap() []key.Binding {
	if s.details || s.filter.editing {
		return s.KeyMap()
	}
	return []key.Binding{ui.Keys.Filter, ui.Keys.LocalOnly, ui.Keys.Details, ui.Keys.Up, ui.Keys.Down}
}
func (s *Lagen) SetSize(w, h int) {
	s.body.SetWidth(max(1, w-4))
	s.body.SetHeight(max(1, h-3))
	s.list.SetSize(s.body.Width(), max(1, h-5))
	s.refresh()
}
func (s *Lagen) Info() string {
	local := 0
	for _, l := range s.catalog.Layers {
		if l.Available() {
			local++
		}
	}
	return fmt.Sprintf("%d bekend · %d met lokaal bestand", len(s.catalog.Layers), local)
}
func (s *Lagen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	if s.details {
		if k, ok := msg.(tea.KeyPressMsg); ok && key.Matches(k, ui.Keys.Back) {
			s.details = false
			return s, nil
		}
		return s, scrollInventory(&s.body, msg)
	}
	if s.filter.update(msg) {
		s.refresh()
		return s, nil
	}
	if k, ok := msg.(tea.KeyPressMsg); ok {
		if key.Matches(k, ui.Keys.LocalOnly) {
			s.localOnly = !s.localOnly
			s.refresh()
			return s, nil
		}
		if key.Matches(k, ui.Keys.Enter) {
			if item, ok := s.list.Selected(); ok {
				s.details = true
				s.body.SetContent(components.WrappedText(item.Detail, s.body.Width()))
				s.body.GotoTop()
			}
			return s, nil
		}
	}
	return s, s.list.Update(msg)
}
func (s *Lagen) refresh() {
	layers := append([]core.CatalogLayer(nil), s.catalog.Layers...)
	sort.SliceStable(layers, func(i, j int) bool {
		if layers[i].Available() != layers[j].Available() {
			return layers[i].Available()
		}
		return layers[i].Label < layers[j].Label
	})
	var items []components.InventoryItem
	for _, layer := range layers {
		if s.localOnly && !layer.Available() || !s.filter.matches(layer.Label+" "+layer.ID+" "+layer.SourceID+" "+layer.Official) {
			continue
		}
		items = append(items, layerInventoryItem(layer, s.catalog))
	}
	s.list.SetItems(items)
	if s.details {
		if item, ok := s.list.Selected(); ok {
			s.body.SetContent(components.WrappedText(item.Detail, s.body.Width()))
		}
	}
}
func layerInventoryItem(layer core.CatalogLayer, catalog core.Catalog) components.InventoryItem {
	state := "geen lokale levering"
	status := components.StatusIdle
	if len(layer.Local) > 0 {
		state = "bestand ontbreekt"
		status = components.StatusDown
	}
	if layer.Available() {
		state = "bestand aanwezig"
		status = components.StatusOK
	}
	license := "onbekend"
	source := layer.SourceID
	for _, s := range catalog.Sources {
		if s.ID == source {
			license = licenseOrUnknown(s.License)
			break
		}
	}
	doc := []string{components.Strong(queryClean(layer.Label + " · " + layer.ID)),
		components.Muted(queryClean("Geometrie: " + layer.Geometry + " · bron: " + source)),
		components.Muted(queryClean("Officieel: " + layer.Official)), components.Strong(state),
		components.Muted("Bestand aanwezig ≠ inhoud gevalideerd of volledig"), ""}
	summary := state + " · " + license
	if len(layer.Local) == 0 {
		doc = append(doc, "Alleen beschreven — geen lokale levering", "Licentie bron: "+license)
	}
	if len(layer.Local) == 1 {
		local := layer.Local[0]
		license = licenseOrUnknown(local.License)
		unit:="objecten";if local.Storage=="gebiedenregister" { unit="gebieden" }
		summary = fmt.Sprintf("%s · %s %s (metadata) · %s", state, Thousands(local.Count),unit, license)
	} else if len(layer.Local) > 1 {
		summary = fmt.Sprintf("%s · %d leveringen · licenties per levering", state, len(layer.Local))
	}
	for _, local := range layer.Local {
		localState := "bestand ontbreekt"
		if local.OnDisk {
			localState = "bestand aanwezig"
		}
		unit := "objecten"
		if local.Storage == "gebiedenregister" {
			unit = "gebieden"
		}
		doc = append(doc, queryClean(fmt.Sprintf("%s · %s: %s %s (metadata) · %s", local.Storage, local.Area, Thousands(local.Count), unit, localState)), "Licentie levering: "+queryClean(licenseOrUnknown(local.License)), "")
	}
	compact:=state+" · "+license
	if len(layer.Local)>1 { compact=fmt.Sprintf("%d leveringen · licenties: enter",len(layer.Local)) }
	return components.InventoryItem{ID: layer.ID, Label: queryClean(layer.Label), Summary: queryClean(summary),CompactSummary:queryClean(compact), Detail: strings.Join(doc, "\n"), Status: status}
}
func (s *Lagen) View() string {
	width := s.body.Width()
	if s.details {
		return components.SingleLine(components.Muted("Laagdetails · esc: lijst"), width) + "\n" + s.body.View()
	}
	context := "Alle lagen · groen = bestand; geen validatie"
	if s.localOnly {
		context = "Alleen lokaal · aantallen zijn metadata"
	}
	if s.err != nil {
		context = queryClean(s.err.Error())
	}
	return components.FitRows(s.filter.view(width)+"\n"+components.SingleLine(components.Muted(context), width)+"\n"+s.list.View()+"\n"+components.SingleLine(components.Muted(s.list.Position()), width), width, s.body.Height()+1)
}

type Gebieden struct {
	lib                      *core.Library
	assets                   []core.AssetMeta
	areas                    []core.Area
	loaded, opening, details bool
	err                      error
	filter                   inventoryFilter
	body                     viewport.Model
	list                     components.Inventory
	spinner                  string
}
type AreasMsg struct {
	Areas []core.Area
	Err   error
}

func NewGebieden(lib *core.Library, assets []core.AssetMeta) *Gebieden {
	return &Gebieden{lib: lib, assets: assets, loaded: lib == nil, body: viewport.New(), list: components.NewInventory()}
}
func (s *Gebieden) ID() ScreenID            { return ScreenGebieden }
func (s *Gebieden) EditingFilter() bool     { return s.filter.editing }
func (s *Gebieden) PaletteKeyMap() []key.Binding { return inventoryPaletteKeys(s.KeyMap(),s.list,s.details,s.filter.editing) }
func (s *Gebieden) Busy() bool              { return s.opening }
func (s *Gebieden) SetSpinner(frame string) { s.spinner = frame }
func (s *Gebieden) KeyMap() []key.Binding {
	if s.details {
		return []key.Binding{ui.Keys.Back, ui.Keys.Up, ui.Keys.Down, ui.Keys.PageUp, ui.Keys.PageDown, ui.Keys.Home, ui.Keys.End}
	}
	return inventoryKeys(s.filter.editing)
}
func (s *Gebieden) ShortKeyMap() []key.Binding {
	if s.details || s.filter.editing {
		return s.KeyMap()
	}
	return []key.Binding{ui.Keys.Filter, ui.Keys.Details, ui.Keys.Up, ui.Keys.Down}
}
func (s *Gebieden) SetSize(w, h int) {
	s.body.SetWidth(max(1, w-4))
	s.body.SetHeight(max(1, h-3))
	s.list.SetSize(s.body.Width(), max(1, h-5))
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
		return fmt.Sprintf("%d CBS-gebieden", len(s.areas))
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
	if result, ok := msg.(AreasMsg); ok {
		s.loaded, s.opening, s.areas, s.err = true, false, result.Areas, result.Err
		s.refresh()
		return s, nil
	}
	if s.details {
		if k, ok := msg.(tea.KeyPressMsg); ok && key.Matches(k, ui.Keys.Back) {
			s.details = false
			return s, nil
		}
		return s, scrollInventory(&s.body, msg)
	}
	if s.filter.update(msg) {
		s.refresh()
		return s, nil
	}
	if k, ok := msg.(tea.KeyPressMsg); ok && key.Matches(k, ui.Keys.Enter) {
		if item, ok := s.list.Selected(); ok {
			s.details = true
			s.body.SetContent(components.WrappedText(item.Detail, s.body.Width()))
			s.body.GotoTop()
		}
		return s, nil
	}
	return s, s.list.Update(msg)
}
func areaRank(kind string) int {
	switch kind {
	case core.KindGemeente:
		return 0
	case core.KindWijk:
		return 1
	case core.KindBuurt:
		return 2
	}
	return 3
}
func (s *Gebieden) refresh() {
	areas := append([]core.Area(nil), s.areas...)
	sort.Slice(areas, func(i, j int) bool {
		if areaRank(areas[i].Kind) != areaRank(areas[j].Kind) {
			return areaRank(areas[i].Kind) < areaRank(areas[j].Kind)
		}
		return areas[i].Code < areas[j].Code
	})
	var items []components.InventoryItem
	for _, area := range areas {
		if !s.filter.matches(area.Label()) {
			continue
		}
		snapshot := "snapshot onbekend"
		if s.lib != nil {
			snapshot = s.lib.Snapshot
		}
		items = append(items, components.InventoryItem{ID: area.Code, Label: queryClean(area.Label()),CompactLabel:queryClean(area.Name),CompactSummary:area.Code+" · "+area.Kind, Summary: "CBS-grens · kies met g in Kaart/Query", Status: components.StatusIdle,
			Detail: queryClean(area.Label()) + "\n\nOfficiële CBS-grens\nSnapshot: " + queryClean(snapshot) + "\nGeen automatische selectie in Kaart/Query; kies daar expliciet met g."})
	}
	if len(areas) == 0 && !s.opening {
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
				items = append(items, components.InventoryItem{ID: id, Label: label, Summary: "managed naam · geen CBS-grens", Status: components.StatusIdle, Detail: label + "\nGeen officiële CBS-grens beschikbaar."})
			}
		}
	}
	s.list.SetItems(items)
	if s.details {
		if item, ok := s.list.Selected(); ok {
			s.body.SetContent(components.WrappedText(item.Detail, s.body.Width()))
		}
	}
}
func (s *Gebieden) View() string {
	width := s.body.Width()
	if s.opening {
		return components.Busy(s.spinner, "CBS-gebieden worden gelezen…")
	}
	if s.details {
		return components.SingleLine(components.Muted("Gebiedsdetails · esc: lijst"), width) + "\n" + s.body.View()
	}
	context := "Officiële CBS-grenzen · g in Kaart/Query"
	if len(s.areas) == 0 {
		context = "Geen CBS-gebiedenregister · alleen managed namen"
	}
	if s.lib != nil && len(s.areas) > 0 && s.lib.Manifest.AreaCount != len(s.areas) {
		context = fmt.Sprintf("Manifest noemt %d; daadwerkelijk gelezen: %d", s.lib.Manifest.AreaCount, len(s.areas))
	}
	if s.err != nil {
		context = queryClean(s.err.Error())
	}
	return components.FitRows(s.filter.view(width)+"\n"+components.SingleLine(components.Muted(context), width)+"\n"+s.list.View()+"\n"+components.SingleLine(components.Muted(s.list.Position()), width), width, s.body.Height()+1)
}
