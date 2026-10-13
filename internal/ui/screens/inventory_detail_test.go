package screens

import (
	"fmt"
	"github.com/charmbracelet/x/ansi"
	"strings"
	"testing"
	"waterketen/internal/core"
)

func TestLayerDetailsKeepEveryDeliveryAndLicences(t *testing.T) {
	c := inventoryCatalog()
	for n := range c.Layers {
		if c.Layers[n].ID == "putten" {
			c.Layers[n].Local = append(c.Layers[n].Local, core.CatalogLocal{Storage: "managed", Area: "tweede levering", Count: 7, License: "CC BY-SA 4.0", OnDisk: true})
		}
	}
	s := NewLagen(c, nil)
	s.SetSize(48, 18)
	filterInventory(s, "putten")
	s.Update(keys("enter"))
	s.Update(keys("end"))
	view := ansi.Strip(s.View())
	if !strings.Contains(view, "tweede levering") || !strings.Contains(view, "CC BY-SA 4.0") {
		t.Fatal("compact inventory lost delivery provenance")
	}
	s.Update(keys("esc"))
	if s.details {
		t.Fatal("details did not return to list")
	}
}

func TestSourcesShowPartialResultsWhileProbing(t *testing.T) {
	s := NewBronnen()
	s.SetSize(48, 18)
	s.BeginProbe()
	s.AddProbe(core.SourceHealth{SourceID: "fixture", Label: "Eerste bron", Status: core.HealthUp, License: "CC0", Count: 42})
	view := ansi.Strip(s.View())
	if !strings.Contains(view, "Eerste bron") || !strings.Contains(view, fmt.Sprintf("1/%d", len(core.WiredSources))) {
		t.Fatalf("partial results hidden: %s", view)
	}
	s.Update(keys("enter"))
	if !s.details || !strings.Contains(ansi.Strip(s.View()), "CC0") {
		t.Fatal("partial source cannot be inspected")
	}
}

func TestInventoryFilterCursorEditsInMiddle(t *testing.T) {
	f := inventoryFilter{}
	f.update(keys("/"))
	f.update(pasted("cafxé"))
	f.update(keys("left"))
	f.update(keys("backspace"))
	if f.query != "café" {
		t.Fatalf("mid-string Unicode edit failed: %q", f.query)
	}
}
