package core

// Bounded library viewports: read ONLY the objects of the current viewport out
// of the published snapshot's GeoPackages, through the R-tree index every
// GeoPackage carries. This is the Go counterpart of library_atlas.py.
//
// Two deliberate differences from the Python original, both disclosed in the
// TUI rather than hidden:
//
//  1. NO polygon clipping. Python re-clips every loaded feature against the
//     selected area boundary with shapely. Here the R-tree gives the bounding
//     box, so a feature that merely touches the viewport box is drawn in full.
//     The snapshot itself is already clipped to Rotterdam at build time, so
//     the error is bounded by the viewport edge — and it is why a viewport can
//     report more objects than Python does for the same window.
//  2. NO geometry simplification. Python runs `.simplify(metres_per_dot * .3)`.
//     Core has no simplify pass; the renderer drops dots per cell anyway.

import (
	"database/sql"
	"fmt"
	"math"
	"sort"
	"strings"
)

// gwsAliases maps GWSW's native attribute names onto the city-schema names the
// inspector and the BOB/material colour modes read, so one BOB path serves both
// sources. Mirrors library_atlas.GWSW_ALIASES.
var gwsAliases = map[string]string{
	"BobBeginpuntLeiding": "BOB_ACTUEE",
	"BobEindpuntLeiding":  "BOB_ACTU_1",
	"MateriaalLeiding":    "MATERIAAL1",
	"beginpunt":           "BEGINPUT",
	"eindpunt":            "EINDPUT",
	"type":                "BUISTYPE_N",
}

// defaultViewportLimit caps how many objects one viewport may load per layer.
// The Python atlas samples evenly across the candidates above this, with a
// warning; so do we — silently drawing 5.000 of 87.146 pipes as if that were
// all of them would be a lie.
const defaultViewportLimit = 5000

// LibraryAtlas pins the snapshot and the selected area. The area sets the fit
// bounds, not a panning fence: the view may roam past the boundary (that is how
// neighbouring parts of the city stay reachable) while the "fit" view always
// returns to the chosen area.
type LibraryAtlas struct {
	Library *Library
	Areas   []Area
	Area    Area
}

// NewLibraryAtlas opens the area registry and resolves `code` (a CBS code;
// empty means the whole-of-Rotterdam default). A snapshot without areas.gpkg is
// a failure here, because the library atlas cannot place itself without one.
func NewLibraryAtlas(lib *Library, code string) (*LibraryAtlas, error) {
	if lib == nil {
		return nil, fmt.Errorf("geen kaartbibliotheek")
	}
	areas, err := LoadAreas(lib)
	if err != nil {
		return nil, err
	}
	if len(areas) == 0 {
		return nil, fmt.Errorf("snapshot %s heeft geen gebiedenregister", lib.Snapshot)
	}
	la := &LibraryAtlas{Library: lib, Areas: areas}
	if code == "" {
		area, ok := DefaultArea(areas)
		if !ok {
			return nil, fmt.Errorf("geen standaardgebied beschikbaar")
		}
		la.Area = area
		return la, nil
	}
	area, err := SelectArea(areas, code)
	if err != nil {
		return nil, err
	}
	la.Area = area
	return la, nil
}

// --- viewport reading ------------------------------------------------------

// layerTable opens one snapshot layer and resolves its feature table and
// geometry column. The caller closes the returned *sql.DB.
func (la *LibraryAtlas) layerTable(name string) (*sql.DB, *Table, error) {
	path := la.Library.LayerPath(name)
	if path == "" {
		return nil, nil, fmt.Errorf("laag %s zit niet in deze snapshot", name)
	}
	db, err := Open(path)
	if err != nil {
		return nil, nil, err
	}
	tbl, err := FeatureTable(db)
	if err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("%s: %w", name, err)
	}
	return db, tbl, nil
}

// viewportIDs returns the R-tree ids inside `bounds`, evenly sampled when there
// are more than `limit`, plus the true candidate count so the caller can warn.
func viewportIDs(tbl *Table, bounds BBox, limit int) ([]int64, int, error) {
	tree := `"rtree_` + strings.ReplaceAll(tbl.name, `"`, `""`) + `_` +
		strings.ReplaceAll(tbl.geom, `"`, `""`) + `"`
	if err := tbl.checkIndexed(tree); err != nil {
		return nil, 0, err
	}
	where := `SELECT id FROM ` + tree + ` WHERE maxx>=? AND minx<=? AND maxy>=? AND miny<=?`
	params := []any{bounds.MinX, bounds.MaxX, bounds.MinY, bounds.MaxY}

	var count int
	if err := tbl.db.QueryRow(`SELECT COUNT(*) FROM (`+where+`)`, params...).Scan(&count); err != nil {
		return nil, 0, fmt.Errorf("tel %s: %w", tbl.name, err)
	}
	if count == 0 {
		return nil, 0, nil
	}
	stride := 1
	if limit > 0 && count > limit {
		stride = int(math.Ceil(float64(count) / float64(limit)))
	}
	query := where
	if stride > 1 {
		// Even sampling across the candidate set, ordered by id so the sample is
		// stable between redraws of the same viewport.
		query = `SELECT id FROM (SELECT id, ROW_NUMBER() OVER (ORDER BY id) AS n FROM (` +
			where + `)) WHERE (n-1)%?=0`
		params = append(params, stride)
	}
	rows, err := tbl.db.Query(query, params...)
	if err != nil {
		return nil, 0, fmt.Errorf("selecteer %s: %w", tbl.name, err)
	}
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, 0, err
		}
		ids = append(ids, id)
	}
	return ids, count, rows.Err()
}

// Load reads the viewport of every requested layer and builds one Scene. The
// scene's bounds are the SELECTED AREA's, not the loaded features': that keeps
// the map frame stable while panning, exactly like the Python atlas.
//
// Unknown layers become warnings, never failures — a snapshot missing one layer
// still draws the rest.
func (la *LibraryAtlas) Load(view View, width, height int, names []string, limit int) (*Scene, error) {
	if limit <= 0 {
		limit = defaultViewportLimit
	}
	width, height = max(1, width), max(1, height)
	frame := view.Frame(la.Area.Bounds, width, height)
	window := frame.Bounds
	names = la.visibleNames(names)

	var sources []SceneSource
	var warnings []string
	for _, name := range names {
		if _, ok := la.Library.Manifest.Layers[name]; !ok {
			warnings = append(warnings, name+": niet beschikbaar in snapshot")
			continue
		}
		features, count, err := la.layerFeatures(name, window, frame.MetresPerDot, limit)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("%s: %v", name, err))
			continue
		}
		if len(features) > 0 {
			sources = append(sources, SceneSource{Name: name, Features: features})
		}
		if count > len(features) {
			warnings = append(warnings, fmt.Sprintf(
				"%s: overzicht %d/%d kandidaten; zoom in voor alle objecten",
				name, len(features), count))
		}
	}
	if len(sources) == 0 {
		warnings = append(warnings, "Geen objecten in deze uitsnede")
		return nil, fmt.Errorf("%s", strings.Join(warnings, "; "))
	}
	scene, err := NewScene(la.Area.Label(), sources, warnings)
	if err != nil {
		return nil, err
	}
	// The AREA owns the frame, not whatever happened to load. Python does the
	// same (AtlasData.build(..., bounds=self.bounds)): without it the map would
	// rescale itself on every pan, and an empty viewport would lose its frame.
	if b := la.Area.Bounds; b.MaxX > b.MinX && b.MaxY > b.MinY {
		scene.Bounds = b
	}
	scene.Display = names
	return scene, nil
}

// layerFeatures reads one layer's viewport rows and converts them to Features,
// applying the GWSW aliases so BOB and material read the same in both schemas.
func (la *LibraryAtlas) layerFeatures(name string, window BBox, metresPerDot float64, limit int) ([]Feature, int, error) {
	db, tbl, err := la.layerTable(name)
	if err != nil {
		return nil, 0, err
	}
	defer db.Close()
	ids, count, err := viewportIDs(tbl, window, limit)
	if err != nil {
		return nil, 0, err
	}
	if len(ids) == 0 {
		return nil, 0, nil
	}
	var features []Feature
	err = tbl.Select(ids, func(r Row) {
		props := make(map[string]string, len(r.Properties)+len(gwsAliases))
		for k, v := range r.Properties {
			props[k] = v
		}
		numbers := r.Numbers
		for source, alias := range gwsAliases {
			if _, seen := props[alias]; seen {
				continue
			}
			if v, ok := props[source]; ok {
				props[alias] = v
			}
			if v, ok := numbers[source]; ok {
				if numbers == nil {
					numbers = map[string]float64{}
				}
				numbers[alias] = v
			}
		}
		street := props["STRAATNAAM"]
		if street == "" && name == "straten" {
			street = props["name"]
		}
		features = append(features, Feature{
			Layer:      name,
			Position:   int(r.ID),
			Geometry:   r.Geometry,
			Properties: props,
			Street:     street,
			BOB:        bobOf(numbers),
		})
	})
	if err != nil {
		return nil, 0, err
	}
	sort.Slice(features, func(i, j int) bool { return features[i].Position < features[j].Position })
	return features, count, nil
}

// bobOf reads the pipe invert from the two end columns. 0,00 means "never
// measured" — a point mass, not a height — so it is excluded, and a pipe with
// both ends measured takes the mean. Same rule as the Python pipeline.
func bobOf(numbers map[string]float64) float64 {
	beg, okBeg := numbers["BOB_ACTUEE"]
	end, okEnd := numbers["BOB_ACTU_1"]
	useBeg, useEnd := okBeg && beg != 0, okEnd && end != 0
	switch {
	case useBeg && useEnd:
		return (beg + end) / 2
	case useBeg:
		return beg
	case useEnd:
		return end
	}
	return math.NaN()
}

// --- street search (no geometry loaded) -------------------------------------

// visibleNames answers what `waterketen kaart`/`bibliotheek` already answer: a
// NIL or EMPTY layer set means "the networks on top of nothing", NOT "every layer
// in the snapshot". Loading all ten (gebouwen is 226.272 objects) for an empty
// picker result would be both slow and a lie about what the user asked for.
func (la *LibraryAtlas) visibleNames(names []string) []string {
	if len(names) > 0 {
		return OrderLayers(names)
	}
	return la.Library.HasLayers("riool", "drinkwaterleidingen")
}

// StreetNames searches local OSM street names inside the selected area. It reads
// the name column only: no geometry is touched, so the search answers in ~160 ms
// on a 58.748-row layer. It uses core.MatchStreetQuery, the same matcher as the
// managed world, so both rank hits the same way — and both accept "west zeedijk"
// for "Westzeedijk".
func (la *LibraryAtlas) StreetNames(query string) []string {
	path := la.Library.LayerPath("straten")
	if path == "" {
		return nil
	}
	db, err := Open(path)
	if err != nil {
		return nil
	}
	defer db.Close()
	rows, err := db.Query(`SELECT DISTINCT name FROM straten WHERE name IS NOT NULL`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	empty := strings.TrimSpace(query) == ""
	type hit struct {
		name   string
		prefix bool
	}
	var hits []hit
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil
		}
		if empty {
			hits = append(hits, hit{name: name})
			continue
		}
		if ok, prefix := MatchStreetQuery(query, name); ok {
			hits = append(hits, hit{name, prefix})
		}
	}
	sort.SliceStable(hits, func(i, j int) bool {
		if hits[i].prefix != hits[j].prefix {
			return hits[i].prefix
		}
		return strings.ToLower(hits[i].name) < strings.ToLower(hits[j].name)
	})
	limit := 8
	if len(hits) < limit {
		limit = len(hits)
	}
	names := make([]string, limit)
	for i := range names {
		names[i] = hits[i].name
	}
	return names
}

// StreetView focuses the view on every shape carrying one street name.
func (la *LibraryAtlas) StreetView(name string, width, height int) (View, bool) {
	path := la.Library.LayerPath("straten")
	if path == "" {
		return View{}, false
	}
	db, tbl, err := la.layerTable("straten")
	if err != nil {
		return View{}, false
	}
	defer db.Close()
	var ids []int64
	rows, err := tbl.db.Query(`SELECT rowid FROM straten WHERE name=?`, name)
	if err != nil {
		return View{}, false
	}
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err == nil {
			ids = append(ids, id)
		}
	}
	rows.Close()
	if len(ids) == 0 {
		return View{}, false
	}
	box, ok := BBox{}, false
	_ = tbl.Select(ids, func(r Row) {
		b, has := r.Geometry.Bounds()
		if !has {
			return
		}
		if !ok {
			box, ok = b, true
			return
		}
		box.MinX, box.MinY = math.Min(box.MinX, b.MinX), math.Min(box.MinY, b.MinY)
		box.MaxX, box.MaxY = math.Max(box.MaxX, b.MaxX), math.Max(box.MaxY, b.MaxY)
	})
	if !ok {
		return View{}, false
	}
	return ViewToBox(box, la.Area.Bounds, width, height, 40), true
}

// ViewToBox frames `target` (grown by `margin` metres) inside the area and
// returns the matching zoom relative to the fit view.
func ViewToBox(target, area BBox, width, height int, margin float64) View {
	grown := BBox{target.MinX - margin, target.MinY - margin, target.MaxX + margin, target.MaxY + margin}
	base := FitView(area).Frame(area, width, height).MetresPerDot
	needed := FitView(grown).Frame(grown, width, height).MetresPerDot
	return View{
		CX:   (target.MinX + target.MaxX) / 2,
		CY:   (target.MinY + target.MaxY) / 2,
		Zoom: math.Max(1, math.Min(1024, base/needed)),
	}
}
