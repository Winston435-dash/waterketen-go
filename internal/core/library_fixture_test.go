package core

import (
	"reflect"
	"strings"
	"testing"
)

func TestLibraryAtlasFixtureSamplingIDsAndEmptyViews(t *testing.T) {
	lib := queryTestLibrary(t)
	atlas, err := NewLibraryAtlas(lib, "")
	if err != nil {
		t.Fatal(err)
	}
	view := FitView(atlas.Area.Bounds)
	scene, err := atlas.Load(view, 40, 20, []string{"gemalen"}, 2)
	if err != nil {
		t.Fatal(err)
	}
	features := scene.Layers["gemalen"]
	if len(features) != 2 || features[0].Position == features[1].Position {
		t.Fatalf("sampling lost distinct IDs: %+v", features)
	}
	if len(scene.Warnings) != 1 || !strings.Contains(scene.Warnings[0], "2/4") {
		t.Fatalf("sampling not disclosed: %v", scene.Warnings)
	}
	if scene.Bounds != atlas.Area.Bounds {
		t.Fatal("viewport lost area frame")
	}
	for _, tc := range []struct {
		name   string
		view   View
		layers []string
	}{
		{"empty-selection", view, []string{}},
		{"empty-window", View{CX: 1000, CY: 1000, Zoom: 1}, []string{"gemalen"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			empty, err := atlas.Load(tc.view, 40, 20, tc.layers, 2)
			if err != nil {
				t.Fatal(err)
			}
			if len(empty.Layers) != 0 || empty.Bounds != atlas.Area.Bounds || !reflect.DeepEqual(empty.Display, tc.layers) {
				t.Fatalf("bad empty scene: %+v", empty)
			}
		})
	}
}

func TestExplicitEmptyLayersDoNotBecomeDefaults(t *testing.T) {
	lib := queryTestLibrary(t)
	atlas := &LibraryAtlas{Library: lib}
	if got := atlas.visibleNames([]string{}); got == nil || len(got) != 0 {
		t.Fatalf("empty selection became %v", got)
	}
	if got := OrderLayers([]string{}); got == nil {
		t.Fatal("explicit empty set became nil")
	}
}
