package components

import (
	"fmt"
	"io"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"waterketen/internal/ui"
)

// InventoryItem keeps list ink in the kit and domain text in its screen.
type InventoryItem struct {
	ID, Label, Summary, Detail   string
	CompactLabel, CompactSummary string
	Status                       Status
}

func (i InventoryItem) FilterValue() string { return i.Label + " " + i.Summary }

type inventoryDelegate struct{ compact bool }

func (d inventoryDelegate) Height() int {
	if d.compact {
		return 2
	}
	return 1
}
func (d inventoryDelegate) Spacing() int                        { return 0 }
func (d inventoryDelegate) Update(tea.Msg, *list.Model) tea.Cmd { return nil }
func (d inventoryDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(InventoryItem)
	if !ok {
		return
	}
	mark := "  "
	label := i.Label
	if d.compact && i.CompactLabel != "" {
		label = i.CompactLabel
	}
	if index == m.Index() {
		mark = "› "
		label = Selected(label, lipgloss.Width(label))
	}
	if d.compact {
		summary := i.Summary
		if i.CompactSummary != "" {
			summary = i.CompactSummary
		}
		fmt.Fprint(w, SingleLine(mark+Badge(i.Status, label), m.Width())+"\n"+SingleLine("    "+Muted(summary), m.Width()))
	} else {
		labelW := max(1, m.Width()/3)
		fmt.Fprint(w, SingleLine(mark+Badge(i.Status, SingleLine(label, labelW))+"  "+Muted(i.Summary), m.Width()))
	}
}

// Inventory wraps Bubbles list selection/pagination with one responsive delegate.
// It has no private spinner or filter: the app owns animation and screen filters.
type Inventory struct {
	model         list.Model
	width, height int
}

func NewInventory() Inventory {
	m := list.New(nil, inventoryDelegate{}, 1, 1)
	m.SetFilteringEnabled(false)
	m.SetShowTitle(false)
	m.SetShowFilter(false)
	m.SetShowStatusBar(false)
	m.SetShowHelp(false)
	m.SetShowPagination(false)
	m.DisableQuitKeybindings()
	m.KeyMap.CursorUp = ui.Keys.Up
	m.KeyMap.CursorDown = ui.Keys.Down
	m.KeyMap.NextPage = ui.Keys.PageDown
	m.KeyMap.PrevPage = ui.Keys.PageUp
	m.KeyMap.GoToStart = ui.Keys.Home
	m.KeyMap.GoToEnd = ui.Keys.End
	return Inventory{model: m}
}
func (i *Inventory) SetSize(width, height int) {
	i.width, i.height = max(1, width), max(1, height)
	i.model.SetDelegate(inventoryDelegate{compact: width < 72 && height >= 2})
	i.model.SetSize(i.width, i.height)
}
func (i *Inventory) SetItems(items []InventoryItem) {
	old, had := i.Selected()
	values := make([]list.Item, len(items))
	selected := 0
	for n, item := range items {
		values[n] = item
		if had && item.ID == old.ID {
			selected = n
		}
	}
	i.model.SetItems(values)
	i.model.Select(selected)
}
func (i *Inventory) Selected() (InventoryItem, bool) {
	item, ok := i.model.SelectedItem().(InventoryItem)
	return item, ok
}
func (i *Inventory) Update(msg tea.Msg) tea.Cmd {
	if wheel, ok := msg.(tea.MouseWheelMsg); ok {
		switch wheel.Button {
		case tea.MouseWheelUp:
			i.model.CursorUp()
		case tea.MouseWheelDown:
			i.model.CursorDown()
		}
		return nil
	}
	if msg, ok := msg.(tea.KeyPressMsg); ok {
		if key.Matches(msg, ui.Keys.Home) {
			i.model.Select(0)
			return nil
		}
		if key.Matches(msg, ui.Keys.End) {
			i.model.Select(max(0, len(i.model.Items())-1))
			return nil
		}
	}
	var cmd tea.Cmd
	i.model, cmd = i.model.Update(msg)
	return cmd
}
func (i Inventory) View() string {
	if len(i.model.Items()) == 0 {
		return SingleLine(Muted("Geen resultaten."), i.width)
	}
	rows := strings.Split(i.model.View(), "\n")
	if len(rows) > i.height {
		rows = rows[:i.height]
	}
	for n, row := range rows {
		rows[n] = SingleLine(row, i.width)
	}
	for len(rows) < i.height {
		rows = append(rows, strings.Repeat(" ", i.width))
	}
	return strings.Join(rows, "\n")
}
func (i Inventory) Position() string {
	return fmt.Sprintf("%d/%d · enter: details", min(i.model.Index()+1, len(i.model.Items())), len(i.model.Items()))
}

// SearchInput supplies a consistent visible editing cursor without an idle blink chain.
func SearchInput(prompt string) textinput.Model {
	m := textinput.New()
	m.Prompt = prompt
	m.CharLimit = 512
	m.SetVirtualCursor(true)
	styles := m.Styles()
	styles.Focused.Text = lipgloss.NewStyle().Foreground(ui.ColorText)
	styles.Focused.Prompt = lipgloss.NewStyle().Foreground(ui.ColorPrimary).Bold(true)
	styles.Focused.Placeholder = lipgloss.NewStyle().Foreground(ui.ColorMuted)
	styles.Blurred = styles.Focused
	styles.Cursor.Color = ui.ColorPrimary
	styles.Cursor.Blink = false
	m.SetStyles(styles)
	m.Focus()
	return m
}
