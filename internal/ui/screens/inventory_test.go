package screens

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"waterketen/internal/core"
)

func inventoryCatalog() core.Catalog {
	c := core.NewCatalog()
	for i := range c.Layers {
		if c.Layers[i].ID == "putten" {
			c.Layers[i].Local = []core.CatalogLocal{{Storage: "snapshot", OnDisk: true, Count: 42, License: "CC0", Area: "snapshot fixture"}}
		}
		if c.Layers[i].ID == "riool" {
			c.Layers[i].Local = []core.CatalogLocal{{Storage: "managed", OnDisk: false, Count: 7, Area: "test"}}
		}
	}
	return c
}

func filterInventory(s Screen, query string) {
	s.Update(keys("/"))
	for _, r := range query {
		s.Update(keys(string(r)))
	}
	s.Update(keys("enter"))
}

func TestLayersUnifyCatalogAndSnapshotAndKeepRemoteUnavailable(t *testing.T) {
	s := NewLagen(inventoryCatalog(), nil)
	s.SetSize(100, 30)
	if !strings.Contains(s.Info(), "1 met lokaal bestand") {
		t.Fatalf("wrong inventory count: %s", s.Info())
	}
	filterInventory(s, "putten")
	s.Update(keys("enter"))
	if view := ansi.Strip(s.View()); !strings.Contains(view, "42 objecten (metadata)") || !strings.Contains(view, "bestand aanwezig") {
		t.Fatalf("snapshot omitted: %s", view)
	}
	s.Update(keys("esc")) // details -> list
	s.Update(keys("esc"))
	filterInventory(s, "damo-rwzi")
	if view := ansi.Strip(s.View()); !strings.Contains(view, "geen lokale levering") || strings.Contains(view, "● DAMO") {
		t.Fatalf("remote layer falsely available: %s", view)
	}
	s.Update(keys("esc"))
	filterInventory(s, "rioolbuizen")
	if !strings.Contains(ansi.Strip(s.View()), "bestand ontbreekt") {
		t.Fatal("metadata-only layer looks downloaded")
	}
	s.Update(keys("esc"))
	s.Update(keys("l"))
	if view := ansi.Strip(s.View()); strings.Contains(view, "DAMO") || !strings.Contains(view, "Rioolputten") {
		t.Fatalf("local-only filter wrong: %s", view)
	}
}

func TestAreasShowRealCodesKindsAndCountRatherThanLayerList(t *testing.T) {
	lib := &core.Library{Snapshot: "fixture", Manifest: core.LibraryManifest{AreaCount: 115}}
	s := NewGebieden(lib, nil)
	s.SetSize(100, 30)
	s.Update(AreasMsg{Areas: []core.Area{
		{Code: "BU05990320", Name: "Delfshaven", Kind: core.KindBuurt},
		{Code: "WK059903", Name: "Delfshaven", Kind: core.KindWijk},
		{Code: "GM0599", Name: "Rotterdam", Kind: core.KindGemeente},
	}})
	view := ansi.Strip(s.View())
	for _, want := range []string{"GM0599 · gemeente Rotterdam", "WK059903 · wijk Delfshaven", "BU05990320 · buurt Delfshaven", "daadwerkelijk gelezen: 3"} {
		if !strings.Contains(view, want) {
			t.Fatalf("missing %q: %s", want, view)
		}
	}
	if !strings.Contains(s.Info(), "3 CBS-gebieden") || strings.Contains(view, "features") {
		t.Fatal("area pane shows layer metadata instead of real areas")
	}
	filterInventory(s, "WK059903")
	view = ansi.Strip(s.View())
	if !strings.Contains(view, "WK059903 · wijk") || strings.Contains(view, "BU05990320") {
		t.Fatalf("area code filter wrong: %s", view)
	}
}

func TestAreaLoadFailureRetainsManagedFallbackAndFinishesBusy(t *testing.T) {
	s := NewGebieden(&core.Library{Root: t.TempDir()}, []core.AssetMeta{{Area: "delfshaven", AreaLabel: "Delfshaven"}})
	s.SetSize(100, 30)
	cmd := s.LoadCmd()
	if cmd == nil || !s.Busy() || s.LoadCmd() != nil {
		t.Fatal("lazy area loading not bounded")
	}
	s.Update(cmd())
	if s.Busy() || !strings.Contains(ansi.Strip(s.View()), "Geen CBS-gebiedenregister") || !strings.Contains(ansi.Strip(s.View()), "delfshaven") {
		t.Fatal("missing area registry hides failure or fallback")
	}
}

func TestInventoryFilterAcceptsUnicodeAndCancelRestoresPrevious(t *testing.T) {
	f := inventoryFilter{query: "old"}
	f.update(keys("/"))
	f.update(pasted(" café"))
	f.update(keys("backspace"))
	if f.query != "old caf" {
		t.Fatalf("bad unicode backspace: %q", f.query)
	}
	f.update(keys("esc"))
	if f.editing || f.query != "old" {
		t.Fatal("filter cancel did not restore old value")
	}
	f.update(keys("esc"))
	if f.query != "" {
		t.Fatal("escape did not clear applied filter")
	}
}

func TestInventoryPanesFitAllSizesAndReachLastArea(t *testing.T) {
	var areas []core.Area
	for i := 0; i < 115; i++ {
		areas = append(areas, core.Area{Code: fmt.Sprintf("BU%08d", i), Name: fmt.Sprintf("Buurt %03d", i), Kind: core.KindBuurt})
	}
	for _, box := range [][2]int{{100, 30}, {48, 26}, {24, 8}, {10, 3}} {
		layers := NewLagen(inventoryCatalog(), errors.New("test diagnostic"))
		regions := NewGebieden(&core.Library{Manifest: core.LibraryManifest{AreaCount: 115}}, nil)
		regions.Update(AreasMsg{Areas: areas})
		for _, s := range []Screen{layers, regions} {
			s.SetSize(box[0], box[1])
			s.Update(keys("end"))
			view := s.View()
			if len(strings.Split(view, "\n")) > max(2, box[1]-2) {
				t.Fatalf("%T %v: view too tall", s, box)
			}
			for _, line := range strings.Split(view, "\n") {
				if ansi.StringWidth(line) > max(1, box[0]-4) {
					t.Fatalf("%T %v: view too wide", s, box)
				}
			}
		}
		if !strings.Contains(ansi.Strip(regions.View()), "114") && box[0] >= 24 {
			t.Fatal("last area not reachable")
		}
	}
}

func TestAnalysisSeparatesManagedCalculationFromSnapshotMetadata(t *testing.T) {
	lib := &core.Library{Snapshot: "fixture"}
	// No geometry read or download is needed to disclose snapshot scope.
	if err := json.Unmarshal([]byte(`{"layers":{"putten":{"file":"missing.gpkg","feature_count":42}}}`), &lib.Manifest); err != nil {
		t.Fatal(err)
	}
	s := NewAnalyse("", nil)
	s.SetLibrary(lib, core.CatalogWithLocal(nil, lib))
	s.SetData([]*LayerData{{Area: "test", Layer: "riool", Features: []core.Feature{{}}}})
	s.SetSize(100, 50)
	view := ansi.Strip(s.View())
	for _, want := range []string{"Managed — berekend", "Snapshot — metadata, geen geometrieanalyse", "42 objecten volgens metadata", "bestand ontbreekt"} {
		if !strings.Contains(view, want) {
			t.Fatalf("missing scope %q: %s", want, view)
		}
	}
}

func TestProfileOnlyCyclesManagedBOBLayers(t *testing.T) {
	s := NewProfiel("", nil)
	s.SetData([]*LayerData{{Layer: "water", Spec: core.LayerSpecFor("drinkwater")}, {Layer: "riool", Spec: core.LayerSpecFor("riool")}})
	if len(s.data) != 1 || s.data[0].Layer != "riool" {
		t.Fatal("profile still cycles layers without BOB")
	}
}
