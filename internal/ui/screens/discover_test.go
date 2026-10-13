package screens

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"waterketen/internal/core"
)

func TestOntdekReadsSelectedQuestionAndTogglesTechnicalDetails(t *testing.T) {
	s := NewOntdek(t.TempDir())
	s.SetSize(48, 24)
	s.Update(keys("down")) // drinkwater
	s.Update(keys("enter"))
	text := ansi.Strip(s.View())
	if !strings.Contains(text, "Drinkwaterverzorgingsgebied") || strings.Contains(text, "https://") {
		t.Fatalf("selected question is not readable without API noise:\n%s", text)
	}
	s.Update(keys("t"))
	s.Update(keys("end"))
	if !strings.Contains(ansi.Strip(s.View()), "returnCountOnly") {
		t.Fatal("technical details are not reachable by scrolling")
	}
	s.Update(keys("esc"))
	if !strings.Contains(ansi.Strip(s.View()), "Wie levert hier drinkwater?") {
		t.Fatal("back did not restore question list")
	}
	s.Update(keys("]"))
	if s.section != core.CatalogSources {
		t.Fatal("category key did not select sources")
	}
}

func TestOntdekReflowsReadingAtSmallSizes(t *testing.T) {
	for _, size := range []struct{ w, h int }{{118, 34}, {48, 24}, {30, 10}, {24, 6}} {
		s := NewOntdek(t.TempDir())
		s.SetSize(size.w, size.h)
		s.Update(keys("enter"))
		s.Update(keys("t"))
		lines := strings.Split(s.View(), "\n")
		if len(lines) > size.h-2 {
			t.Errorf("%dx%d: too many inner rows: %d", size.w, size.h, len(lines))
		}
		for _, line := range lines {
			if ansi.StringWidth(line) > size.w-4 {
				t.Errorf("%dx%d: overflowing inner line: %q", size.w, size.h, line)
			}
		}
		s.Update(keys("pgdown"))
		s.SetSize(48, 24)
		if strings.Contains(ansi.Strip(s.View()), "onbekend catalogusitem") {
			t.Fatal("resize lost selection")
		}
	}
}
