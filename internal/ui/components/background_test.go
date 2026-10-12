package components

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/cellbuf"
	"github.com/muesli/termenv"

	"waterketen/internal/core"
)

func TestAppSurfacePaintsEveryCellBlackAfterNestedStyleResets(t *testing.T) {
	old := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(old)
	content := Strong("Title") + "   " + Muted("text") + "\n" +
		RenderAtlas(&core.Render{Cells: [][]core.Cell{{{Rune: '⣿', Style: core.StyleForLayer("riool")}, {Rune: ' ', Style: core.StyleBlank}}}}) + "\n" +
		"before\x1b[0;39mcombined reset\x1b[49mdefault reset   "
	out := AppSurface(content)
	if ansi.Strip(out) != ansi.Strip(content) {
		t.Fatal("background changed content or geometry")
	}
	for _, line := range strings.Split(out, "\n") {
		assertBlackCells(t, line)
	}
	if !strings.Contains(out, "38;2;56;217;255") {
		t.Fatal("background changed core's riool foreground")
	}
}

func TestAppSurfaceUsesRGBBlackEvenWithAnANSI256Profile(t *testing.T) {
	old := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.ANSI256)
	defer lipgloss.SetColorProfile(old)
	out := AppSurface("   \n x ")
	if !strings.Contains(out, "48;2;0;0;0") {
		t.Fatalf("background uses configurable palette black instead of RGB black: %q", out)
	}
	for _, line := range strings.Split(out, "\n") {
		assertBlackCells(t, line)
	}
}

func assertBlackCells(t *testing.T, line string) {
	t.Helper()
	p := ansi.GetParser()
	defer ansi.PutParser(p)
	var pen cellbuf.Style
	var state byte
	for len(line) > 0 {
		seq, width, n, next := ansi.DecodeSequence(line, state, p)
		if strings.HasPrefix(seq, "\x1b[") && p.Command() == 'm' {
			cellbuf.ReadStyle(p.Params(), &pen)
		}
		if width > 0 {
			if pen.Bg == nil {
				t.Fatalf("%q exposes the terminal's default background", seq)
			}
			r, g, b, _ := pen.Bg.RGBA()
			if r != 0 || g != 0 || b != 0 {
				t.Fatalf("%q has non-black background: %v", seq, pen.Bg)
			}
		}
		line, state = line[n:], next
	}
}
