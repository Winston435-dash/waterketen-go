// Package core holds the pure waterketen domain logic: sources, health
// probing, metadata, library, analysis, atlas rendering. NOTHING in here may
// import bubbletea, bubbles or lipgloss — strings that reach the screen come
// from report.go as styled-neutral Line values and the screens style them
// with tokens from internal/ui.
package core

// Source is one data origin, either wired into the pipeline or detected only.
type Source struct {
	ID        string `json:"source_id"`
	Label     string `json:"label"`
	URL       string `json:"url"`
	ProbeKind string `json:"probe_kind"` // "wfs" | "wfs_damo" | "arcgis"
}

// WiredSources are the sources the CLI can actually fetch. Mirrors the Python
// WIRED_SOURCES table (src/waterketen_cli/tui/app.py) one-to-one.
var WiredSources = []Source{
	{
		ID:        "gws_netwerk",
		Label:     "GWSW netwerk",
		URL:       "https://geodata.gwsw.nl/geoserver/ows",
		ProbeKind: "wfs",
	},
	{
		ID:        "damo_leiding",
		Label:     "DAMO waterketen",
		URL:       "https://service.pdok.nl/hwh/waterschappen-waterketen-damo/wfs/v1_0",
		ProbeKind: "wfs_damo",
	},
	{
		ID:    "vewin",
		Label: "Vewin drinkwater",
		URL: "https://services5.arcgis.com/ZcRj7hEl3ya9tYHV/arcgis/rest/services/" +
			"Waterbedrijven_Nederland/FeatureServer/0",
		ProbeKind: "arcgis",
	},
}

// UnwiredSource is detected + documented but NOT wired into the registry.
type UnwiredSource struct {
	ID    string
	Label string
	URL   string
	Note  string
}

// UnwiredSources is the display-only table. Never invent counts for these.
var UnwiredSources = []UnwiredSource{
	{
		ID:    "rioned_national",
		Label: "RIONED landelijk stedelijk water",
		URL: "https://service.pdok.nl/rioned/beheer-stedelijk-watersystemen-gwsw" +
			"/atom/downloads/stedelijkwater.gpkg",
		Note: "~2,96 GB bulk — alleen expliciet, nooit impliciet",
	},
	{
		ID:    "klic_rotterdam",
		Label: "KLIC-afgeleid Rotterdam (ArcGIS)",
		URL: "https://services9.arcgis.com/nqW2A97fCpeCbw6Z/arcgis/rest/services/" +
			"Gemeente_Rotterdam_Rioolbuizen/FeatureServer/0",
		Note: "provenance: persoonlijk/Hogeschool-account — fallback, niet primair",
	},
	{
		ID:    "gws_aanvullend",
		Label: "GWSW overige Rotterdam-lagen",
		URL:   "https://geodata.gwsw.nl/geoserver/ows",
		Note:  "27 lagen gedetecteerd; alleen beheer/netwerk/waterschap bedraad",
	},
}
