package core

// Gebied (area) registry for the published library snapshot: the 115 CBS
// gemeente/wijk/buurt boundaries that panel 5 uses as its "world". This is the
// Go counterpart of library.py's MapLibrary.areas()/select_area(), minus the
// shapely polygon clipping — the snapshot is already clipped to Rotterdam at
// build time, which is a guarantee we can rely on instead of re-clipping.

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

// Kind values carried by the snapshot's `kind` column.
const (
	KindGemeente = "gemeente"
	KindWijk     = "wijk"
	KindBuurt    = "buurt"
)

// DefaultAreaCode is the area panel 5 opens on, matching the Python library
// atlas: the whole municipality, not one neighbourhood.
const DefaultAreaCode = "GM0599"

// Area is one selectable CBS area. Only the bounding box is retained: the
// boundary polygon is not needed for loading (the snapshot's layers are
// pre-clipped to Rotterdam) and keeping 115 polygons around would buy nothing.
type Area struct {
	Code   string
	Name   string
	Kind   string
	Bounds BBox
}

// Label is the picker/list rendering. The CBS CODE comes first: it is the unique
// identifier and the thing the user types, and it is short enough to survive the
// truncation a phone screen forces. The name follows, so the list reads
// "WK059903 · wijk Delfshaven" instead of burying the code after a long name.
// This is the canonical contract — see TestAreaLabelLeadsWithCode.
func (a Area) Label() string { return fmt.Sprintf("%s · %s %s", a.Code, a.Kind, a.Name) }

// NormaliseName compares human names and codes without changing the
// authoritative code itself: NFKC-ish casefold, '-' to space, collapse runs of
// whitespace. It mirrors library.normalise_name so "delfshaven", "Delfshaven"
// and "Delfshaven " all resolve to the same area.
func NormaliseName(value string) string {
	folded := strings.Map(func(r rune) rune {
		switch {
		case r == '-':
			return ' '
		case unicode.IsSpace(r):
			return ' '
		case unicode.IsUpper(r):
			return unicode.ToLower(r)
		}
		return r
	}, value)
	return strings.Join(strings.Fields(folded), " ")
}

// LoadAreas reads the snapshot's areas.gpkg. A missing file is not an error:
// older snapshots predate the area registry, and the TUI then falls back to
// the managed-asset atlas exactly like the Python app.
func LoadAreas(lib *Library) ([]Area, error) {
	if lib == nil {
		return nil, nil
	}
	path := filepath.Join(lib.Root, "areas.gpkg")
	if fi, err := os.Stat(path); err != nil || fi.IsDir() {
		return nil, nil
	}
	db, err := Open(path)
	if err != nil {
		return nil, fmt.Errorf("open areas.gpkg: %w", err)
	}
	defer db.Close()
	tbl, err := FeatureTable(db)
	if err != nil {
		return nil, fmt.Errorf("areas.gpkg: %w", err)
	}
	type areaIdentity struct{ code, name, kind string }
	indices := make(map[areaIdentity]int)
	var out []Area
	err = tbl.Each(func(r Row) {
		code, name, kind := r.Properties["code"], r.Properties["name"], r.Properties["kind"]
		if code == "" || name == "" || kind == "" {
			return
		}
		b, ok := r.Geometry.Bounds()
		if !ok {
			return
		}
		key := areaIdentity{code, name, kind}
		if i, exists := indices[key]; exists {
			out[i].Bounds, _ = mergeBBox(out[i].Bounds, true, b)
			return
		}
		indices[key] = len(out)
		out = append(out, Area{Code: code, Name: name, Kind: kind, Bounds: b})
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Kind != out[j].Kind {
			return out[i].Kind < out[j].Kind
		}
		return out[i].Name < out[j].Name
	})
	return out, nil
}

// SelectArea resolves a name or CBS code to exactly one area. An ambiguous
// name (Delfshaven is both a wijk and a buurt) is an error listing the codes,
// never a silent pick — the same rule as the Python select_area.
func SelectArea(areas []Area, name string) (Area, error) {
	want := NormaliseName(name)
	var matched []Area
	for _, a := range areas {
		if a.Code == name || NormaliseName(a.Name) == want {
			matched = append(matched, a)
		}
	}
	switch len(matched) {
	case 0:
		return Area{}, fmt.Errorf("onbekend gebied: %s; zie paneel 3 (Gebieden)", name)
	case 1:
		return matched[0], nil
	default:
		codes := make([]string, len(matched))
		for i, a := range matched {
			codes[i] = a.Code
		}
		return Area{}, fmt.Errorf("gebied %s is dubbelzinnig; kies een code: %s",
			name, strings.Join(codes, ", "))
	}
}

// DefaultArea returns the whole-of-Rotterdam area, falling back to the first
// area when a snapshot does not carry GM0599. Never crashes on an odd
// snapshot: a map of something beats no map.
func DefaultArea(areas []Area) (Area, bool) {
	for _, a := range areas {
		if a.Code == DefaultAreaCode {
			return a, true
		}
	}
	if len(areas) > 0 {
		return areas[0], true
	}
	return Area{}, false
}
