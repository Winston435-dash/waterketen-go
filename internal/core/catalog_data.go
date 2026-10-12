package core

// NewCatalog keeps the explanatory vocabulary shared by CLI and TUI. URLs,
// licences and probe layer names come from the existing operational registries.
func NewCatalog() Catalog {
	c := Catalog{
		Questions: []CatalogQuestion{
			{"riool", "Waar liggen rioolbuizen?", "Bekijk de geregistreerde ligging; een publicatie bewijst niet dat iedere buis is opgenomen.", []string{"riool"}, nil},
			{"drinkwater", "Wie levert hier drinkwater?", "Een verzorgingsgebied vertelt wie er levert, niet waar de buizen liggen.", []string{"drinkwater"}, nil},
			{"drinkwaterleidingen", "Waar liggen drinkwaterleidingen?", "De gemeentelijke Evides-publicatie toont een deel van de leidinggeometrie.", []string{"drinkwaterleidingen"}, nil},
			{"oppervlaktewatergemalen", "Waar liggen gemalen voor oppervlaktewater?", "Deze gemeentelijke laag registreert oppervlaktewatergemalen; rioolgemalen staan apart.", []string{"oppervlaktewatergemalen"}, nil},
			{"afvalwater", "Waar wordt afvalwater gepompt of gezuiverd?", "Putten, gemalen en afvoergebieden vertellen verschillende delen van het verhaal. Nabijheid is geen bewezen verbinding.", []string{"putten", "gemalen", "damo-rioolgemaal", "damo-zuiveringseenheid", "damo-rwzi", "damo-lozingspunt", "rioned-rioolgemaal", "rioned-rwzi"}, []string{"rwzi", "overstort"}},
			{"watercontext", "Welk water, welke kering en welk peilgebied liggen bij dit gebied?", "Oppervlaktewater, keringen en peilgebieden beschrijven de omgeving van het netwerk; ze zijn geen registratie van de buizen zelf.", []string{"hwh-oppervlaktewaterlichaam", "hwh-waterkering", "hwh-peilgebied", "hwh-kunstwerken-gemaal"}, []string{"imwa"}},
			{"diepte", "Hoe hoog of diep liggen de buizen?", "BOB is buisbodemhoogte ten opzichte van NAP. Voor diepte onder een straat is ook het maaiveld nodig.", []string{"riool", "hoogte"}, []string{"maaiveld"}},
			{"omgeving", "Wat ligt er boven en rond het netwerk?", "Straten, water en wijkgrenzen geven context; ze zijn geen registratie van ondergrondse leidingen.", []string{"straten", "water", "gebieden"}, []string{"bbox"}},
		},
		Terms: []CatalogTerm{
			{"gwsw", "GWSW — Gegevenswoordenboek Stedelijk Water", "Een standaard voor de namen, eigenschappen en relaties van stedelijke waterobjecten; geen API-protocol.", "Een buis kan een materiaal en begin- en eindpunt hebben volgens GWSW."},
			{"damo", "DAMO — Datamodel van de waterschappen", "Een model waarmee waterschappen objecten en hun eigenschappen beschrijven; geen aanbieder of downloadformaat.", "Een DAMO-rioolgemaal en een GWSW-rioolgemaal kunnen hetzelfde fysieke object beschrijven."},
			{"rioned", "Stichting RIONED", "Een organisatie voor stedelijk waterbeheer, betrokken bij GWSW; geen kaartlaag of protocol.", "RIONED-gegevens kunnen via PDOK worden gepubliceerd."},
			{"pdok", "PDOK — Publieke Dienstverlening Op de Kaart", "Een voorziening die geografische gegevens van verschillende aanbieders toegankelijk maakt.", "Een waterschap kan gegevens leveren die via een PDOK-dienst beschikbaar zijn."},
			{"wfs", "WFS — Web Feature Service", "Een standaard om kaartobjecten en hun eigenschappen op te vragen.", "GetFeature met resultType=hits vraagt alleen het aantal objecten in een laag."},
			{"ogc", "OGC API Features", "Een standaard voor het opvragen van kaartobjecten via collecties en webadressen. Een collectie is een gegevenslaag.", "/collections/waterschap_leiding/items bevat objecten uit die collectie."},
			{"arcgis", "ArcGIS REST — FeatureServer / MapServer", "Een webinterface van ArcGIS. Een FeatureServer publiceert objecten; sommige MapServer-lagen ondersteunen ook objectvragen.", "Een /query-aanvraag met returnCountOnly=true geeft een telling, geen buisvormen."},
			{"feature", "Feature — kaartobject", "Een object met geometrie (kaartvorm) en eigenschappen, zoals een buis, put of gebied.", "Een rioolbuis is een lijn; het materiaal is een eigenschap van die lijn."},
			{"layer", "Laag / collectie", "Een verzameling samenhangende kaartobjecten. Eén API-dienst kan meerdere lagen aanbieden.", "Een laag met verzorgingsgebieden bevat vlakken, geen drinkwaterleidingen."},
			{"endpoint", "API-endpoint", "Een webadres waar een programma een bepaalde vraag kan stellen. Vaak zijn ook parameters nodig.", "Een kale WFS-URL kan een fout geven; een volledige GetFeature-aanvraag noemt ook de laag."},
			{"bob", "BOB — binnenonderkant buis", "De hoogte van de binnenkant van de buisbodem ten opzichte van NAP; niet automatisch de diepte onder de straat.", "Begin- en eindhoogte horen bij verschillende buisuiteinden. In de lokale rioollevering behandelt de tool 0,00 als onbekend."},
			{"nap", "NAP — Normaal Amsterdams Peil", "Het landelijke referentieniveau waarmee hoogtes worden vergeleken. Het is niet het lokale straatniveau.", "Een positieve BOB kan onder een hoger gelegen straat liggen; het teken alleen bewijst geen fout."},
			{"maaiveld", "Maaiveld", "Het oppervlak van de grond. Voor straatdiepte moet je de buisbodem met het straatniveau vergelijken.", "AHN-waarden boven open water kunnen het wateroppervlak voorstellen, niet de straat."},
			{"ahn", "AHN — Actueel Hoogtebestand Nederland", "Hoogtegegevens van Nederland. Het DTM beschrijft het terrein; het DSM bevat ook objecten boven het terrein.", "Voor maaiveld gebruikt de Python-tool het DTM, niet een gebouwhoogte uit het DSM."},
			{"wcs", "WCS — Web Coverage Service", "Een standaard om rastergegevens (een rooster met waarden) op te vragen.", "Een AHN-aanvraag kan een GeoTIFF met hoogtes opleveren, geen losse rioolbuizen."},
			{"crs", "CRS — coördinatenreferentiesysteem", "De afspraak die bepaalt hoe coördinaten op de aarde worden geplaatst.", "RD New (EPSG:28992) gebruikt meters; WGS 84 (EPSG:4326) gebruikt geografische graden."},
			{"bbox", "Bbox — bounding box", "Een rechthoekig zoekvak. Objecten in dat vak liggen niet noodzakelijk binnen de echte wijkgrens.", "De hoeken van een wijkvak kunnen stukken van buurgemeenten of buurwijken bevatten."},
			{"gpkg", "GeoPackage (GPKG)", "Een databasebestand voor geografische objecten en eigenschappen, dat lokaal zonder netwerk kan worden gelezen.", "De CLI en TUI lezen dezelfde beheerde GeoPackages; het bestand is geen API."},
			{"rwzi", "RWZI — rioolwaterzuiveringsinstallatie", "Een installatie die afvalwater zuivert. Een zuiveringseenheid als gebied is niet hetzelfde als een RWZI-punt.", "Een vlak dat het aangesloten gebied beschrijft, markeert niet de locatie van de zuiveringsinstallatie."},
			{"oppervlaktewatergemaal", "Oppervlaktewatergemaal", "Een installatie die water in het oppervlaktewatersysteem verplaatst voor peil- of waterbeheer; geen afvalwaterpomp.", "De gemeentelijke kaartlaag typeert oppervlaktewatergemalen apart van rioolgemalen."},
			{"klic", "KLIC — Kabels en Leidingen Informatie Centrum", "Informatie-uitwisseling via het Kadaster over kabels en leidingen, onder meer voor graafwerk.", "Deze openbare kaartcatalogus vervangt geen vereiste KLIC-melding."},
			{"snapshot", "Snapshot — vastgelegde levering", "Een bewaarde versie van gegevens op een bepaald moment. De ophaaldatum hoeft niet de meet- of brondatum te zijn.", "Een recent opgehaald bestand kan oude metingen bevatten."},
			// Added 12 okt '26 with the IMWA/DAMO/RIONED object types.
			{"imwa", "IMWA — Informatiemodel Water", "Het informatiemodel waarmee waterschappen hun waterobjecten, kunstwerken en beheergebieden beschrijven.", "Een waterkering en een peilgebied zijn IMWA-objecttypen."},
			{"kering", "Kering — waterkering", "Een constructie die water tegenhoudt, zoals een dijk, dam of kade; een waterkeringsectie is een deel daarvan.", "Een geregistreerde kering is geen actuele toetsstatus."},
			{"peilgebied", "Peilgebied", "Een gebied met één afgesproken waterpeil; het peilbesluitgebied is het deel waarvoor formeel een peilbesluit geldt.", "Het peilgebied vertelt welk peil hoort bij de sloot naast een rioolbuis."},
			{"overstort", "Overstort", "Een punt waar rioolwater bij hevige regen of verstopping op het oppervlaktewater komt.", "Een DAMO-lozingspunt kan een overstort zijn, maar de bron beschrijft het gedrag niet."},
			{"iba", "IBA — individuele behandeling afvalwater", "Een kleinschalig systeem dat afvalwater van één of enkele woningen behandelt, buiten het riool om.", "Een IBA is geen rioolaansluiting."},
		},
	}
	// Provider, purpose and limitations are editorial; operational fields remain
	// owned by sources.go so a URL or licence change cannot drift between views.
	descriptions := map[string][3]string{
		"gws_netwerk":        {"GWSW-publicatie Rotterdam", "Geregistreerde rioolbuizen en hun eigenschappen.", "Geen bewijs van fysieke volledigheid; publieke data is geen KLIC-vervanging."},
		"damo_leiding":       {"Het Waterschapshuis / waterschappen via PDOK", "Waterketenobjecten volgens DAMO, zoals leidingen, putten en gemalen.", "Kan overlappen met GWSW. Niet iedere beschreven objectsoort is via WFS beschikbaar."},
		"vewin":              {"Vewin / ArcGIS-publicatie", "Verzorgingsgebieden van drinkwaterbedrijven.", "Gebieden, geen leidingen. Gebruiksvoorwaarden zijn niet expliciet vastgesteld."},
		"rioned_gwsw":        {"Stichting RIONED via PDOK", "Waterketenobjecten volgens GWSW, toegankelijk als afzonderlijke collecties.", "De kleine probe geeft geen totaalaantal; landelijke levering zegt niets over wijkdekking."},
		"zh_drinkwater":      {"Provincie Zuid-Holland", "Drinkwaterbedrijven, winnings- en beschermingsinformatie en een afzonderlijke transportlevering.", "Verzorgingsgebieden kunnen dezelfde soort gegevens als Vewin leveren; dit is geen extra leidingnet. Licenties gelden per dataset."},
		"rws_regiogebieden":  {"Rijkswaterstaat", "Grenzen van diensten en districten van Rijkswaterstaat.", "Beheergebieden zijn geen drinkwaternet. De dienst vermeldt geen expliciete licentie."},
		"hwh_hydrografie":    {"Het Waterschapshuis via PDOK", "Een geharmoniseerd netwerk van waterlopen volgens INSPIRE.", "CC BY-NC-ND verbiedt commercieel gebruik en afgeleide werken; bewust niet aangesloten."},
		"rioned_national":    {"Stichting RIONED via PDOK", "Een landelijke GWSW-levering als één GeoPackage-bestand.", "Een bulkbestand van circa 2,99 GB; nooit automatisch downloaden. Niet verwarren met DAMO."},
		"klic_rotterdam":     {"Persoonlijk/Hogeschool-account op ArcGIS", "Een KLIC-afgeleide rioolpublicatie met andere herkomst dan de primaire bron.", "Fallback, niet de primaire gemeentelijke registratie. Rechten en herkomst eerst beoordelen."},
		"gws_aanvullend":     {"GWSW-publicatie Rotterdam", "Aanvullende Rotterdamse lagen, waaronder putten en pompconstructies.", "De pompconstructielaag kan zowel rioolgemalen als putconstructies bevatten."},
		"rws_nwp_drinkwater": {"Rijkswaterstaat", "Drinkwaterinnamepunten en beschermingsinformatie.", "Geen volledig leidingnet; licentie eerst verifiëren."},
		// Added 12 okt '26. Provider/purpose/limit are editorial; URL, licence
		// and probe layer stay owned by sources.go.
		"damo_rioolgemaal":             {"Het Waterschapshuis / waterschappen via PDOK", "Rioolgemalen die waterschappen volgens DAMO registreren.", "Losse pomp-onderdelen staan niet in deze WFS. Kan hetzelfde fysieke object zijn als een GWSW-gemaal."},
		"damo_rwzi":                    {"Het Waterschapshuis / waterschappen via PDOK", "Rioolwaterzuiveringsinstallaties als puntobject.", "Eén punt per installatie; capaciteit of prestatie staat er niet in. Niet optellen bij de CBS-statistiek."},
		"damo_lozingspunt":             {"Het Waterschapshuis / waterschappen via PDOK", "Punten waar afvalwater op het oppervlaktewater loost, inclusief overstorten.", "Een lozingspunt is geen bewezen overstortwerking; de bron beschrijft de lozing, niet het gedrag."},
		"damo_overnamepunt":            {"Het Waterschapshuis / waterschappen via PDOK", "Punten waar gemeentelijk riool overgaat naar het waterschapssysteem.", "Markeert de ketenoverdracht, niet de volledige route van het water daarna."},
		"rioned_rioolgemaal":           {"Stichting RIONED via PDOK", "Rioolgemalen als eigen GWSW-collectie.", "Lost wat de GWSW beheer_pomp-laag vermengt: die telt gemalen en putconstructies bij elkaar op."},
		"rioned_rwzi":                  {"Stichting RIONED via PDOK", "RWZI's als eigen GWSW-collectie.", "Alleen de installatie; procesgegevens staan in de CBS-tabellen, niet hier."},
		"rioned_iba":                   {"Stichting RIONED via PDOK", "Individuele behandeling afvalwater: kleinschalige systemen buiten het riool.", "Een IBA is geen rioolaansluiting; dekking hangt af van wat de bronhouder aanlevert."},
		"hwh_kunstwerken_imwa":         {"Het Waterschapshuis via PDOK", "Door waterschappen beheerde kunstwerken volgens IMWA: gemalen, stuwen, sluizen, duikers.", "De probe telt alleen de gemaal-laag; de dienst bevat 18 objectsoorten. Registratie zegt niets over actuele bediening of toestand."},
		"hwh_oppervlaktewateren_imwa":  {"Het Waterschapshuis via PDOK", "Landelijke laag van oppervlaktewaterlichamen en oevers volgens IMWA.", "Beheerssemantiek, geen topografie: de BGT-waterdelen zijn geen duplicaat. Geen debiet of waterstand."},
		"hwh_keringen_imwa":            {"Het Waterschapshuis via PDOK", "Waterkeringen en waterkeringsecties van de waterschappen.", "Een geregistreerde kering is geen actuele toetsstatus; gebruik de legger voor juridische uitspraken."},
		"hwh_waterbeheergebieden_imwa": {"Het Waterschapshuis via PDOK", "Peilgebieden, peilbesluitgebieden en afvoer-/aanvoergebieden.", "De probe telt de peilgebiedlaag. Grenzen verschillen van bestuurlijke waterschapsgrenzen."},
		"hwh_waterschapsgrenzen_imso":  {"Het Waterschapshuis via PDOK", "Bestuurlijke waterschapsgrenzen volgens IMSO.", "Alleen als WMS beschikbaar: wel te tekenen, niet als objecten op te vragen."},
	}
	for _, s := range WiredSources {
		d := descriptions[s.ID]
		c.Sources = append(c.Sources, CatalogSource{Source: s, Provider: d[0], Summary: d[1], Limit: d[2], Access: "Live broncontrole aangesloten; ophalen waar ondersteund via Python. Dit is geen lokale downloadstatus."})
	}
	for _, u := range UnwiredSources {
		d := descriptions[u.ID]
		s := Source{ID: u.ID, Label: u.Label, URL: u.URL, License: u.License}
		if u.ID == "gws_aanvullend" {
			s.ProbeKind = "wfs"
			s.TypeName = "Rotterdam-beheer:beheer_put"
		}
		if u.ID == "klic_rotterdam" {
			s.ProbeKind = "arcgis"
		}
		c.Sources = append(c.Sources, CatalogSource{Source: s, Provider: d[0], Summary: d[1], Limit: d[2], Access: "Geen live broncontrole aangesloten: " + u.Note})
	}
	for _, spec := range LayerSpecs {
		if spec.SourceID == "evides" {
			c.Sources = append(c.Sources, CatalogSource{Source: Source{ID: "evides", Label: "Gemeentelijke Evides-publicatie", URL: spec.SourceURL, License: "zie bron (gemeente Rotterdam/Evides)", ProbeKind: "arcgis"}, Provider: "Gemeente Rotterdam; leidingen afkomstig van Evides", Summary: "Lijngeometrie van een deel van de drinkwaterleidingen, met opwarmingsrisico.", Limit: "Volgens de bron alleen leidingen die niet binnen tien jaar worden vervangen; geen volledige netwerk- of buisattributenregistratie.", Access: "Ophalen via Python; geen afzonderlijke live probe in Go."})
		}
	}
	c.Sources = append(c.Sources,
		CatalogSource{Source: Source{ID: "osm", Label: "OpenStreetMap (lokale PBF)", License: "ODbL"}, Provider: "OpenStreetMap-bijdragers", Summary: "Bovengrondse context uit een lokaal geïmporteerd provinciaal bestand.", Limit: "Geen volledige officiële objectregistratie of ondergrondse leidingkaart.", Access: "Python importeert een lokale PBF; Go leest de bestaande snapshot. Geen API-aanvraag."},
		CatalogSource{Source: Source{ID: "ahn", Label: "AHN DTM (WCS)", URL: "https://service.pdok.nl/rws/ahn/wcs/v1_0", License: "onbekend", ProbeKind: ""}, Provider: "AHN via PDOK", Summary: "Terreinhoogtes als raster, voor vergelijking met de buisbodem.", Limit: "Boven open water is een hoogtesample niet automatisch het straatmaaiveld. Geen licentie in dit register vastgesteld.", Access: "De Python-opdracht diepte downloadt AHN expliciet; geen Go-kaartlaag."},
		CatalogSource{Source: Source{ID: "cbs", Label: "CBS wijk- en buurtgrenzen (2025)", URL: "https://service.pdok.nl/cbs/wijkenbuurten/2025/wfs/v1_0", License: "onbekend", ProbeKind: "wfs", TypeName: "wijkenbuurten:wijken"}, Provider: "CBS via PDOK", Summary: "Gemeente-, wijk- en buurtgrenzen met officiële gebiedscodes.", Limit: "2025 is het standaard bouwjaar in Python, niet een claim over het jaar van de lokale snapshot. Licentie hier niet vastgesteld.", Access: "Python haalt de grenzen op bij bibliotheekbouw; Go leest de snapshot."},
		CatalogSource{Source: Source{ID: "rotterdam_oppervlaktewatergemalen", Label: "Gemeente Rotterdam — oppervlaktewatergemalen", URL: "https://services.arcgis.com/zP1tGdLpGvt2qNJ6/arcgis/rest/services/Gemalen_derde_ronde/FeatureServer/0", License: "onbekend — geen licentieveld in de dienst", ProbeKind: "arcgis"}, Provider: "Gemeente Rotterdam", Summary: "Punten van oppervlaktewatergemalen, apart van rioolgemalen.", Limit: "De dienst specificeert geen licentie of volledigheid; gebruik de bronmetadata en beschouw de telling niet als compleetheidsbewijs.", Access: "Wordt als aparte laag opgenomen bij Python bibliotheekbouw met --waterketen; Go leest daarna de nieuwe snapshot."},
	)
	c.Layers = []CatalogLayer{
		{ID: "riool", Label: "Rioolbuizen", Official: "Rotterdam-beheer:beheer_leiding", Summary: "Lijnen met de geregistreerde ligging en eigenschappen van rioolbuizen.", Question: "Waar liggen de geregistreerde rioolbuizen?", Geometry: "lijnen", SourceID: "gws_netwerk", Limit: "Geen bewijs van een volledig fysiek netwerk; geen KLIC-vervanging.", Use: "Leesbaar in Go als managed data of kaartbibliotheek; ophalen via Python.", Terms: []string{"gwsw", "wfs", "bob", "nap"}},
		{ID: "drinkwater", Label: "Drinkwaterverzorgingsgebied", Official: "Waterbedrijven_Nederland / FeatureServer/0", Summary: "Gebieden die aangeven welk drinkwaterbedrijf waar levert; geen leidingen.", Question: "Wie levert hier drinkwater?", Geometry: "vlakken", SourceID: "vewin", Limit: "Vertelt niet waar buizen liggen. Zuid-Holland biedt een tweede levering van dezelfde soort gebieden, geen extra netwerk.", Use: "Managed laag in Go; geen automatische omzetting naar een leidingkaart.", Terms: []string{"layer", "arcgis"}},
		{ID: "drinkwaterleidingen", Label: "Drinkwaterleidingen (Evides)", Official: "Kans_opwarming_drinkwater_Rotterdam / FeatureServer/0", Summary: "Gepubliceerde leidinggeometrie van Evides via gemeente Rotterdam.", Question: "Waar liggen de gepubliceerde drinkwaterleidingen?", Geometry: "lijnen", SourceID: "evides", Limit: "Alleen leidingen die volgens de bron niet binnen tien jaar worden vervangen. Geen volledig net; risico is een opwarmingsklasse, geen materiaal of buisdiepte.", Use: "Go leest de lokale levering; Python haalt deze op.", Terms: []string{"feature", "arcgis", "klic"}},
		{ID: "putten", Label: "Rioolputten", Official: "Rotterdam-beheer:beheer_put", Summary: "Geregistreerde putten als kaartobjecten.", Question: "Waar liggen geregistreerde rioolputten?", Geometry: "punten of objectgeometrie uit de levering", SourceID: "gws_aanvullend", Limit: "Een nabijgelegen buis is niet automatisch met de put verbonden; daarvoor zijn objectrelaties nodig.", Use: "Go-kaartbibliotheek, indien deze laag is opgenomen.", Terms: []string{"feature", "gwsw"}},
		{ID: "gemalen", Label: "Gemalen / pompconstructies", Official: "Rotterdam-beheer:beheer_pomp", Summary: "Constructies waarin afvalwater kan worden opgepompt.", Question: "Waar zijn pompconstructies geregistreerd?", Geometry: "punten of objectgeometrie uit de levering", SourceID: "gws_aanvullend", Limit: "Deze bronlaag mengt rioolgemalen en putconstructies; een kaartobject is niet per definitie een gemaal of losse pomp.", Use: "Go-kaartbibliotheek, indien aanwezig. RIONED heeft ook afzonderlijke collecties voor rioolgemaal en rwzi.", Terms: []string{"gwsw", "rwzi"}},
		{ID: "hoogte", Label: "Terreinhoogte (AHN DTM)", Official: "dtm_05m", Summary: "Een raster met terreinhoogtes om met buisbodemhoogtes te vergelijken.", Question: "Hoe ver ligt de buisbodem onder het straatmaaiveld?", Geometry: "raster; geen buizen", SourceID: "ahn", Limit: "BOB alleen is geen diepte. Wateroppervlak en straatniveau moeten apart worden behandeld.", Use: "Alleen de Python-opdracht diepte; deze doet een expliciete download.", Terms: []string{"ahn", "wcs", "nap"}},
		{ID: "gebieden", Label: "Gemeente, wijken en buurten", Official: "wijkenbuurten:gemeenten / wijkenbuurten:wijken / wijkenbuurten:buurten", RequestLayer: "wijkenbuurten:wijken", Summary: "Officiële CBS-gebiedsgrenzen voor ruimtelijke selectie.", Question: "Welke wijk of buurt onderzoek ik?", Geometry: "vlakken", SourceID: "cbs", Limit: "Een bbox is niet de exacte gebiedsgrens. Een zoekvak-telling bewijst geen wijkdekking.", Use: "Gebieden kiezen uit de lokale kaartbibliotheek; bouwen via Python.", Terms: []string{"bbox", "crs"}},
	}
	c.Layers = append(c.Layers,
		CatalogLayer{ID: "damo-leiding", Label: "DAMO leidingen", Official: "waterschappen-waterketen-damo:leiding", Summary: "Geregistreerde waterketenleidingen van waterschappen volgens DAMO.", Question: "Welke waterschapsleidingen zijn gepubliceerd?", Geometry: "lijnen", SourceID: "damo_leiding", Limit: "Kan dezelfde fysieke objecten bevatten als GWSW; tellingen niet optellen als nieuw netwerk.", Use: "Broncontrole in Go; geen afzonderlijke DAMO-leidinglaag in de huidige Go-kaart.", Terms: []string{"damo", "wfs"}},
		CatalogLayer{ID: "rioned-leiding", Label: "RIONED waterschapsleidingen", Official: "waterschap_leiding", Summary: "Een afzonderlijke collectie van waterschapsleidingen volgens GWSW.", Question: "Welke leidingen publiceert RIONED via PDOK?", Geometry: "lijnen", SourceID: "rioned_gwsw", Limit: "Een antwoord met limit=1 is geen totaaltelling en bewijst geen wijkdekking.", Use: "Broncontrole in Go; niet automatisch als extra netwerk op de kaart geladen.", Terms: []string{"rioned", "gwsw", "ogc"}},
		CatalogLayer{ID: "zh-drinkwater-gebied", Label: "ZH drinkwaterverzorgingsgebied", Official: "water:DRINKWATER_BEDRIJVEN", Summary: "Een tweede levering van drinkwaterverzorgingsgebieden, naast Vewin.", Question: "Wie levert hier drinkwater?", Geometry: "vlakken", SourceID: "zh_drinkwater", Limit: "Geen leidingen en geen nieuw netwerk. Deze levering kan dienen als kruiscontrole op Vewin.", Use: "Broncontrole in Go; geen extra leidingkaart.", Terms: []string{"wfs", "layer"}},
	)
	c.Layers = append(c.Layers, CatalogLayer{ID: "rws-regiogebieden", Label: "RWS regiogebieden", Official: "regiogebieden_rijkswaterstaat / MapServer/0", Geometry: "vlakken", SourceID: "rws_regiogebieden",
		Summary: "Beheergrenzen van Rijkswaterstaatdiensten en districten.", Question: "Welke Rijkswaterstaatregio beheert dit gebied?",
		Limit: "Geen CBS-wijkgrenzen, verzorgingsgebied of leidingnet. De dienst publiceert geen licentieveld.",
		Use:   "Broncontrole in Go; geen lokale snapshotlaag zolang deze niet expliciet is opgenomen.", Terms: []string{"arcgis", "layer"}})
	for _, base := range []struct{ id, label, official, summary, geometry string }{
		{"straten", "Straten", "OSM highway", "Straatgeometrie als bovengrondse oriëntatie.", "lijnen"},
		{"water", "Oppervlaktewater", "OSM natural=water / landuse=reservoir", "Vlakken voor water zoals singels, havens en reservoirs.", "vlakken"},
		{"groen", "Groen", "OSM natural / landuse / leisure-selectie", "Geselecteerde groene gebieden zoals parken en bossen.", "vlakken"},
		{"gebouwen", "Gebouwen", "OSM building", "Gebouwcontouren als bovengrondse context.", "vlakken"},
		{"spoor", "Spoor", "OSM railway", "Spoor-, tram- en metrolijnen uit OpenStreetMap.", "lijnen"},
	} {
		c.Layers = append(c.Layers, CatalogLayer{ID: base.id, Label: base.label, Official: base.official, Summary: base.summary, Question: "Wat ligt er boven en rond het netwerk?", Geometry: base.geometry, SourceID: "osm", Limit: "Bovengrondse context, geen ondergrondse leidingen of volledigheidsbewijs.", Use: "Go-kaartbibliotheek, indien geïmporteerd via Python.", Terms: []string{"feature", "gpkg"}})
	}
	for _, extra := range []struct{ id, label, official, summary, geometry, source, limit string }{
		{"damo-put", "DAMO putten", "waterschappen-waterketen-damo:put", "Waterschapsputten als afzonderlijke levering.", "punten", "damo_leiding", "Mogelijke overlap met GWSW; niet automatisch samenvoegen."},
		{"damo-rioolgemaal", "DAMO rioolgemalen", "waterschappen-waterketen-damo:rioolgemaal", "Geregistreerde rioolgemalen van waterschappen.", "punten", "damo_rioolgemaal", "2.815 gemeten. Gemalen zijn geen losse pomp-onderdelen; pomp staat niet in deze WFS."},
		{"damo-rioleringsgebied", "DAMO rioleringsgebieden", "waterschappen-waterketen-damo:rioleringsgebied", "Gebieden die de afvoer van afvalwater beschrijven.", "vlakken", "damo_leiding", "Geen buizennetwerk; een afvoergebied is een vlak."},
		{"damo-zuiveringseenheid", "DAMO zuiveringseenheden", "waterschappen-waterketen-damo:zuiveringseenheid", "Gebieden die als zuiveringseenheid zijn geregistreerd.", "vlakken", "damo_leiding", "Geen RWZI-punten en geen bewezen route van een individuele buis."},
		{"zh-drinkwater-productieketen", "ZH drinkwaterproductieketen", "water:DRINKWATER_PRODUCTIEKETEN", "Productie- en winningslocaties voor drinkwater.", "punten", "zh_drinkwater", "Geen drinkwaterleidingen; PDM geldt voor deze beschreven dataset, niet automatisch voor alle dienstlagen."},
		{"zh-drinkwater-zoekgebieden", "ZH drinkwaterzoekgebieden", "water:DRINKWATER_ZOEKGEBIEDEN", "Zoekruimte voor drinkwaterwinning.", "vlakken", "zh_drinkwater", "Geen verzorgingsgebied of bewijs dat hier al water wordt gewonnen."},
		{"zh-grondwaterbescherming", "ZH grondwaterbeschermingsgebieden", "water:BESCHERMINGSGEBIED_GRONDWATER", "Gebieden voor bescherming van grondwater.", "vlakken", "zh_drinkwater", "Geen leidingnet; de grens toont bescherming, niet de ligging van buizen."},
		{"zh-delfland-transport", "ZH Delfland leidingsegmenten", "water:HHDELFLAND_LEIDINGSEGMENT", "Een afzonderlijke publicatie van transportleidingsegmenten.", "lijnen", "zh_drinkwater", "Spaarzame levering met mogelijke overlap; geen nieuw drinkwaternet. Bbox-matches kunnen buiten de wijk liggen."},
	} {
		c.Layers = append(c.Layers, CatalogLayer{ID: extra.id, Label: extra.label, Official: extra.official, Summary: extra.summary, Question: "Welke aanvullende waterketengegevens zijn gepubliceerd?", Geometry: extra.geometry, SourceID: extra.source, Limit: extra.limit, Use: "Afzonderlijke downloadlaag in Python; niet automatisch geïntegreerd in de Go-kaart of analyses.", Terms: []string{"wfs", "damo", "bbox"}})
	}
	// The IMWA services carry many object types per dataset, and the probe
	// counts exactly one layer. Every probed layer needs its own entry here
	// (catalog_test.go fails on an unexplained probe layer); the other types a
	// waterketen question asks about are added alongside so the register does
	// not silently hide them.
	for _, extra := range []struct{ id, label, official, summary, geometry, source, limit string }{
		{"hwh-kunstwerken-gemaal", "Kunstwerken IMWA — gemalen", "waterschappen-kunstwerken-imwa:gemaal", "Door waterschappen beheerde gemalen in het watersysteem.", "punten", "hwh_kunstwerken_imwa", "8.565 gemeten. Een gemaal hier is niet automatisch een rioolgemaal."},
		{"hwh-kunstwerken-stuw", "Kunstwerken IMWA — stuwen", "waterschappen-kunstwerken-imwa:stuw", "Stuwen die het waterpeil in een watergang regelen.", "punten", "hwh_kunstwerken_imwa", "53.142 gemeten: de grootste groep. Een stuw regelt peil, geen afvalwatertransport."},
		{"hwh-kunstwerken-sluis", "Kunstwerken IMWA — sluizen", "waterschappen-kunstwerken-imwa:sluis", "Sluizen in het waterbeheer en de scheepvaart.", "punten", "hwh_kunstwerken_imwa", "889 gemeten. Geen zuivering of gemaal; een eigen objecttype."},
		{"hwh-kunstwerken-duiker", "Kunstwerken IMWA — duikers", "waterschappen-kunstwerken-imwa:duiker", "Duikers: watergangen die onder een weg of kering door kruisen.", "punten", "hwh_kunstwerken_imwa", "Een duiker is geen rioolbuis, ook al ligt hij ondergronds."},
		{"hwh-kunstwerken-vispassage", "Kunstwerken IMWA — vispassages", "waterschappen-kunstwerken-imwa:vispassage", "Voorzieningen die vismigratie langs een kunstwerk mogelijk maken.", "punten", "hwh_kunstwerken_imwa", "Ecologische inrichting; geen waterketenobject."},
		{"hwh-oppervlaktewaterlichaam", "Oppervlaktewaterlichaam (IMWA)", "waterschappen-oppervlaktewateren-imwa:oppervlaktewaterlichaam", "Landelijke laag van oppervlaktewaterlichamen volgens IMWA.", "vlakken", "hwh_oppervlaktewateren_imwa", "1.430.861 gemeten: beheerssemantiek, geen topografie. Geen debiet of waterstand."},
		{"hwh-waterkering", "Waterkering (IMWA)", "waterschappen-keringen-imwa:waterkering", "Waterkeringen zoals dijken, dammen en kades.", "lijnen en vlakken", "hwh_keringen_imwa", "28.396 gemeten. Registratie is geen actuele toetsstatus."},
		{"hwh-waterkeringsectie", "Waterkeringsectie (IMWA)", "waterschappen-keringen-imwa:waterkeringsectie", "Secties waarin een waterkering is opgedeeld.", "lijnen", "hwh_keringen_imwa", "21.825 gemeten: een sectie hoort bij één kering, dus niet optellen als extra kering."},
		{"hwh-peilgebied", "Peilgebied (IMWA)", "waterschappen-waterbeheergebieden-imwa:peilgebied", "Gebieden met één afgesproken waterpeil.", "vlakken", "hwh_waterbeheergebieden_imwa", "37.418 gemeten. Grenzen vallen niet samen met waterschapsgrenzen."},
		{"hwh-peilbesluitgebied", "Peilbesluitgebied (IMWA)", "waterschappen-waterbeheergebieden-imwa:peilbesluitgebied", "Het deel waarvoor formeel een peilbesluit geldt.", "vlakken", "hwh_waterbeheergebieden_imwa", "3.937 gemeten: het besluitdeel, geen groter peilgebied."},
	} {
		c.Layers = append(c.Layers, CatalogLayer{ID: extra.id, Label: extra.label, Official: extra.official, Summary: extra.summary, Question: "Welk water, welke kering en welk peilgebied liggen bij dit gebied?", Geometry: extra.geometry, SourceID: extra.source, Limit: extra.limit, Use: "Broncontrole in Go; geen automatische kaartlaag of analyse.", Terms: []string{"imwa", "kering", "peilgebied", "wfs"}})
	}
	// DAMO and RIONED each publish a family of object types. The types the
	// register already described (put, rioleringsgebied, zuiveringseenheid) keep
	// their source; these are the ones that were missing.
	for _, extra := range []struct{ id, label, official, summary, geometry, source, limit string }{
		{"damo-rwzi", "DAMO RWZI's", "waterschappen-waterketen-damo:rwzi", "Rioolwaterzuiveringsinstallaties als puntobject.", "punten", "damo_rwzi", "317 gemeten: één punt per installatie, zonder capaciteit of prestatie."},
		{"damo-lozingspunt", "DAMO lozingspunten", "waterschappen-waterketen-damo:lozingspunt", "Lozingen op oppervlaktewater, inclusief overstorten.", "punten", "damo_lozingspunt", "3.236 gemeten. Een lozingspunt bewijst geen overstortwerking."},
		{"damo-overnamepunt", "DAMO overnamepunten", "waterschappen-waterketen-damo:overnamepunt", "Overdracht van gemeentelijk riool naar het waterschapssysteem.", "punten", "damo_overnamepunt", "2.024 gemeten: het koppelpunt, niet de route daarna."},
		{"damo-mantelbuis", "DAMO mantelbuizen", "waterschappen-waterketen-damo:mantelbuis", "Beschermbuis rond een leiding, bijvoorbeeld bij een kruising.", "lijnen", "damo_leiding", "3.603 gemeten. Een mantelbuis is geen transportleiding."},
		{"damo-ontluchter", "DAMO ontluchters", "waterschappen-waterketen-damo:ontluchter", "Ontluchting op een transportleiding.", "punten", "damo_leiding", "4.257 gemeten. Onderdeel van het transportsysteem, geen rioolput."},
		{"damo-leidingsegment", "DAMO leidingsegmenten", "waterschappen-waterketen-damo:leidingsegment", "Segmenten van waterketenleidingen, los van het leidingobject.", "lijnen", "damo_leiding", "28.221 gemeten. Segment en leiding beschrijven deels dezelfde buis; niet optellen."},
		{"rioned-rioolgemaal", "RIONED rioolgemalen", "waterschap_rioolgemaal", "Rioolgemalen als eigen GWSW-collectie.", "punten", "rioned_rioolgemaal", "Alleen liveness: PDOK geeft geen totaaltelling."},
		{"rioned-rwzi", "RIONED RWZI's", "waterschap_rwzi", "RWZI's als eigen GWSW-collectie.", "punten", "rioned_rwzi", "Alleen liveness; procesgegevens staan niet in deze collectie."},
		{"rioned-iba", "RIONED IBA's", "waterschap_iba", "Kleinschalige afvalwaterbehandeling buiten het riool.", "punten", "rioned_iba", "Dekking hangt af van wat de bronhouder aanlevert."},
		{"rioned-put", "RIONED putten", "waterschap_put", "Putten als eigen GWSW-collectie.", "punten", "rioned_gwsw", "Alleen liveness; nabijheid van een buis bewijst geen verbinding."},
		{"rioned-lozingspunt", "RIONED lozingspunten", "waterschap_lozingspunt", "Lozingen als eigen GWSW-collectie.", "punten", "rioned_gwsw", "Losse objecten; geen bewijs van overstortgedrag."},
		{"rioned-overnamepunt", "RIONED overnamepunten", "waterschap_overnamepunt", "Overnamepunten als eigen GWSW-collectie.", "punten", "rioned_gwsw", "Koppelpunt in de keten, geen volledige route."},
	} {
		c.Layers = append(c.Layers, CatalogLayer{ID: extra.id, Label: extra.label, Official: extra.official, Summary: extra.summary, Question: "Welke objectsoorten publiceren de waterschappen en RIONED?", Geometry: extra.geometry, SourceID: extra.source, Limit: extra.limit, Use: "Broncontrole in Go; niet automatisch op de kaart of in analyses.", Terms: []string{"damo", "rioned", "overstort", "iba"}})
	}
	c.Layers = append(c.Layers, CatalogLayer{
		ID: "oppervlaktewatergemalen", Label: "Oppervlaktewatergemalen",
		Official: "Gemalen_derde_ronde / FeatureServer/0",
		Summary:  "Geregistreerde pompinstallaties voor oppervlaktewater, niet rioolgemalen.",
		Question: "Waar liggen oppervlaktewatergemalen?", Geometry: "punten",
		SourceID: "rotterdam_oppervlaktewatergemalen",
		Limit:    "De gemeentelijke dienst leverde bij controle 23 punten van dit type in Rotterdam, waarvan één in de wijk Delfshaven. De dienst vermeldt geen licentie of volledigheid.",
		Use:      "Go-kaartlaag zodra opgenomen in een lokale bibliotheeksnapshot via Python bibliotheekbouw met --waterketen.",
		Terms:    []string{"feature", "arcgis", "oppervlaktewatergemaal"},
	})
	for i := range c.Layers {
		c.Layers[i].Local = []CatalogLocal{}
	}
	return c
}
