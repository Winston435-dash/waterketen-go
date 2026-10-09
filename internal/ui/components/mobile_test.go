package components

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

func keyOf(k, desc string) key.Binding {
	return key.NewBinding(key.WithKeys(k), key.WithHelp(k, desc))
}

// TestSingleLineIsExactlyOneRow is the mobile contract for the app's fixed
// anatomy. The header and the command bar each get ONE row; if either renders a
// second row the whole screen shifts and the footer falls off the bottom.
//
// This exists because lipgloss Width() pads but does not truncate, and Padding is
// added ON TOP of Width — so the old `Width(n).Padding(0,1)` header rendered n+2
// cells and wrapped. Measured at 30x12: header=2 rows, bar=2 rows, total 14 on a
// 12-row terminal.
func TestSingleLineIsExactlyOneRow(t *testing.T) {
	long := strings.Repeat("waterketen — Kaart · heel lang ", 8)
	for _, width := range []int{1, 8, 20, 30, 40, 72, 118, 200} {
		out := SingleLine(long, width)
		if n := strings.Count(out, "\n"); n != 0 {
			t.Errorf("breedte %d: SingleLine gaf %d newlines", width, n)
		}
		if w := lipgloss.Width(out); w != width {
			t.Errorf("breedte %d: SingleLine gaf %d cellen — moet er precies %d zijn",
				width, w, width)
		}
	}
}

// TestSingleLineKeepsAnsiIntact: the truncation must not cut an escape sequence
// in half, or the terminal paints the rest of the screen in the wrong colour.
func TestSingleLineKeepsAnsiIntact(t *testing.T) {
	styled := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff7bf5")).Bold(true).
		Render(strings.Repeat("pink", 20))
	out := SingleLine(styled, 12)
	if strings.Contains(out, "\x1b[") && !strings.HasSuffix(strings.TrimRight(out, " "), "m") {
		// The last escape must be a complete sequence, not a truncated one.
		idx := strings.LastIndex(out, "\x1b[")
		if !strings.ContainsAny(out[idx+2:], "m") {
			t.Errorf("afgekapte escape-sequentie in %q", out)
		}
	}
}

// TestSingleLineMarksTruncation: a clipped line ends in "…". Without it a chopped
// CBS code ("BU0599108") reads as a typo instead of a clipped row, which is exactly
// what a phone-size picker full of area codes looked like.
func TestSingleLineMarksTruncation(t *testing.T) {
	out := SingleLine("BU05992459 · buurt Bedrijvenpark Noord-West", 30)
	if !strings.HasSuffix(strings.TrimRight(out, " "), "…") {
		t.Errorf("afgekapte regel eindigt niet op een ellips: %q", out)
	}
	if w := lipgloss.Width(out); w != 30 {
		t.Errorf("afgekapte regel is %d cellen, verwacht 30", w)
	}
	// A line that FITS must not gain an ellipsis.
	fits := SingleLine("BU0599", 30)
	if strings.Contains(fits, "…") {
		t.Errorf("passende regel kreeg een ellips: %q", fits)
	}
}

// TestPanelFitsEveryPhoneSize: a Panel claims to occupy exactly its box, and on a
// phone that claim has to hold or the terminal wraps.
func TestPanelFitsEveryPhoneSize(t *testing.T) {
	body := "∅\nNog geen probe-resultaten\n\nDruk v om de bronnen live te verversen."
	for _, box := range []struct{ w, h int }{
		{118, 36}, {80, 24}, {72, 42}, {48, 26}, {40, 20}, {30, 12}, {24, 8}, {20, 5}, {10, 3},
	} {
		out := Panel("Kaart", body, box.w, box.h, true)
		lines := strings.Split(out, "\n")
		if len(lines) > box.h {
			t.Errorf("%dx%d: panel gaf %d regels", box.w, box.h, len(lines))
		}
		for i, line := range lines {
			if w := lipgloss.Width(line); w > box.w {
				t.Errorf("%dx%d: regel %d is %d cellen breed", box.w, box.h, i, w)
			}
		}
	}
}

// TestPanelTitleSurvivesANarrowTerminal: the title used to be SKIPPED when it did
// not fit, leaving an untitled panel — on a 30-column phone every panel lost its
// name. It must clip instead.
func TestPanelTitleSurvivesANarrowTerminal(t *testing.T) {
	out := Panel("AantalZichtbareObjecten", "x", 24, 6, true)
	first := strings.Split(out, "\n")[0]
	if !strings.Contains(first, "Aantal") {
		t.Errorf("titel ontbreekt op 24 kolommen: %q", first)
	}
	if w := lipgloss.Width(first); w > 24 {
		t.Errorf("titelregel is %d cellen breed, max 24: %q", w, first)
	}
}

func TestModalClampsToScreen(t *testing.T) {
	content := "zoek: west\n\n  Westzeedijk\n  West-Kruiskade\n\ntyp · enter · esc"
	for _, screen := range []struct{ w, h int }{
		{118, 36}, {72, 42}, {48, 26}, {40, 20}, {24, 8},
	} {
		out := Modal("", "Straat zoeken", content, screen.w, screen.h, 72)
		lines := strings.Split(out, "\n")
		if len(lines) > screen.h {
			t.Errorf("%dx%d: modal gaf %d regels", screen.w, screen.h, len(lines))
		}
		for i, line := range lines {
			if w := lipgloss.Width(line); w > screen.w {
				t.Errorf("%dx%d: modalregel %d is %d cellen breed: %q",
					screen.w, screen.h, i, w, line)
			}
		}
	}
}

// TestCommandBarIsOneRowAtEveryWidth: the footer is a fixed one-row slot, so a
// wrapping bar is the difference between a usable phone layout and a broken one.
func TestCommandBarIsOneRowAtEveryWidth(t *testing.T) {
	bar := NewCommandBar()
	bindings := []key.Binding{
		keyOf("left", "west"), keyOf("right", "oost"), keyOf("up", "omhoog"),
		keyOf("down", "omlaag"), keyOf("+", "zoom in"), keyOf("-", "zoom uit"),
		keyOf("f", "passend"), keyOf("[", "← laag"), keyOf("]", "laag →"),
		keyOf("o", "lagen samen"), keyOf("l", "straatnamen"), keyOf("/", "straat zoeken"),
		keyOf("g", "gebied"), keyOf("t", "lagen"), keyOf("q", "afsluiten"),
	}
	for _, width := range []int{8, 20, 30, 40, 48, 60, 72, 90, 118, 200} {
		out := bar.View(width, bindings)
		if n := strings.Count(out, "\n"); n != 0 {
			t.Errorf("breedte %d: commandobar gaf %d newlines", width, n)
		}
		if w := lipgloss.Width(out); w != width {
			t.Errorf("breedte %d: commandobar gaf %d cellen", width, w)
		}
	}
}
