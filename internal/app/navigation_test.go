package app

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"

	"waterketen/internal/ui/screens"
)

// Navigation must keep the current panel visible, even when only its label fits.
func TestHeaderNavigationFitsTerminal(t *testing.T) {
	for _, width := range []int{118, 100, 84, 48, 30, 24} {
		m := New(t.TempDir())
		m.width, m.height, m.ready = width, 26, true
		m = m.applySize()
		for _, entry := range screens.Screens {
			next, _ := m.Update(appKey(entry.Key))
			m = next.(Model)
			if m.screens[m.active].ID() != entry.ID {
				t.Fatalf("key %s did not select %s", entry.Key, entry.Title)
			}
			header := m.renderHeader()
			if strings.Contains(header, "\n") || lipgloss.Width(header) > width {
				t.Errorf("%d columns: header wraps: %q", width, header)
			}
			if !strings.Contains(header, entry.Key+" "+entry.Title) {
				t.Errorf("%d columns: active panel missing: %q", width, header)
			}
			if width >= 100 {
				for _, tab := range screens.Screens {
					if !strings.Contains(header, tab.Key+" "+tab.Title) {
						t.Errorf("%d columns: navigation is missing %s: %q", width, tab.Title, header)
					}
				}
			}
		}
	}
}
