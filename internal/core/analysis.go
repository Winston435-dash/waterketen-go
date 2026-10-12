package core

import (
	"fmt"
	"math"
	"sort"
)

// LayerData is a loaded managed layer, shared by the CLI and the TUI.
type LayerData struct {
	Area, Layer string
	Spec        *LayerSpec
	Features    []Feature
}

// LoadLayer reads one managed GeoPackage. Never return partial data as success
// when SQLite fails partway through a read.
func LoadLayer(a AssetMeta) (*LayerData, error) {
	if !a.ProcessedExists() {
		return nil, fmt.Errorf("geen processed data voor %s/%s", a.Area, a.Layer)
	}
	db, err := Open(a.ProcessedPath)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	table, err := FeatureTable(db)
	if err != nil {
		return nil, err
	}
	data := &LayerData{Area: a.Area, Layer: a.Layer, Spec: LayerSpecFor(a.Layer)}
	err = table.Each(func(row Row) {
		data.Features = append(data.Features, Feature{
			Layer: a.Layer, Position: len(data.Features), Geometry: row.Geometry,
			Properties: row.Properties, Street: row.Properties["STRAATNAAM"], BOB: bobOf(row.Numbers),
		})
	})
	if err != nil {
		return nil, fmt.Errorf("lees %s/%s: %w", a.Area, a.Layer, err)
	}
	return data, nil
}

// LoadManagedLayers keeps readable layers while retaining diagnostics for every
// failed layer, so the UI can distinguish missing/unreadable data from no assets.
func LoadManagedLayers(assets []AssetMeta) ([]*LayerData, []error) {
	var data []*LayerData
	var failures []error
	for _, asset := range assets {
		layer, err := LoadLayer(asset)
		if err != nil {
			failures = append(failures, err)
			continue
		}
		data = append(data, layer)
	}
	return data, failures
}

type LayerSummary struct {
	Features, Lines, Polygons, Points, Other, Streets int
	TotalM                                            float64
	BOBMeasured                                       int
	BOBMin, BOBMedian, BOBMax                         float64
}

// Summarize computes domain statistics once; callers own formatting and styling.
func Summarize(features []Feature) LayerSummary {
	s := LayerSummary{Features: len(features)}
	streets := map[string]bool{}
	var bobs []float64
	for _, feature := range features {
		switch feature.Geometry.Kind {
		case "linestring", "multilinestring":
			s.Lines++
		case "polygon", "multipolygon":
			s.Polygons++
		case "point", "multipoint":
			s.Points++
		default:
			s.Other++
		}
		s.TotalM += feature.Geometry.Length()
		if !math.IsNaN(feature.BOB) && !math.IsInf(feature.BOB, 0) {
			bobs = append(bobs, feature.BOB)
		}
		if feature.Street != "" {
			streets[feature.Street] = true
		}
	}
	s.Streets, s.BOBMeasured = len(streets), len(bobs)
	if len(bobs) > 0 {
		sort.Float64s(bobs)
		s.BOBMin, s.BOBMax = bobs[0], bobs[len(bobs)-1]
		s.BOBMedian = bobs[len(bobs)/2]
		if len(bobs)%2 == 0 {
			s.BOBMedian = (bobs[len(bobs)/2-1] + s.BOBMedian) / 2
		}
	}
	return s
}
