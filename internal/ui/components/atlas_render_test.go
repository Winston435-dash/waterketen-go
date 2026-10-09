package components

import (
	"strconv"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"waterketen/internal/core"
)

// TestRenderAtlasEmitsNoBackground is the transparency guard for the map: the
// atlas must set ink, bold and underline only, never a background, so the
// terminal's own background shows through. It forces a truecolor profile (tests
// run without a TTY, where lipgloss would otherwise strip all colour) so the
// assertion is made against the SGR the app really emits.
func TestRenderAtlasEmitsNoBackground(t *testing.T) {
	old := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(old)

	r := &core.Render{Cells: [][]core.Cell{
		{{Rune: '⣿', Style: core.StyleForLayer("riool")}, {Rune: ' ', Style: core.StyleBlank}},
		{{Rune: 'N', Style: core.StyleNorth}, {Rune: 'X', Style: core.StyleHighlight}},
	}}
	out := RenderAtlas(r)
	if hasBackgroundSGR(out) {
		t.Errorf("atlas zet een achtergrond: %q", out)
	}
	// The riool data ink must actually reach the terminal (38;2;56;217;255).
	if !strings.Contains(out, "38;2;56;217;255") {
		t.Errorf("riool-ink (38;2;56;217;255) ontbreekt in %q", out)
	}
}

// TestRenderAtlasPreservesGridAndRuns checks the text survives: stripping SGR
// must give back exactly the grid, newlines included.
func TestRenderAtlasPreservesGridAndRuns(t *testing.T) {
	old := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(old)

	const n = 12
	cells := make([][]core.Cell, 2)
	cells[0] = make([]core.Cell, n)
	cells[1] = make([]core.Cell, n)
	for x := 0; x < n; x++ {
		cells[0][x] = core.Cell{Rune: '─', Style: core.StyleForLayer("riool")}
		cells[1][x] = core.Cell{Rune: '.', Style: core.StyleBlank}
	}
	out := RenderAtlas(&core.Render{Cells: cells})
	want := strings.Repeat("─", n) + "\n" + strings.Repeat(".", n)
	if got := stripSGR(out); got != want {
		t.Errorf("grid-tekst = %q, wil %q", got, want)
	}
}

// hasBackgroundSGR reports whether a string sets a background, parsed properly:
// a foreground triplet like 38;2;23;43;62 contains 43, so a naive per-parameter
// scan reports a false background. This consumes 38/48 extended forms.
func hasBackgroundSGR(s string) bool {
	for i := 0; i+1 < len(s); i++ {
		if s[i] != 0x1b || s[i+1] != '[' {
			continue
		}
		j := i + 2
		for j < len(s) && s[j] != 'm' {
			j++
		}
		params := strings.Split(s[i+2:j], ";")
		for k := 0; k < len(params); k++ {
			n, err := strconv.Atoi(params[k])
			if err != nil {
				continue
			}
			switch {
			case n == 38: // foreground extended: skip 38;5;n or 38;2;r;g;b
				if k+1 < len(params) {
					switch params[k+1] {
					case "5":
						k += 2
					case "2":
						k += 4
					}
				}
			case n == 48: // background extended — the thing we forbid
				return true
			case n >= 40 && n <= 49, n >= 100 && n <= 107:
				return true
			}
		}
		i = j
	}
	return false
}

func stripSGR(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == 0x1b && i+1 < len(s) && s[i+1] == '[' {
			j := i + 2
			for j < len(s) && s[j] != 'm' {
				j++
			}
			i = j
			continue
		}
		b.WriteByte(s[i])
	}
	return b.String()
}
