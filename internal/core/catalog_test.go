package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCatalogKeepsSnapshotSourceDateSeparateFromPublication(t *testing.T) {
	root := t.TempDir()
	snapshot := "20261007T195520Z-252b4bbc"
	dir := filepath.Join(root, "library", "snapshots", snapshot)
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "library", "current.json"), []byte(`{"snapshot":"`+snapshot+`"}`), 0600); err != nil {
		t.Fatal(err)
	}
	// No GeoPackage: the reader must not confuse metadata with a data file.
	manifest := `{"created_at":"2026-10-07T20:02:36Z","crs":"EPSG:28992","layers":{"straten":{"file":"straten.gpkg","feature_count":42,"source":{"source_date":"2026-10-01T12:00:00Z","license":"ODbL","url":"file:///fixture.pbf"}}}}`
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), []byte(manifest), 0600); err != nil {
		t.Fatal(err)
	}
	catalog, err := LoadCatalog(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range catalog.Layers {
		if l.ID != "straten" {
			continue
		}
		if len(l.Local) != 1 || l.Local[0].SourceDate != "2026-10-01T12:00:00Z" || l.Local[0].OnDisk {
			t.Fatalf("lost provenance: %+v", l.Local)
		}
		lines, err := catalog.Explain(CatalogLayers, "straten", false)
		if err != nil {
			t.Fatal(err)
		}
		var text strings.Builder
		for _, line := range lines {
			text.WriteString(line.Text + "\n")
		}
		for _, want := range []string{"metadata zonder bestand", "42 objecten volgens metadata", "07-10-2026 22:02 CEST", "01-10-2026 14:00 CEST"} {
			if !strings.Contains(text.String(), want) {
				t.Errorf("missing %q: %s", want, text.String())
			}
		}
	}
}

func TestCatalogAreaRequestUsesOneOfficialLayer(t *testing.T) {
	lines, err := NewCatalog().Explain(CatalogLayers, "gebieden", true)
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range lines {
		if strings.HasPrefix(line.Text, "Volledige aanvraag:") {
			if !strings.Contains(line.Text, "typeNames=wijkenbuurten:wijken&") {
				t.Fatalf("umbrella label leaked into request: %s", line.Text)
			}
			return
		}
	}
	t.Fatal("missing complete request")
}

func TestCatalogCoversRegisteredSourcesAndHasValidReferences(t *testing.T) {
	c := NewCatalog()
	sources, layers, terms := map[string]bool{}, map[string]bool{}, map[string]bool{}
	for _, source := range c.Sources {
		if sources[source.ID] || source.Provider == "" || source.Summary == "" || source.Limit == "" {
			t.Errorf("incomplete or duplicate source: %+v", source)
		}
		sources[source.ID] = true
	}
	for _, source := range WiredSources {
		if !sources[source.ID] {
			t.Errorf("missing wired source %s", source.ID)
		}
	}
	for _, source := range UnwiredSources {
		if !sources[source.ID] {
			t.Errorf("missing unwired source %s", source.ID)
		}
	}
	for _, term := range c.Terms {
		if terms[term.ID] || term.Meaning == "" || term.Example == "" {
			t.Errorf("incomplete or duplicate term %s", term.ID)
		}
		terms[term.ID] = true
	}
	for _, layer := range c.Layers {
		if layers[layer.ID] || !sources[layer.SourceID] || layer.Official == "" || layer.Limit == "" {
			t.Errorf("incomplete or duplicate layer: %+v", layer)
		}
		layers[layer.ID] = true
		for _, term := range layer.Terms {
			if !terms[term] {
				t.Errorf("layer %s has unknown term %s", layer.ID, term)
			}
		}
	}
	for _, q := range c.Questions {
		for _, layer := range q.Layers {
			if !layers[layer] {
				t.Errorf("question %s has unknown layer %s", q.ID, layer)
			}
		}
	}
	for _, source := range WiredSources {
		if source.TypeName == "" {
			continue
		}
		found := false
		for _, layer := range c.Layers {
			if layer.Official == source.TypeName || layer.RequestLayer == source.TypeName {
				found = true
			}
		}
		if !found {
			t.Errorf("missing explanation of probed layer %s", source.TypeName)
		}
	}
	for _, spec := range LayerSpecs {
		if !layers[spec.Slug] {
			t.Errorf("missing managed layer %s", spec.Slug)
		}
	}
	for _, name := range DrawOrder {
		if !layers[name] {
			t.Errorf("missing map layer %s", name)
		}
	}
}
