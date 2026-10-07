package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"waterketen/internal/core"
	"waterketen/internal/ui"
	"waterketen/internal/ui/components"
)

// Lagen is panel 2: the registered layers plus what is actually on disk.
type Lagen struct {
	assets []core.AssetMeta
	w, h   int
}

func NewLagen(assets []core.AssetMeta) *Lagen { return &Lagen{assets: assets} }

func (s *Lagen) ID() ScreenID { return ScreenLagen }
func (s *Lagen) KeyMap() []key.Binding {
	return []key.Binding{ui.Keys.NextPanel, ui.Keys.Quit}
}
func (s *Lagen) SetSize(w, h int) { s.w, s.h = w, h }
func (s *Lagen) Info() string     { return fmt.Sprintf("%d lagen", len(core.LayerSpecs)) }

func (s *Lagen) Update(msg tea.Msg) (Screen, tea.Cmd) { return s, nil }

func (s *Lagen) View() string {
	managed := map[string][]string{}
	for _, a := range s.assets {
		managed[a.Layer] = append(managed[a.Layer],
			fmt.Sprintf("%s: %s", a.Area, thousands(a.FeatureCount)))
	}
	var b strings.Builder
	for _, spec := range core.LayerSpecs {
		status := components.StatusIdle
		if files := managed[spec.Slug]; len(files) > 0 {
			status = components.StatusOK
		}
		b.WriteString(components.Badge(status, spec.Label))
		b.WriteByte('\n')
		b.WriteString(components.Muted("   " + spec.Geometry + " · " + spec.SourceID))
		b.WriteByte('\n')
		if files := managed[spec.Slug]; len(files) > 0 {
			for _, f := range files {
				b.WriteString(components.Muted("   → " + f))
				b.WriteByte('\n')
			}
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

// Gebieden is panel 3: registered CBS areas from the library, else managed assets.
type Gebieden struct {
	lib    *core.Library
	assets []core.AssetMeta
	w, h   int
}

func NewGebieden(lib *core.Library, assets []core.AssetMeta) *Gebieden {
	return &Gebieden{lib: lib, assets: assets}
}

func (s *Gebieden) ID() ScreenID { return ScreenGebieden }
func (s *Gebieden) KeyMap() []key.Binding {
	return []key.Binding{ui.Keys.NextPanel, ui.Keys.Quit}
}
func (s *Gebieden) SetSize(w, h int) { s.w, s.h = w, h }
func (s *Gebieden) Info() string {
	if s.lib != nil {
		return fmt.Sprintf("%d gebieden (snapshot %s)", s.lib.Manifest.AreaCount, s.lib.Snapshot)
	}
	seen := map[string]struct{}{}
	for _, a := range s.assets {
		seen[a.Area] = struct{}{}
	}
	return fmt.Sprintf("%d gebieden", len(seen))
}

func (s *Gebieden) Update(msg tea.Msg) (Screen, tea.Cmd) { return s, nil }

func (s *Gebieden) View() string {
	if s.lib == nil {
		seen := map[string]struct{}{}
		for _, a := range s.assets {
			seen[a.Area] = struct{}{}
		}
		if len(seen) == 0 {
			return components.EmptyState(
				"Geen managed data",
				"Haal data op: waterketen fetch --gebied delfshaven --laag riool",
				s.w,
			)
		}
		var b strings.Builder
		for a := range seen {
			b.WriteString(components.Strong(a))
			b.WriteByte('\n')
		}
		return strings.TrimRight(b.String(), "\n")
	}
	var b strings.Builder
	b.WriteString(components.Strong("Snapshot " + s.lib.Snapshot))
	b.WriteByte('\n')
	b.WriteString(components.Muted(fmt.Sprintf("%d gebieden · CRS %s",
		s.lib.Manifest.AreaCount, s.lib.Manifest.CRS)))
	b.WriteByte('\n')
	b.WriteByte('\n')
	names := make([]string, 0, len(s.lib.Manifest.Layers))
	for n := range s.lib.Manifest.Layers {
		names = append(names, n)
	}
	for _, n := range names {
		rec := s.lib.Manifest.Layers[n]
		b.WriteString(components.Muted(fmt.Sprintf("  %-22s %s features", n, thousands(rec.FeatureCount))))
		b.WriteByte('\n')
	}
	return strings.TrimRight(b.String(), "\n")
}
