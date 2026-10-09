package screens

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"waterketen/internal/core"
	"waterketen/internal/ui"
	"waterketen/internal/ui/components"
)

// Ontdek explains existing registrations without probing or downloading data.
// On a phone Enter opens the reader; on desktop the same answer is previewed
// alongside the list. Reading mode gives the arrows to the document viewport.
type Ontdek struct {
	catalog                  core.Catalog
	loadErr                  error
	section                  core.CatalogSection
	selected                 int
	reading, technical, wide bool
	width                    int
	menu, document           viewport.Model
}

func NewOntdek(root string) *Ontdek {
	catalog, err := core.LoadCatalog(root)
	return &Ontdek{catalog: catalog, loadErr: err, menu: viewport.New(1, 1), document: viewport.New(1, 1)}
}
func (s *Ontdek) ID() ScreenID { return ScreenOntdek }
func (s *Ontdek) Info() string {
	mode := "kies een vraag"
	if s.reading {
		mode = "lezen"
	}
	if s.technical {
		mode += " · technisch"
	}
	return mode
}
func (s *Ontdek) KeyMap() []key.Binding {
	keys := []key.Binding{ui.Keys.Up, ui.Keys.Down, ui.Keys.Enter, ui.Keys.NextTopic, ui.Keys.PrevTopic}
	if s.reading {
		keys = []key.Binding{ui.Keys.Up, ui.Keys.Down, ui.Keys.PageUp, ui.Keys.PageDown, ui.Keys.Home, ui.Keys.End, ui.Keys.Technical, ui.Keys.Back}
	}
	return append(keys, ui.Keys.NextPanel, ui.Keys.Quit)
}
func (s *Ontdek) SetSize(w, h int) {
	s.width = max(1, w-4)
	s.wide = s.width >= 80
	menuWidth, docWidth := s.width, s.width
	if s.wide {
		menuWidth = min(38, s.width/3)
		docWidth = s.width - menuWidth - 3
	}
	s.menu.Width, s.document.Width = menuWidth, docWidth
	s.menu.Height, s.document.Height = max(1, h-4), max(1, h-4)
	s.refresh()
}

func (s *Ontdek) refresh() {
	entries := s.catalog.Entries(s.section)
	if len(entries) == 0 {
		return
	}
	s.selected = min(max(0, s.selected), len(entries)-1)
	menu, starts := components.CatalogMenu(entries, s.selected, s.menu.Width)
	s.menu.SetContent(menu)
	start := starts[s.selected]
	end := s.menu.TotalLineCount()
	if s.selected+1 < len(starts) {
		end = starts[s.selected+1] - 1
	}
	if start < s.menu.YOffset {
		s.menu.SetYOffset(start)
	} else if end > s.menu.YOffset+s.menu.Height {
		s.menu.SetYOffset(max(start, end-s.menu.Height))
	}
	lines, err := s.catalog.Explain(s.section, entries[s.selected].ID, s.technical)
	if err != nil {
		lines = []core.CatalogLine{{Text: err.Error(), Kind: "warning"}}
	}
	if s.loadErr != nil {
		lines = append([]core.CatalogLine{{Text: "Lokale metadata kon niet volledig worden gelezen: " + s.loadErr.Error(), Kind: "warning"}}, lines...)
	}
	s.document.SetContent(components.CatalogDocument(lines, s.document.Width))
}

func (s *Ontdek) Update(msg tea.Msg) (Screen, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(msg, ui.Keys.NextTopic), key.Matches(msg, ui.Keys.PrevTopic):
			delta := 1
			if key.Matches(msg, ui.Keys.PrevTopic) {
				delta = -1
			}
			s.section = core.CatalogSection((int(s.section) + delta + len(core.CatalogSections)) % len(core.CatalogSections))
			s.selected, s.reading = 0, false
			s.menu.GotoTop()
			s.document.GotoTop()
			s.refresh()
			return s, nil
		case key.Matches(msg, ui.Keys.Enter):
			s.reading = true
			return s, nil
		case key.Matches(msg, ui.Keys.Back):
			s.reading = false
			return s, nil
		case key.Matches(msg, ui.Keys.Technical):
			s.technical = !s.technical
			s.document.GotoTop()
			s.refresh()
			return s, nil
		}
		if !s.reading {
			switch {
			case key.Matches(msg, ui.Keys.Up):
				s.selected--
			case key.Matches(msg, ui.Keys.Down):
				s.selected++
			default:
				return s, nil
			}
			s.document.GotoTop()
			s.refresh()
			return s, nil
		}
	}
	if s.reading {
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch {
			case key.Matches(msg, ui.Keys.Home):
				s.document.GotoTop()
				return s, nil
			case key.Matches(msg, ui.Keys.End):
				s.document.GotoBottom()
				return s, nil
			}
		}
		var cmd tea.Cmd
		s.document, cmd = s.document.Update(msg)
		return s, cmd
	}
	return s, nil
}

func (s *Ontdek) View() string {
	body := s.menu.View()
	if s.wide {
		body = components.CatalogColumns(body, s.document.View(), s.menu.Width)
	} else if s.reading {
		body = s.document.View()
	}
	hint := "Enter: lees antwoord · [ / ]: categorie"
	if s.reading {
		hint = fmt.Sprintf("↑↓: lezen · t: techniek · esc: lijst · %.0f%%", s.document.ScrollPercent()*100)
	}
	return components.CatalogTabs(s.section, s.width) + "\n" + body + "\n" + components.SingleLine(components.Muted(hint), s.width)
}
