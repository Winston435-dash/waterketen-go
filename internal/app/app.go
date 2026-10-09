// Package app is the top-level Bubble Tea shell. It owns the fixed anatomy
// every screen renders inside:
//
//	Header   — app title, screen title, current screen key hints
//	Content  — the active screen's full-width Panel
//	CommandBar — bottom line of screen-scoped keys
//
// Screens never draw their own borders; they get `SetSize` with the OUTER
// panel box and return body content from `View` which the App wraps.
//
// The App also owns the ONE busy-animation spinner. It ticks the spinner only
// while a screen reports bounded work in flight (`Busy()`), broadcasts each
// frame to the screens that display it, and lets the tick chain die the moment
// all work is done — so the app never repaints while idle.
package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"waterketen/internal/core"
	"waterketen/internal/ui"
	"waterketen/internal/ui/components"
	"waterketen/internal/ui/screens"
)

// Model is the one Bubble Tea model for the whole app.
type Model struct {
	root    string
	screens []screens.Screen
	active  int
	width   int
	height  int
	spinner spinner.Model
	ticking bool // the spinner's tick chain is alive
	drawing bool
	drawGen int
	data    []*screens.LayerData
	cmdBar  components.CommandBar
	ready   bool
}

// spinnerScreener is implemented by the screens that display the busy spinner.
// A screen returns Busy() true while it has bounded work in flight; the app
// ticks the spinner for exactly as long as any screen does.
type spinnerScreener interface {
	SetSpinner(frame string)
	Busy() bool
}

// New builds the initial model. Probes and data-loading start via Init.
func New(root string) Model {
	if root == "" {
		root = core.DataRoot()
	}
	return Model{
		root:    root,
		screens: screens.New(root),
		cmdBar:  components.NewCommandBar(),
		spinner: components.NewSpinner(),
		ticking: true, // Init returns the first tick; the chain runs from there.
	}
}

// Init kicks off the live probe and the data load — both off the event loop —
// and starts the busy spinner.
func (m Model) Init() tea.Cmd {
	// Mark the initial work in flight. These mutations reach the running
	// program's screens through the shared slice, so they persist. The probe
	// also returns its bar's first frame command, which must go to the runtime.
	cmds := []tea.Cmd{screens.ProbeCmd(), loadAllDataCmd(m.root), m.spinner.Tick}
	for _, s := range m.screens {
		switch p := s.(type) {
		case *screens.Bronnen:
			cmds = append(cmds, p.BeginProbe())
		case *screens.AnalyseScreen:
			p.SetBusy(true)
		}
	}
	return tea.Batch(cmds...)
}

// startSpinner turns the busy spinner on, starting its tick chain if it is not
// already alive. Returns nil when the chain already runs, so a caller can drop
// it straight into a tea.Batch.
func (m *Model) startSpinner() tea.Cmd {
	if m.ticking {
		return nil
	}
	m.ticking = true
	return m.spinner.Tick
}

// anyBusy reports whether any screen has bounded work in flight.
func (m Model) anyBusy() bool {
	for _, s := range m.screens {
		if b, ok := s.(spinnerScreener); ok && b.Busy() {
			return true
		}
	}
	return false
}

// probeRunning reports whether the Bronnen panel already has a probe run in
// flight, so 'v' cannot start a second run over the top of the first.
func (m Model) probeRunning() bool {
	for _, s := range m.screens {
		if b, ok := s.(*screens.Bronnen); ok {
			return b.Busy()
		}
	}
	return false
}

func loadAllDataCmd(root string) tea.Cmd {
	return func() tea.Msg {
		assets, _ := core.LoadAssets(root)
		return screens.DataMsg(screens.LoadAssets(assets))
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.ready = true
		m = m.applySize()
		return m, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		frame := m.spinner.View()
		for _, s := range m.screens {
			if sp, ok := s.(spinnerScreener); ok {
				sp.SetSpinner(frame)
			}
		}
		if !m.anyBusy() {
			// All work done: let the chain die. startSpinner revives it on the
			// next run, so the app is idle — zero repaints — in between.
			m.ticking = false
			return m, nil
		}
		return m, cmd
	case screens.TickMsg:
		// Bubbles' progress-bar animation frame: hand it to the panel that owns
		// the bar. It replies with the next frame while a run is in flight, so
		// the chain dies on its own once the probe finishes.
		for _, s := range m.screens {
			if b, ok := s.(*screens.Bronnen); ok {
				return m, b.Tick(msg)
			}
		}
		return m, nil
	case screens.HealthPartialMsg:
		for _, s := range m.screens {
			b, ok := s.(*screens.Bronnen)
			if !ok {
				continue
			}
			done, cmd := b.AddProbe(core.SourceHealth(msg))
			if done {
				b.FinishProbe()
				return m, cmd // nil: the bar's frame chain ends here
			}
			return m, tea.Batch(cmd, m.startSpinner())
		}
		return m, nil
	case screens.DataMsg:
		m.data = msg
		for _, s := range m.screens {
			switch p := s.(type) {
			case *screens.AnalyseScreen:
				p.SetData(msg)
			case *screens.ProfielScreen:
				p.SetData(msg)
			case *screens.Kaart:
				p.SetData(msg)
			}
		}
		// The Kaart always has a world to build: the library snapshot when one
		// exists, else the managed assets.
		if k := m.kaart(); k != nil {
			return m, tea.Batch(m.startSpinner(), k.BuildSceneCmd())
		}
		return m, nil
	case screens.LibraryOpenMsg, screens.LibraryLayerMsg, screens.SearchQueryMsg:
		// These belong to the Kaart even when another panel is active: the load
		// was started before the user switched, and dropping it would leave the
		// map spinning forever.
		i, k := m.kaartEntry()
		if k == nil {
			return m, nil
		}
		next, cmd := k.Update(msg)
		m.screens[i] = next
		// A committed viewport is a NEW scene, so the draw must be dispatched
		// here. Relying on afterScreenUpdate alone only triggers a draw for the
		// ACTIVE panel, and it may run before the scene is set — which stranded
		// the map on its spinner with the data already in memory.
		if k.NeedsDraw() {
			m.drawGen++
			m.drawing = true
			return m, tea.Batch(cmd, m.startSpinner(),
				tea.Tick(0, func(time.Time) tea.Msg { return drawFrameMsg{gen: m.drawGen} }))
		}
		return m, tea.Batch(cmd, m.afterScreenUpdate())
	case screens.SceneMsg:
		if k := m.kaart(); k != nil {
			if msg.Err != nil {
				k.SetError(msg.Err)
			} else {
				k.SetScene(msg.Scene)
			}
			m.drawGen++
			m.drawing = true
			return m, tea.Batch(m.startSpinner(),
				tea.Tick(0, func(time.Time) tea.Msg { return drawFrameMsg{gen: m.drawGen} }))
		}
		return m, nil
	case drawFrameMsg:
		if msg.gen != m.drawGen {
			return m, nil // stale
		}
		if k := m.kaart(); k != nil && k.NeedsDraw() {
			return m, k.DrawCmd()
		}
		return m, nil
	case screens.DrawnMsg:
		m.drawing = false
		if k := m.kaart(); k != nil {
			k.Commit(msg.Render)
			if k.NeedsDraw() {
				m.drawing = true
				return m, tea.Batch(m.startSpinner(),
					tea.Tick(0, func(time.Time) tea.Msg { return m.requestDraw() }))
			}
		}
		return m, nil
	case tea.KeyMsg:
		// An open overlay owns EVERY key: without this, typing a street name in
		// the search would also switch panels on 'g', quit on 'q', and pan the
		// map on the arrow keys.
		if k := m.kaart(); k != nil && m.screens[m.active].ID() == screens.ScreenKaart &&
			k.OverlayOpen() {
			next, cmd := k.Update(msg)
			m.screens[m.active] = next
			return m, tea.Batch(cmd, m.afterScreenUpdate())
		}
		// Global keys first — they work regardless of active screen.
		switch {
		case key.Matches(msg, ui.Keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, ui.Keys.Probe):
			if !m.probeRunning() {
				// Reset the panel AND dispatch the probes. BeginProbe only
				// resets state and starts the bar; without ProbeCmd here a
				// refresh would spin forever at 0/3.
				cmds := []tea.Cmd{m.startSpinner(), screens.ProbeCmd()}
				for _, s := range m.screens {
					if b, ok := s.(*screens.Bronnen); ok {
						cmds = append(cmds, b.BeginProbe())
					}
				}
				return m, tea.Batch(cmds...)
			}
			return m, nil
		case key.Matches(msg, ui.Keys.Redraw):
			if k := m.kaart(); k != nil && m.screens[m.active].ID() == screens.ScreenKaart {
				// Redraw means "draw the current view again". It must NOT call
				// SetScene: that dereferences a scene, and the old kCurrentScene
				// stub returned nil, so pressing 'r' panicked the whole app.
				// needsDraw alone makes the app re-render what it already has.
				k.MarkDirty()
				m.drawGen++
				return m, tea.Batch(m.startSpinner(),
					tea.Tick(0, func(time.Time) tea.Msg { return drawFrameMsg{gen: m.drawGen} }))
			}
			return m, nil
		}
		// j/k move between panels here, not inside screens — the screens never
		// both scroll AND switch. Letters stay free for content keys otherwise.
		if key.Matches(msg, ui.Keys.NextPanel) {
			m.active = (m.active + 1) % len(m.screens)
			return m, m.afterScreenUpdate()
		}
		if key.Matches(msg, ui.Keys.PrevPanel) {
			m.active = (m.active - 1 + len(m.screens)) % len(m.screens)
			return m, m.afterScreenUpdate()
		}
		// Number keys 1-6 select the panel directly.
		for i, info := range screens.Screens {
			if msg.String() == info.Key {
				m.active = i
				return m, m.afterScreenUpdate()
			}
		}
		// Forward to the active screen.
		next, cmd := m.screens[m.active].Update(msg)
		m.screens[m.active] = next
		return m, tea.Batch(cmd, m.afterScreenUpdate())
	case tea.MouseMsg:
		next, cmd := m.screens[m.active].Update(msg)
		m.screens[m.active] = next
		return m, tea.Batch(cmd, m.afterScreenUpdate())
	}
	return m, nil
}

// afterScreenUpdate reacts to a screen that may have started or finished work
// while handling an event: it (re)triggers an atlas draw when the map is dirty
// and (re)starts the busy spinner when the active screen went busy.
func (m *Model) afterScreenUpdate() tea.Cmd {
	draw := m.maybeTriggerDraw()
	var spin tea.Cmd
	if b, ok := m.screens[m.active].(spinnerScreener); ok && b.Busy() {
		spin = m.startSpinner()
	}
	return tea.Batch(draw, spin)
}

type drawFrameMsg struct{ gen int }

func (m Model) requestDraw() tea.Msg { return drawFrameMsg{gen: m.drawGen} }

func (m *Model) kaart() *screens.Kaart {
	_, k := m.kaartEntry()
	return k
}

// kaartEntry returns the Kaart screen together with its index, so a message can
// be routed to it without assuming it is the ACTIVE panel.
func (m *Model) kaartEntry() (int, *screens.Kaart) {
	for i, s := range m.screens {
		if k, ok := s.(*screens.Kaart); ok {
			return i, k
		}
	}
	return -1, nil
}

func (m *Model) maybeTriggerDraw() tea.Cmd {
	if m.screens[m.active].ID() != screens.ScreenKaart {
		return nil
	}
	if k := m.kaart(); k != nil && k.NeedsDraw() && !m.drawing {
		m.drawing = true
		m.drawGen++
		gen := m.drawGen
		return tea.Tick(0, func(time.Time) tea.Msg { return drawFrameMsg{gen: gen} })
	}
	return nil
}

func (m Model) applySize() Model {
	w, h := m.panelBox()
	for _, s := range m.screens {
		s.SetSize(w, h)
	}
	return m
}

// layout owns the OUTER content box used by both SetSize and rendering.
// Navigation and commands each occupy one row; content gets all columns.
type layout struct {
	panelW int
	panelH int
}

func (m Model) layout() layout {
	// Clamp the initial 0x0 size without reserving extra rows on short terminals.
	return layout{panelW: max(m.width, 1), panelH: max(m.height-2, 1)}
}

// panelBox computes the content panel's OUTER size (incl. border).
func (m Model) panelBox() (int, int) {
	l := m.layout()
	return l.panelW, l.panelH
}

func (m Model) View() string {
	if !m.ready {
		return "initialiseren…"
	}
	header := m.renderHeader()
	body := m.renderBody()
	bar := m.cmdBar.View(m.width, m.screens[m.active].KeyMap())
	return lipgloss.JoinVertical(lipgloss.Left, header, body, bar)
}

func (m Model) renderBody() string {
	screen := m.screens[m.active]
	l := m.layout()
	title := screenTitle(screen)
	// Keep area, layer and source context next to the content at every size.
	if info := screen.Info(); info != "" {
		title += " · " + info
	}
	return components.Panel(title, screen.View(), l.panelW, l.panelH, true)
}

// renderHeader shows all panel keys when they fit, neighbouring panels on a
// phone, and just the active panel on a very narrow terminal. Never sacrifice
// the current panel's label to the app name or an inactive neighbour.
func (m Model) renderHeader() string {
	accent := lipgloss.NewStyle().Foreground(ui.ColorPrimary).Bold(true)
	muted := lipgloss.NewStyle().Foreground(ui.ColorMuted)
	name := accent.Render("waterketen")
	tabs := make([]string, 0, len(screens.Screens))
	for i, entry := range screens.Screens {
		style := muted
		if i == m.active {
			style = accent
		}
		tabs = append(tabs, style.Render(entry.Key+" "+entry.Title))
	}
	full := " " + name + "  " + strings.Join(tabs, "   ")
	if lipgloss.Width(full) <= m.width {
		return components.SingleLine(full, m.width)
	}
	compact := " " + name + " — " + panelSwitcher(m.active)
	if lipgloss.Width(compact) <= m.width {
		return components.SingleLine(compact, m.width)
	}
	entry := screens.Screens[m.active]
	current := accent.Render(entry.Key + " " + entry.Title)
	compact = " " + name + " — " + current
	if lipgloss.Width(compact) > m.width {
		compact = " " + current
	}
	return components.SingleLine(compact, m.width)
}

// panelSwitcher is the compact header legend: "4 Analyse ‹ 5 Kaart › 6 Profiel".
// Titles and number keys come from the screen registry, never hardcoded labels.
func panelSwitcher(active int) string {
	if active < 0 || active >= len(screens.Screens) {
		return ""
	}
	total := len(screens.Screens)
	entry := func(i int) string {
		s := screens.Screens[(i+total)%total]
		return s.Key + " " + s.Title
	}
	muted := lipgloss.NewStyle().Foreground(ui.ColorMuted).Render
	cur := screens.Screens[active]
	return muted(entry(active-1)+" ‹ ") +
		lipgloss.NewStyle().Foreground(ui.ColorPrimary).Bold(true).
			Render(cur.Key+" "+cur.Title) +
		muted(" › "+entry(active+1))
}

func screenTitle(s screens.Screen) string {
	for _, info := range screens.Screens {
		if info.ID == s.ID() {
			return info.Title
		}
	}
	return fmt.Sprintf("%d", s.ID())
}

// Run starts the TUI. Called from cmd/waterketen-tui/main.go.
func Run(root string) error {
	p := tea.NewProgram(New(root), tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err := p.Run()
	return err
}
