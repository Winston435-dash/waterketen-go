// waterketen is the command-line interface to the managed waterketen store.
// It shares the data directory with the Python CLI (~/.local/share/waterketen)
// and answers the reading questions: sources, status, analyse. Fetching stays
// with the Python CLI for now (see MIGRATION.md).
package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"waterketen/internal/core"
	"waterketen/internal/ui/components"
	"waterketen/internal/ui/screens"
)

func main() {
	root := &cobra.Command{
		Use:   "waterketen",
		Short: "Betrouwbare workflow voor Rotterdamse waterketen-geodata.",
	}
	root.AddCommand(sourcesCmd(), statusCmd(), analyseCmd(), brailleCmd())
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "waterketen:", err)
		os.Exit(1)
	}
}

func dataRoot(cmd *cobra.Command) string {
	if v, _ := cmd.Flags().GetString("data-root"); v != "" {
		return v
	}
	return core.DataRoot()
}

// --- sources -----------------------------------------------------------------

func sourcesCmd() *cobra.Command {
	var asJSON, live bool
	c := &cobra.Command{
		Use:   "sources",
		Short: "toon geregistreerde bronnen en (optioneel) hun live status",
		RunE: func(cmd *cobra.Command, args []string) error {
			type sourceOut struct {
				core.Source
				Status    string `json:"status,omitempty"`
				Count     int    `json:"count,omitempty"`
				LatencyMS int    `json:"latency_ms,omitempty"`
				Error     string `json:"error,omitempty"`
				ProbeKind string `json:"probe_kind"`
				Wired     bool   `json:"wired"`
				Note      string `json:"note,omitempty"`
			}
			var out []sourceOut
			health := map[string]core.SourceHealth{}
			if live {
				for _, h := range core.ProbeAll() {
					health[h.SourceID] = h
				}
			}
			for _, s := range core.WiredSources {
				entry := sourceOut{Source: s, Wired: true, ProbeKind: s.ProbeKind}
				if h, ok := health[s.ID]; ok {
					entry.Status = string(h.Status)
					entry.Count = h.Count
					entry.LatencyMS = h.LatencyMS
					entry.Error = h.Error
				}
				out = append(out, entry)
			}
			for _, u := range core.UnwiredSources {
				out = append(out, sourceOut{
					Source: core.Source{ID: u.ID, Label: u.Label, URL: u.URL, License: u.License},
					Wired:  false, Note: u.Note,
					Status: string(core.HealthNotWired),
				})
			}
			if asJSON {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(out)
			}
			for _, s := range out {
				if s.Wired {
					style := components.StatusOK
					if !live {
						style = components.StatusIdle
					} else if s.Status != string(core.HealthUp) {
						style = components.StatusDown
					}
					line := s.Label
					if live && s.Count >= 0 {
						line += fmt.Sprintf("  ·  %s features  ·  %d ms", thousands(s.Count), s.LatencyMS)
					} else if live && s.Error != "" {
						line += fmt.Sprintf("  ·  DOWN: %s", s.Error)
					}
					line += "  ·  " + s.License
					fmt.Println(components.Badge(style, line))
				} else {
					fmt.Println(components.Badge(components.StatusIdle, s.Label+" — "+s.Note+"  ·  "+s.License))
				}
			}
			return nil
		},
	}
	c.Flags().BoolVar(&asJSON, "json", false, "machine-leesbare output")
	c.Flags().BoolVar(&live, "live", false, "probeer elke bron echt (netwerk)")
	return c
}

// --- status ------------------------------------------------------------------

func statusCmd() *cobra.Command {
	var asJSON bool
	c := &cobra.Command{
		Use:   "status",
		Short: "toon beheerde data (managed) en herkomst — offline",
		RunE: func(cmd *cobra.Command, args []string) error {
			root := dataRoot(cmd)
			assets, err := core.LoadAssets(root)
			if err != nil {
				return err
			}
			lib, err := core.OpenLibrary(root)
			if err != nil {
				return err
			}
			if asJSON {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(map[string]any{
					"assets":  assets,
					"library": lib,
				})
			}
			fmt.Printf("Data-root: %s\n\n", root)
			if len(assets) == 0 {
				fmt.Println(components.Muted("Geen managed data. Haal op met: waterketen fetch …"))
			} else {
				fmt.Println(components.Strong("Managed:"))
				for _, a := range assets {
					exists := ""
					if a.ProcessedExists() {
						exists = "  (op schijf)"
					} else {
						exists = "  (METADATA ZONDER BESTAND)"
					}
					fmt.Printf("  %s/%-22s  %6s features  %s%s\n",
						a.Area, a.Layer, thousands(a.FeatureCount), a.FetchedAt[:10], exists)
				}
			}
			fmt.Println()
			if lib == nil {
				fmt.Println(components.Muted("Geen kaartbibliotheek — bouw met Python: waterketen bibliotheek bouw …"))
			} else {
				fmt.Println(components.Strong(fmt.Sprintf("Bibliotheek-snapshot: %s", lib.Snapshot)))
				fmt.Printf("  gebieden: %s  ·  CRS: %s\n", thousands(lib.Manifest.AreaCount), lib.Manifest.CRS)
				names := make([]string, 0, len(lib.Manifest.Layers))
				for n := range lib.Manifest.Layers {
					names = append(names, n)
				}
				sortStrings(names)
				for _, n := range names {
					rec := lib.Manifest.Layers[n]
					fmt.Printf("  %-22s %8s features\n", n, thousands(rec.FeatureCount))
				}
			}
			return nil
		},
	}
	c.Flags().BoolVar(&asJSON, "json", false, "machine-leesbare output")
	c.Flags().String("data-root", "", "overschrijf $WATERKETEN_DATA_DIR")
	return c
}

// --- analyse -----------------------------------------------------------------

func analyseCmd() *cobra.Command {
	var asJSON bool
	c := &cobra.Command{
		Use:   "analyse --gebied GEBIED --laag LAAG",
		Short: "analyseer een managed laag: netwerk, lengte, diepte",
		RunE: func(cmd *cobra.Command, args []string) error {
			gebied, _ := cmd.Flags().GetString("gebied")
			laag, _ := cmd.Flags().GetString("laag")
			root := dataRoot(cmd)
			assets, err := core.LoadAssets(root)
			if err != nil {
				return err
			}
			var match *core.AssetMeta
			for i := range assets {
				if assets[i].Area == gebied && assets[i].Layer == laag {
					match = &assets[i]
					break
				}
			}
			if match == nil {
				return fmt.Errorf("geen managed data voor %s/%s", gebied, laag)
			}
			d, err := screens.LoadLayer(*match)
			if err != nil {
				return err
			}
			spec := d.Spec
			var totalLen float64
			var lines, polys, pts int
			var bobs []float64
			streets := map[string]int{}
			for _, f := range d.Features {
				switch f.Geometry.Kind {
				case "linestring", "multilinestring":
					lines++
				case "polygon", "multipolygon":
					polys++
				default:
					pts++
				}
				totalLen += f.Geometry.Length()
				if !math.IsNaN(f.BOB) {
					bobs = append(bobs, f.BOB)
				}
				if f.Street != "" {
					streets[f.Street]++
				}
			}
			if asJSON {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				out := map[string]any{
					"area": d.Area, "layer": d.Layer,
					"features": len(d.Features),
					"lines":    lines, "polygons": polys, "points": pts,
					"total_m": totalLen,
				}
				if len(bobs) > 0 {
					out["bob_measured"] = len(bobs)
				}
				return enc.Encode(out)
			}
			label := d.Layer
			if spec != nil {
				label = spec.Label
			}
			fmt.Println(components.Strong(fmt.Sprintf("%s/%s — %s", d.Area, d.Layer, label)))
			fmt.Printf("  features: %s (lijnen %s · vlakken %s · punten %s)\n",
				thousands(len(d.Features)), thousands(lines), thousands(polys), thousands(pts))
			if spec != nil && !spec.Linear {
				fmt.Printf("  omtrek totaal: %.2f km\n", totalLen/1000)
			} else if lines > 0 {
				fmt.Printf("  lengte totaal: %.2f km\n", totalLen/1000)
				fmt.Printf("  gemiddelde leiding: %.1f m\n", totalLen/float64(lines))
			}
			if spec != nil && spec.HasBOB {
				if len(bobs) == 0 {
					fmt.Println(components.Error("  geen BOB gemeten (0,00 = geen meting)"))
				} else {
					lo, hi, sum := bobs[0], bobs[0], 0.0
					for _, v := range bobs {
						lo = math.Min(lo, v)
						hi = math.Max(hi, v)
						sum += v
					}
					fmt.Printf("  BOB: gemeten op %s van %s (%.1f%%)\n",
						thousands(len(bobs)), thousands(len(d.Features)),
						100*float64(len(bobs))/float64(len(d.Features)))
					fmt.Printf("  BOB: min %.2f · mediaan %s · max %.2f\n",
						lo, medianOf(bobs), hi)
				}
			}
			if len(streets) > 0 {
				fmt.Printf("  straten met buizen: %s\n", thousands(len(streets)))
			}
			return nil
		},
	}
	c.Flags().String("gebied", "", "gebied (bijv. delfshaven)")
	c.Flags().String("laag", "", "laag (riool|drinkwater|drinkwaterleidingen)")
	c.Flags().BoolVar(&asJSON, "json", false, "machine-leesbaar")
	c.MarkFlagRequired("gebied")
	c.MarkFlagRequired("laag")
	return c
}

// --- braille -----------------------------------------------------------------

func brailleCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "braille --gebied GEBIED --laag LAAG",
		Short: "teken managed data als braillekaart in de terminal",
		RunE: func(cmd *cobra.Command, args []string) error {
			gebied, _ := cmd.Flags().GetString("gebied")
			laag, _ := cmd.Flags().GetString("laag")
			w, _ := cmd.Flags().GetInt("breedte")
			h, _ := cmd.Flags().GetInt("hoogte")
			root := dataRoot(cmd)
			assets, _ := core.LoadAssets(root)
			for _, a := range assets {
				if a.Area == gebied && a.Layer == laag {
					d, err := screens.LoadLayer(a)
					if err != nil {
						return err
					}
					scene, err := core.NewScene(d.Area, []core.SceneSource{{Name: laag, Features: d.Features}}, nil)
					if err != nil {
						return err
					}
					view := core.FitView(scene.Bounds)
					r := core.RenderScene(scene, view, w, h, laag, true, "netwerk", true, false, nil, "", nil, true)
					// Plain braille out, no colours — the Python tool's contract.
					for _, row := range r.Cells {
						var b strings.Builder
						for _, c := range row {
							b.WriteRune(c.Rune)
						}
						fmt.Println(b.String())
					}
					return nil
				}
			}
			return fmt.Errorf("geen managed data voor %s/%s", gebied, laag)
		},
	}
	c.Flags().String("gebied", "", "gebied")
	c.Flags().String("laag", "", "laag")
	c.Flags().Int("breedte", 110, "kolommen")
	c.Flags().Int("hoogte", 38, "regels")
	c.MarkFlagRequired("gebied")
	c.MarkFlagRequired("laag")
	return c
}

// --- helpers -----------------------------------------------------------------

func thousands(n int) string {
	s := strconv.Itoa(n)
	var out []byte
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, '.')
		}
		out = append(out, byte(c))
	}
	return string(out)
}

func medianOf(v []float64) string {
	if len(v) == 0 {
		return "—"
	}
	s := append([]float64(nil), v...)
	// insertion sort — plenty fast at our sizes
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j-1] > s[j]; j-- {
			s[j-1], s[j] = s[j], s[j-1]
		}
	}
	m := s[len(s)/2]
	if len(s)%2 == 0 {
		m = (s[len(s)/2-1] + s[len(s)/2]) / 2
	}
	return fmt.Sprintf("%.2f", m)
}

func sortStrings(v []string) {
	for i := 1; i < len(v); i++ {
		for j := i; j > 0 && v[j-1] > v[j]; j-- {
			v[j-1], v[j] = v[j], v[j-1]
		}
	}
}
