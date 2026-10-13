package components

import (
	"strings"
	"github.com/charmbracelet/x/ansi"
)

// FitRows bounds an assembled small-screen slot without introducing wrap rows.
func FitRows(text string,width,height int) string {
	rows:=strings.Split(text,"\n")
	if len(rows)>max(1,height) { rows=rows[:max(1,height)] }
	for i,row:=range rows { rows[i]=SingleLine(row,width) }
	return strings.Join(rows,"\n")
}

// WrappedText preserves styled content in scrollable panes. Unlike SingleLine,
// it never discards the tail of a source name, licence or diagnostic.
func WrappedText(text string, width int) string {
	width = max(1, width)
	// Wrap at words when possible, then bound long tokens (licence identifiers,
	// URLs and unbroken diagnostics) too, even on a one-cell viewport.
	return ansi.Hardwrap(ansi.Wrap(text, width, ""), width, true)
}
