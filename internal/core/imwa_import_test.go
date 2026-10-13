package core

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

const imwaTestFeatures = `[
	{"type":"Feature","id":"test.1","properties":{"guid":"1","naam":"Binnen"},"geometry":{"type":"Point","coordinates":[1,1]}},
	{"type":"Feature","id":"test.2","properties":{"guid":"2","naam":"Alleen bbox"},"geometry":{"type":"Point","coordinates":[9,9]}},
	{"type":"Feature","id":"test.3","properties":{"guid":"3","naam":"Duiker"},"geometry":{"type":"LineString","coordinates":[[1,2],[2,2]]}}
]`

func imwaTestServer(t *testing.T, failType string) *httptest.Server {
	t.Helper()
	var features []json.RawMessage
	if err := json.Unmarshal([]byte(imwaTestFeatures), &features); err != nil {
		t.Fatal(err)
	}
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("service") != "WFS" || q.Get("srsName") != imwaCRS || !strings.HasSuffix(q.Get("bbox"), ",EPSG:28992") {
			t.Error("missing explicit WFS/CRS/bbox", r.URL)
		}
		if strings.HasSuffix(q.Get("typeNames"), ":"+failType) && failType != "" {
			http.Error(w, "failed", http.StatusBadGateway)
			return
		}
		if q.Get("resultType") == "hits" {
			fmt.Fprint(w, `<wfs:FeatureCollection numberMatched="3"/>`)
			return
		}
		if q.Get("sortBy") != "guid" || q.Get("outputFormat") != "application/json" {
			t.Error("missing stable ordering/GeoJSON", r.URL)
		}
		offset, _ := strconv.Atoi(q.Get("startIndex"))
		count, _ := strconv.Atoi(q.Get("count"))
		if offset >= len(features) || count < 1 {
			http.Error(w, "bad page", 400)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"type": "FeatureCollection", "features": features[offset:min(offset+count, len(features))],
			"crs": map[string]any{"properties": map[string]string{"name": "urn:ogc:def:crs:EPSG::28992"}}})
	}))
	t.Cleanup(s.Close)
	return s
}

func TestIMWADownloadPagesWithoutChangingGeometry(t *testing.T) {
	s := imwaTestServer(t, "")
	path := filepath.Join(t.TempDir(), "raw.geojson")
	count, request, err := downloadIMWA(context.Background(), s.Client(), s.URL, "duiker", BBox{MaxX: 10, MaxY: 10}, 2, path)
	if err != nil {
		t.Fatal(err)
	}
	if count != 3 || !strings.Contains(request, "typeNames=waterschappen-kunstwerken-imwa%3Aduiker") {
		t.Fatalf("count/request: %d %s", count, request)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var collection struct{ Features []json.RawMessage }
	json.Unmarshal(data, &collection)
	var expected []json.RawMessage
	json.Unmarshal([]byte(imwaTestFeatures), &expected)
	for i, raw := range collection.Features {
		var got, want any
		json.Unmarshal(raw, &got)
		json.Unmarshal(expected[i], &want)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("source geometry/attributes changed: %s", raw)
		}
	}
	if len(collection.Features) != 3 {
		t.Fatal("lost features")
	}
}

func TestIMWADownloadRejectsBrokenDeliveries(t *testing.T) {
	for _, tc := range []struct{ name, response string }{
		{"exception", `<ExceptionReport/>`},
		{"wrong-crs", `{"type":"FeatureCollection","features":[],"crs":{"properties":{"name":"EPSG:4326"}}}`},
		{"short", `{"type":"FeatureCollection","features":[],"crs":{"properties":{"name":"EPSG:28992"}}}`},
		{"duplicate", `{"type":"FeatureCollection","features":[{"type":"Feature","id":"a","geometry":{"type":"Point","coordinates":[1,1]}}],"crs":{"properties":{"name":"EPSG:28992"}}}`},
		{"null-geometry", `{"type":"FeatureCollection","features":[{"type":"Feature","id":"a","geometry":null}],"crs":{"properties":{"name":"EPSG:28992"}}}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Query().Get("resultType") == "hits" {
					fmt.Fprint(w, `<FeatureCollection numberMatched="3"/>`)
				} else {
					fmt.Fprint(w, tc.response)
				}
			}))
			defer s.Close()
			path := filepath.Join(t.TempDir(), "raw.geojson")
			if _, _, err := downloadIMWA(context.Background(), s.Client(), s.URL, "duiker", BBox{MaxX: 10, MaxY: 10}, 2, path); err == nil {
				t.Fatal("broken delivery accepted")
			}
			if localFileExists(path) {
				t.Fatal("broken delivery written")
			}
		})
	}
}

func imwaImportLibrary(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("ogr2ogr"); err != nil {
		t.Skip("GDAL producer dependency not installed")
	}
	root := t.TempDir()
	snapshot := filepath.Join(root, "library", "snapshots", "original")
	if err := os.MkdirAll(snapshot, 0700); err != nil {
		t.Fatal(err)
	}
	area := `{"type":"FeatureCollection","features":[{"type":"Feature","properties":{"code":"GM0599","name":"Rotterdam","kind":"gemeente"},"geometry":{"type":"Polygon","coordinates":[[[0,0],[10,0],[0,10],[0,0]]]}}]}`
	raw := filepath.Join(root, "area.geojson")
	os.WriteFile(raw, []byte(area), 0600)
	output, err := exec.Command("ogr2ogr", "-f", "GPKG", filepath.Join(snapshot, "areas.gpkg"), raw, "-a_srs", imwaCRS).CombinedOutput()
	if err != nil {
		t.Fatalf("GDAL area fixture: %v: %s", err, output)
	}
	os.WriteFile(filepath.Join(snapshot, "unchanged.gpkg"), []byte("immutable fixture"), 0600)
	manifest := `{"schema":1,"snapshot":"original","crs":"EPSG:28992","created_at":"2026-01-01T00:00:00Z","area_count":1,"extension":{"keep":true},"layers":{"unchanged":{"file":"unchanged.gpkg","feature_count":17,"source":{"license":"ODbL","notes":"preserve","raw_path":"original/raw","source_date":"2025-01-01"}}}}`
	os.WriteFile(filepath.Join(snapshot, "manifest.json"), []byte(manifest), 0600)
	os.WriteFile(filepath.Join(root, "library", "current.json"), []byte(`{"snapshot":"original"}`), 0600)
	return root
}

func TestIMWAImportPublishesClippedMapQueryAndCatalogLayers(t *testing.T) {
	root := imwaImportLibrary(t)
	s := imwaTestServer(t, "")
	result, err := importIMWA(context.Background(), root, s.URL, s.Client(), 2)
	if err != nil {
		t.Fatal(err)
	}
	lib, err := OpenLibrary(root)
	if err != nil {
		t.Fatal(err)
	}
	if lib.Snapshot == "original" || lib.Snapshot != result.Snapshot {
		t.Fatal("snapshot not published")
	}
	manifest, _ := os.ReadFile(filepath.Join(lib.Root, "manifest.json"))
	if !strings.Contains(string(manifest), `"notes": "preserve"`) || !strings.Contains(string(manifest), `"extension"`) {
		t.Fatal("original metadata lost")
	}
	old, _ := os.ReadFile(filepath.Join(root, "library", "snapshots", "original", "manifest.json"))
	if strings.Contains(string(old), "hwh-kunstwerken") {
		t.Fatal("original snapshot modified")
	}
	for _, object := range imwaMapTypes {
		name := "hwh-kunstwerken-" + object
		if result.Counts[name] != 2 || lib.Manifest.Layers[name].SourceFeatureCount != 3 {
			t.Fatal("bbox-only object survived polygon clip", result)
		}
		query, err := RunAreaQuery(context.Background(), lib, DefaultAreaCode, []string{name})
		if err != nil || query.Layers[0].Status != QueryOK || query.Layers[0].Count != 2 {
			t.Fatalf("import unavailable in query: %+v %v", query, err)
		}
		atlas, err := NewLibraryAtlas(lib, "")
		if err != nil {
			t.Fatal(err)
		}
		scene, err := atlas.Load(FitView(atlas.Area.Bounds), 40, 20, []string{name}, 100)
		if err != nil || len(scene.Layers[name]) != 2 {
			t.Fatalf("import unavailable on map: %+v %v", scene, err)
		}
		for _, layer := range CatalogWithLocal(nil, lib).Layers {
			if layer.ID == name && (!layer.Available() || layer.SourceID != "hwh_kunstwerken_imwa") {
				t.Fatalf("catalog not joined: %+v", layer)
			}
		}
	}
	// A second import must replace staging links, not overwrite old raw/GPKG files.
	previous := lib.Root
	before, _ := os.ReadFile(lib.LayerPath("hwh-kunstwerken-duiker"))
	if _, err := importIMWA(context.Background(), root, s.URL, s.Client(), 2); err != nil {
		t.Fatal("repeat import:", err)
	}
	after, _ := os.ReadFile(filepath.Join(previous, "hwh-kunstwerken-duiker.gpkg"))
	if !reflect.DeepEqual(before, after) {
		t.Fatal("repeat import mutated a published layer")
	}
}

func TestIMWAImportFailureLeavesCurrentSnapshotUntouched(t *testing.T) {
	root := imwaImportLibrary(t)
	s := imwaTestServer(t, "sluis")
	if _, err := importIMWA(context.Background(), root, s.URL, s.Client(), 2); err == nil {
		t.Fatal("partial import published")
	}
	lib, err := OpenLibrary(root)
	if err != nil || lib.Snapshot != "original" {
		t.Fatalf("current snapshot changed: %+v %v", lib, err)
	}
	entries, _ := os.ReadDir(filepath.Join(root, "library", "snapshots"))
	if len(entries) != 1 {
		t.Fatal("failed staging snapshot left behind")
	}
}

func TestIMWAImportKeepsValidEmptyLayers(t *testing.T) {
	root := imwaImportLibrary(t)
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `<FeatureCollection numberMatched="0"/>`)
	}))
	defer s.Close()
	result, err := importIMWA(context.Background(), root, s.URL, s.Client(), 2)
	if err != nil {
		t.Fatal(err)
	}
	lib, err := OpenLibrary(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, object := range imwaMapTypes {
		name := "hwh-kunstwerken-" + object
		if _, ok := result.Counts[name]; !ok || !localFileExists(lib.LayerPath(name)) {
			t.Fatal("empty layer vanished", name)
		}
	}
}

func TestIMWAImportCancellationAndChangedCountsDoNotPublish(t *testing.T) {
	root := imwaImportLibrary(t)
	s := imwaTestServer(t, "")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := importIMWA(ctx, root, s.URL, s.Client(), 2); err == nil {
		t.Fatal("canceled import published")
	}
	hits := 0
	changing := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("resultType") == "hits" {
			fmt.Fprintf(w, `<FeatureCollection numberMatched="%d"/>`, hits)
			hits++
		}
	}))
	defer changing.Close()
	if _, _, err := downloadIMWA(context.Background(), changing.Client(), changing.URL, "duiker", BBox{MaxX: 10, MaxY: 10}, 2, filepath.Join(t.TempDir(), "raw.geojson")); err == nil {
		t.Fatal("changed source count accepted")
	}
	lib, err := OpenLibrary(root)
	if err != nil || lib.Snapshot != "original" {
		t.Fatalf("failed import published: %+v %v", lib, err)
	}
}
