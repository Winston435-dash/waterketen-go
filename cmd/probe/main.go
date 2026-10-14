package main

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"waterketen/internal/core"
)

func main() {
	paths := []string{
		"/home/hermes/.local/share/waterketen/processed/delfshaven-riool.gpkg",
		"/home/hermes/.local/share/waterketen/processed/delfshaven-drinkwaterleidingen.gpkg",
		"/home/hermes/.local/share/waterketen/processed/delfshaven-drinkwater.gpkg",
	}
	var layers []core.SceneSource
	for _, p := range paths {
		name := base(p)
		n, feats, err := readFeatures(p, name)
		must(err)
		fmt.Printf("%-32s %6d rijen -> %6d features met geometrie\n", name, n, len(feats))
		layers = append(layers, core.SceneSource{Name: name, Features: feats})
	}

	scene, err := core.NewScene("delfshaven", layers, nil)
	must(err)
	fmt.Printf("scene bounds: %.0f,%.0f → %.0f,%.0f  (%.0f × %.0f m)\n",
		scene.Bounds.MinX, scene.Bounds.MinY, scene.Bounds.MaxX, scene.Bounds.MaxY,
		scene.Bounds.MaxX-scene.Bounds.MinX, scene.Bounds.MaxY-scene.Bounds.MinY)
	fmt.Printf("straten herkend: %d; materialen: %d\n", len(scene.Streets), len(scene.Materials))
	keys := make([]string, 0, len(scene.Materials))
	for k := range scene.Materials {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf("  materiaal %-20s %s\n", k, scene.Materials[k])
	}
}

func readFeatures(path, name string) (int, []core.Feature, error) {
	db, err := core.Open(path)
	if err != nil {
		return 0, nil, err
	}
	defer db.Close()
	tbl, err := core.FeatureTable(db)
	if err != nil {
		return 0, nil, err
	}
	n, err := tbl.Count()
	if err != nil {
		return 0, nil, err
	}
	var feats []core.Feature
	if err := tbl.Each(func(r core.Row) {
		f := core.Feature{
			Layer:      name,
			Position:   len(feats),
			Geometry:   r.Geometry,
			Properties: r.Properties,
			BOB:        bobOf(r),
		}
		if name == "riool" || name == "straten" {
			f.Street = r.Properties["STRAATNAAM"]
			if f.Street == "" {
				f.Street = r.Properties["name"]
			}
		}
		feats = append(feats, f)
	}); err != nil {
		return 0, nil, err
	}
	return n, feats, nil
}

func bobOf(r core.Row) float64 {
	a, okA := r.Numbers["BOB_ACTUEE"]
	b, okB := r.Numbers["BOB_ACTU_1"]
	// 0,00 betekent "nooit gemeten", uitfilteren
	useA := okA && a != 0
	useB := okB && b != 0
	switch {
	case useA && useB:
		return (a + b) / 2
	case useA:
		return a
	case useB:
		return b
	}
	return math.NaN()
}

func base(p string) string {
	name := strings.TrimSuffix(filepath.Base(p), ".gpkg")
	if i := strings.IndexByte(name, '-'); i >= 0 {
		return name[i+1:]
	}
	return name
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "FOUT:", err)
		os.Exit(1)
	}
}
