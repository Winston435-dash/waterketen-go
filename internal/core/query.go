package core

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
)

// QueryLayer is a selectable snapshot layer, optionally restricted to a source
// object type. GWSW's mixed gemalen delivery is not a count of sewer pumps.
type QueryLayer struct {
	ID, Label, Layer, TypeURI, SourceID string
	Available                           bool
}

type QueryStatus string

const (
	QueryOK          QueryStatus = "ok"
	QueryUnavailable QueryStatus = "unavailable"
	QueryFailed      QueryStatus = "failed"
)

type QueryLayerResult struct {
	Layer                          QueryLayer
	Status                         QueryStatus
	Count                          int // -1 when unavailable or failed; zero is a successful empty query
	IDs                            []int64
	Error                          string
	SourceURL, License, SourceDate string
}

type QueryResult struct {
	Area                  Area
	Snapshot, PublishedAt string
	Layers                []QueryLayerResult
}

// QueryLayers keeps water-infrastructure choices first and adds every other
// published vector layer. Alternatives stay separate; there is no grand total.
func QueryLayers(lib *Library) []QueryLayer {
	out := []QueryLayer{
		{ID: "rioolgemalen", Label: "Rioolgemalen", Layer: "gemalen", TypeURI: "http://data.gwsw.nl/1.5/totaal/Rioolgemaal", SourceID: "gws_aanvullend"},
		{ID: "putten", Label: "Rioolputten", Layer: "putten", SourceID: "gws_aanvullend"},
		{ID: "riool", Label: "Rioolbuizen", Layer: "riool", SourceID: "gws_netwerk"},
		{ID: "oppervlaktewatergemalen", Label: "Oppervlaktewatergemalen", Layer: "oppervlaktewatergemalen", SourceID: "rotterdam_oppervlaktewatergemalen"},
		{ID: "drinkwaterleidingen", Label: "Drinkwaterleidingen (Evides)", Layer: "drinkwaterleidingen", SourceID: "evides"},
	}
	if lib == nil {
		return out
	}
	seen := map[string]bool{}
	for i := range out {
		_, out[i].Available = lib.Manifest.Layers[out[i].Layer]
		seen[out[i].ID] = true
	}
	catalog := NewCatalog()
	var extra []QueryLayer
	for id := range lib.Manifest.Layers {
		if seen[id] {
			continue
		}
		l := QueryLayer{ID: id, Label: id, Layer: id, Available: true}
		for _, c := range catalog.Layers {
			if c.ID == id {
				l.Label, l.SourceID = c.Label, c.SourceID
				break
			}
		}
		extra = append(extra, l)
	}
	sort.Slice(extra, func(i, j int) bool { return extra[i].Label < extra[j].Label })
	return append(out, extra...)
}

// RunAreaQuery reads the immutable snapshot offline. The R-tree only reduces
// candidates: full, unsampled geometries are tested against the official polygon,
// including holes and boundary touches. Results retain IDs, not all geometries.
func RunAreaQuery(ctx context.Context, lib *Library, code string, layers []string) (*QueryResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if lib == nil {
		return nil, fmt.Errorf("geen kaartbibliotheek beschikbaar")
	}
	if len(layers) == 0 {
		return nil, fmt.Errorf("kies minstens één laag")
	}
	available := map[string]QueryLayer{}
	for _, l := range QueryLayers(lib) {
		available[l.ID] = l
	}
	seen := map[string]bool{}
	for _, id := range layers {
		if _, ok := available[id]; !ok || seen[id] {
			return nil, fmt.Errorf("onbekende of dubbele querylaag: %s", id)
		}
		seen[id] = true
	}
	area, geometry, err := queryArea(ctx, lib, code)
	if err != nil {
		return nil, err
	}
	filter, err := newAreaFilter(geometry)
	if err != nil {
		return nil, err
	}
	result := &QueryResult{Area: area, Snapshot: lib.Snapshot, PublishedAt: lib.Manifest.CreatedAt}
	for _, id := range layers {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		row := queryCountLayer(ctx, lib, available[id], area.Bounds, filter)
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		result.Layers = append(result.Layers, row)
	}
	return result, nil
}

func queryArea(ctx context.Context, lib *Library, code string) (Area, Geometry, error) {
	areas, err := LoadAreas(lib)
	if err != nil {
		return Area{}, Geometry{}, err
	}
	area, err := SelectArea(areas, code)
	if err != nil {
		return Area{}, Geometry{}, err
	}
	db, err := Open(filepath.Join(lib.Root, "areas.gpkg"))
	if err != nil {
		return Area{}, Geometry{}, err
	}
	defer db.Close()
	table, err := queryTable(ctx, db)
	if err != nil {
		return Area{}, Geometry{}, err
	}
	var geometry Geometry
	err = table.eachQueryRow(ctx, `WHERE "code"=?`, []any{area.Code}, func(row Row) { geometry = row.Geometry })
	if err != nil {
		return Area{}, Geometry{}, err
	}
	if geometry.Kind == "" {
		return Area{}, Geometry{}, fmt.Errorf("gebiedsgrens %s ontbreekt", area.Code)
	}
	return area, geometry, nil
}

func queryCountLayer(ctx context.Context, lib *Library, layer QueryLayer, bounds BBox, filter *areaFilter) QueryLayerResult {
	result := QueryLayerResult{Layer: layer, Status: QueryUnavailable, Count: -1}
	record, ok := lib.Manifest.Layers[layer.Layer]
	if !ok {
		result.Error = "niet in deze snapshot"
		return result
	}
	result.SourceURL, result.License, result.SourceDate = record.Source.URL, record.Source.License, record.Source.Date
	path := lib.LayerPath(layer.Layer)
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		result.Error = "niet lokaal beschikbaar"
		return result
	}
	fail := func(err error) QueryLayerResult {
		result.Status, result.Error, result.IDs = QueryFailed, err.Error(), nil
		return result
	}
	if err != nil {
		return fail(err)
	}
	if info.IsDir() {
		return fail(fmt.Errorf("laagbestand is een map"))
	}
	db, err := Open(path)
	if err != nil {
		return fail(err)
	}
	defer db.Close()
	table, err := queryTable(ctx, db)
	if err != nil {
		return fail(err)
	}
	tree := quoteQueryIdentifier("rtree_" + table.name + "_" + table.geom)
	if err := table.checkIndexed(tree); err != nil {
		return fail(err)
	}
	where := `WHERE rowid IN (SELECT id FROM ` + tree + ` WHERE maxx>=? AND minx<=? AND maxy>=? AND miny<=?)`
	args := []any{bounds.MinX, bounds.MaxX, bounds.MinY, bounds.MaxY}
	if layer.TypeURI != "" {
		columns, err := table.columns()
		if err != nil {
			return fail(err)
		}
		// SQLite can treat a missing double-quoted column as a string literal.
		// Without an explicit check, an unknown schema looks like zero pumps.
		if !slices.Contains(columns.names, "type") {
			return fail(fmt.Errorf("objecttypekolom ontbreekt in %s", table.name))
		}
		where += ` AND "type"=?`
		args = append(args, layer.TypeURI)
	}
	err = table.eachQueryRow(ctx, where, args, func(row Row) {
		if filter.intersects(row.Geometry) {
			result.IDs = append(result.IDs, row.ID)
		}
	})
	if err != nil {
		return fail(err)
	}
	result.Status, result.Count = QueryOK, len(result.IDs)
	return result
}

// ReadQueryObjects pages through the exact IDs from a completed query. A page
// limit affects display only, never the count; source IDs and properties survive.
func ReadQueryObjects(ctx context.Context, lib *Library, result QueryLayerResult, offset, limit int) ([]Row, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if lib == nil || result.Status != QueryOK {
		return nil, fmt.Errorf("geen leesbaar queryresultaat")
	}
	if offset < 0 || limit <= 0 || limit > 100 {
		return nil, fmt.Errorf("ongeldige objectpagina")
	}
	if offset >= len(result.IDs) {
		return nil, nil
	}
	ids := result.IDs[offset:min(offset+limit, len(result.IDs))]
	db, err := Open(lib.LayerPath(result.Layer.Layer))
	if err != nil {
		return nil, err
	}
	defer db.Close()
	table, err := queryTable(ctx, db)
	if err != nil {
		return nil, err
	}
	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	where := `WHERE rowid IN (` + strings.TrimSuffix(strings.Repeat("?,", len(ids)), ",") + `)`
	var rows []Row
	err = table.eachQueryRow(ctx, where, args, func(row Row) { rows = append(rows, row) })
	if err != nil {
		return nil, err
	}
	if len(rows) != len(ids) {
		return nil, fmt.Errorf("snapshotobjecten ontbreken; voer de query opnieuw uit")
	}
	return rows, nil
}

func queryTable(ctx context.Context, db *sql.DB) (*Table, error) {
	table, err := FeatureTable(db)
	if err != nil {
		return nil, err
	}
	var srid int
	if err := db.QueryRowContext(ctx, `SELECT srs_id FROM gpkg_geometry_columns WHERE table_name=? AND column_name=?`, table.name, table.geom).Scan(&srid); err != nil {
		return nil, err
	}
	if srid != 28992 {
		return nil, fmt.Errorf("query verwacht EPSG:28992, vond EPSG:%d", srid)
	}
	return table, nil
}

func quoteQueryIdentifier(name string) string { return `"` + strings.ReplaceAll(name, `"`, `""`) + `"` }

// Unlike viewport reads, exact queries must fail on an unreadable geometry,
// rather than silently report an incomplete count. QueryContext owns cancellation.
func (t *Table) eachQueryRow(ctx context.Context, where string, args []any, fn func(Row)) error {
	cols, err := t.columns()
	if err != nil {
		return err
	}
	const idColumn = "_waterketen_query_rowid"
	columns := []string{`rowid AS ` + quoteQueryIdentifier(idColumn)}
	for _, name := range cols.names {
		if name == idColumn {
			return fmt.Errorf("gereserveerde querykolom: %s", name)
		}
		columns = append(columns, quoteQueryIdentifier(name))
	}
	cols.names = append([]string{idColumn}, cols.names...)
	rows, err := t.db.QueryContext(ctx, `SELECT `+strings.Join(columns, ",")+` FROM `+quoteQueryIdentifier(t.name)+` `+where+` ORDER BY rowid`, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		if err := ctx.Err(); err != nil {
			return err
		}
		row, ok, err := t.scanRow(cols, rows)
		if err != nil {
			return err
		}
		bounds, valid := row.Geometry.Bounds()
		if !ok || !valid || math.IsNaN(bounds.MinX) || math.IsNaN(bounds.MinY) || math.IsNaN(bounds.MaxX) || math.IsNaN(bounds.MaxY) || math.IsInf(bounds.MinX, 0) || math.IsInf(bounds.MinY, 0) || math.IsInf(bounds.MaxX, 0) || math.IsInf(bounds.MaxY, 0) {
			return fmt.Errorf("onleesbare geometrie in %s; telling afgebroken", t.name)
		}
		row.ID, err = strconv.ParseInt(row.Properties[idColumn], 10, 64)
		if err != nil {
			return err
		}
		delete(row.Properties, idColumn)
		delete(row.Numbers, idColumn)
		fn(row)
	}
	return rows.Err()
}
