// Package app is the top-level Bubble Tea shell. It owns the fixed anatomy
// every screen renders inside:
//
//	Header   — app title, screen title, current screen key hints
//	Content  — sidebar + the active screen's Panel
//	CommandBar — bottom line of screen-scoped keys
//
// Screens never draw their own borders; they get `SetSize` with the OUTER
// panel box and return body content from `View` which the App wraps.
package app

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"waterketen/internal/core"
	"waterketen/internal/ui"
	"waterketen/internal/ui/components"
	"waterketen/internal/ui/screens"
)

// Model is the one Bubble Tea model for the whole app.
type Model struct {
	root     string
	screens  []screens.Screen
	active   int
	width    int
	height   int
	probing  bool
	drawing  bool
	drawGen  int
	data     []*screens.LayerData
	cmdBar   components.CommandBar
	ready    bool
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
	}
}

// Init kicks off the live probe and the data load — both off the event loop.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		screens.ProbeCmd(),
		loadAllDataCmd(m.root),
	)
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
	case screens.HealthMsg:
		m.probing = false
		for _, s := range m.screens {
			if b, ok := s.(*screens.Bronnen); ok {
				b.SetProbing(false)
				b.SetHealth(msg)
			}
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
		// The Kaart needs its Scene built now that data exists.
		if k := m.kaart(); k != nil {
			return m, k.BuildSceneCmd()
		}
		return m, nil
	case screens.SceneMsg:
		if k := m.kaart(); k != nil {
			if msg.Err != nil {
				k.SetError(msg.Err)
			} else {
				k.SetScene(msg.Scene)
			}
			m.drawing = false
			m.drawGen++
			gen := m.drawGen
			m.drawing = true
			return m, tea.Tick(0, func(time.Time) tea.Msg { return drawFrameMsg{gen: gen} })
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
				return m, tea.Tick(0, func(time.Time) tea.Msg { return m.requestDraw() })
			}
		}
		return m, nil
	case tea.KeyMsg:
		// Global keys first — they work regardless of active screen.
		switch {
		case key.Matches(msg, ui.Keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, ui.Keys.Probe):
			if !m.probing {
				m.probing = true
				for _, s := range m.screens {
					if b, ok := s.(*screens.Bronnen); ok {
						b.SetProbing(true)
					}
				}
				return m, screens.ProbeCmd()
			}
			return m, nil
		case key.Matches(msg, ui.Keys.Redraw):
			if k := m.kaart(); k != nil && m.screens[m.active].ID() == screens.ScreenKaart {
				k.SetScene(kCurrentScene(k))
				m.drawGen++
				gen := m.drawGen
				return m, tea.Tick(0, func(time.Time) tea.Msg { return drawFrameMsg{gen: gen} })
			}
			return m, nil
		}
		// j/k move between panels here, not inside screens — the screens never
		// both scroll AND switch. Letters stay free for content keys otherwise.
		if key.Matches(msg, ui.Keys.NextPanel) {
			m.active = (m.active + 1) % len(m.screens)
			return m, m.maybeTriggerDraw()
		}
		if key.Matches(msg, ui.Keys.PrevPanel) {
			m.active = (m.active - 1 + len(m.screens)) % len(m.screens)
			return m, m.maybeTriggerDraw()
		}
		// Number keys 1-6 select the panel directly.
		for i, info := range screens.Screens {
			if msg.String() == info.Key {
				m.active = i
				return m, m.maybeTriggerDraw()
			}
		}
		// Forward to the active screen.
		next, cmd := m.screens[m.active].Update(msg)
		m.screens[m.active] = next
		if m.screens[m.active].ID() == screens.ScreenKaart {
			if k := m.kaart(); k != nil && k.NeedsDraw() && !m.drawing {
				m.drawing = true
				m.drawGen++
				gen := m.drawGen
				return m, tea.Tick(0, func(time.Time) tea.Msg { return drawFrameMsg{gen: gen} })
			}
		}
		return m, cmd
	case tea.MouseMsg:
		next, cmd := m.screens[m.active].Update(msg)
		m.screens[m.active] = next
		if m.screens[m.active].ID() == screens.ScreenKaart {
			if k := m.kaart(); k != nil && k.NeedsDraw() && !m.drawing {
				m.drawing = true
				m.drawGen++
				gen := m.drawGen
				return m, tea.Tick(0, func(time.Time) tea.Msg { return drawFrameMsg{gen: gen} })
			}
		}
		return m, cmd
	}
	return m, nil
}

type drawFrameMsg struct{ gen int }

func (m Model) requestDraw() tea.Msg { return drawFrameMsg{gen: m.drawGen} }

func (m *Model) kaart() *screens.Kaart {
	for _, s := range m.screens {
		if k, ok := s.(*screens.Kaart); ok {
			return k
		}
	}
	return nil
}

func kCurrentScene(k *screens.Kaart) *core.Scene {
	// redraw after a manual 'r' — keep the last scene, just mark dirty.
	return nil // signal: re-running BuildSceneCmd is wasted; SetScene(nil→nil)
	// is a no-op but the drawGen bump above retriggers a draw of the existing
	// view. The Kaart already owns drawing on NeedsDraw.
}

func (m Model) maybeTriggerDraw() tea.Cmd {
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

// panelBox computes the content panel's OUTER size (incl. border).
func (m Model) panelBox() (int, int) {
	bodyH := m.height - 2 // header + commandbar rows
	if bodyH < 3 {
		bodyH = 3
	}
	bodyW := m.width - ui.SidebarWidth
	if bodyW < 20 {
		bodyW = m.width // narrow terminal: sidebar collapses
	}
	return bodyW, bodyH
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

func (m Model) renderHeader() string {
	active := screens.Screens[m.active]
	title := lipgloss.NewStyle().
		Foreground(ui.ColorPrimary).Bold(true).
		Render("waterketen")
	scr := lipgloss.NewStyle().
		Foreground(ui.ColorText).
		Render(" — " + active.Title)
	return lipgloss.NewStyle().
		Width(m.width).
		Background(ui.ColorSurfaceHi).
		Padding(0, ui.SpacingXS).
		Render(title + scr)
}

func (m Model) renderBody() string {
	screen := m.screens[m.active]
	w, h := m.panelBox()
	narrow := m.width < 80
	var mainPanel string
	if narrow {
		mainPanel = components.Panel(screenTitle(screen), screen.View(), m.width, h, true)
		return mainPanel
	}
	info := screen.Info()
	side := screens.Sidebar(screen.ID(), screens.Screens, info, h)
	mainPanel = components.Panel(screenTitle(screen), screen.View(), w, h, true)
	return lipgloss.JoinHorizontal(lipgloss.Top, side, mainPanel)
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
