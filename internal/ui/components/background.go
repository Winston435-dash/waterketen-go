package components

import (
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/cellbuf"
	"github.com/muesli/termenv"

	"waterketen/internal/ui"
)

var backgroundRenderer = func() *lipgloss.Renderer {
	r := lipgloss.NewRenderer(io.Discard)
	// Palette entry 0 can be tinted by the terminal theme. Use RGB black,
	// rather than the terminal's configurable ANSI "black" or default colour.
	r.SetColorProfile(termenv.TrueColor)
	return r
}()

// AppSurface supplies the shell's background without changing content or data
// ink. Every row is self-contained because Bubble Tea redraws rows independently.
// Nested bold/text styles often reset all SGR attributes; reapply the background
// after those resets so padding and text never expose the terminal's default.
func AppSurface(view string) string {
	paint := backgroundRenderer.NewStyle().Background(ui.ColorBackground).Render(" ")
	background, _, _ := strings.Cut(paint, " ")
	p := ansi.GetParser()
	defer ansi.PutParser(p)
	var base cellbuf.Style
	ansi.DecodeSequence(background, 0, p)
	cellbuf.ReadStyle(p.Params(), &base)

	var b strings.Builder
	for i, line := range strings.Split(view, "\n") {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(background)
		pen := base
		var state byte
		for len(line) > 0 {
			seq, _, n, next := ansi.DecodeSequence(line, state, p)
			b.WriteString(seq)
			if strings.HasPrefix(seq, "\x1b[") && p.Command() == 'm' {
				cellbuf.ReadStyle(p.Params(), &pen)
				if pen.Bg == nil {
					b.WriteString(background)
					pen.Bg = base.Bg
				}
			}
			line, state = line[n:], next
		}
		b.WriteString("\x1b[0m")
	}
	return b.String()
}
