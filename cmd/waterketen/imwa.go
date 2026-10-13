package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"waterketen/internal/core"
)

func libraryCmd() *cobra.Command {
	c := &cobra.Command{Use: "bibliotheek", Short: "beheer de lokale kaartbibliotheek"}
	var asJSON bool
	imwa := &cobra.Command{
		Use:   "imwa",
		Short: "haal IMWA-duikers, -gemalen, -sluizen en -stuwen op voor Rotterdam",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer stop()
			result, err := core.ImportIMWA(ctx, dataRoot(cmd))
			if err != nil {
				return err
			}
			if asJSON {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(result)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Gepubliceerd: %s\n", result.Snapshot)
			for _, name := range core.OrderLayers([]string{"hwh-kunstwerken-duiker", "hwh-kunstwerken-gemaal", "hwh-kunstwerken-sluis", "hwh-kunstwerken-stuw"}) {
				fmt.Fprintf(cmd.OutOrStdout(), "  %s: %d objecten\n", name, result.Counts[name])
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Herstart de TUI; kies de IMWA-lagen met t op de kaart. Bron: Het Waterschapshuis / PDOK, CC0.")
			return nil
		},
	}
	imwa.Flags().String("data-root", "", "overschrijf $WATERKETEN_DATA_DIR")
	imwa.Flags().BoolVar(&asJSON, "json", false, "machine-leesbare output")
	c.AddCommand(imwa)
	return c
}
