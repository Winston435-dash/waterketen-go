package screens

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"

	"waterketen/internal/core"
	"waterketen/internal/ui"
	"waterketen/internal/ui/components"
)

type LayerData = core.LayerData

// --- Analyse: panel 4 --------------------------------------------------------
//
// The analysis screen computes network, length and depth over every managed
// layer. The Python rule that caught real bugs: a polygon layer has no pipes,
// so a per-pipe length is nonsense — the report says "omtrek" for those.

// AnalyseScreen renders the analysis of all loaded layers as a scrollable list.
type AnalyseScreen struct {
	assets   []core.AssetMeta
	data     []*LayerData
	busy     bool   // the initial managed-data load is in flight
	spinner  string // current app busy-spinner frame
	w, h     int
	scroll   int
	lines    []string // cached when data changes, never recomputed by scrolling
	report   []string // data-only report, unchanged when diagnostics arrive
	failures []error
	library  *core.Library
	catalog  core.Catalog
}

func NewAnalyse(root string, assets []core.AssetMeta) *AnalyseScreen {
	return &AnalyseScreen{assets: assets}
}

func (s *AnalyseScreen) ID() ScreenID { return ScreenAnalyse }
func (s *AnalyseScreen) KeyMap() []key.Binding {
	return []key.Binding{ui.Keys.Up, ui.Keys.Down, ui.Keys.PageUp, ui.Keys.PageDown, ui.Keys.Home, ui.Keys.End, ui.Keys.Quit}
}
func (s *AnalyseScreen) SetSize(w, h int) { s.w, s.h = w, h }
func (s *AnalyseScreen) Info() string {
	if s.library != nil {
		return fmt.Sprintf("%d managed · %d snapshotmetadata", len(s.data), len(s.library.Manifest.Layers))
	}
	return fmt.Sprintf("%d managed lagen berekend", len(s.data))
}

func (s *AnalyseScreen) SetLibrary(lib *core.Library, catalog core.Catalog) {
	s.library, s.catalog = lib, catalog
	s.report = s.renderLines()
	s.refreshLines()
}

// DataMsg carries loaded analysis data back to the app.
type DataMsg struct {
	Layers []*LayerData
	Errors []error
}

// SetData stores the loaded assets, clears the busy flag and resets the scroll.
func (s *AnalyseScreen) SetData(d []*LayerData) {
	s.data, s.scroll, s.busy = d, 0, false
	s.report = s.renderLines()
	s.refreshLines()
}

func (s *AnalyseScreen) SetErrors(errors []error) {
	s.failures = errors
	s.refreshLines()
}

func (s *AnalyseScreen) refreshLines() {
	s.lines = append([]string(nil), s.report...)
	for _, err := range s.failures {
		s.lines = append(s.lines, components.Error("Leesfout: "+err.Error()))
	}
}

// Busy reports bounded work in flight: the managed assets are being read.
func (s *AnalyseScreen) Busy() bool { return s.busy }

// SetBusy marks the initial managed-data load as in flight.
func (s *AnalyseScreen) SetBusy(v bool) { s.busy = v }

// SetSpinner receives the app's current busy-spinner frame.
func (s *AnalyseScreen) SetSpinner(frame string) { s.spinner = frame }

func (s *AnalyseScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch m := msg.(type) {
	case tea.KeyPressMsg:
		contentH := max(1, s.h-4)
		total := len(s.lines)
		switch {
		case key.Matches(m, ui.Keys.Down):
			s.scroll = min(s.scroll+1, max(0, total-contentH))
		case key.Matches(m, ui.Keys.Up):
			s.scroll = max(0, s.scroll-1)
		case key.Matches(m, ui.Keys.PageDown):
			s.scroll = min(s.scroll+contentH, max(0, total-contentH))
		case key.Matches(m, ui.Keys.PageUp):
			s.scroll = max(0, s.scroll-contentH)
		case key.Matches(m, ui.Keys.Home):
			s.scroll = 0
		case key.Matches(m, ui.Keys.End):
			s.scroll = max(0, total-contentH)
		}
	}
	return s, nil
}

func (s *AnalyseScreen) View() string {
	if s.busy {
		return components.Busy(s.spinner, "Lagen worden gelezen…")
	}
	if len(s.data) == 0 && len(s.failures) == 0 && s.library == nil {
		return components.EmptyState(
			"Geen managed data",
			"Haal data op: waterketen fetch --gebied delfshaven --laag riool",
			s.w,
		)
	}
	lines := s.lines
	contentH := max(1, s.h-4)
	if s.scroll > len(lines)-contentH {
		s.scroll = max(0, len(lines)-contentH)
	}
	window := lines[s.scroll:]
	if len(window) > contentH {
		window = window[:contentH]
	}
	body := strings.Join(window, "\n")
	if s.scroll > 0 || len(lines) > contentH {
		body += "\n" + components.Muted(fmt.Sprintf("  (%d/%d regels — pijltjes scrollen)", s.scroll+1, len(lines)))
	}
	return body
}

// renderLines is the pure report: analysis into strings. It never styles
// beyond Muted via components — colour decisions stay in the components.
func (s *AnalyseScreen) renderLines() []string {
	var out []string
	if len(s.data) > 0 {
		out = append(out, components.Strong("Managed — berekend over de gebiedsbestanden"),
			components.Muted("Dit is niet de gemeentebrede snapshot; kies paneel 8 voor een gebiedstelling."), "")
	}
	for _, d := range s.data {
		spec := d.Spec
		header := fmt.Sprintf("%s/%s", d.Area, d.Layer)
		if spec != nil {
			header = fmt.Sprintf("%s/%s — %s", d.Area, d.Layer, spec.Label)
		}
		out = append(out, components.Strong(header))

		summary := core.Summarize(d.Features)
		out = append(out, fmt.Sprintf("  features: %s  ·  lijnen %s · vlakken %s · punten %s",
			Thousands(summary.Features), Thousands(summary.Lines), Thousands(summary.Polygons), Thousands(summary.Points)))
		if spec != nil && !spec.Linear {
			// polygon layer: perimeter, never "gemiddeld N m per buis"
			out = append(out, fmt.Sprintf("  omtrek totaal: %.2f km", summary.TotalM/1000))
		} else {
			out = append(out, fmt.Sprintf("  lengte totaal: %.2f km", summary.TotalM/1000))
			if summary.Lines > 0 {
				out = append(out, fmt.Sprintf("  gemiddelde leiding: %.1f m", summary.TotalM/float64(summary.Lines)))
			}
		}
		if spec != nil && spec.HasBOB {
			if summary.BOBMeasured == 0 {
				out = append(out, components.Error("  geen BOB gemeten (0,00 = geen meting)"))
			} else {
				coverage := float64(summary.BOBMeasured) / float64(summary.Features) * 100
				out = append(out, fmt.Sprintf(
					"  BOB: gemeten op %s van %s buizen (%.1f%%) · min %.2f · mediaan %s · max %.2f",
					Thousands(summary.BOBMeasured), Thousands(summary.Features), coverage,
					summary.BOBMin, fmt.Sprintf("%.2f", summary.BOBMedian), summary.BOBMax))
			}
		}
		if summary.Streets > 0 {
			out = append(out, fmt.Sprintf("  straten met buizen: %s", Thousands(summary.Streets)))
		}
		out = append(out, "")
	}
	if len(out) > 0 {
		out = out[:len(out)-1] // trailing blank line
	}
	if s.library != nil {
		out = append(out, "", components.Strong("Snapshot — metadata, geen geometrieanalyse"),
			components.Muted("Snapshot "+s.library.Snapshot),
			components.Muted("Registratie ≠ volledig netwerk. Voor exacte gebiedsaantallen: paneel 8 Query."))
		for _, name := range s.library.LayerNames() {
			label, state := name, "bestand ontbreekt"
			for _, layer := range s.catalog.Layers {
				if layer.ID != name {
					continue
				}
				label = layer.Label
				for _, local := range layer.Local {
					if local.Storage == "snapshot" && local.OnDisk {
						state = "bestand aanwezig"
					}
				}
				break
			}
			rec := s.library.Manifest.Layers[name]
			out = append(out, components.Muted(queryClean(fmt.Sprintf("  %s: %s objecten volgens metadata · %s", label, Thousands(rec.FeatureCount), state))))
		}
	}
	return out
}

// --- Profiel: panel 6 -------------------------------------------------------
//
// Cross-section of one street, deepest pipes first. The Python section.py
// draws every pipe as a segment across its real x-extent with maaiveld above
// and depth increasing downward.

type ProfielScreen struct {
	root     string
	assets   []core.AssetMeta
	data     []*LayerData
	index    int // which (area, layer) is active
	w, h     int
	body     viewport.Model
	street   string
	streets  []string
	byStreet map[string][]core.Feature
	picker   *Picker
}

func NewProfiel(root string, assets []core.AssetMeta) *ProfielScreen {
	return &ProfielScreen{root: root, assets: assets, body: viewport.New()}
}

func (s *ProfielScreen) ID() ScreenID { return ScreenProfiel }
func (s *ProfielScreen) PaletteKeyMap() []key.Binding {
	if len(s.streets) == 0 {
		return nil
	}
	return s.KeyMap()
}
func (s *ProfielScreen) KeyMap() []key.Binding {
	if s.picker != nil {
		return s.picker.KeyMap()
	}
	return []key.Binding{ui.Keys.Street, ui.Keys.Up, ui.Keys.Down, ui.Keys.PageUp, ui.Keys.PageDown, ui.Keys.Home, ui.Keys.End, ui.Keys.PrevLayer, ui.Keys.NextLayer, ui.Keys.Quit}
}
func (s *ProfielScreen) ShortKeyMap() []key.Binding {
	if s.picker != nil {
		return s.picker.KeyMap()
	}
	return []key.Binding{ui.Keys.Street, ui.Keys.Up, ui.Keys.Down, ui.Keys.End}
}
func (s *ProfielScreen) OverlayOpen() bool { return s.picker != nil }
func (s *ProfielScreen) SetSize(w, h int) {
	s.w, s.h = w, h
	s.body.SetWidth(max(1, w-4))
	s.body.SetHeight(max(1, h-2))
	if s.picker != nil {
		s.picker.SetSize(s.body.Width(), s.body.Height())
	}
	s.refreshProfile()
}
func (s *ProfielScreen) Info() string {
	if len(s.data) == 0 {
		return "geen managed data"
	}
	return fmt.Sprintf("BOB-overzicht · managed %d/%d", s.index+1, len(s.data))
}

// SetData shares the same loaded data as Analyse — one load, two panels.
func (s *ProfielScreen) SetData(data []*LayerData) {
	s.data = nil
	for _, layer := range data {
		if layer.Spec != nil && layer.Spec.HasBOB {
			s.data = append(s.data, layer)
		}
	}
	s.index = min(s.index, max(0, len(s.data)-1))
	s.groupProfile()
}

func (s *ProfielScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	if s.picker != nil {
		res, done := s.picker.Update(msg)
		if done {
			s.picker = nil
			if res.Confirm {
				s.street = res.Key
				s.body.GotoTop()
				s.refreshProfile()
			}
		}
		return s, nil
	}
	switch m := msg.(type) {
	case tea.KeyPressMsg:
		if len(s.data) == 0 {
			return s, nil
		}
		switch {
		case key.Matches(m, ui.Keys.NextLayer):
			s.index = (s.index + 1) % len(s.data)
			s.groupProfile()
		case key.Matches(m, ui.Keys.PrevLayer):
			s.index = (s.index - 1 + len(s.data)) % len(s.data)
			s.groupProfile()
		case key.Matches(m, ui.Keys.Street):
			p := &Picker{mode: pickOne, title: "Straat kiezen · BOB", hint: "typ zoekt · enter kiest · esc terug"}
			for _, street := range s.streets {
				p.all = append(p.all, pickItem{key: street, label: fmt.Sprintf("%s · %d buizen", street, len(s.byStreet[street]))})
			}
			p.SetSize(s.body.Width(), s.body.Height())
			s.picker = p
			return s, nil
		}
	}
	return s, scrollInventory(&s.body, msg)
}

func (s *ProfielScreen) View() string {
	if s.picker != nil {
		return s.picker.View()
	}
	if len(s.data) == 0 {
		return components.EmptyState(
			"Geen managed BOB-lagen",
			"Kaart- en querylagen zijn niet automatisch profieldata; laad een managed rioollaag met BOB.",
			s.w,
		)
	}
	if len(s.streets) == 0 {
		return components.EmptyState(
			"Geen straatnamen of metingen",
			"Deze laag heeft geen BOB of geen STRAATNAAM gevuld",
			s.w,
		)
	}
	return s.body.View()
}

// Group once on data/layer changes, not during repaint or scrolling.
func (s *ProfielScreen) groupProfile() {
	s.byStreet = map[string][]core.Feature{}
	s.streets = nil
	s.street = ""
	if len(s.data) > 0 {
		for _, f := range s.data[s.index].Features {
			if f.Street != "" && !math.IsNaN(f.BOB) && !math.IsInf(f.BOB, 0) {
				s.byStreet[f.Street] = append(s.byStreet[f.Street], f)
			}
		}
	}
	for street := range s.byStreet {
		s.streets = append(s.streets, street)
	}
	sort.Strings(s.streets)
	lowest := math.Inf(1)
	for _, street := range s.streets {
		fs := s.byStreet[street]
		sort.SliceStable(fs, func(i, j int) bool { return fs[i].BOB < fs[j].BOB })
		if fs[0].BOB < lowest {
			lowest = fs[0].BOB
			s.street = street
		}
	}
	s.body.GotoTop()
	s.refreshProfile()
}

func (s *ProfielScreen) refreshProfile() {
	fs := s.byStreet[s.street]
	lines := []string{components.Strong(fmt.Sprintf("%s · %d buizen", s.street, len(fs))),
		components.Muted("BOB in m NAP · geen ruimtelijk profiel"),
		components.Muted("Geen diepte onder maaiveld"),
		components.Muted("Laagste BOB eerst · s kiest straat"), ""}
	for i, f := range fs {
		lines = append(lines, fmt.Sprintf("%4d.  %8.2f m NAP", i+1, f.BOB))
	}
	lines = append(lines, components.Muted(fmt.Sprintf("Einde · alle %d buizen weergegeven", len(fs))))
	s.body.SetContent(components.WrappedText(strings.Join(lines, "\n"), max(1, s.body.Width())))
}
