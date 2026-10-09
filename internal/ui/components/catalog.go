package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/glamour"
	glamouransi "github.com/charmbracelet/glamour/ansi"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"waterketen/internal/core"
	"waterketen/internal/ui"
)

// CatalogDocument translates the catalog's semantic lines to Markdown, then
// renders it with Glamour at the viewport width. Styles use foreground ink only.
func CatalogDocument(lines []core.CatalogLine, width int) string {
	width = max(1, width)
	renderer, err := glamour.NewTermRenderer(
		glamour.WithStyles(catalogMarkdownStyles()),
		glamour.WithColorProfile(lipgloss.ColorProfile()),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return catalogPlainDocument(lines, width)
	}
	rendered, err := renderer.Render(catalogMarkdown(lines))
	if err != nil {
		return catalogPlainDocument(lines, width)
	}
	return strings.TrimSpace(rendered)
}

func catalogMarkdown(lines []core.CatalogLine) string {
	var markdown strings.Builder
	headings := 0
	for _, line := range lines {
		text := escapeMarkdownText(line.Text)
		switch line.Kind {
		case "heading":
			headings++
			level := "###"
			if headings == 1 {
				level = "##"
			}
			fmt.Fprintf(&markdown, "%s %s\n\n", level, text)
		case "warning":
			fmt.Fprintf(&markdown, "> **Let op:** %s\n\n", text)
		case "meta":
			fmt.Fprintf(&markdown, "_%s_\n\n", text)
		default:
			fmt.Fprintf(&markdown, "%s\n\n", text)
		}
	}
	return markdown.String()
}

func escapeMarkdownText(text string) string {
	var escaped strings.Builder
	for _, r := range text {
		if r == '\n' || r == '\r' || r == '\t' {
			escaped.WriteRune(' ')
			continue
		}
		if strings.ContainsRune("\\`*_{}[]<>()#+-.!|~", r) {
			escaped.WriteRune('\\')
		}
		escaped.WriteRune(r)
	}
	return escaped.String()
}

func catalogMarkdownStyles() glamouransi.StyleConfig {
	bold, italic := true, true
	zero, one := uint(0), uint(1)
	primary := adaptiveMarkdownColor(ui.ColorPrimary)
	text := adaptiveMarkdownColor(ui.ColorText)
	muted := adaptiveMarkdownColor(ui.ColorMuted)
	warning := adaptiveMarkdownColor(ui.ColorWarning)
	ok := adaptiveMarkdownColor(ui.ColorOK)
	underline := true
	return glamouransi.StyleConfig{
		Document: glamouransi.StyleBlock{
			StylePrimitive: glamouransi.StylePrimitive{Color: &text},
			Margin:         &zero,
		},
		Heading: glamouransi.StyleBlock{
			StylePrimitive: glamouransi.StylePrimitive{Color: &primary, Bold: &bold},
		},
		BlockQuote: glamouransi.StyleBlock{
			StylePrimitive: glamouransi.StylePrimitive{Color: &warning},
			Indent:         &one,
			IndentToken:    stringValue("│ "),
		},
		Emph:   glamouransi.StylePrimitive{Color: &muted, Italic: &italic},
		Strong: glamouransi.StylePrimitive{Bold: &bold},
		Link:   glamouransi.StylePrimitive{Color: &primary, Underline: &underline},
		Code: glamouransi.StyleBlock{
			StylePrimitive: glamouransi.StylePrimitive{Color: &ok},
		},
	}
}

func adaptiveMarkdownColor(color lipgloss.AdaptiveColor) string {
	if lipgloss.HasDarkBackground() {
		return color.Dark
	}
	return color.Light
}

func stringValue(value string) *string { return &value }

func catalogPlainDocument(lines []core.CatalogLine, width int) string {
	rows := make([]string, 0, len(lines))
	for _, line := range lines {
		rows = append(rows, ansi.Wrap(line.Text, width, ""))
	}
	return strings.Join(rows, "\n")
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
