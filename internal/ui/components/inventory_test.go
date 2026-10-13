package components

import (
	tea "charm.land/bubbletea/v2"
	"fmt"
	"github.com/charmbracelet/x/ansi"
	"strings"
	"testing"
)

func TestInventoryResponsiveSelectionAndIdentity(t *testing.T) {
	for _, width := range []int{24, 30, 48, 96} {
		view := NewInventory()
		view.SetSize(width, 8)
		var items []InventoryItem
		for n := 0; n < 40; n++ {
			items = append(items, InventoryItem{ID: fmt.Sprint(n), Label: fmt.Sprintf("Laag %02d", n), Summary: "bestand aanwezig · CC0", Status: StatusOK})
		}
		view.SetItems(items)
		view.Update(tea.KeyPressMsg{Code: tea.KeyEnd})
		if item, _ := view.Selected(); item.ID != "39" {
			t.Fatal("end did not select final item")
		}
		if !strings.Contains(ansi.Strip(view.View()), "Laag 39") {
			t.Fatal("final selection hidden")
		}
		for _, line := range strings.Split(view.View(), "\n") {
			if ansi.StringWidth(line) != width {
				t.Fatal("inventory width drifted")
			}
		}
		items = append([]InventoryItem{{ID: "new", Label: "Nieuwe bron"}}, items...)
		view.SetItems(items)
		if item, _ := view.Selected(); item.ID != "39" {
			t.Fatal("streaming results lost selected identity")
		}
		view.SetItems(nil)
		if _, ok := view.Selected(); ok {
			t.Fatal("empty filter retained stale selection")
		}
	}
}
