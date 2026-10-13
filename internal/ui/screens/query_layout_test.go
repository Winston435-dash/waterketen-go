package screens

import (
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"strings"
	"testing"
	"waterketen/internal/core"
)

func queryWithResults(t *testing.T, w, h int) *Query {
	s := readyQuery(t)
	s.SetSize(w, h)
	s.Update(keys("enter"))
	layers := []core.QueryLayerResult{
		{Layer: core.QueryLayer{ID: "rioolgemalen", Label: "Rioolgemalen"}, Status: core.QueryOK, Count: 13, IDs: []int64{1}},
		{Layer: core.QueryLayer{ID: "putten", Label: "Rioolputten"}, Status: core.QueryOK, Count: 4184},
		{Layer: core.QueryLayer{ID: "riool", Label: "Rioolbuizen"}, Status: core.QueryOK, Count: 4998},
	}
	s.Update(QueryDoneMsg{Generation: s.generation, Result: &core.QueryResult{Area: s.area, Layers: layers}})
	return s
}

func TestQueryResultsVisibleFirstAtEverySize(t *testing.T) {
	for _, size := range []struct{ w, h int }{{120, 34}, {80, 22}, {48, 18}, {30, 10}} {
		s := queryWithResults(t, size.w, size.h)
		view := ansi.Strip(s.View())
		for _, want := range []string{"Rioolgemalen", "Rioolputten", "Rioolbuizen"} {
			if !strings.Contains(view, want) {
				t.Fatalf("%dx%d results hidden: %s", size.w, size.h, view)
			}
		}
		s.Update(keys("down"))
		_, cmd := s.Update(keys("enter"))
		if cmd == nil {
			t.Fatal("object selection not actionable")
		}
		s.Update(keys("esc"))
		if !s.resultFocus || !strings.Contains(ansi.Strip(s.View()), "›") {
			t.Fatal("return did not restore result selection")
		}
		s.SetSize(30, 10)
		if !strings.Contains(ansi.Strip(s.View()), "›") {
			t.Fatal("resize hid result selection")
		}
	}
}

func TestQueryCompactFormShowsRunAction(t *testing.T) {
	s := readyQuery(t)
	s.SetSize(30, 10)
	if !strings.Contains(ansi.Strip(s.View()), "Enter: query uitvoeren") {
		t.Fatal("run action hidden below form")
	}
}

func TestQueryHasSeparateRerunAction(t *testing.T) {
	s := queryWithResults(t, 48, 18)
	_, cmd := s.Update(tea.KeyPressMsg{Code: 'r', Mod: tea.ModCtrl})
	if cmd == nil || !s.running || s.result != nil {
		t.Fatal("rerun did not dispatch independent query")
	}
}
