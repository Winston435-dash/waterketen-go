package components

import "github.com/charmbracelet/x/ansi"

// WrappedText preserves styled content in scrollable panes. Unlike SingleLine,
// it never discards the tail of a source name, licence or diagnostic.
func WrappedText(text string, width int) string {
	width = max(1, width)
	// Wrap at words when possible, then bound long tokens (licence identifiers,
	// URLs and unbroken diagnostics) too, even on a one-cell viewport.
	return ansi.Hardwrap(ansi.Wrap(text, width, ""), width, true)
}
