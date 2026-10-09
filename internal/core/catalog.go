package core

import (
	"fmt"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"
)

// Catalog is explanatory, not a fetch registry. Reading it never probes a service
// or downloads data; being described here does not make a layer map-ready.
type Catalog struct {
	Questions []CatalogQuestion `json:"questions"`
	Sources   []CatalogSource   `json:"sources"`
	Layers    []CatalogLayer    `json:"layers"`
	Terms     []CatalogTerm     `json:"terms"`
}

type CatalogQuestion struct {
	ID      string   `json:"id"`
	Title   string   `json:"title"`
	Summary string   `json:"summary"`
	Layers  []string `json:"layers"`
	Terms   []string `json:"terms"`
}

type CatalogSource struct {
	Source
	Provider string `json:"provider"`
	Summary  string `json:"summary"`
	Limit    string `json:"limit"`
	Access   string `json:"access"`
}

type CatalogLayer struct {
	ID           string         `json:"id"`
	Label        string         `json:"label"`
	Official     string         `json:"official_name"`
	RequestLayer string         `json:"request_layer,omitempty"` // representative layer for a grouped catalog entry
	Summary      string         `json:"summary"`
	Question     string         `json:"question"`
	Geometry     string         `json:"geometry"`
	SourceID     string         `json:"source_id"`
	Limit        string         `json:"limit"`
	Use          string         `json:"use"`
	Terms        []string       `json:"terms"`
	Local        []CatalogLocal `json:"local"`
}

// Counts and dates are explicitly metadata, not claims of physical completeness.
type CatalogLocal struct {
	Area       string `json:"area"`
	OnDisk     bool   `json:"on_disk"`
	Count      int    `json:"metadata_count"`
	FetchedAt  string `json:"fetched_at"`
	SourceDate string `json:"source_date"`
	CRS        string `json:"crs"`
	Status     string `json:"status"`
	SourceURL  string `json:"source_url"`
	License    string `json:"license"`
}

type CatalogTerm struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Meaning string `json:"meaning"`
	Example string `json:"example"`
}

type CatalogSection int

const (
	CatalogQuestions CatalogSection = iota
	CatalogSources
	CatalogLayers
	CatalogTerms
)

var CatalogSections = []string{"Vragen", "Bronnen", "Lagen", "Begrippen"}

type CatalogEntry struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Summary string `json:"summary"`
}
type CatalogLine struct {
	Text string `json:"text"`
	Kind string `json:"kind"` // heading, warning, meta, or plain
}

func (c Catalog) Entries(section CatalogSection) []CatalogEntry {
	var out []CatalogEntry
	switch section {
	case CatalogQuestions:
		for _, q := range c.Questions {
			out = append(out, CatalogEntry{q.ID, q.Title, q.Summary})
		}
	case CatalogSources:
		for _, s := range c.Sources {
			out = append(out, CatalogEntry{s.ID, s.Label, s.Summary})
		}
	case CatalogLayers:
		for _, l := range c.Layers {
			out = append(out, CatalogEntry{l.ID, l.Label, l.Summary})
		}
	case CatalogTerms:
		for _, t := range c.Terms {
			out = append(out, CatalogEntry{t.ID, t.Name, t.Meaning})
		}
	}
	return out
}

// LoadCatalog reads only existing metadata and file presence, never geometries.
func LoadCatalog(root string) (Catalog, error) {
	c := NewCatalog()
	assets, err := LoadAssets(root)
	if err != nil {
		return c, err
	}
	lib, err := OpenLibrary(root)
	if err != nil {
		return c, err
	}
	for i := range c.Layers {
		l := &c.Layers[i]
		l.Local = []CatalogLocal{}
		for _, a := range assets {
			if a.Layer == l.ID {
				l.Local = append(l.Local, CatalogLocal{Area: a.Area, OnDisk: a.ProcessedExists(), Count: a.FeatureCount, FetchedAt: a.FetchedAt, SourceURL: a.SourceURL, License: a.SourceLicense, CRS: a.CRS, Status: a.Status})
			}
		}
		if lib != nil {
			if rec, ok := lib.Manifest.Layers[l.ID]; ok {
				_, statErr := os.Stat(lib.LayerPath(l.ID))
				l.Local = append(l.Local, CatalogLocal{Area: "kaartbibliotheek / " + lib.Snapshot, OnDisk: statErr == nil, Count: rec.FeatureCount, FetchedAt: lib.Manifest.CreatedAt, SourceDate: rec.Source.Date, SourceURL: rec.Source.URL, License: rec.Source.License, CRS: lib.Manifest.CRS, Status: "snapshot; datum = publicatie, niet brondatum"})
			}
		}
	}
	return c, nil
}

func (c Catalog) Explain(section CatalogSection, id string, technical bool) ([]CatalogLine, error) {
	var out []CatalogLine
	add := func(text, kind string) { out = append(out, CatalogLine{Text: text, Kind: kind}) }
	addTerms := func(ids []string) {
		if len(ids) == 0 {
			return
		}
		add("Begrippen bij dit antwoord", "heading")
		for _, id := range ids {
			for _, term := range c.Terms {
				if term.ID == id {
					add(term.Name+" — "+term.Meaning, "plain")
				}
			}
		}
	}
	switch section {
	case CatalogQuestions:
		for _, q := range c.Questions {
			if q.ID != id {
				continue
			}
			add(q.Title, "heading")
			add(q.Summary, "plain")
			for _, layerID := range q.Layers {
				lines, err := c.Explain(CatalogLayers, layerID, technical)
				if err != nil {
					return nil, err
				}
				out = append(out, lines...)
			}
			addTerms(q.Terms)
			return out, nil
		}
	case CatalogLayers:
		for _, l := range c.Layers {
			if l.ID != id {
				continue
			}
			add(l.Label, "heading")
			add("Laag-ID: "+l.ID+" · officieel: "+l.Official, "meta")
			add("Wat is het? "+l.Summary, "plain")
			add("Beantwoordt: "+l.Question, "plain")
			add("Op de kaart: "+l.Geometry, "plain")
			add("Wat vertelt het niet? "+l.Limit, "warning")
			add("In de tool: "+l.Use, "meta")
			for _, s := range c.Sources {
				if s.ID == l.SourceID {
					add("Herkomst: "+s.Provider+" → "+s.Label, "plain")
					add("Licentie: "+s.License+" — "+licenseExplanation(s.License), "plain")
				}
			}
			add("Lokaal aanwezig / actualiteit", "heading")
			if len(l.Local) == 0 {
				add("Geen lokale levering gevonden. Brondatum en dekking zijn onbekend.", "meta")
			}
			for _, local := range l.Local {
				state := "metadata zonder bestand"
				if local.OnDisk {
					state = "bestand aanwezig"
				}
				add(fmt.Sprintf("%s: %s; %d objecten volgens metadata", local.Area, state, local.Count), "plain")
				add("Opgehaald/gepubliceerd: "+catalogDate(local.FetchedAt)+"; brondatum: "+catalogDate(local.SourceDate)+"; CRS: "+known(local.CRS), "meta")
				add("Verwerking: "+known(local.Status)+". Dit bewijst geen volledig fysiek netwerk.", "meta")
				add("Licentie lokale levering: "+known(local.License), "meta")
				if technical && local.SourceURL != "" {
					add("Herkomstadres lokale levering: "+local.SourceURL, "meta")
				}
			}
			addTerms(l.Terms)
			if technical {
				for _, s := range c.Sources {
					if s.ID != l.SourceID {
						continue
					}
					source := s.Source
					if source.ProbeKind == "wfs" || source.ProbeKind == "wfs_damo" || source.ProbeKind == "ogc" {
						source.TypeName = l.Official
						if l.RequestLayer != "" {
							source.TypeName = l.RequestLayer
							add("Voorbeeldlaag in deze aanvraag: "+l.RequestLayer, "meta")
						}
					}
					out = append(out, catalogRequestLines(source)...)
				}
			}
			return out, nil
		}
	case CatalogSources:
		for _, s := range c.Sources {
			if s.ID != id {
				continue
			}
			add(s.Label, "heading")
			add("Bron-ID: "+s.ID+" · aanbieder: "+s.Provider, "meta")
			add(s.Summary, "plain")
			add("Beperking: "+s.Limit, "warning")
			add("Toegang in de tool: "+s.Access, "plain")
			add("Licentie: "+s.License+" — "+licenseExplanation(s.License), "plain")
			add("Dienst bereikbaar ≠ gedownload ≠ zichtbaar op de kaart ≠ volledig netwerk.", "warning")
			for _, l := range c.Layers {
				if l.SourceID == s.ID {
					add("Laag: "+l.Label+" ("+l.ID+") — "+l.Summary, "plain")
				}
			}
			if technical {
				out = append(out, catalogRequestLines(s.Source)...)
			}
			return out, nil
		}
	case CatalogTerms:
		for _, t := range c.Terms {
			if t.ID != id {
				continue
			}
			add(t.Name, "heading")
			add(t.Meaning, "plain")
			add("Voorbeeld: "+t.Example, "plain")
			return out, nil
		}
	}
	return nil, fmt.Errorf("onbekend catalogusitem %q; bekijk eerst de lijst", id)
}

func known(s string) string {
	if s == "" {
		return "onbekend"
	}
	return s
}

// Preserve raw dates in JSON; render timestamps in the user's Amsterdam timezone.
func catalogDate(value string) string {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return known(value)
	}
	zone, err := time.LoadLocation("Europe/Amsterdam")
	if err != nil {
		return value
	}
	return parsed.In(zone).Format("02-01-2006 15:04 MST")
}

func licenseExplanation(s string) string {
	switch {
	case s == "CC0":
		return "vrij hergebruik; bewaar de herkomst"
	case strings.Contains(s, "CC BY-SA"):
		return "naamsvermelding en gelijk delen bij afgeleide werken"
	case strings.Contains(s, "CC BY-NC-ND"):
		return "geen commercieel gebruik en geen afgeleide werken"
	case strings.Contains(s, "PDM"):
		return "publiek-domeinmarkering; geen auteursrechtelijke beperkingen bekend"
	case s == "ODbL":
		return "naamsvermelding; voor afgeleide databanken kunnen deelvoorwaarden gelden"
	default:
		return "gebruiksrechten niet vastgesteld; controleer de voorwaarden bij de bron"
	}
}

// Show the actual probe request, not a bare service URL that may return HTTP 400.
// This function only builds descriptions; it never sends these requests.
func catalogRequestLines(s Source) []CatalogLine {
	out := []CatalogLine{{"Technische verdieping", "heading"}, {"Basisadres (parameters zijn nog nodig): " + s.URL, "meta"}}
	if s.ProbeKind == "" {
		if strings.HasSuffix(s.URL, ".gpkg") {
			return append(out, CatalogLine{"Bestandsdownload: " + s.URL, "plain"}, CatalogLine{"Een groot GeoPackage-bestand; alleen expliciet downloaden. Geen live status of telling gecontroleerd.", "warning"})
		}
		return append(out, CatalogLine{"Geen objectaanvraag vastgelegd; dit is broninformatie, geen beloofde downloadactie.", "meta"})
	}
	request := probeURL(s)
	response := "XML met een objecttelling; geen geometrie"
	if s.ProbeKind == "arcgis" {
		response = "JSON met een objecttelling; geen geometrie"
	}
	if s.ProbeKind == "ogc" {
		response = "JSON met maximaal één kaartobject; totaalaantal onbekend"
	}
	out = append(out, CatalogLine{"Volledige aanvraag: " + request, "plain"}, CatalogLine{"Antwoord: " + response, "plain"}, CatalogLine{"Bereik: de opgegeven bronlaag, zonder wijkfilter. Geen bewijs van wijkdekking.", "warning"})
	u, err := url.Parse(request)
	if err == nil {
		meanings := map[string]string{"service": "standaard waarmee we de dienst aanspreken", "version": "versie van die standaard", "request": "soort vraag aan de dienst", "typeNames": "officiële laag-ID", "resultType": "hits telt alleen, zonder de objecten te downloaden", "where": "1=1 selecteert alle objecten", "returnCountOnly": "true vraagt alleen een telling", "f": "formaat van het antwoord", "limit": "maximum aantal objecten in dit antwoord, niet het totaalaantal"}
		keys := make([]string, 0, len(u.Query()))
		for k := range u.Query() {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			out = append(out, CatalogLine{k + "=" + u.Query().Get(k) + " — " + known(meanings[k]), "meta"})
		}
	}
	if s.ProbeKind == "ogc" {
		out = append(out, CatalogLine{"/collections/" + s.TypeName + "/items: objecten uit deze collectie (gegevenslaag).", "meta"})
	}
	return out
}
