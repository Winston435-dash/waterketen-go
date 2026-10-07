package screens

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"waterketen/internal/core"
	"waterketen/internal/ui"
	"waterketen/internal/ui/components"
)

// LayerData is one loaded (area, layer) pair ready for analysis.
type LayerData struct {
	Area     string
	Layer    string
	Spec     *core.LayerSpec
	Features []core.Feature
}

// LoadLayer reads one managed asset's processed GPKG into a Feature slice.
// Returns an error when the asset has no on-disk file or the table cannot
// be read — the analysis screen shows it, never crashes.
func LoadLayer(a core.AssetMeta) (*LayerData, error) {
	if !a.ProcessedExists() {
		return nil, fmt.Errorf("geen processed data voor %s/%s", a.Area, a.Layer)
	}
	db, err := core.Open(a.ProcessedPath)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	tbl, err := core.FeatureTable(db)
	if err != nil {
		return nil, err
	}
	data := &LayerData{
		Area:  a.Area,
		Layer: a.Layer,
		Spec:  core.LayerSpecFor(a.Layer),
	}
	tbl.Each(func(r core.Row) {
		f := core.Feature{
			Layer:      a.Layer,
			Position:   len(data.Features),
			Geometry:   r.Geometry,
			Properties: r.Properties,
			Street:     r.Properties["STRAATNAAM"],
		}
		// BOB: 0,00 betekent "nooit gemeten"; gebruik beide uiteinden apart.
		// One put gets two values when begin + end share a manhole; the mean
		// of the two columns per pipe is the established simplification.
		beg, okBeg := r.Numbers["BOB_ACTUEE"]
		end, okEnd := r.Numbers["BOB_ACTU_1"]
		useBeg := okBeg && beg != 0
		useEnd := okEnd && end != 0
		switch {
		case useBeg && useEnd:
			f.BOB = (beg + end) / 2
		case useBeg:
			f.BOB = beg
		case useEnd:
			f.BOB = end
		default:
			f.BOB = math.NaN()
		}
		data.Features = append(data.Features, f)
	})
	return data, nil
}

// LoadAssets reads every on-disk managed asset.
func LoadAssets(assets []core.AssetMeta) []*LayerData {
	var out []*LayerData
	for _, a := range assets {
		if !a.ProcessedExists() {
			continue
		}
		d, err := LoadLayer(a)
		if err == nil {
			out = append(out, d)
		}
	}
	return out
}

// --- Analyse: panel 4 --------------------------------------------------------
//
// The analysis screen computes network, length and depth over every managed
// layer. The Python rule that caught real bugs: a polygon layer has no pipes,
// so a per-pipe length is nonsense — the report says "omtrek" for those.

type AnalysisLine struct {
	Text  string
	Muted bool
}

// AnalyseScreen renders the analysis of all loaded layers as a scrollable list.
type AnalyseScreen struct {
	assets []core.AssetMeta
	data   []*LayerData
	w, h   int
	scroll int
}

func NewAnalyse(root string, assets []core.AssetMeta) *AnalyseScreen {
	return &AnalyseScreen{assets: assets}
}

func (s *AnalyseScreen) ID() ScreenID { return ScreenAnalyse }
func (s *AnalyseScreen) KeyMap() []key.Binding {
	return []key.Binding{ui.Keys.Up, ui.Keys.Down, ui.Keys.PageUp, ui.Keys.PageDown, ui.Keys.Quit}
}
func (s *AnalyseScreen) SetSize(w, h int) { s.w, s.h = w, h }
func (s *AnalyseScreen) Info() string {
	return fmt.Sprintf("%d lagen geladen", len(s.data))
}

// DataMsg carries loaded analysis data back to the app.
type DataMsg []*LayerData

// LoadDataCmd loads every managed asset off the event loop.
func LoadDataCmd(assets []core.AssetMeta) tea.Cmd {
	return func() tea.Msg { return DataMsg(LoadAssets(assets)) }
}

// SetData stores the loaded assets and resets the scroll.
func (s *AnalyseScreen) SetData(d []*LayerData) { s.data, s.scroll = d, 0 }

func (s *AnalyseScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch m := msg.(type) {
	case tea.KeyMsg:
		contentH := s.h - 4 // Panel border + padding
		total := len(s.renderLines())
		switch {
		case key.Matches(m, ui.Keys.Down):
			s.scroll = min2(s.scroll+1, max2(0, total-contentH))
		case key.Matches(m, ui.Keys.Up):
			s.scroll = max2(0, s.scroll-1)
		case key.Matches(m, ui.Keys.PageDown):
			s.scroll = min2(s.scroll+contentH, max2(0, total-contentH))
		case key.Matches(m, ui.Keys.PageUp):
			s.scroll = max2(0, s.scroll-contentH)
		case key.Matches(m, ui.Keys.Home):
			s.scroll = 0
		case key.Matches(m, ui.Keys.End):
			s.scroll = max2(0, total-contentH)
		}
	}
	return s, nil
}

func (s *AnalyseScreen) View() string {
	if len(s.data) == 0 {
		return components.EmptyState(
			"Geen managed data",
			"Haal data op: waterketen fetch --gebied delfshaven --laag riool",
			s.w,
		)
	}
	lines := s.renderLines()
	contentH := s.h - 4
	if s.scroll > len(lines)-contentH {
		s.scroll = max2(0, len(lines)-contentH)
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
	for _, d := range s.data {
		spec := d.Spec
		header := fmt.Sprintf("%s/%s", d.Area, d.Layer)
		if spec != nil {
			header = fmt.Sprintf("%s/%s — %s", d.Area, d.Layer, spec.Label)
		}
		out = append(out, components.Strong(header))

		var totalLen float64
		var lines, polys, points int
		var bobs []float64
		streets := map[string]int{}
		for _, f := range d.Features {
			switch f.Geometry.Kind {
			case "linestring", "multilinestring":
				lines++
			case "polygon", "multipolygon":
				polys++
			case "point":
				points++
			}
			totalLen += f.Geometry.Length()
			if !math.IsNaN(f.BOB) {
				bobs = append(bobs, f.BOB)
			}
			if f.Street != "" {
				streets[f.Street]++
			}
		}
		out = append(out, fmt.Sprintf("  features: %s  ·  lijnen %s · vlakken %s · punten %s",
			thousands(len(d.Features)), thousands(lines), thousands(polys), thousands(points)))
		if spec != nil && !spec.Linear {
			// polygon layer: perimeter, never "gemiddeld N m per buis"
			out = append(out, fmt.Sprintf("  omtrek totaal: %.2f km", totalLen/1000))
		} else {
			out = append(out, fmt.Sprintf("  lengte totaal: %.2f km", totalLen/1000))
			if lines > 0 {
				out = append(out, fmt.Sprintf("  gemiddelde leiding: %.1f m", totalLen/float64(lines)))
			}
		}
		if spec != nil && spec.HasBOB {
			if len(bobs) == 0 {
				out = append(out, components.Error("  geen BOB gemeten (0,00 = geen meting)"))
			} else {
				minB, maxB, sum := bobs[0], bobs[0], 0.0
				for _, v := range bobs {
					minB = math.Min(minB, v)
					maxB = math.Max(maxB, v)
					sum += v
				}
				coverage := float64(len(bobs)) / float64(len(d.Features)) * 100
				out = append(out, fmt.Sprintf(
					"  BOB: gemeten op %s van %s buizen (%.1f%%) · min %.2f · mediaan %s · max %.2f",
					thousands(len(bobs)), thousands(len(d.Features)), coverage,
					minB, median(bobs), maxB))
			}
		}
		if len(streets) > 0 {
			out = append(out, fmt.Sprintf("  straten met buizen: %s", thousands(len(streets))))
		}
		out = append(out, "")
	}
	if len(out) > 0 {
		out = out[:len(out)-1] // trailing blank line
	}
	return out
}

func median(v []float64) string {
	if len(v) == 0 {
		return "—"
	}
	sorted := append([]float64(nil), v...)
	sort2(sorted)
	m := sorted[len(sorted)/2]
	if len(v)%2 == 0 {
		m = (sorted[len(v)/2-1] + sorted[len(v)/2]) / 2
	}
	return fmt.Sprintf("%.2f", m)
}

func sort2(v []float64) {
	for i := 1; i < len(v); i++ {
		for j := i; j > 0 && v[j-1] > v[j]; j-- {
			v[j-1], v[j] = v[j], v[j-1]
		}
	}
}

// --- Profiel: panel 6 -------------------------------------------------------
//
// Cross-section of one street, deepest pipes first. The Python section.py
// draws every pipe as a segment across its real x-extent with maaiveld above
// and depth increasing downward.

type ProfielScreen struct {
	root   string
	assets []core.AssetMeta
	data   []*LayerData
	index  int // which (area, layer) is active
	w, h   int
}

func NewProfiel(root string, assets []core.AssetMeta) *ProfielScreen {
	return &ProfielScreen{root: root, assets: assets}
}

func (s *ProfielScreen) ID() ScreenID { return ScreenProfiel }
func (s *ProfielScreen) KeyMap() []key.Binding {
	return []key.Binding{ui.Keys.PrevLayer, ui.Keys.NextLayer, ui.Keys.Quit}
}
func (s *ProfielScreen) SetSize(w, h int) { s.w, s.h = w, h }
func (s *ProfielScreen) Info() string {
	if len(s.data) == 0 {
		return "geen managed data"
	}
	return fmt.Sprintf("laag %d/%d", s.index+1, len(s.data))
}

// SetData shares the same loaded data as Analyse — one load, two panels.
func (s *ProfielScreen) SetData(d []*LayerData) { s.data = d }

func (s *ProfielScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch m := msg.(type) {
	case tea.KeyMsg:
		if len(s.data) == 0 {
			return s, nil
		}
		switch {
		case key.Matches(m, ui.Keys.NextLayer):
			s.index = (s.index + 1) % len(s.data)
		case key.Matches(m, ui.Keys.PrevLayer):
			s.index = (s.index - 1 + len(s.data)) % len(s.data)
		}
	}
	return s, nil
}

func (s *ProfielScreen) View() string {
	if len(s.data) == 0 {
		return components.EmptyState(
			"Geen managed data",
			"Haal data op: waterketen fetch --gebied delfshaven --laag riool",
			s.w,
		)
	}
	d := s.data[s.index]
	spec := d.Spec
	if spec == nil || !spec.HasBOB {
		return components.EmptyState(
			fmt.Sprintf("%s/%s heeft geen BOB-waarden", d.Area, d.Layer),
			"Alleen rioolbuizen hebben een buisbodem; kies een andere laag met [ of ]",
			s.w,
		)
	}
	// Group pipes by street, keep the street with the deepest pipe.
	byStreet := map[string][]core.Feature{}
	for _, f := range d.Features {
		if f.Street == "" || math.IsNaN(f.BOB) {
			continue
		}
		byStreet[f.Street] = append(byStreet[f.Street], f)
	}
	if len(byStreet) == 0 {
		return components.EmptyState(
			"Geen straatnamen of metingen",
			"Deze laag heeft geen BOB of geen STRAATNAAM gevuld",
			s.w,
		)
	}
	deepest := ""
	deepestBOB := math.Inf(1)
	for street, fs := range byStreet {
		for _, f := range fs {
			if f.BOB < deepestBOB {
				deepestBOB = f.BOB
				deepest = street
			}
		}
	}
	features := byStreet[deepest]
	var b strings.Builder
	b.WriteString(components.Strong(fmt.Sprintf("%s — diepste straat (%d buizen)", deepest, len(features))))
	b.WriteByte('\n')
	b.WriteString(components.Muted(fmt.Sprintf("diepste buisbodem: %.2f m NAP", deepestBOB)))
	b.WriteByte('\n')
	b.WriteByte('\n')
	// Simple horizontal section: a dot per pipe at its BOB height. Depth runs
	// downward: deeper pipes end up LOWER on screen.
	// (Later refinement: real per-pipe x-extents, maaiveld band above.)
	sorted := append([]core.Feature(nil), features...)
	sortByBOB(sorted)
	lo, hi := math.Inf(1), math.Inf(-1)
	for _, f := range sorted {
		lo = math.Min(lo, f.BOB)
		hi = math.Max(hi, f.BOB)
	}
	if hi-lo < 0.1 {
		hi = lo + 0.1
	}
	rows := s.h - 8
	if rows < 4 {
		rows = 4
	}
	for _, f := range sorted {
		if len(sorted) > 30 && rows > 0 && strings.Count(b.String(), "\n") > rows {
			remaining := len(sorted) - strings.Count(b.String(), "●")
			b.WriteString(components.Muted(fmt.Sprintf("  … en %d meer", remaining)))
			break
		}
		depth := (f.BOB - lo) / (hi - lo) // 0 = shallowest, 1 = deepest
		mark := components.Badge(components.StatusOK, "")
		bar := strings.Repeat(" ", int(depth*float64(s.w-20)))
		b.WriteString(fmt.Sprintf("%8.2f %s%s\n", f.BOB, bar, mark))
	}
	return strings.TrimRight(b.String(), "\n")
}

func sortByBOB(v []core.Feature) {
	for i := 1; i < len(v); i++ {
		for j := i; j > 0; j-- {
			a, b := v[j-1].BOB, v[j].BOB
			if math.IsNaN(a) || (!math.IsNaN(b) && b < a) {
				v[j-1], v[j] = v[j], v[j-1]
			}
		}
	}
}

func min2(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max2(a, b int) int {
	if a > b {
		return a
	}
	return b
}
