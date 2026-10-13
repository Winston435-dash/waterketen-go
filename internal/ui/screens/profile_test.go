package screens

import (
	"fmt"
	"github.com/charmbracelet/x/ansi"
	"math"
	"strings"
	"testing"
	"waterketen/internal/core"
)

func TestProfileAllBOBRowsReachableAndStreetPickerCancels(t *testing.T) {
	for _, size := range []struct{ w, h int }{{120, 34}, {48, 18}, {30, 10}} {
		s := NewProfiel("", nil)
		s.SetSize(size.w, size.h)
		fs := []core.Feature{{Street: "Onbekend", BOB: math.NaN()}, {Street: "Andere straat", BOB: 3}}
		for i := 0; i < 54; i++ {
			fs = append(fs, core.Feature{Street: "Tjalklaan", BOB: -7 + float64(i)/10})
		}
		s.SetData([]*LayerData{{Layer: "riool", Spec: core.LayerSpecFor("riool"), Features: fs}})
		if !strings.Contains(ansi.Strip(s.View()), "Tjalklaan") {
			t.Fatal("default street lost")
		}
		s.Update(keys("end"))
		view := ansi.Strip(s.View())
		if !strings.Contains(view, "54.") || !strings.Contains(view, "Einde") {
			t.Fatalf("%dx%d final BOB unreachable: %s", size.w, size.h, view)
		}
		s.Update(keys("home"))
		s.Update(keys("s"))
		s.Update(keys("Andere"))
		s.Update(keys("esc"))
		if s.street != "Tjalklaan" {
			t.Fatal("cancel changed street")
		}
		s.Update(keys("s"))
		s.Update(pasted("Andere straat"))
		s.Update(keys("enter"))
		if s.street != "Andere straat" || !strings.Contains(ansi.Strip(s.View()), fmt.Sprintf("%.2f", 3.0)) {
			t.Fatal("street choice did not change BOB list")
		}
	}
}
