package core

// Explicit producer for the four IMWA object types used by the water-context
// map. Nothing is fetched by opening the catalog or starting the TUI.

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var imwaMapTypes = []string{"duiker", "gemaal", "sluis", "stuw"}

const imwaCRS = "EPSG:28992"

type IMWAImportResult struct {
	Snapshot string         `json:"snapshot"`
	Counts   map[string]int `json:"counts"`
}

// ImportIMWA publishes a NEW snapshot, retaining the current layers and their
// complete metadata. All downloads, polygon clipping and R-tree checks must
// succeed before current.json is atomically replaced. The old snapshot stays
// usable, including by an already running TUI. GDAL is a producer dependency,
// not a dependency of offline map/query reading.
func ImportIMWA(ctx context.Context, root string) (*IMWAImportResult, error) {
	var source Source
	for _, s := range WiredSources {
		if s.ID == "hwh_kunstwerken_imwa" {
			source = s
			break
		}
	}
	return importIMWA(ctx, root, source.URL, &http.Client{Timeout: 60 * time.Second}, 1000)
}

func importIMWA(ctx context.Context, root, endpoint string, client *http.Client, pageSize int) (*IMWAImportResult, error) {
	if _, err := exec.LookPath("ogr2ogr"); err != nil {
		return nil, fmt.Errorf("IMWA-import vereist ogr2ogr (GDAL): %w", err)
	}
	libraryRoot := filepath.Join(root, "library")
	lock := filepath.Join(libraryRoot, ".imwa-import.lock")
	if err := os.Mkdir(lock, 0700); err != nil {
		return nil, fmt.Errorf("kan IMWA-import niet vergrendelen (andere import actief of geen bibliotheek): %w", err)
	}
	defer os.Remove(lock)
	currentPath := filepath.Join(libraryRoot, "current.json")
	current, err := os.ReadFile(currentPath)
	if err != nil {
		return nil, err
	}
	lib, err := OpenLibrary(root)
	if err != nil {
		return nil, err
	}
	if lib == nil || lib.Manifest.CRS != imwaCRS {
		return nil, fmt.Errorf("IMWA-import vereist een bestaande kaartbibliotheek in %s", imwaCRS)
	}
	area, _, err := queryArea(ctx, lib, DefaultAreaCode)
	if err != nil {
		return nil, fmt.Errorf("gemeentegrens voor IMWA-import: %w", err)
	}
	manifestData, err := os.ReadFile(filepath.Join(lib.Root, "manifest.json"))
	if err != nil {
		return nil, err
	}
	// RawMessage preserves producer fields unknown to Go (notes, raw_path,
	// future schema extensions), unlike marshaling LibraryManifest.
	var manifest map[string]json.RawMessage
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		return nil, err
	}
	var layers map[string]json.RawMessage
	if err := json.Unmarshal(manifest["layers"], &layers); err != nil || layers == nil {
		return nil, fmt.Errorf("manifest heeft geen lagenregister")
	}
	stage, err := os.MkdirTemp(filepath.Join(libraryRoot, "snapshots"), ".imwa-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(stage) // Only our unpublished staging directory.
	now := time.Now().UTC()
	snapshot := now.Format("20060102T150405Z") + "-imwa-" + strings.TrimPrefix(filepath.Base(stage), ".imwa-")
	destination := filepath.Join(libraryRoot, "snapshots", snapshot)
	// Hard links only for immutable, unchanged files. New layer files and the
	// manifest are always written to fresh paths, never through these links.
	if err := linkSnapshot(lib.Root, stage); err != nil {
		return nil, err
	}
	result := &IMWAImportResult{Snapshot: snapshot, Counts: map[string]int{}}
	for _, object := range imwaMapTypes {
		name := "hwh-kunstwerken-" + object
		rawDir := filepath.Join(stage, "imwa-raw")
		if err := os.MkdirAll(rawDir, 0700); err != nil {
			return nil, err
		}
		rawPath := filepath.Join(rawDir, name+".geojson")
		// Unlink a copied previous delivery before writing this fresh one.
		if err := removeImportFile(rawPath); err != nil {
			return nil, err
		}
		count, requestURL, err := downloadIMWA(ctx, client, endpoint, object, area.Bounds, pageSize, rawPath)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		file := name + ".gpkg"
		path := filepath.Join(stage, file)
		if err := removeImportFile(path); err != nil {
			return nil, err
		}
		args := []string{"-f", "GPKG", path, rawPath, "-nln", name, "-a_srs", imwaCRS,
			"-dim", "XY", "-nlt", "GEOMETRY", "-lco", "SPATIAL_INDEX=YES",
			"-clipsrc", filepath.Join(lib.Root, "areas.gpkg"), "-clipsrcwhere", "code='" + DefaultAreaCode + "'"}
		output, err := exec.CommandContext(ctx, "ogr2ogr", args...).CombinedOutput()
		if err != nil {
			return nil, fmt.Errorf("%s: polygonclip/GDAL: %w: %.500s", name, err, output)
		}
		kept, err := validateIMWALayer(path)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		if kept > count {
			return nil, fmt.Errorf("%s: meer objecten na clip dan in bronantwoord", name)
		}
		record := map[string]any{"file": file, "feature_count": kept, "source_feature_count": count,
			"source": map[string]any{"url": endpoint, "license": "CC0", "source_date": nil,
				"type_name":  "waterschappen-kunstwerken-imwa:" + object,
				"fetched_at": now.Format(time.RFC3339), "request_url": requestURL,
				"raw_path": filepath.Join(destination, "imwa-raw", name+".geojson"),
				"notes":    "Het Waterschapshuis via PDOK; bbox-levering geclipt op CBS-gemeente GM0599. Alternatieve registratie, niet optellen bij gemeentelijke/GWSW-gemalen. Brondatum onbekend; ophaaldatum is geen brondatum."}}
		layers[name], err = json.Marshal(record)
		if err != nil {
			return nil, err
		}
		result.Counts[name] = kept
	}
	manifest["layers"], _ = json.Marshal(layers)
	manifest["snapshot"], _ = json.Marshal(snapshot)
	manifest["created_at"], _ = json.Marshal(now.Format(time.RFC3339))
	manifest["parent_snapshot"], _ = json.Marshal(lib.Snapshot)
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(filepath.Join(stage, "manifest.json"), append(data, '\n'), 0600); err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	latest, err := os.ReadFile(currentPath)
	if err != nil || string(latest) != string(current) {
		return nil, fmt.Errorf("huidige snapshot veranderde tijdens import; niets gepubliceerd")
	}
	if err := os.Rename(stage, destination); err != nil {
		return nil, err
	}
	pointer, err := os.CreateTemp(libraryRoot, ".current-imwa-")
	if err != nil {
		return nil, err
	}
	defer os.Remove(pointer.Name())
	err = json.NewEncoder(pointer).Encode(map[string]string{"snapshot": snapshot})
	if err == nil {
		err = pointer.Sync()
	}
	closeErr := pointer.Close()
	if err != nil {
		return nil, err
	}
	if closeErr != nil {
		return nil, closeErr
	}
	if err := os.Rename(pointer.Name(), currentPath); err != nil {
		return nil, err
	}
	return result, nil
}

func removeImportFile(path string) error {
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func linkSnapshot(source, destination string) error {
	return filepath.WalkDir(source, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(source, path)
		if err != nil || rel == "." || rel == "manifest.json" {
			return err
		}
		out := filepath.Join(destination, rel)
		if entry.IsDir() {
			return os.Mkdir(out, 0700)
		}
		if !entry.Type().IsRegular() {
			return fmt.Errorf("snapshot bevat geen gewoon bestand: %s", rel)
		}
		return os.Link(path, out)
	})
}

func validateIMWALayer(path string) (int, error) {
	db, err := Open(path)
	if err != nil {
		return 0, err
	}
	defer db.Close()
	table, err := FeatureTable(db)
	if err != nil {
		return 0, err
	}
	var srid int
	if err := db.QueryRow(`SELECT srs_id FROM gpkg_geometry_columns`).Scan(&srid); err != nil || srid != 28992 {
		return 0, fmt.Errorf("laag heeft geen RD-geometrie")
	}
	if err := table.checkIndexed(quoteQueryIdentifier("rtree_" + table.name + "_" + table.geom)); err != nil {
		return 0, err
	}
	count := 0
	valid := true
	if err := table.Each(func(r Row) {
		bounds, ok := r.Geometry.Bounds()
		valid = valid && ok
		for _, coordinate := range []float64{bounds.MinX, bounds.MinY, bounds.MaxX, bounds.MaxY} {
			valid = valid && !math.IsNaN(coordinate) && !math.IsInf(coordinate, 0)
		}
		count++
	}); err != nil {
		return 0, err
	}
	if !valid {
		return 0, fmt.Errorf("laag bevat ongeldige geometriecoördinaten")
	}
	rows, err := table.Count()
	if err != nil {
		return 0, err
	}
	if count != rows {
		return 0, fmt.Errorf("laag bevat objecten zonder leesbare geometrie")
	}
	return count, nil
}

func imwaRequest(endpoint, object string, bounds BBox) (*url.URL, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("service", "WFS")
	q.Set("version", "2.0.0")
	q.Set("request", "GetFeature")
	q.Set("typeNames", "waterschappen-kunstwerken-imwa:"+object)
	q.Set("srsName", imwaCRS)
	q.Set("bbox", fmt.Sprintf("%.8f,%.8f,%.8f,%.8f,%s", bounds.MinX, bounds.MinY, bounds.MaxX, bounds.MaxY, imwaCRS))
	u.RawQuery = q.Encode()
	return u, nil
}

func imwaResponse(ctx context.Context, client *http.Client, u *url.URL) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "waterketen-go/IMWA-import")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	const maxBytes = 32 << 20
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxBytes+1))
	if err != nil {
		return nil, err
	}
	if len(data) > maxBytes {
		return nil, fmt.Errorf("WFS-antwoord groter dan 32 MiB")
	}
	return data, nil
}

func downloadIMWA(ctx context.Context, client *http.Client, endpoint, object string, bounds BBox, pageSize int, path string) (int, string, error) {
	if pageSize <= 0 {
		return 0, "", fmt.Errorf("ongeldige WFS-paginagrootte")
	}
	u, err := imwaRequest(endpoint, object, bounds)
	if err != nil {
		return 0, "", err
	}
	q := u.Query()
	q.Set("resultType", "hits")
	u.RawQuery = q.Encode()
	data, err := imwaResponse(ctx, client, u)
	if err != nil {
		return 0, "", err
	}
	total, err := countFromText(string(data), "wfs")
	if err != nil || total > 100000 {
		return 0, "", fmt.Errorf("ongeldige of te grote bbox-telling: %d (%v)", total, err)
	}
	q.Del("resultType")
	q.Set("outputFormat", "application/json")
	q.Set("sortBy", "guid")
	q.Set("count", strconv.Itoa(pageSize))
	var features []json.RawMessage
	seen := map[string]bool{}
	for offset := 0; offset < total; {
		q.Set("startIndex", strconv.Itoa(offset))
		u.RawQuery = q.Encode()
		data, err := imwaResponse(ctx, client, u)
		if err != nil {
			return 0, "", err
		}
		var page struct {
			Type     string            `json:"type"`
			Features []json.RawMessage `json:"features"`
			CRS      struct {
				Properties struct {
					Name string `json:"name"`
				} `json:"properties"`
			} `json:"crs"`
		}
		if err := json.Unmarshal(data, &page); err != nil {
			return 0, "", fmt.Errorf("geen GeoJSON: %w", err)
		}
		crs := page.CRS.Properties.Name
		if page.Type != "FeatureCollection" || (crs != "urn:ogc:def:crs:EPSG::28992" && crs != imwaCRS) {
			return 0, "", fmt.Errorf("WFS-antwoord is geen FeatureCollection in RD")
		}
		if len(page.Features) == 0 || len(page.Features) > pageSize || offset+len(page.Features) > total {
			return 0, "", fmt.Errorf("onvolledige of gewijzigde WFS-paginering: %d/%d", offset, total)
		}
		for _, raw := range page.Features {
			var f struct {
				ID       string          `json:"id"`
				Type     string          `json:"type"`
				Geometry json.RawMessage `json:"geometry"`
			}
			if err := json.Unmarshal(raw, &f); err != nil || f.Type != "Feature" || f.ID == "" || seen[f.ID] || len(f.Geometry) == 0 || string(f.Geometry) == "null" {
				return 0, "", fmt.Errorf("dubbel of onleesbaar bronobject: %s", f.ID)
			}
			seen[f.ID] = true
			features = append(features, raw)
		}
		offset += len(page.Features)
	}
	// Do not publish a partial delivery if the source changed size while paging.
	q.Del("startIndex")
	q.Del("outputFormat")
	q.Del("sortBy")
	q.Del("count")
	q.Set("resultType", "hits")
	u.RawQuery = q.Encode()
	data, err = imwaResponse(ctx, client, u)
	if err != nil {
		return 0, "", err
	}
	finalTotal, err := countFromText(string(data), "wfs")
	if err != nil || finalTotal != total {
		return 0, "", fmt.Errorf("WFS-telling veranderde tijdens ophalen: %d → %d (%v)", total, finalTotal, err)
	}
	q.Del("resultType")
	q.Set("outputFormat", "application/json")
	q.Set("sortBy", "guid")
	q.Set("count", strconv.Itoa(pageSize))
	// A valid empty delivery must still produce a GeoPackage, not disappear.
	if features == nil {
		features = []json.RawMessage{}
	}
	collection := map[string]any{"type": "FeatureCollection", "features": features,
		"crs": map[string]any{"type": "name", "properties": map[string]string{"name": "urn:ogc:def:crs:EPSG::28992"}}}
	data, err = json.Marshal(collection)
	if err != nil {
		return 0, "", err
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		return 0, "", err
	}
	q.Del("startIndex")
	u.RawQuery = q.Encode()
	return total, u.String(), nil
}
