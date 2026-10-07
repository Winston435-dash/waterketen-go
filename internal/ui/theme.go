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
// cyan accent) while staying legible on black terminals.
var (
	ColorPrimary   = lipgloss.AdaptiveColor{Light: "#0b6e8c", Dark: "#38d9ff"} // cyan accent (riool lines, active UI)
	ColorMuted     = lipgloss.AdaptiveColor{Light: "#5a6b78", Dark: "#627d91"} // secondary text, street ink
	ColorText      = lipgloss.AdaptiveColor{Light: "#1a2530", Dark: "#d2e2ee"} // body text
	ColorError     = lipgloss.AdaptiveColor{Light: "#b3261e", Dark: "#ff6d64"} // probe failures, impossible values
	ColorWarning   = lipgloss.AdaptiveColor{Light: "#8c6d0b", Dark: "#ffa95c"} // drinkwater accent doubles as warning
	ColorOK        = lipgloss.AdaptiveColor{Light: "#1e7d32", Dark: "#6fe38f"} // source up, probe healthy
	ColorSurface   = lipgloss.AdaptiveColor{Light: "#f2f5f7", Dark: "#0b1725"} // panel background
	ColorSurfaceHi = lipgloss.AdaptiveColor{Light: "#e2eaef", Dark: "#102235"} // raised surface (HUD, help)
	ColorBorder    = lipgloss.AdaptiveColor{Light: "#8fb3c4", Dark: "#24445c"} // calm border
	ColorActiveBg  = lipgloss.AdaptiveColor{Light: "#d0e4ee", Dark: "#173850"} // selected row / active panel
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

// SidebarWidth is the fixed width of the screen's panel-switcher column.
const SidebarWidth = 24
