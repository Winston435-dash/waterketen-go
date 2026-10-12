package screens

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

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
	failures []error
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
	return fmt.Sprintf("%d lagen geladen", len(s.data))
}

// DataMsg carries loaded analysis data back to the app.
type DataMsg struct {
	Layers []*LayerData
	Errors []error
}

// SetData stores the loaded assets, clears the busy flag and resets the scroll.
func (s *AnalyseScreen) SetData(d []*LayerData) {
	s.data, s.scroll, s.busy = d, 0, false
	s.lines = s.renderLines()
}

func (s *AnalyseScreen) SetErrors(errors []error) {
	s.failures = errors
	s.lines = s.renderLines()
}

// Busy reports bounded work in flight: the managed assets are being read.
func (s *AnalyseScreen) Busy() bool { return s.busy }

// SetBusy marks the initial managed-data load as in flight.
func (s *AnalyseScreen) SetBusy(v bool) { s.busy = v }

// SetSpinner receives the app's current busy-spinner frame.
func (s *AnalyseScreen) SetSpinner(frame string) { s.spinner = frame }

func (s *AnalyseScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch m := msg.(type) {
	case tea.KeyMsg:
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
	if len(s.data) == 0 && len(s.failures) == 0 {
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
	for _, err := range s.failures {
		out = append(out, components.Error("Leesfout: "+err.Error()))
	}
	return out
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
	sort.SliceStable(sorted, func(i, j int) bool {
		a, b := sorted[i].BOB, sorted[j].BOB
		if math.IsNaN(a) {
			return false
		}
		if math.IsNaN(b) {
			return true
		}
		return a < b
	})
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
		bar := strings.Repeat(" ", int(depth*float64(max(0, s.w-20))))
		b.WriteString(fmt.Sprintf("%8.2f %s%s\n", f.BOB, bar, mark))
	}
	return strings.TrimRight(b.String(), "\n")
}
