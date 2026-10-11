// Package core holds the pure waterketen domain logic: sources, health
// probing, metadata, library, analysis, atlas rendering. NOTHING in here may
// import bubbletea, bubbles or lipgloss — strings that reach the screen come
// from report.go as styled-neutral Line values and the screens style them
// with tokens from internal/ui.
package core

// Source is one data origin, either wired into the pipeline or detected only.
type Source struct {
	ID      string `json:"source_id"`
	Label   string `json:"label"`
	URL     string `json:"url"`
	License string `json:"license"`
	// ProbeKind selects the probe implementation in health.go:
	// "wfs" | "wfs_damo" | "arcgis" | "ogc".
	ProbeKind string `json:"probe_kind"`
	// TypeName is the layer/collection a probe must name. The GWSW layer name
	// sent to DAMO returns HTTP 400, so it belongs with the source and not
	// hard-coded in the probe.
	TypeName string `json:"type_name,omitempty"`
}

// WiredSources are the sources this app probes live, in panel order. It is a
// superset of the Python CLI's fetch registry (src/waterketen_cli/tui/app.py):
// an entry here is registered and probeable, but ingest is still Python's —
// the IMWA/DAMO/RIONED object types below have no Python fetch slug yet, so a
// green row means "the source answers", never "this is on disk".
//
// Register entries that were checked live (8 okt '26) and are deliberately NOT
// here, so they are not re-investigated: data.overheid.nl/dataset/79777 is the
// English duplicate of 81464 and returns 404; data.overheid.nl/dataset/11606
// ("GWSW — Beheer stedelijk water systemen") is gone, its living equivalent is
// the unwired bulk GPKG below. The community datarequest 7077 is not a dataset
// at all — its "direct GeoJSON download" is exactly the wired Vewin endpoint.
//
// Dead-and-deliberately-not-registered (12 okt '26): the HWH service names in
// PDOK's "vernieuwing" article of 22 dec 2023 — waterschapsgrenzenimso,
// zoneringenimwa, kunstwerkenimwa, oppervlaktewaterenimwa,
// waterbeheergebiedenimwa and keringenimwa — all answer HTTP 404 under
// /hwh/<naam>/wms/v1_0. The living names are `waterschappen-<naam>` on v2_0
// (keringen v3_0); sources_test.go fails if a dead name sneaks back in.
var WiredSources = []Source{
	{
		ID:        "gws_netwerk",
		Label:     "GWSW netwerk",
		URL:       "https://geodata.gwsw.nl/geoserver/ows",
		License:   "CC0",
		ProbeKind: "wfs",
		TypeName:  "Rotterdam-beheer:beheer_leiding",
	},
	{
		ID:        "damo_leiding",
		Label:     "DAMO waterketen",
		URL:       "https://service.pdok.nl/hwh/waterschappen-waterketen-damo/wfs/v1_0",
		License:   "CC BY-SA 4.0",
		ProbeKind: "wfs_damo",
		TypeName:  "waterschappen-waterketen-damo:leiding",
	},
	{
		ID:    "vewin",
		Label: "Vewin drinkwater",
		URL: "https://services5.arcgis.com/ZcRj7hEl3ya9tYHV/arcgis/rest/services/" +
			"Waterbedrijven_Nederland/FeatureServer/0",
		License:   "zie bron (Vewin/ArcGIS)",
		ProbeKind: "arcgis",
	},
	{
		// RIONED's national GWSW service, PDOK, OGC API Features. CC0. It carries
		// the same objects as the register entry "Waterschappen Waterketen GWSW",
		// including rioolgemaal, rwzi and iba as their own collections — which the
		// GWSW beheer_pomp layer does NOT separate.
		ID:        "rioned_gwsw",
		Label:     "RIONED Waterketen GWSW (OGC)",
		URL:       "https://api.pdok.nl/rioned/waterschappen-waterketen-gwsw/ogc/v1_0",
		License:   "CC0",
		ProbeKind: "ogc",
		TypeName:  "waterschap_leiding",
	},
	{
		// Provincie Zuid-Holland "drinkwaterkaart" WFS: 60 layers, of which
		// DRINKWATER_BEDRIJVEN is the ten utility service areas (Evides over
		// Delfshaven) — the independent second delivery next to Vewin, not a
		// pipe network. Also carries production chains, 6-hour zones, strategic
		// reserves and the sparse Delfland sewer registration.
		ID:        "zh_drinkwater",
		Label:     "Zuid-Holland drinkwater (WFS)",
		URL:       "https://geodata.zuid-holland.nl/geoserver/water/wfs",
		License:   "Publiek domein (PDM 1.0)",
		ProbeKind: "wfs",
		TypeName:  "water:DRINKWATER_BEDRIJVEN",
	},
	{
		// Rijkswaterstaat GDR region boundaries: management services and
		// districts (dry + wet), 48 each. Delivered as a MapServer, not a
		// FeatureServer — the FeatureServer form of this name answers HTTP 500.
		// License is genuinely not published in the service, so the panel must
		// say so rather than imply one.
		ID:    "rws_regiogebieden",
		Label: "RWS regiogebieden (ArcGIS)",
		URL: "https://geo.rijkswaterstaat.nl/arcgis/rest/services/GDR/" +
			"regiogebieden_rijkswaterstaat/MapServer/0",
		License:   "onbekend — geen licentieveld in de dienst",
		ProbeKind: "arcgis",
	},
}

// UnwiredSource is detected + documented but NOT wired into the registry.
type UnwiredSource struct {
	ID      string
	Label   string
	URL     string
	License string
	Note    string
}

// UnwiredSources is the display-only table. Never invent counts for these.
var UnwiredSources = []UnwiredSource{
	{
		// INSPIRE-harmonised watercourse network. Registered but deliberately NOT
		// wired: the PDOK landing page states CC BY-NC-ND 4.0, which forbids
		// commercial use and derivatives — a materially different licence from the
		// CC0 datasets here. Decision for the user, not the agent.
		ID:      "hwh_hydrografie",
		Label:   "Waterschappen Hydrografie (INSPIRE)",
		URL:     "https://api.pdok.nl/hwh/waterschappen-hydrografie/ogc/v1",
		License: "CC BY-NC-ND 4.0 (geen commercieel gebruik, geen afgeleiden)",
		Note:    "bewust niet bedraad — licentie eerst beslissen",
	},
	{
		// Measured Content-Length 2.993.788.928 B = 2,99 GB (2,79 GiB); the earlier
		// 2,96 GB figure was an estimate. Same content as DAMO, as one file.
		ID:    "rioned_national",
		Label: "RIONED landelijk stedelijk water",
		URL: "https://service.pdok.nl/rioned/beheer-stedelijk-watersystemen-gwsw" +
			"/atom/downloads/stedelijkwater.gpkg",
		License: "CC0",
		Note:    "~2,99 GB bulk — alleen expliciet, nooit impliciet",
	},
	{
		ID:    "klic_rotterdam",
		Label: "KLIC-afgeleid Rotterdam (ArcGIS)",
		URL: "https://services9.arcgis.com/nqW2A97fCpeCbw6Z/arcgis/rest/services/" +
			"Gemeente_Rotterdam_Rioolbuizen/FeatureServer/0",
		License: "zie bron (persoonlijk/Hogeschool-account)",
		Note:    "provenance: persoonlijk/Hogeschool-account — fallback, niet primair",
	},
	{
		ID:      "gws_aanvullend",
		Label:   "GWSW overige Rotterdam-lagen",
		URL:     "https://geodata.gwsw.nl/geoserver/ows",
		License: "CC0",
		Note:    "27 lagen gedetecteerd; alleen beheer/netwerk/waterschap bedraad",
	},
	{
		// On the same RWS host as the region boundaries this is the more relevant
		// neighbour for drinking water: intake points, points and protection
		// zones. Not wired because the service publishes no licence field.
		ID:    "rws_nwp_drinkwater",
		Label: "RWS drinkwater (NWP)",
		URL: "https://geo.rijkswaterstaat.nl/arcgis/rest/services/GDR/" +
			"nwp_drinkwater/MapServer",
		License: "onbekend — geen licentieveld in de dienst",
		Note:    "innamepunten, punten en beschermingszones — licentie eerst verifiëren",
	},
}
