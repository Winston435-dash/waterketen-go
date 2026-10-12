package core

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestCatalogInventoryIncludesUnknownLocalLayersWithoutDuplicates(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "present.gpkg"), []byte("file presence only"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "directory.gpkg"), 0700); err != nil {
		t.Fatal(err)
	}
	lib := &Library{Root: root, Snapshot: "test"}
	if err := json.Unmarshal([]byte(`{"layers":{"new-layer":{"file":"present.gpkg","feature_count":42,"source":{"license":"CC0"}},"missing-layer":{"file":"missing.gpkg"},"directory-layer":{"file":"directory.gpkg"}}}`), &lib.Manifest); err != nil {
		t.Fatal(err)
	}
	assets := []AssetMeta{{Area: "test", Layer: "new-layer", LayerLabel: "New local layer", ProcessedPath: filepath.Join(root, "present.gpkg"), FeatureCount: 7},
		{Area: "other", Layer: "new-layer", ProcessedPath: filepath.Join(root, "missing.gpkg")}}
	c := CatalogWithLocal(assets, lib)
	seen := map[string]bool{}
	for _, layer := range c.Layers {
		if seen[layer.ID] {
			t.Fatalf("duplicate %s", layer.ID)
		}
		seen[layer.ID] = true
		if layer.ID == "new-layer" {
			if layer.Label != "New local layer" || len(layer.Local) != 3 || !layer.Available() {
				t.Fatalf("inventory lost deliveries: %+v", layer)
			}
			if layer.Local[0].Storage != "managed" || layer.Local[2].Storage != "snapshot" || layer.Local[2].Count != 42 || layer.Local[2].License != "CC0" {
				t.Fatalf("provenance merged incorrectly: %+v", layer.Local)
			}
		}
		if (layer.ID == "missing-layer" || layer.ID == "directory-layer" || layer.ID == "damo-rwzi") && layer.Available() {
			t.Fatalf("metadata/remote/directory treated as data: %s", layer.ID)
		}
	}
	for _, id := range []string{"new-layer", "missing-layer", "directory-layer", "damo-rwzi", "hwh-peilgebied", "rws-regiogebieden"} {
		if !seen[id] {
			t.Fatalf("hidden layer %s", id)
		}
	}
}

func TestNewSnapshotLayerWorksInMapAndQueryWithoutAllowlist(t *testing.T) {
	lib := queryTestLibrary(t)
	if err := json.Unmarshal([]byte(`{"layers":{"future-points":{"file":"gemalen.gpkg","feature_count":4}}}`), &lib.Manifest); err != nil {
		t.Fatal(err)
	}
	if names := lib.LayerNames(); !reflect.DeepEqual(names, []string{"gemalen", "future-points"}) {
		t.Fatalf("inventory order %v", names)
	}
	atlas, err := NewLibraryAtlas(lib, "")
	if err != nil {
		t.Fatal(err)
	}
	view := FitView(atlas.Area.Bounds)
	scene, err := atlas.Load(view, 40, 20, []string{"future-points"}, 5000)
	if err != nil {
		t.Fatal(err)
	}
	render := RenderScene(scene, view, 40, 20, "future-points", false, "netwerk", false, false, nil, "", nil, true)
	if render.Counts["future-points"] != 4 || len(render.Legend) != 1 || render.Legend[0].Symbol != "•" {
		t.Fatalf("new layer not rendered: counts=%v legend=%v", render.Counts, render.Legend)
	}
	if CellColour(render.Legend[0].Style) != genericLayerInk {
		t.Fatal("new layer legend uses misleading sewer ink")
	}
	result, err := RunAreaQuery(context.Background(), lib, "WK000001", []string{"future-points"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Layers[0].Status != QueryOK || result.Layers[0].Count != 3 {
		t.Fatalf("new layer not queryable: %+v", result.Layers[0])
	}
}

func TestEveryProbedSourceHasALayerExplanation(t *testing.T) {
	c := NewCatalog()
	for _, source := range WiredSources {
		found := false
		for _, layer := range c.Layers {
			if layer.SourceID == source.ID {
				found = true
			}
		}
		if !found {
			t.Errorf("source %s has no layer explanation", source.ID)
		}
	}
}
