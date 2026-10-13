package screens

import (
	"github.com/charmbracelet/x/ansi"
	"strings"
	"testing"
	"waterketen/internal/core"
)

func TestMapUsesAvailableHeightAndLegendCanBeReadOnPhone(t *testing.T) {
	for _, size := range []struct{ w, h int }{{120, 34}, {80, 22}, {48, 18}, {30, 10}} {
		s := NewKaart(t.TempDir(), nil, nil)
		s.SetSize(size.w, size.h)
		geom := core.Geometry{Kind: "line", Line: core.LineString{{X: 1, Y: 1}, {X: 90, Y: 35}}}
		s.SetScene(&core.Scene{Bounds: core.BBox{MinX: 0, MinY: 0, MaxX: 100, MaxY: 40}, Layers: map[string][]core.Feature{"riool": {{Layer: "riool", Geometry: geom}}}, Index: map[string]core.GridIndex{"riool": core.NewGridIndex([]core.Geometry{geom})}})
		msg := s.DrawCmd(1)().(DrawnMsg)
		s.Commit(msg.Render)
		view := ansi.Strip(s.View())
		if n := strings.Count(view, "\n") + 1; n != size.h-2 {
			t.Fatalf("%dx%d: map+legend=%d rows, want %d", size.w, size.h, n, size.h-2)
		}
		if size.h < 20 && !strings.Contains(view, "L opent") {
			t.Fatal("compact legend has no route to detail")
		}
		s.Update(keys("L"))
		if !s.OverlayOpen() || !strings.Contains(ansi.Strip(s.View()), "riool") {
			t.Fatal("full legend did not open")
		}
		s.Update(keys("esc"))
		if s.OverlayOpen() {
			t.Fatal("legend did not close")
		}
	}
}
