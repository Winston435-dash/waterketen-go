package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"waterketen/internal/core"
	"waterketen/internal/ui"
)

// CatalogDocument wraps before colouring, including long endpoint URLs, so the
// reading viewport never silently clips explanatory text on a phone.
func CatalogDocument(lines []core.CatalogLine, width int) string {
	var rows []string
	for _, line := range lines {
		text := ansi.Wrap(line.Text, max(1, width), "")
		switch line.Kind {
		case "heading":
			rows = append(rows, "", Strong(text))
		case "warning":
			rows = append(rows, lipgloss.NewStyle().Foreground(ui.ColorWarning).Render(text))
		case "meta":
			rows = append(rows, Muted(text))
		default:
			rows = append(rows, text)
		}
	}
	return strings.TrimLeft(strings.Join(rows, "\n"), "\n")
}

// CatalogMenu returns the document and each entry's first row for selection-aware
// scrolling. Full titles wrap instead of hiding the question behind an ellipsis.
func CatalogMenu(entries []core.CatalogEntry, selected, width int) (string, []int) {
	var rows []string
	var starts []int
	for i, entry := range entries {
		starts = append(starts, len(rows))
		mark := "  "
		if i == selected {
			mark = "› "
		}
		wrapped := strings.Split(ansi.Wrap(mark+entry.Title, max(1, width), ""), "\n")
		for _, row := range wrapped {
			if i == selected {
				row = Selected(row, width)
			}
			rows = append(rows, row)
		}
		rows = append(rows, "")
	}
	return strings.TrimRight(strings.Join(rows, "\n"), "\n"), starts
}

func CatalogTabs(section core.CatalogSection, width int) string {
	if width < 64 {
		return SingleLine(Strong(fmt.Sprintf("%s (%d/%d)", core.CatalogSections[section], section+1, len(core.CatalogSections))), width)
	}
	var labels []string
	for i, label := range core.CatalogSections {
		if core.CatalogSection(i) == section {
			label = Selected(label, len(label))
		} else {
			label = Muted(label)
		}
		labels = append(labels, label)
	}
	return SingleLine(strings.Join(labels, " · "), width)
}

func CatalogColumns(menu, document string, menuWidth int) string {
	return lipgloss.JoinHorizontal(lipgloss.Top, lipgloss.NewStyle().Width(menuWidth).Render(menu), Muted(" │ "), document)
}
