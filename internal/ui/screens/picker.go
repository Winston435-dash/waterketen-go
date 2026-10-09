package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"waterketen/internal/core"
	"waterketen/internal/ui"
	"waterketen/internal/ui/components"
)

// pickMode is what a picker does with its list. It is a field rather than a
// check on the title: keying behaviour off display text is one rename away from
// silently changing how the overlay works.
type pickMode int

const (
	pickOne    pickMode = iota // area: search, cursor, enter confirms
	pickToggle                 // layers: cursor + space, enter applies the set
)

// Picker is the shared searchable-list overlay behind 'g' (gebied) and 't'
// (lagen). It is a full-screen model of its own — the app gives it every key —
// and it renders inside the same Panel the map uses, so it looks like a mode,
// not a separate app.
//
// The list stays short on purpose (12 rows): the snapshot carries 115 areas and
// scrolling 115 with j/k is worse than typing three letters.
type Picker struct {
	mode     pickMode
	title    string
	hint     string
	all      []pickItem
	filtered []pickItem
	cursor   int
	query    string
	w, h     int
}

type pickItem struct {
	key   string // CBS code, or layer slug
	label string
	mark  bool // pickToggle only: is it currently on?
}

// NewAreaPicker offers the snapshot's areas, prefix-first on name or code.
func NewAreaPicker(areas []core.Area) *Picker {
	items := make([]pickItem, len(areas))
	for i, a := range areas {
		items[i] = pickItem{key: a.Code, label: a.Label()}
	}
	return &Picker{
		mode:  pickOne,
		title: "Gebied kiezen",
		hint:  "typ om te zoeken · enter kiest · esc terug",
		all:   items,
	}
}

// NewLayerPicker offers the layers the snapshot actually carries, with the
// currently visible ones pre-marked. `order` fixes the display order so the
// list matches DrawOrder rather than map iteration order.
func NewLayerPicker(available, visible, order []string) *Picker {
	on := map[string]bool{}
	for _, n := range visible {
		on[n] = true
	}
	var items []pickItem
	for _, n := range order {
		for _, a := range available {
			if a == n {
				items = append(items, pickItem{key: n, label: n, mark: on[n]})
				break
			}
		}
	}
	return &Picker{
		mode:  pickToggle,
		title: "Kaartlagen",
		hint:  "spatie schakelt · enter past toe · esc terug",
		all:   items,
	}
}

// SetSize records the screen size and (re)filters. It must exist as a method:
// OpenPicker calls it, and without it the picker would render with w=h=0.
func (p *Picker) SetSize(w, h int) {
	p.w, p.h = w, h
	p.refilter()
}

func (p *Picker) Searchable() bool { return p.mode == pickOne }
func (p *Picker) Toggling() bool   { return p.mode == pickToggle }

// PickerResult is the outcome of closing a picker. Confirm is false when the
// user backed out, in which case the caller must change nothing: reading
// "marked layers" after an escape would apply a half-finished edit.
//
// Toggled says WHICH question was answered — the caller must not infer it from
// "are there marked items", because an empty layer selection is a legitimate
// answer ("draw nothing") and would be mistaken for an area pick.
type PickerResult struct {
	Toggled bool
	Key     string
	Marked  []string
	Confirm bool
}

func (p *Picker) result(confirm bool) PickerResult {
	if !confirm {
		return PickerResult{}
	}
	var marked []string
	for _, it := range p.all {
		if it.mark {
			marked = append(marked, it.key)
		}
	}
	res := PickerResult{Toggled: p.Toggling(), Marked: marked, Confirm: true}
	if p.cursor >= 0 && p.cursor < len(p.filtered) {
		res.Key = p.filtered[p.cursor].key
	}
	return res
}

func (p *Picker) refilter() {
	if !p.Searchable() {
		p.filtered = append([]pickItem(nil), p.all...)
		p.clampCursor()
		return
	}
	q := strings.ToLower(strings.TrimSpace(p.query))
	if q == "" {
		p.filtered = append([]pickItem(nil), p.all...)
		p.clampCursor()
		return
	}
	p.filtered = p.filtered[:0]
	for _, it := range p.all {
		if strings.Contains(strings.ToLower(it.label), q) {
			p.filtered = append(p.filtered, it)
		}
	}
	p.clampCursor()
}

func (p *Picker) clampCursor() {
	if p.cursor >= len(p.filtered) {
		p.cursor = len(p.filtered) - 1
	}
	if p.cursor < 0 {
		p.cursor = 0
	}
}

func (p *Picker) move(delta int) {
	if len(p.filtered) == 0 {
		return
	}
	p.cursor = (p.cursor + delta + len(p.filtered)) % len(p.filtered)
}

// Update handles one key and reports the outcome plus whether to close.
//
// enter APPLIES the picker's state in both modes: it confirms the highlighted
// area, or applies the toggled layer set. esc always cancels. Toggling a
// checkbox without any way to say "done" was the first bug here — the set could
// never be applied, and reaching for esc threw it away.
func (p *Picker) Update(msg tea.KeyMsg) (PickerResult, bool) {
	switch {
	case key.Matches(msg, ui.Keys.Back):
		return p.result(false), true
	case key.Matches(msg, ui.Keys.Enter):
		return p.result(true), true
	case key.Matches(msg, ui.Keys.Up), msg.String() == "ctrl+k":
		p.move(-1)
		return PickerResult{}, false
	case key.Matches(msg, ui.Keys.Down), msg.String() == "ctrl+j":
		p.move(1)
		return PickerResult{}, false
	case msg.String() == " " && p.Toggling():
		p.toggle()
		return PickerResult{}, false
	case msg.Type == tea.KeyBackspace:
		if p.Searchable() && p.query != "" {
			p.query = p.query[:len(p.query)-1]
			p.refilter()
		}
		return PickerResult{}, false
	case msg.Type == tea.KeyRunes:
		if p.Searchable() {
			p.query += string(msg.Runes)
			p.refilter()
		}
		return PickerResult{}, false
	}
	return PickerResult{}, false
}

func (p *Picker) toggle() {
	if p.cursor < 0 || p.cursor >= len(p.filtered) {
		return
	}
	slug := p.filtered[p.cursor].key
	for i := range p.all {
		if p.all[i].key == slug {
			p.all[i].mark = !p.all[i].mark
		}
	}
	p.filtered[p.cursor].mark = !p.filtered[p.cursor].mark
}

// View renders the overlay as a centered Panel. The caller decides what (if
// anything) shows through: nothing is painted behind it.
func (p *Picker) View() string {
	var b strings.Builder
	if p.Searchable() {
		b.WriteString(components.Strong("zoek: " + p.query))
		b.WriteByte('\n')
	} else {
		b.WriteString(components.MutedText("lagen die de kaart tekent"))
		b.WriteByte('\n')
	}
	b.WriteByte('\n')
	if len(p.filtered) == 0 {
		b.WriteString(components.MutedText("  geen treffers"))
		b.WriteByte('\n')
	}
	const window = 12
	start := 0
	if p.cursor >= window {
		start = p.cursor - window + 1
	}
	for i := start; i < len(p.filtered) && i < start+window; i++ {
		it := p.filtered[i]
		line := "  " + it.label
		if p.Toggling() {
			mark := "☐"
			if it.mark {
				mark = "☑"
			}
			line = "  " + mark + " " + it.label
		}
		if i == p.cursor {
			b.WriteString(components.Selected(line, max(20, p.w-14)))
		} else {
			b.WriteString(components.MutedText(line))
		}
		b.WriteByte('\n')
	}
	b.WriteByte('\n')
	b.WriteString(components.MutedText(p.hint))
	title := p.title
	if p.Searchable() && p.query != "" {
		title = fmt.Sprintf("%s · %d van %d", p.title, len(p.filtered), len(p.all))
	} else if p.Toggling() {
		n := 0
		for _, it := range p.all {
			if it.mark {
				n++
			}
		}
		title = fmt.Sprintf("%s · %d van %d aan", p.title, n, len(p.all))
	}
	return components.Modal("", title, b.String(), p.w, p.h, min(72, max(30, p.w-8)))
}
