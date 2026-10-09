package screens

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"waterketen/internal/core"
	"waterketen/internal/ui"
	"waterketen/internal/ui/components"
)

// Kaart is panel 5: the live atlas. It renders the core.Render grid through
// the style table — the ONLY place where core's semantic style keys become
// lipgloss styles. All the colours come from theme tokens; per-layer ink
// lives in the components package as a fixed map.
type Kaart struct {
	root      string
	lib       *core.Library
	assets    []core.AssetMeta
	data      []*LayerData
	scene     *core.Scene
	view      core.View
	hasView   bool
	active    string
	overlay   bool
	mode      string
	labels    bool
	selected  *core.Feature
	street    string
	w, h      int
	render    *core.Render
	needsDraw bool
	spinner   string
	err       string
	search    string
	searching bool
}

// Busy reports bounded work in flight: the atlas is being drawn.
func (s *Kaart) Busy() bool { return s.scene != nil && (s.render == nil || s.needsDraw) }

// SetSpinner receives the app's current busy-spinner frame.
func (s *Kaart) SetSpinner(frame string) { s.spinner = frame }

// NewKaart initialises the atlas screen. Loads via the loader commands off
// the event loop; the screen stays responsive while reading.
func NewKaart(root string, lib *core.Library, assets []core.AssetMeta) *Kaart {
	return &Kaart{
		root: root, lib: lib, assets: assets,
		overlay: true, mode: "netwerk", labels: true, active: "riool",
		needsDraw: true,
	}
}

func (s *Kaart) ID() ScreenID { return ScreenKaart }
func (s *Kaart) KeyMap() []key.Binding {
	return []key.Binding{
		ui.Keys.PanLeft, ui.Keys.PanRight, ui.Keys.Up, ui.Keys.Down,
		ui.Keys.ZoomIn, ui.Keys.ZoomOut, ui.Keys.Fit,
		ui.Keys.PrevLayer, ui.Keys.NextLayer, ui.Keys.Overlay, ui.Keys.Labels,
		ui.Keys.Search, ui.Keys.Quit,
	}
}
func (s *Kaart) SetSize(w, h int) {
	if w != s.w || h != s.h {
		s.needsDraw = true
	}
	s.w, s.h = w, h
}
func (s *Kaart) Info() string {
	if s.scene == nil {
		return s.err
	}
	if s.render == nil {
		return "tekenen…"
	}
	drawn := 0
	for _, n := range s.render.Counts {
		drawn += n
	}
	return fmt.Sprintf("%s · %s features zichtbaar", s.active, thousandsShort(drawn))
}

// SetData receives the loaded layer data.
func (s *Kaart) SetData(d []*LayerData) {
	s.data = d
	s.scene = nil
	s.needsDraw = true
}

// SceneMsg carries a built scene back into the app loop.
type SceneMsg struct {
	Scene *core.Scene
	Err   error
}

// BuildSceneCmd constructs the atlas scene off the event loop.
func (s *Kaart) BuildSceneCmd() tea.Cmd {
	data := s.data
	if len(data) == 0 {
		return func() tea.Msg { return SceneMsg{Err: fmt.Errorf("geen managed data")} }
	}
	return func() tea.Msg {
		var layers []core.SceneSource
		for _, d := range data {
			layers = append(layers, core.SceneSource{Name: d.Layer, Features: d.Features})
		}
		scene, err := core.NewScene(data[0].Area, layers, nil)
		if err != nil {
			return SceneMsg{Err: err}
		}
		return SceneMsg{Scene: scene}
	}
}

// SetScene commits a built scene and triggers a redraw.
func (s *Kaart) SetScene(scene *core.Scene) {
	s.scene = scene
	s.err = ""
	s.view = core.FitView(scene.Bounds)
	s.hasView = true
	s.needsDraw = true
}

// SetError shows a load failure inside the panel.
func (s *Kaart) SetError(err error) { s.err = err.Error(); s.scene = nil }

// DrawnMsg carries a finished render back into the app loop.
type DrawnMsg struct{ Render *core.Render }

// DrawCmd renders off the event loop. It always renders the CURRENT view —
// a resize/pan/zoom issued during the draw invalidates via NeedsDraw and the
// app re-queues.
func (s *Kaart) DrawCmd() tea.Cmd {
	if s.scene == nil || !s.hasView {
		return nil
	}
	scene, view := s.scene, s.view
	w, h := s.innerSize()
	active, overlay, mode, labels := s.active, s.overlay, s.mode, s.labels
	selected, street := s.selected, s.street
	return func() tea.Msg {
		r := core.RenderScene(scene, view, w, h,
			active, overlay, mode, labels, true, selected, street, nil, true)
		return DrawnMsg{Render: &r}
	}
}

// Commit stores the finished render.
func (s *Kaart) Commit(r *core.Render) { s.render = r; s.needsDraw = false }

// NeedsDraw reports whether the current view has no committed render yet.
func (s *Kaart) NeedsDraw() bool { return s.needsDraw }

func (s *Kaart) innerSize() (int, int) {
	return max2(1, s.w-4), max2(1, s.h-8)
}

// ActiveLayerChoices returns the cycle order for [ and ].
func (s *Kaart) ActiveLayerChoices() []string {
	if s.scene == nil {
		return nil
	}
	var out []string
	for _, l := range s.scene.VisibleLayers(s.active, true, nil) {
		out = append(out, l)
	}
	return out
}

func (s *Kaart) cycleLayer(step int) {
	choices := s.ActiveLayerChoices()
	if len(choices) < 2 {
		return
	}
	i := 0
	for j, c := range choices {
		if c == s.active {
			i = j
			break
		}
	}
	s.active = choices[(i+step+len(choices))%len(choices)]
	s.selected = nil
	s.needsDraw = true
}

func (s *Kaart) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch m := msg.(type) {
	case tea.KeyMsg:
		if s.scene == nil {
			return s, nil
		}
		w, h := s.innerSize()
		switch {
		case key.Matches(m, ui.Keys.PanLeft):
			s.view.Pan(-1, 0, s.scene.Bounds, w, h)
		case key.Matches(m, ui.Keys.PanRight):
			s.view.Pan(1, 0, s.scene.Bounds, w, h)
		case key.Matches(m, ui.Keys.Up):
			s.view.Pan(0, 1, s.scene.Bounds, w, h)
		case key.Matches(m, ui.Keys.Down):
			s.view.Pan(0, -1, s.scene.Bounds, w, h)
		case key.Matches(m, ui.Keys.ZoomIn):
			s.view.ZoomAt(1.5, float64(w)/2, float64(h)/2, s.scene.Bounds, w, h)
		case key.Matches(m, ui.Keys.ZoomOut):
			s.view.ZoomAt(2.0/3.0, float64(w)/2, float64(h)/2, s.scene.Bounds, w, h)
		case key.Matches(m, ui.Keys.Fit):
			s.view = core.FitView(s.scene.Bounds)
			s.street = ""
			s.selected = nil
		case key.Matches(m, ui.Keys.Overlay):
			s.overlay = !s.overlay
			s.selected = nil
		case key.Matches(m, ui.Keys.Labels):
			s.labels = !s.labels
		case key.Matches(m, ui.Keys.NextLayer):
			s.cycleLayer(1)
			return s, nil // cycleLayer sets needsDraw
		case key.Matches(m, ui.Keys.PrevLayer):
			s.cycleLayer(-1)
			return s, nil
		default:
			return s, nil
		}
		s.needsDraw = true
		return s, nil
	case tea.MouseMsg:
		if s.scene == nil || !s.hasView {
			return s, nil
		}
		w, h := s.innerSize()
		switch m.Button {
		case tea.MouseButtonWheelUp:
			s.view.ZoomAt(1.25, float64(m.X)-2, float64(m.Y)-2, s.scene.Bounds, w, h)
		case tea.MouseButtonWheelDown:
			s.view.ZoomAt(0.8, float64(m.X)-2, float64(m.Y)-2, s.scene.Bounds, w, h)
		case tea.MouseButtonLeft:
			if m.Action != tea.MouseActionPress {
				return s, nil
			}
			x, y := s.view.Frame(s.scene.Bounds, w, h).World(float64(m.X)-2, float64(m.Y)-2)
			if f := s.scene.Nearest(x, y, s.view.Frame(s.scene.Bounds, w, h).MetresPerDot*5,
				s.active, s.overlay, nil); f != nil {
				s.selected = f
				s.street = ""
				s.needsDraw = true
			}
		}
		return s, nil
	}
	return s, nil
}

func (s *Kaart) View() string {
	if s.err != "" {
		return components.Error("Atlas: " + s.err)
	}
	if s.scene == nil {
		return components.EmptyState(
			"Geen kaartdata",
			"Laad een laag met: waterketen fetch --gebied delfshaven --laag riool",
			s.w,
		)
	}
	if s.render == nil || s.needsDraw {
		return components.Busy(s.spinner, "Kaart wordt getekend…")
	}
	return RenderGrid(s.render, s.w-4)
}

// RenderGrid is the only core→lipgloss translation point for atlas output.
// The legend's colours live in the per-layer hex table in core's render
// (geometry logic); here we turn each semantic style key into concrete
// lipgloss styling using the UI tokens where they apply.
func RenderGrid(r *core.Render, maxWidth int) string {
	var b strings.Builder
	for y, row := range r.Cells {
		var cur string
		var curStyle core.CellStyle
		flush := func() {
			if cur == "" {
				return
			}
			b.WriteString(styleFor(curStyle).Render(cur))
			cur = ""
		}
		for x, cell := range row {
			if cell.Style != curStyle {
				flush()
				curStyle = cell.Style
			}
			cur += string(cell.Rune)
			_ = x
		}
		flush()
		if y < len(r.Cells)-1 {
			b.WriteByte('\n')
		}
	}
	_ = maxWidth
	return b.String()
}

// layerInk is the map-ink hex per layer. This table is data — it's the
// colour the ATLAS has established per layer, not UI chrome. It lives next
// to the grid translation because that's the single place both views agree on.
var layerInk = map[string]string{
	"riool":               "#38d9ff",
	"drinkwaterleidingen": "#ffa95c",
	"drinkwater":          "#476c91",
	"straten":             "#627d91",
	"water":               "#416e91",
	"spoor":               "#b19bbf",
	"putten":              "#38d9ff",
	"gemalen":             "#cc94ff",
	"groen":               "#5a8f6f",
	"gebouwen":            "#4a5a68",
}

var bobInk = map[string]string{
	"very-deep": "#b28cff",
	"deep":      "#589dff",
	"mid":       "#38d9ff",
	"shallow":   "#8ee6a8",
	"above":     "#ffcc73",
	"unknown":   "#73879d",
}

func styleFor(s core.CellStyle) lipgloss.Style {
	str := string(s)
	switch s {
	case core.StyleBlank, core.StyleGrid:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#172b3e"))
	case core.StyleNorth:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#a8c4df")).Bold(true)
	case core.StyleScale:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#b3d3e9")).Bold(true)
	case core.StyleLabel:
		// Transparent label: background stays the terminal's own.
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#b9d5e7")).
			Bold(true)
	case core.StyleHighlight:
		// Highlight by ink + bold only — a filled box would fight the terminal
		// background the rest of the UI now lets through.
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ffffff")).
			Bold(true).Underline(true)
	}
	if strings.HasPrefix(str, "layer:") {
		if hex, ok := layerInk[strings.TrimPrefix(str, "layer:")]; ok {
			return lipgloss.NewStyle().Foreground(lipgloss.Color(hex))
		}
	}
	if strings.HasPrefix(str, "bob:") {
		if hex, ok := bobInk[strings.TrimPrefix(str, "bob:")]; ok {
			return lipgloss.NewStyle().Foreground(lipgloss.Color(hex))
		}
	}
	if strings.HasPrefix(str, "hex:") {
		return lipgloss.NewStyle().Foreground(lipgloss.Color(strings.TrimPrefix(str, "hex:")))
	}
	if strings.HasPrefix(str, "material:") {
		// materials share the fixed palette — resolve via any stable hash
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#9ddeda"))
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#8ee6a8"))
}

func thousandsShort(n int) string {
	if n >= 1000 {
		return fmt.Sprintf("%.1fk", float64(n)/1000)
	}
	return fmt.Sprintf("%d", n)
}

var _ = math.Inf // keep the import honest until the inspector lands
