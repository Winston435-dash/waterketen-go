package components

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestWrappedTextPreservesEntireLicencesAtSmallWidths(t *testing.T) {
	text := "licentie: CC BY-NC-ND 4.0 (geen commercieel gebruik, geen afgeleiden)"
	for _, width := range []int{1, 10, 44, 96} {
		wrapped := WrappedText(Muted(text), width)
		var words []string
		for _, line := range strings.Split(wrapped, "\n") {
			if ansi.StringWidth(line) > width {
				t.Fatalf("width %d: oversized line %q", width, line)
			}
			words = append(words, ansi.Strip(line))
		}
		// Wrapping may consume spaces at a line boundary, never characters.
		got := strings.ReplaceAll(strings.Join(words, ""), " ", "")
		if got != strings.ReplaceAll(text, " ", "") {
			t.Fatalf("width %d: text lost: %q", width, got)
		}
	}
}
