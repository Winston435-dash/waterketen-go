package components

import (
	"strings"
	"testing"
)

func TestProgressCount(t *testing.T) {
	for _, tc := range []struct {
		done, total int
		want        string
	}{
		{0, 6, "0/6"},
		{3, 6, "3/6"},
		{6, 6, "6/6"},
		{9, 6, "6/6"},  // clamped
		{-1, 6, "0/6"}, // clamped
		{0, 0, "0/1"},  // never divide by a zero total
	} {
		if got := ProgressCount(tc.done, tc.total); got != tc.want {
			t.Errorf("ProgressCount(%d,%d)=%q, verwacht %q", tc.done, tc.total, got, tc.want)
		}
	}
}

// TestNewProgressRenders guards the wrapper: a freshly built bar must render a
// line of exactly the requested cell width (Bubbles pads the bar to Width).
func TestNewProgressRenders(t *testing.T) {
	p := NewProgress(30)
	out := stripANSI(p.View())
	if out == "" {
		t.Fatal("lege progress-bar")
	}
	if n := len([]rune(out)); n != 30 {
		t.Errorf("breedte=%d cellen, verwacht 30 (%q)", n, out)
	}
}

// stripANSI removes CSI escape sequences so the test can count cells.
func stripANSI(s string) string {
	var b strings.Builder
	in := false
	for _, r := range s {
		if r == 0x1b {
			in = true
			continue
		}
		if in {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				in = false
			}
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}
