package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// DataRoot is the managed waterketen data directory, shared with the Python
// CLI (~/.local/share/waterketen). Both binaries read the same store.
func DataRoot() string {
	if root := os.Getenv("WATERKETEN_DATA_DIR"); root != "" {
		return root
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", "waterketen")
}

// AssetMeta mirrors src/waterketen_cli/pipeline.py's metadata JSON for one
// managed (area, layer) asset. Fields stay optional: the writer own this
// format, the reader tolerates what it does not know.
type AssetMeta struct {
	Status             string `json:"status"`
	Area               string `json:"area"`
	AreaLabel          string `json:"area_label"`
	Layer              string `json:"layer"`
	LayerLabel         string `json:"layer_label"`
	Source             string `json:"source"`
	SourceKind         string `json:"source_kind"`
	SourceURL          string `json:"source_url"`
	SourceLicense      string `json:"source_license"`
	SourceTypeName     string `json:"source_type_name"`
	GeometryNote       string `json:"geometry_note"`
	FetchedAt          string `json:"fetched_at"`
	CRS                string `json:"crs"`
	SourceFeatureCount int    `json:"source_feature_count"`
	FeatureCount       int    `json:"feature_count"`
	ProcessedPath      string `json:"processed_path"`

	// Derived on load, not in JSON.
	path string
}

// LayerSpec describes a registered layer (static, mirrors LAYER_SPECS).
type LayerSpec struct {
	Slug      string
	Label     string
	SourceID  string
	Linear    bool   // false for polygon layers ("omtrek", nooit "gemiddeld N m per buis")
	HasBOB    bool   // true when the layer carries BobBeginpuntLeiding/BobEindpuntLeiding
	Geometry  string // "lijnen" | "vlakken"
	SourceURL string
}

// LayerSpecs is the static layer registry (kept in sync with pipeline.py).
var LayerSpecs = []LayerSpec{
	{Slug: "riool", Label: "Rioolbuizen", SourceID: "gws_netwerk", Linear: true, HasBOB: true,
		Geometry: "lijnen", SourceURL: "https://geodata.gwsw.nl/geoserver/ows"},
	{Slug: "drinkwater", Label: "Drinkwater verzorgingsgebied", SourceID: "vewin",
		Linear: false, HasBOB: false, Geometry: "vlakken",
		SourceURL: "https://services5.arcgis.com/ZcRj7hEl3ya9tYHV/arcgis/rest/services/Waterbedrijven_Nederland/FeatureServer/0"},
	{Slug: "drinkwaterleidingen", Label: "Drinkwaterleidingen (Evides)", SourceID: "evides",
		Linear: true, HasBOB: false, Geometry: "lijnen",
		SourceURL: "https://services.arcgis.com/zP1tGdLpGvt2qNJ6/arcgis/rest/services/Kans_opwarming_drinkwater_Rotterdam/FeatureServer/0"},
}

// LayerSpecFor returns the spec for a slug, or nil.
func LayerSpecFor(slug string) *LayerSpec {
	for i := range LayerSpecs {
		if LayerSpecs[i].Slug == slug {
			return &LayerSpecs[i]
		}
	}
	return nil
}

// LoadAssets scans $DATA_ROOT/metadata/*.json. A corrupt or non-asset JSON is
// skipped silently — one bad file must not take the whole panel down.
func LoadAssets(root string) ([]AssetMeta, error) {
	entries, err := os.ReadDir(filepath.Join(root, "metadata"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []AssetMeta
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		path := filepath.Join(root, "metadata", e.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var meta AssetMeta
		if err := json.Unmarshal(data, &meta); err != nil {
			continue
		}
		if meta.Area == "" || meta.Layer == "" {
			continue
		}
		meta.path = path
		out = append(out, meta)
	}
	prio := map[string]int{"riool": 0, "drinkwaterleidingen": 1, "drinkwater": 2}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Area != out[j].Area {
			return out[i].Area < out[j].Area
		}
		if pi, pj := prio[out[i].Layer], prio[out[j].Layer]; pi != pj {
			return pi < pj
		}
		return out[i].Layer < out[j].Layer
	})
	return out, nil
}

// ProcessedExists reports whether the asset's data file is actually on disk.
func (a AssetMeta) ProcessedExists() bool {
	if a.ProcessedPath == "" {
		return false
	}
	_, err := os.Stat(a.ProcessedPath)
	return err == nil
}

// --- Library (the snapshot store under $DATA_ROOT/library) ------------------

// LibraryManifest is the manifest.json of one library snapshot.
type LibraryManifest struct {
	Schema         int    `json:"schema"`
	Snapshot       string `json:"snapshot"`
	CRS            string `json:"crs"`
	CreatedAt      string `json:"created_at"`
	BoundarySource string `json:"boundary_source"`
	AreaCount      int    `json:"area_count"`
	Layers         map[string]struct {
		File               string `json:"file"`
		FeatureCount       int    `json:"feature_count"`
		SourceFeatureCount int    `json:"source_feature_count"`
		Source             struct {
			URL     string `json:"url"`
			License string `json:"license"`
			Date    string `json:"source_date"`
		} `json:"source"`
	} `json:"layers"`
}

// Library points at the currently published snapshot.
type Library struct {
	Root     string
	Snapshot string
	Manifest LibraryManifest
}

// OpenLibrary resolves current.json → snapshot → manifest.json. Returns nil
// (no error) when no library exists — the TUI then falls back to the managed
// assets, exactly like the Python app.
func OpenLibrary(root string) (*Library, error) {
	currentPath := filepath.Join(root, "library", "current.json")
	data, err := os.ReadFile(currentPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var current struct {
		Snapshot string `json:"snapshot"`
	}
	if err := json.Unmarshal(data, &current); err != nil {
		return nil, fmt.Errorf("lees current.json: %w", err)
	}
	snapRoot := filepath.Join(root, "library", "snapshots", current.Snapshot)
	manifestData, err := os.ReadFile(filepath.Join(snapRoot, "manifest.json"))
	if err != nil {
		return nil, fmt.Errorf("lees manifest: %w", err)
	}
	var manifest LibraryManifest
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		return nil, fmt.Errorf("manifest corrupt: %w", err)
	}
	return &Library{Root: snapRoot, Snapshot: current.Snapshot, Manifest: manifest}, nil
}

// LayerPath returns the .gpkg path for a library layer, "" when absent.
func (l *Library) LayerPath(name string) string {
	entry, ok := l.Manifest.Layers[name]
	if !ok {
		return ""
	}
	return filepath.Join(l.Root, entry.File)
}

// HasLayers reports which of `names` exist in the manifest, preserving order.
func (l *Library) HasLayers(names ...string) []string {
	var out []string
	for _, n := range names {
		if _, ok := l.Manifest.Layers[n]; ok {
			out = append(out, n)
		}
	}
	return out
}
