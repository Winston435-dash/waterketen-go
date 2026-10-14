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

func TestQueryPickersFitTheContentViewport(t *testing.T) {
	for _, command := range []string{"g", "l"} {
		t.Run(command, func(t *testing.T) {
			s := readyQuery(t)
			s.Update(keys(command))
			fitsHost(t, "querypicker", s.View(), s.width, s.body.Height())
			s.SetSize(48, 24)
			fitsHost(t, "querypicker na resize", s.View(), s.width, s.body.Height())
		})
	}
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
	if !strings.Contains(s.View(), "lagen voor de gebiedstelling") {
		t.Fatal("query layer picker describes drawing instead of counting")
	}
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

func TestQueryObjectPagingKeepsSelectionWhenReturningFromDetails(t *testing.T) {
	s := readyQuery(t)
	s.Update(keys("enter"))
	ids := make([]int64, 84)
	for i := range ids {
		ids[i] = int64(i + 1)
	}
	s.Update(QueryDoneMsg{Generation: s.generation, Result: &core.QueryResult{Area: s.area, Layers: []core.QueryLayerResult{{Layer: core.QueryLayer{ID: "riool", Layer: "riool", Label: "Rioolbuizen"}, Status: core.QueryOK, Count: len(ids), IDs: ids}}}})
	_, cmd := s.Update(keys("enter"))
	if cmd == nil {
		t.Fatal("object page was not dispatched")
	}
	rows := make([]core.Row, 50)
	for i := range rows {
		rows[i] = core.Row{ID: ids[i], Properties: map[string]string{"naam": "fixture-object"}}
	}
	s.Update(QueryObjectsMsg{Generation: s.generation, Offset: 0, Rows: rows})
	for i := 0; i < 47; i++ {
		s.Update(keys("down"))
	}
	s.Update(keys("enter"))
	if !strings.Contains(s.View(), "Object 48") {
		t.Fatal("selected object did not open")
	}
	s.Update(keys("esc"))
	if !strings.Contains(s.View(), "› 48") {
		t.Fatal("returning from details hid the selected object")
	}
	_, cmd = s.Update(keys("end"))
	if cmd == nil || s.pageOffset != 50 {
		t.Fatal("last page not dispatched beyond first 50 objects")
	}
	last := make([]core.Row, len(ids)-50)
	for i := range last {
		last[i] = core.Row{ID: ids[50+i]}
	}
	s.Update(QueryObjectsMsg{Generation: s.generation, Offset: 50, Rows: last})
	if !strings.Contains(s.View(), "› 84") || !strings.Contains(s.View(), "Pagina 51–84 van 84") {
		t.Fatal("final selection or page range is hidden")
	}
	s.Update(keys("enter"))
	if !strings.Contains(s.View(), "Object 84") {
		t.Fatal("final object is unreachable")
	}
}

func TestQueryIgnoresDuplicateCompletions(t *testing.T) {
	s := readyQuery(t)
	s.Update(keys("enter"))
	msg := QueryDoneMsg{Generation: s.generation, Result: &core.QueryResult{Area: s.area, Layers: []core.QueryLayerResult{{Status: core.QueryOK, Count: 1}}}}
	s.Update(msg)
	s.Update(keys("esc"))
	s.Update(msg)
	if s.resultFocus {
		t.Fatal("duplicate completion stole focus back from the form")
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

func TestCancellingInitialLoadAllowsRetry(t *testing.T) {
	s := NewQuery(t.TempDir())
	if cmd := s.LoadCmd(); cmd == nil || !s.Busy() {
		t.Fatal("initial metadata load was not dispatched")
	}
	s.Update(keys("esc"))
	if s.loaded {
		t.Fatal("cancelled initial load was marked complete")
	}
	if cmd := s.LoadCmd(); cmd == nil || !s.Busy() {
		t.Fatal("cancelled initial load could not be retried")
	}
}
