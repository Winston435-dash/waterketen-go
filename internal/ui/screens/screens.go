// Package screens holds the canonical app panels: Bronnen, Lagen, Gebieden,
// Analyse, Kaart, Profiel. Each screen renders via ui/components only —
// it never draws its own borders, padding or colors, and it renders through
// the fixed anatomy Header → Content → CommandBar via the App in app.go.
package screens

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"waterketen/internal/core"
	"waterketen/internal/ui"
	"waterketen/internal/ui/components"
)

// ScreenID identifies one of the six canonical panels.
type ScreenID int

const (
	ScreenBronnen ScreenID = iota
	ScreenLagen
	ScreenGebieden
	ScreenAnalyse
	ScreenKaart
	ScreenProfiel
)

// ScreenInfo is the stable identity of one screen in the sidebar.
type ScreenInfo struct {
	ID    ScreenID
	Key   string // single keystroke ("1".."6")
	Title string
}

// Screens is the canonical registry, in sidebar order.
var Screens = []ScreenInfo{
	{ScreenBronnen, "1", "Bronnen"},
	{ScreenLagen, "2", "Lagen"},
	{ScreenGebieden, "3", "Gebieden"},
	{ScreenAnalyse, "4", "Analyse"},
	{ScreenKaart, "5", "Kaart"},
	{ScreenProfiel, "6", "Profiel"},
}

// Screen is the contract every panel satisfies. `Update` returns the new
// state and an optional command; the app-level loop passes events down.
type Screen interface {
	ID() ScreenID
	KeyMap() []key.Binding // what the CommandBar shows
	SetSize(w, h int)      // OUTER panel size incl. border
	Update(msg tea.Msg) (Screen, tea.Cmd)
	View() string // Panel-ready content, no border
	Info() string // sidebar footer info ("3 bedraad, 2 gedetecteerd")
}

// New builds the initial screen set, given the loaded domain state.
func New(root string) []Screen {
	assets, _ := core.LoadAssets(root)
	lib, _ := core.OpenLibrary(root)
	return []Screen{
		NewBronnen(),
		NewLagen(assets),
		NewGebieden(lib, assets),
		NewAnalyse(root, assets),
		NewKaart(root, lib, assets),
		NewProfiel(root, assets),
	}
}

// --- Sidebar ----------------------------------------------------------------

// Sidebar renders the panel switcher. It IS the legend for panel keys, so
// the CommandBar never repeats them.
func Sidebar(active ScreenID, infos []ScreenInfo, infoText string, height int) string {
	var b bytes.Buffer
	for _, s := range infos {
		label := fmt.Sprintf("[%s] %s", s.Key, s.Title)
		if s.ID == active {
			// Active panel is marked by ink + bold, not a filled bar, so the
			// terminal background stays visible.
			b.WriteString(lipgloss.NewStyle().
				Foreground(ui.ColorPrimary).
				Bold(true).
				Width(ui.SidebarWidth - 2).
				Render(label))
		} else {
			b.WriteString(lipgloss.NewStyle().
				Foreground(ui.ColorMuted).
				Width(ui.SidebarWidth - 2).
				Render(label))
		}
		b.WriteByte('\n')
	}
	body := b.String()
	if infoText != "" {
		body += "\n" + components.Muted(infoText)
	}
	return components.Panel("Panelen", body, ui.SidebarWidth, height, true)
}

// --- Bronnen ---------------------------------------------------------------

// Bronnen is panel 1: live status of every source.
type Bronnen struct {
	health  []core.SourceHealth
	probing bool
	// Incremental probe state: `probing` means a run is in flight; `done`
	// counts the wired results that have come back so far of `want`.
	done    int
	want    int
	spinner string
	prog    progress.Model // Bubbles' spring-animated bar for the probe run
	w, h    int
}

func NewBronnen() *Bronnen {
	return &Bronnen{prog: components.NewProgress(40)}
}

func (s *Bronnen) ID() ScreenID { return ScreenBronnen }
func (s *Bronnen) KeyMap() []key.Binding {
	return []key.Binding{ui.Keys.Probe, ui.Keys.NextPanel, ui.Keys.Quit}
}
func (s *Bronnen) SetSize(w, h int) {
	s.w, s.h = w, h
	if s.probing {
		s.prog.Width = s.barWidth()
	}
}
func (s *Bronnen) Info() string {
	if len(s.health) == 0 {
		return "nog niet geprobeerd"
	}
	up, down, unwired := 0, 0, 0
	for _, h := range s.health {
		switch h.Status {
		case core.HealthUp:
			up++
		case core.HealthDown:
			down++
		default:
			unwired++
		}
	}
	return fmt.Sprintf("%d up, %d down, %d gedetecteerd", up, down, unwired)
}

// HealthPartialMsg carries ONE wired source's probe result. The app forwards
// it to Bronnen, which counts it and — once all wired sources are in —
// assembles the final list. Streaming per source is what lets the panel draw
// real progress instead of a single "bezig…" string.
type HealthPartialMsg core.SourceHealth

// ProbeSourceCmd probes one wired source off the event loop.
func ProbeSourceCmd(src core.Source) tea.Cmd {
	return func() tea.Msg { return HealthPartialMsg(core.ProbeSource(src)) }
}

// ProbeCmd probes every wired source concurrently, one message per source, so
// the panel can show real progress as each answer arrives.
func ProbeCmd() tea.Cmd {
	cmds := make([]tea.Cmd, 0, len(core.WiredSources))
	for _, src := range core.WiredSources {
		cmds = append(cmds, ProbeSourceCmd(src))
	}
	return tea.Batch(cmds...)
}

// TickMsg is Bubbles' progress-bar animation frame. It is aliased here so the
// app can route it to the owning screen without importing the progress package
// itself.
type TickMsg = progress.FrameMsg

// barWidth is the cell width of the probe bar: capped so it never runs the
// whole panel, floored so it stays readable in a narrow terminal.
func (s *Bronnen) barWidth() int {
	w := s.w - 6
	if w > 40 {
		w = 40
	}
	if w < 20 {
		w = 20
	}
	return w
}

// BeginProbe starts a run: clears the count and marks every wired source as
// still pending. It returns the progress bar's first frame command — Bubbles'
// progress bar animates by returning FrameMsgs that must reach the runtime, so
// the app batches this in.
func (s *Bronnen) BeginProbe() tea.Cmd {
	s.health = nil
	s.probing = true
	s.done = 0
	s.want = len(core.WiredSources)
	s.prog.Width = s.barWidth()
	return s.prog.SetPercent(0)
}

// AddProbe records one finished probe and reports whether the run is complete.
// While the run goes on it retargets the bar (returning the frame command that
// makes it glide to the new value on its spring) and reports done=false.
func (s *Bronnen) AddProbe(h core.SourceHealth) (done bool, cmd tea.Cmd) {
	s.health = append(s.health, h)
	if h.Status != core.HealthNotWired {
		s.done++
	}
	if s.want > 0 && s.done >= s.want {
		return true, nil
	}
	return false, s.prog.SetPercent(float64(s.done) / float64(max(s.want, 1)))
}

// Tick feeds an animation message to the panel's own progress bar. The app
// routes every message here; the bar only replies while a run is in flight, so
// once the probe is done the bar's frame chain dies and the app goes idle.
func (s *Bronnen) Tick(msg tea.Msg) tea.Cmd {
	if !s.probing {
		return nil
	}
	m, cmd := s.prog.Update(msg)
	if pm, ok := m.(progress.Model); ok {
		s.prog = pm
	}
	return cmd
}

// FinishProbe appends the display-only (unwired) sources and ends the run.
func (s *Bronnen) FinishProbe() {
	s.health = core.AssembleHealth(s.health)
	s.probing = false
}

// SetHealth stores a full probe result (used for a pre-baked list). The wired
// records are shown up-first, matching what a live run settles into.
func (s *Bronnen) SetHealth(h []core.SourceHealth) {
	wired := make([]core.SourceHealth, 0, len(h))
	for _, r := range h {
		if r.Status != core.HealthNotWired {
			wired = append(wired, r)
		}
	}
	sort.SliceStable(wired, func(i, j int) bool {
		return wired[i].Status == core.HealthUp && wired[j].Status != core.HealthUp
	})
	s.health = core.AssembleHealth(wired)
	s.probing = false
}

// Busy reports bounded work in flight: a probe run is counting up.
func (s *Bronnen) Busy() bool { return s.probing }

// SetSpinner receives the app's current busy-spinner frame. Screens do not spin
// themselves; the app broadcasts one frame to every screen that shows it.
func (s *Bronnen) SetSpinner(frame string) { s.spinner = frame }

func (s *Bronnen) Update(msg tea.Msg) (Screen, tea.Cmd) { return s, nil }

func (s *Bronnen) View() string {
	if s.probing {
		// Bubbles' own animated bar: it glides to each new value on a spring.
		// The counter beside it is the real thing — n sources answered of the
		// real total — so the animation is decoration over honest progress.
		return components.Busy(s.spinner, "Bronnen worden geprobeerd…") + "\n\n" +
			"   " + s.prog.View() + " " + components.ProgressCount(s.done, s.want)
	}
	if len(s.health) == 0 {
		return components.EmptyState(
			"Nog geen probe-resultaten",
			fmt.Sprintf("Druk %s om de bronnen live te verversen.", "v"),
			s.w,
		)
	}
	var b strings.Builder
	for _, h := range s.health {
		if h.Status == core.HealthNotWired {
			continue // unwired section below
		}
		status := components.StatusOK
		if h.Status != core.HealthUp {
			status = components.StatusDown
		}
		count := "—"
		if h.Count >= 0 {
			count = thousands(h.Count)
		}
		line := fmt.Sprintf("%-22s %11s  %5d ms", h.Label, count, h.LatencyMS)
		b.WriteString(components.Badge(status, line))
		b.WriteByte('\n')
		// Licence belongs next to the source it applies to: CC0, CC-BY-SA and
		// "unknown" are different permissions, not decoration.
		b.WriteString(components.Muted("   licentie: " + licenseOrUnknown(h.License)))
		b.WriteByte('\n')
		if h.Error != "" {
			b.WriteString("   ")
			b.WriteString(components.Error(h.Error))
			b.WriteByte('\n')
		}
	}
	unwired := []core.SourceHealth{}
	for _, h := range s.health {
		if h.Status == core.HealthNotWired {
			unwired = append(unwired, h)
		}
	}
	if len(unwired) > 0 {
		b.WriteByte('\n')
		b.WriteString(components.Strong("Gedetecteerd, niet bedraad:"))
		b.WriteByte('\n')
		for _, h := range unwired {
			b.WriteString(components.Muted("  " + h.Label + " — " + h.Note +
				"  ·  " + licenseOrUnknown(h.License)))
			b.WriteByte('\n')
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

// licenseOrUnknown renders an absent licence as unknown instead of an empty
// gap: a source whose terms are not published must say so.
func licenseOrUnknown(license string) string {
	if license == "" {
		return "onbekend"
	}
	return license
}

func thousands(n int) string {
	s := fmt.Sprintf("%d", n)
	var out []byte
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, '.')
		}
		out = append(out, byte(c))
	}
	return string(out)
}

var _ = help.New // silence unused-import during staged development

// --- placeholder screens: implemented one file each -------------------------
