package screens

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"waterketen/internal/core"
)

func TestMapPanUsesOneVisibleSpanAtEveryZoom(t *testing.T) {
	for _, library := range []bool{false, true} {
		for _, zoom := range []float64{1, 4} {
			s := NewKaart("", nil, nil)
			s.SetSize(100, 30)
			bounds := core.BBox{MinX: 0, MinY: 0, MaxX: 1000, MaxY: 1000}
			s.SetScene(&core.Scene{Bounds: bounds})
			s.view.Zoom = zoom
			if library {
				s.atlas = &core.LibraryAtlas{Area: core.Area{Bounds: bounds}}
			}
			before := s.view
			var width, height int
			if library {
				width, height = s.innerSize()
			} else {
				width, height = s.w, s.h
			}
			frame := before.Frame(bounds, width, height)
			want := (frame.Bounds.MaxX - frame.Bounds.MinX) * .18
			s.Update(tea.KeyMsg{Type: tea.KeyRight})
			if got := s.view.CX - before.CX; math.Abs(got-want) > 1e-9 || s.view.CY != before.CY {
				t.Fatalf("library=%v zoom=%v: moved %v, want %v", library, zoom, got, want)
			}
		}
	}
}

func regressionMap(t *testing.T) *Kaart {
	t.Helper()
	s := NewKaart("", nil, nil)
	s.SetSize(100, 30)
	scene, err := core.NewScene("test", []core.SceneSource{
		{Name: "riool", Features: []core.Feature{{Layer: "riool", Position: 1, Geometry: core.Geometry{Kind: "point", Point: core.Point{X: 0, Y: 0}}}}},
		{Name: "drinkwaterleidingen", Features: []core.Feature{{Layer: "drinkwaterleidingen", Position: 2, Geometry: core.Geometry{Kind: "point", Point: core.Point{X: 10, Y: 10}}}}},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	scene.Display = []string{"riool", "drinkwaterleidingen"}
	s.SetScene(scene)
	s.layers = []string{"riool", "drinkwaterleidingen"}
	s.atlas = &core.LibraryAtlas{Library: &core.Library{}, Area: core.Area{Bounds: scene.Bounds}}
	return s
}

func TestSoloLayerChangesReloadAndOverrideOldDisplay(t *testing.T) {
	s := regressionMap(t)
	_, cmd := s.Update(keys("o"))
	if cmd == nil || s.loads != 1 {
		t.Fatal("solo toggle did not reload")
	}
	s.CommitLoad(s.scene)
	drawn := s.DrawCmd(1)().(DrawnMsg)
	if len(drawn.Render.Counts) != 1 || drawn.Render.Counts["riool"] != 1 {
		t.Fatalf("solo still draws old display: %v", drawn.Render.Counts)
	}
	_, cmd = s.Update(keys("]"))
	if cmd == nil || s.active != "drinkwaterleidingen" || s.loads != 1 {
		t.Fatal("layer change did not reload new layer")
	}
	s.CommitLoad(s.scene)
	drawn = s.DrawCmd(2)().(DrawnMsg)
	if len(drawn.Render.Counts) != 1 || drawn.Render.Counts["drinkwaterleidingen"] != 1 {
		t.Fatalf("wrong solo layer: %v", drawn.Render.Counts)
	}
	_, cmd = s.Update(keys("o"))
	if cmd == nil {
		t.Fatal("return to samen did not reload")
	}
	s.CommitLoad(s.scene)
	if got := s.DrawCmd(3)().(DrawnMsg).Render.Counts; len(got) != 2 {
		t.Fatalf("samen did not restore layer set: %v", got)
	}
}

func TestEmptyMapSelectionStaysEmptyThroughRendering(t *testing.T) {
	s := regressionMap(t)
	s.SetLayers(nil)
	if s.layers == nil || len(s.visibleLayerNames()) != 0 {
		t.Fatal("empty selection became defaults")
	}
	s.CommitLoad(s.scene)
	if got := s.DrawCmd(1)().(DrawnMsg).Render.Counts; len(got) != 0 {
		t.Fatalf("empty selection draws %v", got)
	}
	s.Update(keys("o"))
	if len(s.visibleLayerNames()) != 0 {
		t.Fatal("solo restored an unselected network")
	}
}

func TestMapLoadErrorsFinishBusyWorkAndRetryStaleView(t *testing.T) {
	s := regressionMap(t)
	s.loads, s.needsDraw = 1, true
	s.FailLoad(errors.New("broken"))
	if s.Busy() {
		t.Fatal("failed load spins forever")
	}
	s.loads, s.stale = 1, true
	if cmd := s.FailLoad(errors.New("old view broken")); cmd == nil || s.loads != 1 {
		t.Fatal("failed superseded view dropped latest request")
	}
	s.SetError(errors.New("no data"))
	s.loads = 0
	if s.Busy() {
		t.Fatal("empty managed scene spins forever")
	}
}

func TestAnalysisAndProfileSurviveSmallTerminalsWithData(t *testing.T) {
	data := []*LayerData{{Area: "test", Layer: "riool", Spec: core.LayerSpecFor("riool"), Features: []core.Feature{{Street: "test", BOB: -2}, {Street: "test", BOB: -1}}}}
	for _, box := range [][2]int{{100, 30}, {30, 2}, {10, 20}, {1, 1}} {
		t.Run(fmt.Sprint(box), func(t *testing.T) {
			analysis, profile := NewAnalyse("", nil), NewProfiel("", nil)
			analysis.SetData(data)
			profile.SetData(data)
			analysis.SetSize(box[0], box[1])
			profile.SetSize(box[0], box[1])
			analysis.View()
			profile.View()
			analysis.Update(tea.KeyMsg{Type: tea.KeyPgDown})
			analysis.View()
		})
	}
}

func TestAnalysisCachesReportsAndShowsReadFailures(t *testing.T) {
	s := NewAnalyse("", nil)
	s.SetSize(100, 30)
	data := []*LayerData{{Area: "test", Layer: "riool", Features: []core.Feature{{BOB: math.NaN()}}}}
	s.SetData(data)
	before := s.View()
	// Rendering/scrolling must not traverse feature data after SetData.
	data[0].Features = nil
	if s.View() != before {
		t.Fatal("report recomputed during View")
	}
	s.SetErrors([]error{errors.New("test failure")})
	if !strings.Contains(s.View(), "test failure") {
		t.Fatal("load error hidden")
	}
}

func TestBronnenScrollsToAllSourcesAndLicences(t *testing.T) {
	s := NewBronnen()
	s.SetSize(48, 12)
	var health []core.SourceHealth
	for i := 0; i < 25; i++ {
		health = append(health, core.SourceHealth{Label: fmt.Sprintf("Source %02d", i), License: "CC0", Status: core.HealthUp, Count: i})
	}
	s.SetHealth(health)
	if !strings.Contains(s.View(), "Source 00") || strings.Contains(s.View(), "Source 24") {
		t.Fatal("bad first page")
	}
	s.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	if s.body.YOffset == 0 {
		t.Fatal("page down did not scroll")
	}
	s.Update(tea.KeyMsg{Type: tea.KeyEnd})
	if !strings.Contains(s.View(), "Gedetecteerd") {
		t.Fatal("end did not reach unwired sources")
	}
	// Walk every row, verifying bounded rendering and reachability.
	s.Update(tea.KeyMsg{Type: tea.KeyHome})
	seenLast := false
	for i := 0; i < s.body.TotalLineCount(); i++ {
		view := s.View()
		seenLast = seenLast || strings.Contains(view, "Source 24")
		for _, line := range strings.Split(view, "\n") {
			if lipgloss.Width(line) > 44 {
				t.Fatal("source row exceeds viewport width")
			}
		}
		s.Update(tea.KeyMsg{Type: tea.KeyDown})
	}
	if !seenLast {
		t.Fatal("last wired source cannot be reached")
	}
	s.Update(tea.KeyMsg{Type: tea.KeyHome})
	if s.body.YOffset != 0 {
		t.Fatal("home did not restore top")
	}
}
