package main

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"waterketen/internal/core"
)

// discoverCmd is an offline reader of the same explanations the TUI renders.
func discoverCmd() *cobra.Command {
	var question, source, layer, term, overview string
	var technical, asJSON, markdown bool
	cmd := &cobra.Command{
		Use:   "ontdek",
		Short: "begrijp bronnen, lagen en begrippen vanuit je vraag — offline",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			catalog, err := core.LoadCatalog(dataRoot(cmd))
			if err != nil {
				return err
			}
			output := newCommandOutput(cmd.OutOrStdout())
			if markdown {
				fmt.Fprintln(output, "# Waterketen ontdekken\n\nGegenereerd uit de gedeelde CLI-/TUI-catalogus. Deze uitleg beschrijft de geregistreerde bronnen en bekende lagen, niet alle lagen van elke externe dienst. Lezen doet geen download of live controle.\n\nAanbieder → API-dienst (toegang) → laag (verzameling) → feature (kaartobject) → eigenschap (bijvoorbeeld materiaal). Een GeoPackage is een lokaal bestand, geen API.")
				for section := core.CatalogQuestions; section <= core.CatalogTerms; section++ {
					fmt.Fprintf(output, "\n## %s\n", core.CatalogSections[section])
					for _, entry := range catalog.Entries(section) {
						fmt.Fprintf(output, "\n### %s (`%s`)\n", entry.Title, entry.ID)
						lines, err := catalog.Explain(section, entry.ID, technical)
						if err != nil {
							return err
						}
						for _, line := range lines {
							prefix := ""
							if line.Kind == "heading" {
								prefix = "#### "
							}
							if line.Kind == "warning" {
								prefix = "> "
							}
							fmt.Fprintf(output, "\n%s%s\n", prefix, line.Text)
						}
					}
				}
				return output.Err()
			}
			section, id := core.CatalogQuestions, ""
			for _, choice := range []struct {
				section core.CatalogSection
				id      string
			}{
				{core.CatalogQuestions, question}, {core.CatalogSources, source},
				{core.CatalogLayers, layer}, {core.CatalogTerms, term},
			} {
				if choice.id != "" {
					section, id = choice.section, choice.id
				}
			}
			if id != "" {
				lines, err := catalog.Explain(section, id, technical)
				if err != nil {
					return err
				}
				if asJSON {
					return json.NewEncoder(cmd.OutOrStdout()).Encode(lines)
				}
				for _, line := range lines {
					if line.Kind == "heading" {
						fmt.Fprintln(output)
					}
					fmt.Fprintln(output, line.Text)
				}
				if !technical && section != core.CatalogTerms {
					fmt.Fprintln(output, "\nVoeg --technisch toe voor de volledige API-aanvraag en parameters.")
				}
				return output.Err()
			}
			if asJSON && overview == "" {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(catalog)
			}
			if overview != "" {
				found := false
				for i, name := range []string{"vragen", "bronnen", "lagen", "begrippen"} {
					if name == overview {
						section, found = core.CatalogSection(i), true
					}
				}
				if !found {
					return fmt.Errorf("onbekend overzicht %q; kies vragen, bronnen, lagen of begrippen", overview)
				}
			}
			entries := catalog.Entries(section)
			if asJSON {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(entries)
			}
			fmt.Fprintln(output, core.CatalogSections[section])
			for _, entry := range entries {
				fmt.Fprintf(output, "\n%s — %s\n  %s\n", entry.ID, entry.Title, entry.Summary)
			}
			fmt.Fprintln(output, "\nKies een ID met --vraag, --bron, --laag of --begrip. --overzicht toont de bijbehorende lijst.")
			return output.Err()
		},
	}
	cmd.Flags().StringVar(&question, "vraag", "", "vraag-ID (bijv. riool of drinkwater)")
	cmd.Flags().StringVar(&source, "bron", "", "bron-ID (zie --overzicht bronnen)")
	cmd.Flags().StringVar(&layer, "laag", "", "laag-ID (zie --overzicht lagen)")
	cmd.Flags().StringVar(&term, "begrip", "", "begrip-ID (bijv. bob, nap of wfs)")
	cmd.Flags().StringVar(&overview, "overzicht", "", "vragen, bronnen, lagen of begrippen")
	cmd.Flags().String("data-root", "", "lees een andere bestaande datastore")
	cmd.Flags().BoolVar(&technical, "technisch", false, "toon API-aanvraag en parameteruitleg; voert geen aanvraag uit")
	cmd.Flags().BoolVar(&asJSON, "json", false, "catalogus of geselecteerde uitleg als JSON")
	cmd.Flags().BoolVar(&markdown, "markdown", false, "exporteer de gedeelde catalogus als documentatie")
	cmd.MarkFlagsMutuallyExclusive("markdown", "json")
	cmd.MarkFlagsMutuallyExclusive("vraag", "bron", "laag", "begrip", "overzicht", "markdown")
	return cmd
}
