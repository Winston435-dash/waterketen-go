package components

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"

	"waterketen/internal/core"
)

func TestCatalogDocumentRendersMarkdownWithoutPaintingBackgrounds(t *testing.T) {
	old := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(old)

	got := CatalogDocument([]core.CatalogLine{
		{Text: "Riool onbekend", Kind: "heading"},
		{Text: "Gebruik laag-id en controleer de bron.", Kind: "plain"},
		{Text: "Controleer de bronvoorwaarden.", Kind: "warning"},
		{Text: "Gepubliceerd op 09-10-2026.", Kind: "meta"},
	}, 40)

	plain := ansi.Strip(got)
	for _, text := range []string{"Riool onbekend", "laag-id", "controleer", "Let op: Controleer de bronvoorwaarden"} {
		if !strings.Contains(plain, text) {
			t.Errorf("rendered document lost %q: %q", text, plain)
		}
	}
	if strings.Contains(plain, "## ") || strings.Contains(plain, "_Gepubliceerd") {
		t.Errorf("markdown syntax was not rendered: %q", plain)
	}
	if !strings.Contains(got, ";1m") {
		t.Fatalf("Glamour did not apply markdown emphasis: %q", got)
	}
	if !strings.Contains(got, "\x1b[") {
		t.Fatal("Glamour output has no terminal styling")
	}
	if hasBackgroundSGR(got) {
		t.Fatalf("markdown rendering paints a terminal background: %q", got)
	}
	for _, line := range strings.Split(got, "\n") {
		if width := ansi.StringWidth(line); width > 40 {
			t.Errorf("rendered line is %d cells wide, want <= 40: %q", width, line)
		}
	}
}
