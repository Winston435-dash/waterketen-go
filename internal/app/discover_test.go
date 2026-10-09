package app

import (
	"testing"
	"waterketen/internal/ui/screens"
)

func TestStartsWithAccessibleQuestions(t *testing.T) {
	m := New(t.TempDir())
	if m.screens[m.active].ID() != screens.ScreenOntdek {
		t.Fatal("app should open the question-driven catalog")
	}
}
