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
	var technical, asJSON bool
	cmd := &cobra.Command{
		Use: "ontdek",
		Short: "begrijp bronnen, lagen en begrippen vanuit je vraag — offline",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			catalog, err := core.LoadCatalog(dataRoot(cmd))
			if err != nil { return err }
			section, id := core.CatalogQuestions, ""
			for _, choice := range []struct {section core.CatalogSection; id string}{
				{core.CatalogQuestions, question}, {core.CatalogSources, source},
				{core.CatalogLayers, layer}, {core.CatalogTerms, term},
			} {
				if choice.id != "" { section, id = choice.section, choice.id }
			}
			if id != "" {
				lines, err := catalog.Explain(section, id, technical)
				if err != nil { return err }
				if asJSON { return json.NewEncoder(cmd.OutOrStdout()).Encode(lines) }
				for _, line := range lines {
					if line.Kind == "heading" { fmt.Fprintln(cmd.OutOrStdout()) }
					fmt.Fprintln(cmd.OutOrStdout(), line.Text)
				}
				if !technical && section != core.CatalogTerms { fmt.Fprintln(cmd.OutOrStdout(), "\nVoeg --technisch toe voor de volledige API-aanvraag en parameters.") }
				return nil
			}
			if asJSON && overview == "" { return json.NewEncoder(cmd.OutOrStdout()).Encode(catalog) }
			if overview != "" {
				found := false
				for i, name := range []string{"vragen", "bronnen", "lagen", "begrippen"} {
					if name == overview { section, found = core.CatalogSection(i), true }
				}
				if !found { return fmt.Errorf("onbekend overzicht %q; kies vragen, bronnen, lagen of begrippen", overview) }
			}
			entries := catalog.Entries(section)
			if asJSON { return json.NewEncoder(cmd.OutOrStdout()).Encode(entries) }
			fmt.Fprintln(cmd.OutOrStdout(), core.CatalogSections[section])
			for _, entry := range entries { fmt.Fprintf(cmd.OutOrStdout(), "\n%s — %s\n  %s\n", entry.ID, entry.Title, entry.Summary) }
			fmt.Fprintln(cmd.OutOrStdout(), "\nKies een ID met --vraag, --bron, --laag of --begrip. --overzicht toont de bijbehorende lijst.")
			return nil
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
	cmd.MarkFlagsMutuallyExclusive("vraag", "bron", "laag", "begrip", "overzicht")
	return cmd
}
