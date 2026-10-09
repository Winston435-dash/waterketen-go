// Package ui is the waterketen design system: tokens plus semantic components.
//
// CONTRACT — every screen in this app follows:
//
//  1. Never create borders/colors directly. Use components from ui/components.
//  2. Use theme tokens only (ColorPrimary, SpacingSM, ...). No raw hex/adaptive
//     colors and no literal spacing numbers in screens/.
//  3. Every screen renders Header → Content → CommandBar.
//
// This file is the only place where a color value or a spacing number may be
// written down.
package ui

import "github.com/charmbracelet/lipgloss"

// --- Color tokens -----------------------------------------------------------

// The palette matches the existing waterketen visuals (atlas deep blue with a
// cyan accent) while staying legible on black terminals. The accent is the
// Charm-pink from the Bubble Tea/Bubbles stack; the data-ink of the atlas
// (layer colours) lives in core and stays its own thing.
//
// SURFACES ARE TRANSPARENT. The UI never paints a background: only ink colours
// and panel borders are set, so the terminal's own background shows through.
// The three surface tokens are kept for the rare case a component genuinely
// needs a fill, but nothing should reach for them by default.
var (
	ColorPrimary   = lipgloss.AdaptiveColor{Light: "#c81ec8", Dark: "#ff7bf5"} // Charm-pink accent (active UI, busy)
	ColorMuted     = lipgloss.AdaptiveColor{Light: "#5a6b78", Dark: "#627d91"} // secondary text, street ink
	ColorText      = lipgloss.AdaptiveColor{Light: "#1a2530", Dark: "#d2e2ee"} // body text
	ColorError     = lipgloss.AdaptiveColor{Light: "#b3261e", Dark: "#ff6d64"} // probe failures, impossible values
	ColorWarning   = lipgloss.AdaptiveColor{Light: "#8c6d0b", Dark: "#ffa95c"} // drinkwater accent doubles as warning
	ColorOK        = lipgloss.AdaptiveColor{Light: "#1e7d32", Dark: "#6fe38f"} // source up, probe healthy
	ColorSurface   = lipgloss.AdaptiveColor{Light: "#f2f5f7", Dark: "#0b1725"} // reserved fill — NOT used by default
	ColorSurfaceHi = lipgloss.AdaptiveColor{Light: "#e2eaef", Dark: "#102235"} // reserved fill — NOT used by default
	ColorBorder    = lipgloss.AdaptiveColor{Light: "#8fb3c4", Dark: "#3a5f7a"} // calm border
	ColorActiveBg  = lipgloss.AdaptiveColor{Light: "#d0e4ee", Dark: "#173850"} // reserved fill — NOT used by default
)

// Progress-bar gradient endpoints. These are Bubbles' own default gradient
// (violet → pink), so the bar looks like the one in the Charm docs.
const (
	GradientBarA = "#5A56E0"
	GradientBarB = "#EE6FF8"
)

// --- Spacing tokens (terminal cells) ----------------------------------------

const (
	SpacingXS = 1 // one cell: padding inside badges/tight containers
	SpacingSM = 2 // two cells: default inner padding
	SpacingMD = 3 // section gaps
	SpacingLG = 4 // screen-level margins
)

// --- Border tokens ----------------------------------------------------------

var (
	BorderPanel  = lipgloss.RoundedBorder() // every content panel
	BorderActive = lipgloss.ThickBorder()   // the focused panel only
)

// --- Busy-animation tokens --------------------------------------------------

const (
	// SpinnerFPS is how often the app advances its one busy spinner while
	// bounded work runs. It is a token so no screen picks its own frame rate;
	// 10 fps is visibly alive without repainting on every tick of a
	// high-refresh terminal.
	SpinnerFPS = 10
	// SpinnerFrames is the single spinner alphabet for the whole app, one rune
	// per frame. Components render a frame bare; the label spacing is added by
	// components.Busy so the alphabet stays pure.
	SpinnerFrames = "⣾⣽⣻⢿⡿⣟⣯⣷"
)

// SpinnerFrame returns frame n of the app spinner (wrapping). Pure, so the
// app's tick handler stays free of presentation details.
func SpinnerFrame(n int) string {
	frames := []rune(SpinnerFrames)
	return string(frames[((n%len(frames))+len(frames))%len(frames)])
}

// SidebarWidth is the fixed width of the screen's panel-switcher column.
const SidebarWidth = 24
