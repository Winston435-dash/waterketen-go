// Package screens holds the canonical app panels: Bronnen, Lagen, Gebieden,
// Analyse, Kaart, Profiel. Each screen renders via ui/components only —
// it never draws its own borders, padding or colors, and it renders through
// the fixed anatomy Header → Content → CommandBar via the App in app.go.
package screens

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
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
	KeyMap() []key.Binding       // what the CommandBar shows
	SetSize(w, h int)            // OUTER panel size incl. border
	Update(msg tea.Msg) (Screen, tea.Cmd)
	View() string                // Panel-ready content, no border
	Info() string                // sidebar footer info ("3 bedraad, 2 gedetecteerd")
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
			b.WriteString(lipgloss.NewStyle().
				Foreground(ui.ColorText).
				Background(ui.ColorActiveBg).
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
	health []core.SourceHealth
	probing bool
	w, h   int
}

func NewBronnen() *Bronnen { return &Bronnen{} }

func (s *Bronnen) ID() ScreenID { return ScreenBronnen }
func (s *Bronnen) KeyMap() []key.Binding {
	return []key.Binding{ui.Keys.Probe, ui.Keys.NextPanel, ui.Keys.Quit}
}
func (s *Bronnen) SetSize(w, h int) { s.w, s.h = w, h }
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

// HealthMsg carries probe results back to the app.
type HealthMsg []core.SourceHealth

// ProbeCmd runs the probes off the event loop and posts the results.
func ProbeCmd() tea.Cmd {
	return func() tea.Msg { return HealthMsg(core.ProbeAll()) }
}

// SetHealth stores the last probe result.
func (s *Bronnen) SetHealth(h []core.SourceHealth) { s.health = h }

// SetProbing toggles the "probing now" indicator.
func (s *Bronnen) SetProbing(v bool) { s.probing = v }

func (s *Bronnen) Update(msg tea.Msg) (Screen, tea.Cmd) { return s, nil }

func (s *Bronnen) View() string {
	if s.probing {
		return components.Muted("Bronnen worden geprobeerd…")
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
			b.WriteString(components.Muted("  " + h.Label + " — " + h.Note))
			b.WriteByte('\n')
		}
	}
	return strings.TrimRight(b.String(), "\n")
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
