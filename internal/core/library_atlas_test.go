package core

import (
	"strings"
	"testing"
)

// TestLibraryAtlasOpensOnWholeRotterdam is the regression net for panel 5's
// default world. The Python library atlas opens on GM0599 (the whole
// municipality), NOT on a neighbourhood — and this test fails if that ever
// silently becomes a wijk, because then the map would frame a different city
// with no visible error.
func TestLibraryAtlasOpensOnWholeRotterdam(t *testing.T) {
	lib := testLibrary(t)
	atlas, err := NewLibraryAtlas(lib, "")
	if err != nil {
		t.Fatal(err)
	}
	if atlas.Area.Code != DefaultAreaCode {
		t.Fatalf("standaardgebied = %s, verwacht %s", atlas.Area.Code, DefaultAreaCode)
	}
	if atlas.Area.Kind != KindGemeente {
		t.Errorf("standaardgebied is een %s, verwacht gemeente", atlas.Area.Kind)
	}
	b := atlas.Area.Bounds
	if w := (b.MaxX - b.MinX) / 1000; w < 44 || w > 47 {
		t.Errorf("standaardwereld is %.1f km breed, verwacht ~45,5 (heel Rotterdam)", w)
	}
	// Delfshaven is an order of magnitude smaller, and lies INSIDE the default
	// world. Both facts matter: the first is why the default is a city view, the
	// second is why 'g' is a zoom-in rather than a jump to another map.
	dfw, err := NewLibraryAtlas(lib, "WK059903")
	if err != nil {
		t.Fatal(err)
	}
	d := dfw.Area.Bounds
	if ratio := (b.MaxX - b.MinX) / (d.MaxX - d.MinX); ratio < 10 {
		t.Errorf("Rotterdam is maar %.1fx zo breed als Delfshaven, verwacht >10x", ratio)
	}
	if d.MinX < b.MinX || d.MaxX > b.MaxX || d.MinY < b.MinY || d.MaxY > b.MaxY {
		t.Errorf("Delfshaven %+v ligt niet binnen Rotterdam %+v", d, b)
	}
}

// TestLibraryAtlasViewportIsBounded is the honest-counting guard: a viewport
// must load a bounded number of objects, report the TRUE candidate count, and
// warn when it sampled. Loading 87.146 riool pipes for one screen and calling it
// "the neighbourhood" is the failure this pins.
func TestLibraryAtlasViewportIsBounded(t *testing.T) {
	lib := testLibrary(t)
	atlas, err := NewLibraryAtlas(lib, DefaultAreaCode)
	if err != nil {
		t.Fatal(err)
	}
	names := lib.HasLayers("riool", "straten")
	if len(names) == 0 {
		t.Skip("snapshot mist riool/straten")
	}
	view := FitView(atlas.Area.Bounds)
	scene, err := atlas.Load(view, 110, 38, names, 5000)
	if err != nil {
		t.Fatal(err)
	}
	// The whole municipality over a 110x38 frame exceeds the cap, so every
	// layer must have warned and nothing may silently claim to be complete.
	if len(scene.Warnings) == 0 {
		t.Error("geen enkele waarschuwing voor een volle-stad-viewport met sampling")
	}
	warned := map[string]bool{}
	for _, w := range scene.Warnings {
		for _, n := range names {
			if strings.Contains(w, n) && strings.Contains(w, "overzicht") {
				warned[n] = true
			}
		}
	}
	for _, n := range names {
		if _, ok := lib.Manifest.Layers[n]; !ok {
			continue
		}
		if !warned[n] {
			t.Errorf("%s: geen overzichts-waarschuwing terwijl er gesampeld is", n)
		}
	}
	// Every drawn feature must be one of the loaded ones, and the loaded set
	// must respect the cap.
	for _, n := range names {
		if got := len(scene.Layers[n]); got > 5000 {
			t.Errorf("%s: %d features geladen, limiet was 5000", n, got)
		}
	}
	// The scene frame is the AREA's, not the loaded features': panning must not
	// rescale the map.
	if scene.Bounds != atlas.Area.Bounds {
		t.Errorf("scene-bounds %+v != gebied-bounds %+v", scene.Bounds, atlas.Area.Bounds)
	}
}

// TestLibraryAtlasZoomShrinksWindowAndCount pins what zooming actually does,
// which is the OPPOSITE of the naive expectation. A viewport read loads objects
// whose bounding box intersects the window, so:
//
//	zoom in  -> the window covers LESS ground -> FEWER candidate objects
//
// That is the behaviour (measured on the real snapshot: 3.413 pipes at wijk fit,
// 691 at 4×, 194 at 8×), and it is why the panel's own warning text says "zoom in
// voor alle objecten": at fit, the 5.000 cap makes the city view a SAMPLE.
func TestLibraryAtlasZoomShrinksWindowAndCount(t *testing.T) {
	lib := testLibrary(t)
	atlas, err := NewLibraryAtlas(lib, "WK059903")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := lib.Manifest.Layers["riool"]; !ok {
		t.Skip("snapshot mist riool")
	}
	view := FitView(atlas.Area.Bounds)
	atFit, err := atlas.Load(view, 110, 38, []string{"riool"}, 5000)
	if err != nil {
		t.Fatal(err)
	}
	zoomed := view
	zoomed.Zoom = 4
	atZoom, err := atlas.Load(zoomed, 110, 38, []string{"riool"}, 5000)
	if err != nil {
		t.Fatal(err)
	}

	frameFit := view.Frame(atlas.Area.Bounds, 110, 38)
	frameZoom := zoomed.Frame(atlas.Area.Bounds, 110, 38)
	if spanZoom := frameZoom.Bounds.MaxX - frameZoom.Bounds.MinX; spanZoom >= frameFit.Bounds.MaxX-frameFit.Bounds.MinX {
		t.Errorf("inzoomen vergrootte het venster (%.0f >= %.0f)",
			spanZoom, frameFit.Bounds.MaxX-frameFit.Bounds.MinX)
	}
	if got, base := len(atZoom.Layers["riool"]), len(atFit.Layers["riool"]); got >= base {
		t.Errorf("inzoomen gaf %d buizen tegen %d op fit — het venster werd kleiner, "+
			"dus er horen MINDER kandidaten te zijn", got, base)
	}

	// The cap warning is the honesty check: at fit the wijk already samples
	// (3.413 of 6.825), so the warning must be present. And the drawn set must
	// never exceed the cap.
	if len(atFit.Layers["riool"]) > 5000 {
		t.Errorf("fit-viewport laadde %d buizen, limiet was 5000", len(atFit.Layers["riool"]))
	}
	found := false
	for _, w := range atFit.Warnings {
		if strings.Contains(w, "overzicht") {
			found = true
		}
	}
	if !found {
		t.Error("fit-viewport sampelde maar waarschuwde niet")
	}
	// The scene frame is the AREA's, not the loaded features': panning must not
	// rescale the map.
	if atFit.Bounds != atlas.Area.Bounds {
		t.Errorf("scene-bounds %+v != gebied-bounds %+v", atFit.Bounds, atlas.Area.Bounds)
	}
}

// TestLibraryAtlasLayerSetIsOrdered: the panel's chosen layer set must come out
// in DrawOrder, so the map paints background under network — and an unknown
// slug is dropped rather than drawn blind.
func TestLibraryAtlasLayerSetIsOrdered(t *testing.T) {
	got := OrderLayers([]string{"riool", "straten", "onbekend", "water"})
	want := []string{"water", "straten", "riool"}
	if len(got) != len(want) {
		t.Fatalf("OrderLayers = %v, verwacht %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("OrderLayers = %v, verwacht %v", got, want)
		}
	}
	if len(OrderLayers(nil)) != 0 {
		t.Error("lege lijst hoort leeg te blijven")
	}
}
