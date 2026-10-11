package core

import (
	"slices"
	"testing"
)

// Benchmark the actual full-city library scene used by the profiling probe.
// It skips on machines without the managed snapshot, keeping the normal test
// suite portable while allowing reproducible local render measurements.
func BenchmarkRenderSceneRealLibrary(b *testing.B) {
	library, err := OpenLibrary(DataRoot())
	if err != nil {
		b.Fatal(err)
	}
	if library == nil {
		b.Skip("geen managed library snapshot")
	}
	atlas, err := NewLibraryAtlas(library, "")
	if err != nil {
		b.Fatal(err)
	}
	layers := library.HasLayers("straten", "water", "riool", "drinkwaterleidingen")
	view := FitView(atlas.Area.Bounds)
	scene, err := atlas.Load(view, 106, 30, layers, 0)
	if err != nil {
		b.Fatal(err)
	}
	featureCount := 0
	for _, features := range scene.Layers {
		featureCount += len(features)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		rendered := RenderScene(scene, view, 106, 30, "riool", true, "netwerk", true, true, nil, "", nil, true)
		if len(rendered.Cells) != 30 {
			b.Fatal("render returned unexpected height")
		}
	}
	b.ReportMetric(float64(featureCount), "features")
	b.ReportMetric(float64(len(scene.Layers["riool"])), "riool")
}

func TestDrawProjectedLineMatchesSegmentWiseProjection(t *testing.T) {
	frame := Frame{Bounds: BBox{MinX: 0, MaxY: 20}, MetresPerDot: 1}
	lines := []LineString{
		nil,
		{{X: 4, Y: 4}},
		{{X: -12, Y: 18}, {X: -8.5, Y: 17.5}, {X: 0.2, Y: 15.5}, {X: 6.7, Y: 12.4}, {X: 19.5, Y: 2.5}, {X: 28, Y: 0}},
		{{X: 5, Y: 30}, {X: 5, Y: 21}, {X: 5, Y: -5}},
	}

	for i, line := range lines {
		got := newColourGrid(12, 6)
		got.style, got.priority = "line", 1
		drawProjectedLine(frame, line, &got)

		want := newColourGrid(12, 6)
		want.style, want.priority = "line", 1
		for j := 1; j < len(line); j++ {
			ax, ay := frame.Project(line[j-1].X, line[j-1].Y)
			bx, by := frame.Project(line[j].X, line[j].Y)
			want.line(ax, ay, bx, by)
		}
		if !slices.Equal(got.grid.Text(), want.grid.Text()) {
			t.Errorf("line %d raster differs from independent per-segment projection", i)
		}
	}
}
