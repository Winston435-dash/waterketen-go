package screens

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"waterketen/internal/core"
	"waterketen/internal/ui"
	"waterketen/internal/ui/components"
)

// Kaart is panel 5: the live atlas. It renders the core.Render grid through
// components.RenderAtlas — the ONE core→lipgloss translation point — so this
// screen imports no lipgloss and holds no colour tables. The atlas data ink is
// core's (core.CellColour); the kit adds chrome tokens and emphasis.
//
// Two data worlds, one panel:
//
//   - LIBRARY (preferred): the published snapshot under $DATA_ROOT/library with
//     its 115 CBS areas. Like the Python app, the panel opens on the whole
//     municipality (GM0599) — 'g' narrows to a wijk or buurt and refits the view.
//     Viewports come from the GeoPackage R-trees, so a 87.146-row layer costs
//     only its viewport. A viewport read is expensive (Rotterdam over four
//     layers ≈ 3,4 s), so it is coalesced: panning while a read is in flight
//     does not queue a second one.
//   - MANAGED (fallback): the per-area assets from `waterketen fetch`, used when
//     no library snapshot exists. Same rendering, no area picker.
type Kaart struct {
	root   string
	lib    *core.Library
	assets []core.AssetMeta

	atlas   *core.LibraryAtlas // nil until the library opens
	layers  []string           // chosen layer set (library world)
	opening bool               // the area registry is being read
	loads   int                // viewport reads in flight (0 or 1)
	stale   bool               // a read finished while another was in flight

	data  []*LayerData
	scene *core.Scene
	view  core.View

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
	picker    *Picker // open overlay, nil when closed
}

// SetSpinner receives the app's current busy-spinner frame.
func (s *Kaart) SetSpinner(frame string) { s.spinner = frame }

// Busy reports bounded work in flight: a viewport read, or a pending draw.
func (s *Kaart) Busy() bool { return s.opening || s.loads > 0 || s.needsDraw }

// NewKaart initialises the atlas screen. Loading happens through the loader
// commands off the event loop; the screen stays responsive while reading.
func NewKaart(root string, lib *core.Library, assets []core.AssetMeta) *Kaart {
	return &Kaart{
		root: root, lib: lib, assets: assets,
		overlay: true, mode: "netwerk", labels: true, active: "riool",
		needsDraw: true,
	}
}

func (s *Kaart) ID() ScreenID { return ScreenKaart }
func (s *Kaart) KeyMap() []key.Binding {
	if s.picker != nil {
		return []key.Binding{ui.Keys.Up, ui.Keys.Down, ui.Keys.Enter, ui.Keys.Back}
	}
	bindings := []key.Binding{
		ui.Keys.PanLeft, ui.Keys.PanRight, ui.Keys.Up, ui.Keys.Down,
		ui.Keys.ZoomIn, ui.Keys.ZoomOut, ui.Keys.Fit,
		ui.Keys.PrevLayer, ui.Keys.NextLayer, ui.Keys.Overlay, ui.Keys.Labels,
		ui.Keys.Search, ui.Keys.Quit,
	}
	if s.atlas != nil {
		bindings = append(bindings, ui.Keys.Area, ui.Keys.Layers)
	}
	return bindings
}

func (s *Kaart) SetSize(w, h int) {
	resized := (w != s.w || h != s.h) && s.hasView
	s.w, s.h = w, h
	if resized {
		// A resize changes the visible world: reload, do not rescale a stale
		// frame. This is the one place a pan-like event costs a disk read.
		s.requestLoad()
	}
}

func (s *Kaart) Info() string {
	if s.picker != nil {
		return "overlay open"
	}
	if s.atlas != nil {
		if s.err != "" {
			return s.err
		}
		return fmt.Sprintf("%s · %d lagen", s.atlas.Area.Code, len(s.layers))
	}
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

// --- loading ---------------------------------------------------------------

// SetData receives the loaded managed-layer data (the fallback world).
func (s *Kaart) SetData(d []*LayerData) {
	s.data = d
	if s.atlas == nil && !s.opening {
		s.scene = nil
		s.needsDraw = true
	}
}

// SceneMsg carries a built managed-asset scene back into the app loop.
type SceneMsg struct {
	Scene *core.Scene
	Err   error
}

// LibraryOpenMsg carries the opened area registry back to the app loop.
type LibraryOpenMsg struct {
	Atlas *core.LibraryAtlas
	Err   error
}

// LibraryLayerMsg carries one loaded library viewport back to the app loop.
type LibraryLayerMsg struct {
	Scene *core.Scene
	Err   error
}

// BuildSceneCmd starts the panel's data path: open the library when there is a
// snapshot, else build a scene from the managed assets.
func (s *Kaart) BuildSceneCmd() tea.Cmd {
	if s.lib != nil {
		s.opening = true
		return func() tea.Msg {
			atlas, err := core.NewLibraryAtlas(s.lib, core.DefaultAreaCode)
			return LibraryOpenMsg{Atlas: atlas, Err: err}
		}
	}
	return s.managedSceneCmd()
}

func (s *Kaart) managedSceneCmd() tea.Cmd {
	if len(s.data) == 0 {
		return func() tea.Msg { return SceneMsg{Err: fmt.Errorf("geen managed data")} }
	}
	data := s.data
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

// LibraryLayerCmd reads the current viewport from the snapshot off the event
// loop. Returns nil when no atlas is open, so a caller can drop it into a batch.
func (s *Kaart) LibraryLayerCmd() tea.Cmd {
	if s.atlas == nil {
		return nil
	}
	atlas, view, names := s.atlas, s.view, s.layers
	w, h := s.innerSize()
	return func() tea.Msg {
		scene, err := atlas.Load(view, w, h, names, 0)
		if err != nil {
			return LibraryLayerMsg{Err: err}
		}
		return LibraryLayerMsg{Scene: scene}
	}
}

// SetAtlas commits an opened library atlas: it adopts the area, the layer set
// and the fit view, then triggers the first viewport read.
func (s *Kaart) SetAtlas(atlas *core.LibraryAtlas) tea.Cmd {
	s.opening = false
	s.atlas = atlas
	s.err = ""
	if s.layers == nil {
		s.layers = atlas.Library.HasLayers("straten", "water", "riool", "drinkwaterleidingen")
	}
	s.view = core.FitView(atlas.Area.Bounds)
	s.hasView = true
	s.scene = nil
	s.render = nil
	return s.requestLoad()
}

// SetArea switches the selected area: refits the view and reloads the viewport.
func (s *Kaart) SetArea(code string) tea.Cmd {
	if s.atlas == nil {
		return nil
	}
	next, err := core.NewLibraryAtlas(s.atlas.Library, code)
	if err != nil {
		s.err = err.Error()
		return nil
	}
	return s.SetAtlas(next)
}

// SetLayers replaces the visible layer set and reloads.
func (s *Kaart) SetLayers(names []string) tea.Cmd {
	s.layers = core.OrderLayers(names)
	if len(s.layers) > 0 && !containsString(s.layers, s.active) {
		s.active = s.layers[0]
	}
	return s.requestLoad()
}

// requestLoad schedules a viewport read. Only one read runs at a time: while one
// is in flight this only marks the view stale, and the finished read re-queues
// because `stale` is set. Without this a held-down arrow key would start a
// multi-second disk read per repeat.
func (s *Kaart) requestLoad() tea.Cmd {
	s.needsDraw = true
	if s.loads > 0 {
		s.stale = true
		return nil
	}
	s.loads++
	return s.LibraryLayerCmd()
}

// CommitLoad stores a finished viewport read and re-queues when the view moved
// while it ran.
func (s *Kaart) CommitLoad(scene *core.Scene) tea.Cmd {
	s.loads = max(0, s.loads-1)
	s.scene = scene
	s.err = ""
	s.needsDraw = true
	if s.stale {
		s.stale = false
		return s.requestLoad()
	}
	return nil
}

// FailLoad records a failed viewport read.
func (s *Kaart) FailLoad(err error) tea.Cmd {
	s.loads = max(0, s.loads-1)
	s.err = err.Error()
	if s.stale {
		s.stale = false
	}
	return nil
}

// --- pickers ---------------------------------------------------------------

// Picker returns the open overlay, if any.
func (s *Kaart) Picker() *Picker { return s.picker }

// OpenAreaPicker opens the 'g' overlay over the snapshot's areas.
func (s *Kaart) OpenAreaPicker() {
	if s.atlas == nil {
		return
	}
	p := NewAreaPicker(s.atlas.Areas)
	p.SetSize(s.w, s.h)
	s.picker = p
}

// OpenLayerPicker opens the 't' overlay over the layers the snapshot carries.
func (s *Kaart) OpenLayerPicker() {
	if s.atlas == nil {
		return
	}
	available := s.atlas.Library.HasLayers(core.DrawOrder...)
	p := NewLayerPicker(available, s.layers, core.DrawOrder)
	p.SetSize(s.w, s.h)
	s.picker = p
}

// ClosePicker closes the overlay; the caller applies any confirmed result.
func (s *Kaart) ClosePicker() { s.picker = nil }

// --- scene / render --------------------------------------------------------

// SetScene commits a built managed-asset scene and triggers a redraw.
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

// DrawCmd renders off the event loop against the CURRENT view — a resize, pan
// or zoom issued during the draw invalidates via NeedsDraw and the app requeues.
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
	if s.atlas != nil {
		return append([]string(nil), s.layers...)
	}
	if s.scene == nil {
		return nil
	}
	return append([]string(nil), s.scene.VisibleLayers(s.active, true, nil)...)
}

// cycleLayer walks the layer cycle. In the library world this only changes the
// ACTIVE layer — narrowing the drawn set is 't', which is a deliberate choice
// rather than a toggle away from the whole point of the map.
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

// --- update ----------------------------------------------------------------

func (s *Kaart) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch m := msg.(type) {
	case LibraryOpenMsg:
		if m.Err != nil {
			// No usable snapshot: fall back to the managed assets rather than
			// leaving panel 5 blank while data exists on disk.
			s.opening = false
			s.err = m.Err.Error()
			return s, s.managedSceneCmd()
		}
		return s, s.SetAtlas(m.Atlas)
	case LibraryLayerMsg:
		if m.Err != nil {
			return s, s.FailLoad(m.Err)
		}
		return s, s.CommitLoad(m.Scene)
	case tea.KeyMsg:
		if s.picker != nil {
			result, done := s.picker.Update(m)
			if !done {
				return s, nil
			}
			s.picker = nil
			if !result.Confirm {
				return s, nil
			}
			if result.Toggled {
				return s, s.SetLayers(result.Marked)
			}
			return s, s.SetArea(result.Key)
		}
		return s, s.handleKey(m)
	case tea.MouseMsg:
		if s.picker != nil {
			return s, nil
		}
		return s, s.handleMouse(m)
	}
	return s, nil
}

func (s *Kaart) handleKey(m tea.KeyMsg) tea.Cmd {
	if s.scene == nil {
		return nil
	}
	w, h := s.innerSize()
	pan := func(dx, dy float64) tea.Cmd {
		dx = dx * (s.scene.Bounds.MaxX - s.scene.Bounds.MinX) * .18
		dy = dy * (s.scene.Bounds.MaxY - s.scene.Bounds.MinY) * .18
		return s.moveView(dx, dy)
	}
	switch {
	case key.Matches(m, ui.Keys.PanLeft):
		return pan(-1, 0)
	case key.Matches(m, ui.Keys.PanRight):
		return pan(1, 0)
	case key.Matches(m, ui.Keys.Up):
		return pan(0, 1)
	case key.Matches(m, ui.Keys.Down):
		return pan(0, -1)
	case key.Matches(m, ui.Keys.ZoomIn):
		return s.zoomView(1.5, float64(w)/2, float64(h)/2, w, h)
	case key.Matches(m, ui.Keys.ZoomOut):
		return s.zoomView(2.0/3.0, float64(w)/2, float64(h)/2, w, h)
	case key.Matches(m, ui.Keys.Fit):
		s.street = ""
		s.selected = nil
		if s.atlas != nil {
			return s.SetAtlas(s.atlas)
		}
		s.view = core.FitView(s.scene.Bounds)
		s.needsDraw = true
		return nil
	case key.Matches(m, ui.Keys.Overlay):
		s.overlay = !s.overlay
		s.selected = nil
	case key.Matches(m, ui.Keys.Labels):
		s.labels = !s.labels
	case key.Matches(m, ui.Keys.Area):
		s.OpenAreaPicker()
		return nil
	case key.Matches(m, ui.Keys.Layers):
		s.OpenLayerPicker()
		return nil
	case key.Matches(m, ui.Keys.NextLayer):
		s.cycleLayer(1)
		return nil
	case key.Matches(m, ui.Keys.PrevLayer):
		s.cycleLayer(-1)
		return nil
	}
	return nil
}

// moveView pans the view and reloads the viewport in the library world. `east`
// and `north` are −1/+1 directions; the library world converts them to metres
// against the AREA bounds, the managed world uses core.View.Pan.
func (s *Kaart) moveView(east, north float64) tea.Cmd {
	if s.atlas != nil {
		b := s.atlas.Area.Bounds
		s.view.CX += east * (b.MaxX - b.MinX) * .18
		s.view.CY += north * (b.MaxY - b.MinY) * .18
		return s.requestLoad()
	}
	s.view.Pan(east, north, s.scene.Bounds, s.w, s.h)
	s.needsDraw = true
	return nil
}

// zoomView zooms in the library world (which changes the viewport, so it
// reloads) or the managed world (which only changes the scale).
func (s *Kaart) zoomView(factor, col, row float64, w, h int) tea.Cmd {
	if s.atlas != nil {
		s.view.ZoomAt(factor, col, row, s.atlas.Area.Bounds, w, h)
		return s.requestLoad()
	}
	s.view.ZoomAt(factor, col, row, s.scene.Bounds, w, h)
	s.needsDraw = true
	return nil
}

func (s *Kaart) handleMouse(m tea.MouseMsg) tea.Cmd {
	if s.scene == nil || !s.hasView {
		return nil
	}
	w, h := s.innerSize()
	bounds := s.scene.Bounds
	if s.atlas != nil {
		bounds = s.atlas.Area.Bounds
	}
	frame := s.view.Frame(bounds, w, h)
	switch m.Button {
	case tea.MouseButtonWheelUp:
		return s.zoomView(1.25, float64(m.X)-2, float64(m.Y)-2, w, h)
	case tea.MouseButtonWheelDown:
		return s.zoomView(0.8, float64(m.X)-2, float64(m.Y)-2, w, h)
	case tea.MouseButtonLeft:
		if m.Action != tea.MouseActionPress {
			return nil
		}
		x, y := frame.World(float64(m.X)-2, float64(m.Y)-2)
		if f := s.scene.Nearest(x, y, frame.MetresPerDot*5, s.active, s.overlay, nil); f != nil {
			s.selected = f
			s.street = ""
			s.needsDraw = true
		}
	}
	return nil
}

// --- view ------------------------------------------------------------------

func (s *Kaart) View() string {
	if s.picker != nil {
		return s.picker.View()
	}
	if s.err != "" {
		return components.Error("Atlas: " + s.err)
	}
	if s.opening || s.loads > 0 {
		return components.Busy(s.spinner, "Kaart wordt geladen…")
	}
	if s.scene == nil {
		return components.EmptyState(
			"Geen kaartdata",
			"Bouw de kaartbibliotheek (waterketen bibliotheek bouw) of haal een laag op",
			s.w,
		)
	}
	if s.render == nil || s.needsDraw {
		return components.Busy(s.spinner, "Kaart wordt getekend…")
	}
	return components.RenderAtlas(s.render)
}

func thousandsShort(n int) string {
	if n >= 1000 {
		return fmt.Sprintf("%.1fk", float64(n)/1000)
	}
	return fmt.Sprintf("%d", n)
}

func containsString(hay []string, needle string) bool {
	for _, h := range hay {
		if h == needle {
			return true
		}
	}
	return false
}
