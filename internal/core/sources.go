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
		// DAMO publishes 18 type names over WFS (measured 12 okt '26). These
		// four are the asset classes the panel could not tell apart: without
		// them a rioolgemaal, an RWZI, an overstort and a ketenoverdracht all
		// collapse into the one "DAMO waterketen" row. `pomp` is deliberately
		// absent — the WFS does not carry it (download service only), so a
		// `pomp` source would report a false zero.
		ID:        "damo_rioolgemaal",
		Label:     "DAMO rioolgemalen",
		URL:       "https://service.pdok.nl/hwh/waterschappen-waterketen-damo/wfs/v1_0",
		License:   "CC BY-SA 4.0",
		ProbeKind: "wfs_damo",
		TypeName:  "waterschappen-waterketen-damo:rioolgemaal",
	},
	{
		ID:        "damo_rwzi",
		Label:     "DAMO RWZI's",
		URL:       "https://service.pdok.nl/hwh/waterschappen-waterketen-damo/wfs/v1_0",
		License:   "CC BY-SA 4.0",
		ProbeKind: "wfs_damo",
		TypeName:  "waterschappen-waterketen-damo:rwzi",
	},
	{
		// Lozingspunten are the overstorten: where the sewer system discharges
		// to surface water. A distinct question from "where are the pipes".
		ID:        "damo_lozingspunt",
		Label:     "DAMO lozingspunten",
		URL:       "https://service.pdok.nl/hwh/waterschappen-waterketen-damo/wfs/v1_0",
		License:   "CC BY-SA 4.0",
		ProbeKind: "wfs_damo",
		TypeName:  "waterschappen-waterketen-damo:lozingspunt",
	},
	{
		// Overnamepunten mark where a municipality's sewer hands over to the
		// waterboard's transport system — the join in the waterketen.
		ID:        "damo_overnamepunt",
		Label:     "DAMO overnamepunten",
		URL:       "https://service.pdok.nl/hwh/waterschappen-waterketen-damo/wfs/v1_0",
		License:   "CC BY-SA 4.0",
		ProbeKind: "wfs_damo",
		TypeName:  "waterschappen-waterketen-damo:overnamepunt",
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
		// RIONED's own collections split what the GWSW `beheer_pomp` layer
		// merges: on 8 okt '26 that layer answered 938 objects = 638
		// Rioolgemaal + 300 putconstructies, so "gemalen" could not be
		// counted from it. An OGC probe is a liveness check only — PDOK
		// omits numberMatched, so these rows show "—" for the count.
		ID:        "rioned_rioolgemaal",
		Label:     "RIONED rioolgemalen",
		URL:       "https://api.pdok.nl/rioned/waterschappen-waterketen-gwsw/ogc/v1_0",
		License:   "CC0",
		ProbeKind: "ogc",
		TypeName:  "waterschap_rioolgemaal",
	},
	{
		ID:        "rioned_rwzi",
		Label:     "RIONED RWZI's",
		URL:       "https://api.pdok.nl/rioned/waterschappen-waterketen-gwsw/ogc/v1_0",
		License:   "CC0",
		ProbeKind: "ogc",
		TypeName:  "waterschap_rwzi",
	},
	{
		// IBA = individuele behandeling afvalwater: the small-scale systems
		// outside the sewer network, a separate collection from RWZI.
		ID:        "rioned_iba",
		Label:     "RIONED IBA's",
		URL:       "https://api.pdok.nl/rioned/waterschappen-waterketen-gwsw/ogc/v1_0",
		License:   "CC0",
		ProbeKind: "ogc",
		TypeName:  "waterschap_iba",
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
		// The Waterschapshuis IMWA family, added 12 okt '26 after the report's
		// URLs turned out dead. Licences verified per NGR record (CC0, "Geen
		// beperkingen") — NOT inherited from `hydrografie` on the same host,
		// which is CC BY-NC-ND 4.0 and stays unwired.
		//
		// The probe counts the named layer, not the whole dataset: this
		// service carries 18 object types (stuw 53.142, brug 92.368, sluis 889,
		// duiker, vispassage…), and only `gemaal` is counted here because it is
		// the water-regulating work a waterketen view asks about.
		ID:        "hwh_kunstwerken_imwa",
		Label:     "Waterschappen Kunstwerken IMWA (gemaal)",
		URL:       "https://service.pdok.nl/hwh/waterschappen-kunstwerken-imwa/wfs/v2_0",
		License:   "CC0",
		ProbeKind: "wfs",
		TypeName:  "waterschappen-kunstwerken-imwa:gemaal",
	},
	{
		// Canonical national surface-water layer: the "what water is this pipe
		// next to" question. 1.430.861 surface-water bodies measured, but the
		// hits-request costs ~10 s — the slowest probe here (waterdeel 2,5 s,
		// oever 0,24 s measured 12 okt '26). The panel streams per source, so
		// this row lands last instead of holding up the other sixteen; the
		// canonical object type is worth that.
		ID:        "hwh_oppervlaktewateren_imwa",
		Label:     "Waterschappen Oppervlaktewateren IMWA",
		URL:       "https://service.pdok.nl/hwh/waterschappen-oppervlaktewateren-imwa/wfs/v2_0",
		License:   "CC0",
		ProbeKind: "wfs",
		TypeName:  "waterschappen-oppervlaktewateren-imwa:oppervlaktewaterlichaam",
	},
	{
		// v3_0, not v2_0: keringen answers 404 on v1_0 and v2_0. 28.396
		// waterkeringen + 21.825 waterkeringsecties measured 12 okt '26.
		ID:        "hwh_keringen_imwa",
		Label:     "Waterschappen Keringen IMWA",
		URL:       "https://service.pdok.nl/hwh/waterschappen-keringen-imwa/wfs/v3_0",
		License:   "CC0",
		ProbeKind: "wfs",
		TypeName:  "waterschappen-keringen-imwa:waterkering",
	},
	{
		// Peilgebieden are the unit in which water levels are managed;
		// 37.418 measured. peilbesluitgebied (3.937) is the formal subset.
		ID:        "hwh_waterbeheergebieden_imwa",
		Label:     "Waterschappen Waterbeheergebieden IMWA (peilgebied)",
		URL:       "https://service.pdok.nl/hwh/waterschappen-waterbeheergebieden-imwa/wfs/v2_0",
		License:   "CC0",
		ProbeKind: "wfs",
		TypeName:  "waterschappen-waterbeheergebieden-imwa:peilgebied",
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
	{
		// The attribution layer: which waterboard manages where. Registered
		// but NOT wired because it is WMS-only — /wfs/v1_0 and /wfs/v2_0 both
		// answer HTTP 404 (measured 12 okt '26), so it can be drawn and never
		// queried as objects. Licence is CC0 in the service's constraints.
		ID:      "hwh_waterschapsgrenzen_imso",
		Label:   "Waterschappen Waterschapsgrenzen IMSO (alleen WMS)",
		URL:     "https://service.pdok.nl/hwh/waterschappen-waterschapsgrenzen-imso/wms/v2_0",
		License: "CC0",
		Note:    "alleen WMS, geen WFS: geen objectvraag; attributie loopt via GWSW/DAMO",
	},
}
