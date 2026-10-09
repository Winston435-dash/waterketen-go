package components

import (
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"

	"waterketen/internal/ui"
)

// TableModel wraps bubbles/table with the project styling fixed. Construct
// with NewTable; never style the table in a screen.
type TableModel struct{ Model table.Model }

// Column is one table column.
type Column struct {
	Title string
	Width int
}

// NewTable builds a DataTable. height is the number of visible rows.
func NewTable(cols []Column, height int) TableModel {
	cs := make([]table.Column, len(cols))
	for i, c := range cols {
		cs[i] = table.Column{Title: c.Title, Width: c.Width}
	}
	t := table.New(
		table.WithColumns(cs),
		table.WithHeight(height),
		table.WithFocused(true),
	)
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(ui.ColorBorder).
		Foreground(ui.ColorMuted).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(ui.ColorPrimary).
		Bold(true)
	s.Cell = s.Cell.Foreground(ui.ColorText)
	t.SetStyles(s)
	return TableModel{Model: t}
}

// Rows replaces all rows. Each row slice MUST match the column count.
func (t *TableModel) Rows(rows ...[]string) {
	rs := make([]table.Row, len(rows))
	for i, r := range rows {
		rs[i] = table.Row(r)
	}
	t.Model.SetRows(rs)
}

// Selected returns the highlighted row, or nil when empty.
func (t *TableModel) Selected() []string {
	r := t.Model.SelectedRow()
	if r == nil {
		return nil
	}
	return []string(r)
}

// View renders the table as a string for embedding in a Panel.
func (t TableModel) View() string { return t.Model.View() }

// EmptyState is the single pattern for "there is nothing here": a muted glyph,
// a bold one-line headline, then the action the user should take (the fetch
// hint, typically). It exists so no screen invents its own sad-empty-box.
func EmptyState(headline, hint string, width int) string {
	body := lipgloss.JoinVertical(lipgloss.Left,
		Muted("∅"),
		Strong(headline),
		"",
		Muted(hint),
	)
	return lipgloss.Place(width, 5, lipgloss.Left, lipgloss.Top, body)
}

var _ = strings.TrimSpace // keep strings import honest for future table work
