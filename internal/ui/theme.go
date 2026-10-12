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

// Charm-pink marks the active UI; neutral secondary ink stays legible on dark
// terminals. Atlas layer colours live in core and do not follow chrome tokens.
//
// The app shell paints a true-black background, independent of the terminal's
// default palette. Components remain transparent over that one shared surface.
var (
	ColorBackground = lipgloss.Color("#000000")                                 // app-wide surface, including blank cells
	ColorPrimary    = lipgloss.AdaptiveColor{Light: "#c81ec8", Dark: "#ff7bf5"} // Charm-pink accent (active UI, busy)
	ColorMuted      = lipgloss.AdaptiveColor{Light: "#536171", Dark: "#9aa8b8"} // secondary UI text
	ColorText       = lipgloss.AdaptiveColor{Light: "#1a2530", Dark: "#d2e2ee"} // body text
	ColorError      = lipgloss.AdaptiveColor{Light: "#b3261e", Dark: "#ff6d64"} // probe failures, impossible values
	ColorWarning    = lipgloss.AdaptiveColor{Light: "#8c6d0b", Dark: "#ffa95c"} // drinkwater accent doubles as warning
	ColorOK         = lipgloss.AdaptiveColor{Light: "#1e7d32", Dark: "#6fe38f"} // source up, probe healthy
	ColorSurface    = lipgloss.AdaptiveColor{Light: "#f2f5f7", Dark: "#0b1725"} // reserved fill — NOT used by default
	ColorSurfaceHi  = lipgloss.AdaptiveColor{Light: "#e2eaef", Dark: "#102235"} // reserved fill — NOT used by default
	ColorBorder     = lipgloss.AdaptiveColor{Light: "#8a99aa", Dark: "#475569"} // calm border
	ColorActiveBg   = lipgloss.AdaptiveColor{Light: "#d0e4ee", Dark: "#173850"} // reserved fill — NOT used by default
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
	BorderActive = lipgloss.RoundedBorder() // focus is accent ink, not a heavier line
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
