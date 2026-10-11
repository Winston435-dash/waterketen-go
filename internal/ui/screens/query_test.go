package screens

import (
	"reflect"
	"strings"
	"testing"

	"waterketen/internal/core"
)

func readyQuery(t *testing.T) *Query {
	t.Helper()
	s := NewQuery(t.TempDir())
	s.SetSize(110, 26)
	if s.LoadCmd() == nil {
		t.Fatal("metadata load not dispatched")
	}
	s.Update(QueryOpenMsg{Generation: s.generation, Library: &core.Library{Snapshot: "test"}, Areas: []core.Area{
		{Code: "GM0599", Name: "Rotterdam", Kind: core.KindGemeente},
		{Code: "WK059903", Name: "Delfshaven", Kind: core.KindWijk},
		{Code: "BU05990320", Name: "Delfshaven", Kind: core.KindBuurt},
	}})
	return s
}

func TestQueryPickersDisambiguateAreasAndApplyCheckboxes(t *testing.T) {
	s := readyQuery(t)
	s.Update(keys("g"))
	if !s.OverlayOpen() {
		t.Fatal("area picker did not open")
	}
	s.Update(keys("delf"))
	s.Update(keys("enter"))
	if s.area.Code != "WK059903" {
		t.Fatalf("picked %+v instead of wijk", s.area)
	}
	s.Update(keys("l"))
	s.Update(keys(" ")) // first checkbox: Rioolgemalen
	s.Update(keys("enter"))
	if !reflect.DeepEqual(s.selected, []string{"putten", "riool"}) {
		t.Fatalf("checkboxes not applied: %v", s.selected)
	}
	s.Update(keys("l"))
	s.Update(keys(" "))
	s.Update(keys("esc"))
	if !reflect.DeepEqual(s.selected, []string{"putten", "riool"}) {
		t.Fatal("Escape applied an unfinished selection")
	}
}

func TestQueryCancelDiscardsLateResults(t *testing.T) {
	s := readyQuery(t)
	_, cmd := s.Update(keys("enter"))
	if cmd == nil || !s.Busy() {
		t.Fatal("run only changed UI, no work was dispatched")
	}
	generation := s.generation
	s.Update(keys("esc"))
	s.Update(QueryDoneMsg{Generation: generation, Result: &core.QueryResult{Area: s.area}})
	if s.Busy() || s.result != nil {
		t.Fatal("cancelled run committed a late result")
	}
	if !strings.Contains(s.View(), "afgebroken") {
		t.Fatal("cancellation is not visible")
	}
}
