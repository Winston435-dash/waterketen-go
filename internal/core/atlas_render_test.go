package core

import (
	"math"
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

func TestRenderSceneLineRasterMatchesIndependentProjection(t *testing.T) {
	bounds := BBox{MinX: 0, MinY: 0, MaxX: 20, MaxY: 20}
	geom := Geometry{Kind: "linestring", Line: LineString{
		{X: -12, Y: 18}, {X: -8.5, Y: 17.5}, {X: 0.2, Y: 15.5},
		{X: 6.7, Y: 12.4}, {X: 19.5, Y: 2.5}, {X: 28, Y: 0},
	}}
	scene := &Scene{
		Bounds: bounds,
		Layers: map[string][]Feature{"riool": {{Layer: "riool", Geometry: geom}}},
		Index:  map[string]GridIndex{"riool": NewGridIndex([]Geometry{geom})},
	}
	const width, height = 12, 6
	rendered := RenderScene(scene, FitView(bounds), width, height, "riool", false, "netwerk", false, false, nil, "", []string{"riool"}, false)

	want := NewDotGrid(width, height)
	for _, line := range geom.Outlines() {
		for i := 1; i < len(line); i++ {
			ax, ay := rendered.Frame.Project(line[i-1].X, line[i-1].Y)
			bx, by := rendered.Frame.Project(line[i].X, line[i].Y)
			lineInt(int(math.Round(ax)), int(math.Round(ay)), int(math.Round(bx)), int(math.Round(by)), want.SetDot)
		}
	}

	checked := 0
	for y, row := range want.bits {
		for x, bits := range row {
			if bits == 0 {
				continue
			}
			got := int(rendered.Cells[y][x].Rune) - brailleBlank
			if got&bits != bits {
				t.Errorf("cell (%d,%d) lacks projected line dots: got %#x, want %#x", x, y, got, bits)
			}
			if rendered.Cells[y][x].Style != StyleForLayer("riool") {
				t.Errorf("cell (%d,%d) style = %q, want riool", x, y, rendered.Cells[y][x].Style)
			}
			checked++
		}
	}
	if checked == 0 {
		t.Fatal("reference line produced no visible dots")
	}
}

func TestRenderSceneShowsSurfaceWaterPumpInDotLegend(t *testing.T) {
	bounds := BBox{MinX: 0, MinY: 0, MaxX: 20, MaxY: 20}
	geom := Geometry{Kind: "point", Point: Point{X: 10, Y: 10}}
	scene := &Scene{
		Bounds: bounds,
		Layers: map[string][]Feature{"oppervlaktewatergemalen": {{Layer: "oppervlaktewatergemalen", Geometry: geom}}},
		Index:  map[string]GridIndex{"oppervlaktewatergemalen": NewGridIndex([]Geometry{geom})},
	}

	rendered := RenderScene(scene, FitView(bounds), 12, 6, "oppervlaktewatergemalen", false, "netwerk", false, false, nil, "", []string{"oppervlaktewatergemalen"}, false)
	if len(rendered.Legend) != 1 || rendered.Legend[0].Symbol != "•" || rendered.Legend[0].Style != StyleForLayer("oppervlaktewatergemalen") {
		t.Fatalf("oppervlaktewatergemaal-legenda = %+v, verwacht gekleurde punt", rendered.Legend)
	}
}

func TestRenderSceneDrawsGemalenAsDots(t *testing.T) {
	bounds := BBox{MinX: 0, MinY: 0, MaxX: 20, MaxY: 20}
	geom := Geometry{Kind: "polygon", Polygon: []LineString{{
		{X: 2, Y: 2}, {X: 18, Y: 2}, {X: 18, Y: 18}, {X: 2, Y: 18}, {X: 2, Y: 2},
	}}}
	scene := &Scene{
		Bounds: bounds,
		Layers: map[string][]Feature{"gemalen": {{Layer: "gemalen", Geometry: geom}}},
		Index:  map[string]GridIndex{"gemalen": NewGridIndex([]Geometry{geom})},
	}

	rendered := RenderScene(scene, FitView(bounds), 12, 6, "gemalen", false, "netwerk", false, false, nil, "", []string{"gemalen"}, false)
	markers := 0
	for _, row := range rendered.Cells {
		for _, cell := range row {
			if cell.Style == StyleForLayer("gemalen") {
				markers++
			}
		}
	}
	if markers != 1 {
		t.Fatalf("gemalen polygon painted %d cells; want one point marker", markers)
	}
	if len(rendered.Legend) != 1 || rendered.Legend[0].Symbol != "•" {
		t.Fatalf("gemalen legend = %+v; want a dot marker", rendered.Legend)
	}
}
