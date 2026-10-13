// waterketen is the command-line interface to the managed waterketen store.
// It shares the data directory with the Python CLI (~/.local/share/waterketen)
// and answers the reading questions: sources, status, analyse. Fetching stays
// with the Python CLI except the explicit IMWA library import (see MIGRATION.md).
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"waterketen/internal/core"
	"waterketen/internal/ui"
	"waterketen/internal/ui/components"
)

func main() {
	root := &cobra.Command{
		Use:   "waterketen",
		Short: "Betrouwbare workflow voor Rotterdamse waterketen-geodata.",
	}
	root.AddCommand(discoverCmd(), sourcesCmd(), statusCmd(), analyseCmd(), brailleCmd(), libraryCmd())
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
				Count     *int   `json:"count,omitempty"`
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
					entry.Count = &h.Count
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
				enc := json.NewEncoder(cmd.OutOrStdout())
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
					if live && s.Count != nil && *s.Count >= 0 {
						line += fmt.Sprintf("  ·  %s features  ·  %d ms", ui.Thousands(*s.Count), s.LatencyMS)
					} else if live && s.Error != "" {
						line += fmt.Sprintf("  ·  DOWN: %s", s.Error)
					}
					line += "  ·  " + s.License
					fmt.Fprintln(cmd.OutOrStdout(), components.Badge(style, line))
				} else {
					fmt.Fprintln(cmd.OutOrStdout(), components.Badge(components.StatusIdle, s.Label+" — "+s.Note+"  ·  "+s.License))
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
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(map[string]any{
					"assets":  assets,
					"library": lib,
				})
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Data-root: %s\n\n", root)
			if len(assets) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), components.Muted("Geen managed data. Haal op met: waterketen fetch …"))
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), components.Strong("Managed:"))
				for _, a := range assets {
					exists := ""
					if a.ProcessedExists() {
						exists = "  (op schijf)"
					} else {
						exists = "  (METADATA ZONDER BESTAND)"
					}
					fmt.Fprintf(cmd.OutOrStdout(), "  %s/%-22s  %6s features  %s%s\n",
						a.Area, a.Layer, ui.Thousands(a.FeatureCount), assetDate(a.FetchedAt), exists)
				}
			}
			fmt.Fprintln(cmd.OutOrStdout())
			if lib == nil {
				fmt.Fprintln(cmd.OutOrStdout(), components.Muted("Geen kaartbibliotheek — bouw met Python: waterketen bibliotheek bouw …"))
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), components.Strong(fmt.Sprintf("Bibliotheek-snapshot: %s", lib.Snapshot)))
				fmt.Fprintf(cmd.OutOrStdout(), "  gebieden: %s  ·  CRS: %s\n", ui.Thousands(lib.Manifest.AreaCount), lib.Manifest.CRS)
				names := make([]string, 0, len(lib.Manifest.Layers))
				for n := range lib.Manifest.Layers {
					names = append(names, n)
				}
				sort.Strings(names)
				for _, n := range names {
					rec := lib.Manifest.Layers[n]
					fmt.Fprintf(cmd.OutOrStdout(), "  %-22s %8s features\n", n, ui.Thousands(rec.FeatureCount))
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
			d, err := core.LoadLayer(*match)
			if err != nil {
				return err
			}
			spec := d.Spec
			summary := core.Summarize(d.Features)
			if asJSON {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				out := map[string]any{
					"area": d.Area, "layer": d.Layer,
					"features": summary.Features,
					"lines":    summary.Lines, "polygons": summary.Polygons, "points": summary.Points,
					"total_m": summary.TotalM,
				}
				if summary.BOBMeasured > 0 {
					out["bob_measured"] = summary.BOBMeasured
				}
				return enc.Encode(out)
			}
			label := d.Layer
			if spec != nil {
				label = spec.Label
			}
			fmt.Fprintln(cmd.OutOrStdout(), components.Strong(fmt.Sprintf("%s/%s — %s", d.Area, d.Layer, label)))
			fmt.Fprintf(cmd.OutOrStdout(), "  features: %s (lijnen %s · vlakken %s · punten %s)\n",
				ui.Thousands(summary.Features), ui.Thousands(summary.Lines), ui.Thousands(summary.Polygons), ui.Thousands(summary.Points))
			if spec != nil && !spec.Linear {
				fmt.Fprintf(cmd.OutOrStdout(), "  omtrek totaal: %.2f km\n", summary.TotalM/1000)
			} else if summary.Lines > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "  lengte totaal: %.2f km\n", summary.TotalM/1000)
				fmt.Fprintf(cmd.OutOrStdout(), "  gemiddelde leiding: %.1f m\n", summary.TotalM/float64(summary.Lines))
			}
			if spec != nil && spec.HasBOB {
				if summary.BOBMeasured == 0 {
					fmt.Fprintln(cmd.OutOrStdout(), components.Error("  geen BOB gemeten (0,00 = geen meting)"))
				} else {
					fmt.Fprintf(cmd.OutOrStdout(), "  BOB: gemeten op %s van %s (%.1f%%)\n",
						ui.Thousands(summary.BOBMeasured), ui.Thousands(summary.Features),
						100*float64(summary.BOBMeasured)/float64(summary.Features))
					fmt.Fprintf(cmd.OutOrStdout(), "  BOB: min %.2f · mediaan %.2f · max %.2f\n",
						summary.BOBMin, summary.BOBMedian, summary.BOBMax)
				}
			}
			if summary.Streets > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "  straten met buizen: %s\n", ui.Thousands(summary.Streets))
			}
			return nil
		},
	}
	c.Flags().String("gebied", "", "gebied (bijv. delfshaven)")
	c.Flags().String("laag", "", "laag (riool|drinkwater|drinkwaterleidingen)")
	c.Flags().BoolVar(&asJSON, "json", false, "machine-leesbaar")
	c.Flags().String("data-root", "", "overschrijf $WATERKETEN_DATA_DIR")
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
			assets, err := core.LoadAssets(root)
			if err != nil {
				return err
			}
			for _, a := range assets {
				if a.Area == gebied && a.Layer == laag {
					d, err := core.LoadLayer(a)
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
						fmt.Fprintln(cmd.OutOrStdout(), b.String())
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
	c.Flags().String("data-root", "", "overschrijf $WATERKETEN_DATA_DIR")
	c.MarkFlagRequired("gebied")
	c.MarkFlagRequired("laag")
	return c
}

// --- helpers -----------------------------------------------------------------

func assetDate(value string) string {
	if date, err := time.Parse(time.RFC3339, value); err == nil {
		return date.Format("2006-01-02")
	}
	return "datum onbekend"
}
