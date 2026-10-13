package components

import (
	"fmt"
	"image/color"
	"strings"

	"charm.land/glamour/v2"
	glamouransi "charm.land/glamour/v2/ansi"
	"charm.land/lipgloss/v2"
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
	primary := adaptiveMarkdownColor(ui.ColorPrimary)
	text := adaptiveMarkdownColor(ui.ColorText)
	muted := adaptiveMarkdownColor(ui.ColorMuted)
	warning := adaptiveMarkdownColor(ui.ColorWarning)
	ok := adaptiveMarkdownColor(ui.ColorOK)
	return glamouransi.StyleConfig{
		Document: glamouransi.StyleBlock{
			StylePrimitive: glamouransi.StylePrimitive{Color: &text},
			Margin:         ptr(uint(0)),
		},
		Heading: glamouransi.StyleBlock{
			StylePrimitive: glamouransi.StylePrimitive{Color: &primary, Bold: ptr(true)},
		},
		BlockQuote: glamouransi.StyleBlock{
			StylePrimitive: glamouransi.StylePrimitive{Color: &warning},
			Indent:         ptr(uint(1)),
			IndentToken:    ptr("│ "),
		},
		Emph:   glamouransi.StylePrimitive{Color: &muted, Italic: ptr(true)},
		Strong: glamouransi.StylePrimitive{Bold: ptr(true)},
		Link:   glamouransi.StylePrimitive{Color: &primary, Underline: ptr(true)},
		Code: glamouransi.StyleBlock{
			StylePrimitive: glamouransi.StylePrimitive{Color: &ok},
		},
	}
}

func adaptiveMarkdownColor(c color.Color) string {
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("#%02x%02x%02x", r>>8, g>>8, b>>8)
}

// ptr returns a pointer to v — for glamour's pointer-taking style config, so a
// bool or uint doesn't need a named local just to take its address.
func ptr[T any](v T) *T { return &v }

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
