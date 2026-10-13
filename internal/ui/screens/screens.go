// Package screens holds the canonical app panels, including the question-driven
// Ontdek catalog. Each screen renders via ui/components only —
// it never draws its own borders, padding or colors, and it renders through
// the fixed anatomy Header → Content → CommandBar via the App in app.go.
package screens

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/progress"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"

	"waterketen/internal/core"
	"waterketen/internal/ui"
	"waterketen/internal/ui/components"
)

// ScreenID identifies one of the canonical panels.
type ScreenID int

const (
	ScreenBronnen ScreenID = iota
	ScreenLagen
	ScreenGebieden
	ScreenAnalyse
	ScreenKaart
	ScreenProfiel
	ScreenOntdek
	ScreenQuery
)

// ScreenInfo is the stable identity of one screen in the header navigation.
type ScreenInfo struct {
	ID    ScreenID
	Key   string // single keystroke from the registry
	Title string
}

// Screens is the canonical registry, in navigation order.
var Screens = []ScreenInfo{
	{ScreenBronnen, "1", "Bronnen"},
	{ScreenLagen, "2", "Lagen"},
	{ScreenGebieden, "3", "Gebieden"},
	{ScreenAnalyse, "4", "Analyse"},
	{ScreenKaart, "5", "Kaart"},
	{ScreenProfiel, "6", "Profiel"},
	{ScreenOntdek, "7", "Ontdek"},
	{ScreenQuery, "8", "Query"},
}

// Screen is the contract every panel satisfies. `Update` returns the new
// state and an optional command; the app-level loop passes events down.
type Screen interface {
	ID() ScreenID
	KeyMap() []key.Binding // what the CommandBar shows
	SetSize(w, h int)      // OUTER panel size incl. border
	Update(msg tea.Msg) (Screen, tea.Cmd)
	View() string // Panel-ready content, no border
	Info() string // panel-title context ("3 bedraad, 2 gedetecteerd")
}

// New builds the initial screen set, given the loaded domain state.
func New(root string) []Screen {
	assets, assetErr := core.LoadAssets(root)
	lib, libErr := core.OpenLibrary(root)
	catalog := core.CatalogWithLocal(assets, lib)
	metadataErr := errors.Join(assetErr, libErr)
	analysis := NewAnalyse(root, assets)
	analysis.SetLibrary(lib, catalog)
	areas := NewGebieden(lib, assets)
	areas.err = libErr
	return []Screen{
		NewBronnen(),
		NewLagen(catalog, metadataErr),
		areas,
		analysis,
		NewKaart(root, lib, assets),
		NewProfiel(root, assets),
		NewOntdekCatalog(catalog, metadataErr),
		NewQuery(root),
	}
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
	body    viewport.Model
	list    components.Inventory
	details bool
}

func NewBronnen() *Bronnen {
	return &Bronnen{prog: components.NewProgress(40), body: viewport.New(viewport.WithWidth(1), viewport.WithHeight(1)),list:components.NewInventory()}
}

func (s *Bronnen) ID() ScreenID { return ScreenBronnen }
func (s *Bronnen) KeyMap() []key.Binding {
	if s.details { return []key.Binding{ui.Keys.Back,ui.Keys.Up,ui.Keys.Down,ui.Keys.PageUp,ui.Keys.PageDown,ui.Keys.Home,ui.Keys.End} }
	return []key.Binding{ui.Keys.Probe,ui.Keys.Details,ui.Keys.Up, ui.Keys.Down, ui.Keys.PageUp, ui.Keys.PageDown, ui.Keys.Home, ui.Keys.End, ui.Keys.NextPanel, ui.Keys.Quit}
}
func (s *Bronnen) ShortKeyMap() []key.Binding { if s.details { return s.KeyMap() };return []key.Binding{ui.Keys.Probe,ui.Keys.Details,ui.Keys.Up,ui.Keys.Down} }
func (s *Bronnen) SetSize(w, h int) {
	s.w, s.h = w, h
	s.body.SetWidth(max(1, w-4))
	s.body.SetHeight(max(1, h-2))
	s.refreshHealth()
	if s.probing {
		s.prog.SetWidth(s.barWidth())
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
	return max(1,min(40,s.body.Width()-1-len(components.ProgressCount(s.want,s.want))))
}

// BeginProbe starts a run: clears the count and marks every wired source as
// still pending. It returns the progress bar's first frame command — Bubbles'
// progress bar animates by returning FrameMsgs that must reach the runtime, so
// the app batches this in.
func (s *Bronnen) BeginProbe() tea.Cmd {
	s.details=false
	s.health = nil
	s.probing = true
	s.done = 0
	s.want = len(core.WiredSources)
	s.prog.SetWidth(s.barWidth())
	s.refreshHealth()
	return s.prog.SetPercent(0)
}

// AddProbe records one finished probe and reports whether the run is complete.
// While the run goes on it retargets the bar (returning the frame command that
// makes it glide to the new value on its spring) and reports done=false.
func (s *Bronnen) AddProbe(h core.SourceHealth) (done bool, cmd tea.Cmd) {
	s.health = append(s.health, h)
	s.refreshHealth()
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
	s.prog = m
	return cmd
}

// FinishProbe appends the display-only (unwired) sources and ends the run.
func (s *Bronnen) FinishProbe() {
	s.health = core.AssembleHealth(s.health)
	s.probing = false
	s.refreshHealth()
	s.body.GotoTop()
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
	s.refreshHealth()
	s.body.GotoTop()
}

// Busy reports bounded work in flight: a probe run is counting up.
func (s *Bronnen) Busy() bool { return s.probing }

// SetSpinner receives the app's current busy-spinner frame. Screens do not spin
// themselves; the app broadcasts one frame to every screen that shows it.
func (s *Bronnen) SetSpinner(frame string) { s.spinner = frame }

func (s *Bronnen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	if s.details {
		if k,ok:=msg.(tea.KeyPressMsg);ok&&key.Matches(k,ui.Keys.Back) { s.details=false;return s,nil }
		return s,scrollInventory(&s.body,msg)
	}
	if k,ok:=msg.(tea.KeyPressMsg);ok&&key.Matches(k,ui.Keys.Enter) {
		if item,ok:=s.list.Selected();ok { s.details=true;s.body.SetContent(components.WrappedText(item.Detail,s.body.Width()));s.body.GotoTop() }
		return s,nil
	}
	return s,s.list.Update(msg)
}

func (s *Bronnen) View() string {
	if s.details { return s.body.View() }
	if s.probing {
		// Bubbles' own animated bar: it glides to each new value on a spring.
		// The counter beside it is the real thing — n sources answered of the
		// real total — so the animation is decoration over honest progress.
		return components.FitRows(components.Busy(s.spinner,"Bronnen worden geprobeerd…")+"\n"+s.prog.View()+" "+components.ProgressCount(s.done,s.want)+"\n"+s.list.View()+"\n"+components.Muted(s.list.Position()),s.body.Width(),s.body.Height())
	}
	if len(s.health) == 0 {
		return components.EmptyState(
			"Nog geen probe-resultaten",
			fmt.Sprintf("Druk %s om de bronnen live te verversen.", "v"),
			s.w,
		)
	}
	return components.FitRows(components.Muted("Groen = bereikbaar, niet lokaal beschikbaar")+"\n"+s.list.View()+"\n"+components.Muted(s.list.Position()),s.body.Width(),s.body.Height())
}

func (s *Bronnen) refreshHealth() {
	rows:=s.body.Height()-2;if s.probing { rows-- }
	s.list.SetSize(s.body.Width(),max(1,rows))
	var items []components.InventoryItem
	for _,h:=range s.health {
		state:="niet bedraad";status:=components.StatusIdle
		if h.Status==core.HealthUp { state="bereikbaar";status=components.StatusOK }
		if h.Status==core.HealthDown { state="onbereikbaar";status=components.StatusDown }
		license:=licenseOrUnknown(h.License)
		count:="—";if h.Count>=0 { count=Thousands(h.Count) }
		summary:=state+" · "+license
		if h.Status!=core.HealthNotWired { summary=fmt.Sprintf("%s · %s dienstobjecten · %d ms · %s",state,count,h.LatencyMS,license) }
		doc:=components.Strong(queryClean(h.Label))+"\nesc: bronnenlijst\n\nStatus: "+state+"\nLicentie: "+queryClean(license)+"\nDiensttelling: "+count+"\nProbe-aantallen kunnen landelijk zijn; dit zijn geen gebiedstellingen.\nBereikbaar ≠ gedownload of lokaal beschikbaar.\nBasis-URL (requests voegen parameters toe): "+queryClean(h.URL)
		if h.Error!="" { doc+="\n"+components.Error(queryClean(h.Error)) }
		if h.Note!="" { doc+="\n"+components.Muted(queryClean(h.Note)) }
		id:=h.SourceID;if id=="" { id=h.Label }
		items=append(items,components.InventoryItem{ID:id,Label:queryClean(h.Label),Summary:queryClean(summary),Detail:doc,Status:status})
	}
	s.list.SetItems(items)
	if s.details { if item,ok:=s.list.Selected();ok { s.body.SetContent(components.WrappedText(item.Detail,s.body.Width())) } }
}

func (s *Bronnen) renderHealth() string {
	var b strings.Builder
	b.WriteString(components.Muted("Groen = dienst bereikbaar, niet gedownload of lokaal beschikbaar."))
	b.WriteByte('\n')
	b.WriteString(components.Muted("Probe-aantallen kunnen landelijk zijn; dit zijn geen gebiedstellingen."))
	b.WriteByte('\n')
	b.WriteByte('\n')
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
			count = Thousands(h.Count)
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
			b.WriteString(components.Muted("  " + h.Label))
			b.WriteByte('\n')
			b.WriteString(components.Muted("   licentie: " + licenseOrUnknown(h.License)))
			b.WriteByte('\n')
			b.WriteString(components.Muted("   " + h.Note))
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

// Thousands renders an int with Dutch thousand separators: 87146 → "87.146".
func Thousands(n int) string {
	return ui.Thousands(n)
}

// --- placeholder screens: implemented one file each -------------------------
