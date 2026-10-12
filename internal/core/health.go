package core

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// HealthStatus is the probe verdict for one source.
type HealthStatus string

const (
	HealthUp       HealthStatus = "up"
	HealthDown     HealthStatus = "down"
	HealthNotWired HealthStatus = "not-wired"
)

// SourceHealth is the result of one live probe against one endpoint. A probe
// never errors out of this package: an unreachable source is a record with
// Status=down, so the panel shows one broken source without losing the rest.
type SourceHealth struct {
	SourceID  string
	Label     string
	URL       string
	License   string
	Status    HealthStatus
	LatencyMS int
	Count     int // -1 = not available
	Error     string
	Note      string // only when Status == HealthNotWired
}

var numberMatchedRe = regexp.MustCompile(`(?:numberMatched|numberOfFeatures|totalFeatures)="(\d+)"`)

// probeURL builds the cheapest "am I alive and how much is there" request.
func probeURL(src Source) string {
	sep := "?"
	if strings.Contains(src.URL, "?") {
		sep = "&"
	}
	switch src.ProbeKind {
	case "arcgis":
		return src.URL + "/query?where=1%3D1&returnCountOnly=true&f=json"
	case "ogc":
		// OGC API Features: /collections/<id>/items. The collection id comes
		// from the source (TypeName), so a new OGC source needs no code change.
		return fmt.Sprintf("%s/collections/%s/items?limit=1&f=json", src.URL, src.TypeName)
	case "wfs":
		// GeoServer's typeName is per source. `beheer_leiding` is the pipeline's
		// network layer; fall back to it only when a source did not name one.
		name := src.TypeName
		if name == "" {
			name = "Rotterdam-beheer:beheer_leiding"
		}
		return src.URL + sep + "service=WFS&version=2.0.0&request=GetFeature" +
			"&typeNames=" + name + "&resultType=hits"
	case "wfs_damo":
		// DAMO has its own workspace; the GWSW layer name is invalid here (HTTP 400).
		name := src.TypeName
		if name == "" {
			name = "waterschappen-waterketen-damo:leiding"
		}
		return src.URL + sep + "service=WFS&version=2.0.0&request=GetFeature" +
			"&typeNames=" + name + "&resultType=hits"
	}
	return src.URL
}

// checkOGC validates that an OGC API Features items response is real JSON with
// a features array — a liveness check only. It deliberately returns no count:
// PDOK's OGC services omit numberMatched, and a probe that requests limit=1
// would report a meaningless "1". Unknown is reported as -1, shown as "—".
func checkOGC(text string) (int, error) {
	var payload struct {
		Features *[]json.RawMessage `json:"features"`
	}
	if err := json.Unmarshal([]byte(text), &payload); err != nil {
		return -1, fmt.Errorf("geen geldige JSON in OGC-antwoord: %.80s", text)
	}
	if payload.Features == nil {
		return -1, fmt.Errorf("geen features in OGC-antwoord: %.80s", text)
	}
	return -1, nil // alive, total unknown
}

// countFromText extracts the feature count from JSON (ArcGIS) or XML (WFS:
// GeoServer answers resultType=hits with XML even when json is requested).
func countFromText(text string, kind string) (int, error) {
	if kind == "arcgis" {
		var payload struct {
			Count *int `json:"count"`
			Error json.RawMessage `json:"error"`
		}
		if err := json.Unmarshal([]byte(text), &payload); err != nil {
			return -1, fmt.Errorf("geen geldige JSON in probe-antwoord: %.80s", text)
		}
		if payload.Count == nil || *payload.Count < 0 || len(payload.Error) > 0 && string(payload.Error) != "null" {
			return -1, fmt.Errorf("geen geldige telling in ArcGIS-antwoord: %.80s", text)
		}
		return *payload.Count, nil
	}
	if kind == "ogc" {
		return checkOGC(text)
	}
	m := numberMatchedRe.FindStringSubmatch(text)
	if m == nil {
		return -1, fmt.Errorf("geen numberMatched in WFS-antwoord: %.80s", text)
	}
	var n int
	if _, err := fmt.Sscanf(m[1], "%d", &n); err != nil {
		return -1, err
	}
	return n, nil
}

var probeClient = &http.Client{Timeout: 25 * time.Second}

// ProbeSource probes one source. Returns the record (down on error) — never raises.
func ProbeSource(src Source) SourceHealth {
	started := time.Now()
	req, err := http.NewRequest(http.MethodGet, probeURL(src), nil)
	if err != nil {
		return down(src, started, err)
	}
	req.Header.Set("User-Agent",
		"waterketen-cli/1.0 (+https://github.com/mvanhorn/printing-press-library)")
	resp, err := probeClient.Do(req)
	if err != nil {
		return down(src, started, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return down(src, started, fmt.Errorf("HTTP %d", resp.StatusCode))
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return down(src, started, err)
	}
	count, err := countFromText(string(body), src.ProbeKind)
	if err != nil {
		return down(src, started, err)
	}
	return SourceHealth{
		SourceID:  src.ID,
		Label:     src.Label,
		URL:       src.URL,
		License:   src.License,
		Status:    HealthUp,
		LatencyMS: int(time.Since(started).Milliseconds()),
		Count:     count,
	}
}

func down(src Source, started time.Time, err error) SourceHealth {
	msg := err.Error()
	if len(msg) > 200 {
		msg = msg[:200]
	}
	return SourceHealth{
		SourceID:  src.ID,
		Label:     src.Label,
		URL:       src.URL,
		License:   src.License,
		Status:    HealthDown,
		LatencyMS: int(time.Since(started).Milliseconds()),
		Count:     -1,
		Error:     msg,
	}
}

// UnwiredHealth returns the display-only records for the detected-but-not-wired
// sources. No network access; Count is always -1. Both ProbeAll and the TUI's
// incremental probe go through this, so the two cannot drift.
func UnwiredHealth() []SourceHealth {
	out := make([]SourceHealth, 0, len(UnwiredSources))
	for _, u := range UnwiredSources {
		out = append(out, SourceHealth{
			SourceID: u.ID,
			Label:    u.Label,
			URL:      u.URL,
			License:  u.License,
			Status:   HealthNotWired,
			Count:    -1,
			Note:     u.Note,
		})
	}
	return out
}

// AssembleHealth orders the per-source probe results and appends the unwired
// records — the single definition of "the health list", shared by ProbeAll and
// the TUI (which collects the wired probes one message at a time).
func AssembleHealth(wired []SourceHealth) []SourceHealth {
	out := append([]SourceHealth(nil), wired...)
	return append(out, UnwiredHealth()...)
}

// ProbeAll probes every wired source in parallel; order in the result follows
// WiredSources, independent of completion order.
func ProbeAll() []SourceHealth {
	results := make([]SourceHealth, len(WiredSources))
	done := make(chan struct {
		i int
		h SourceHealth
	}, len(WiredSources))
	for i, src := range WiredSources {
		go func(i int, s Source) {
			done <- struct {
				i int
				h SourceHealth
			}{i, ProbeSource(s)}
		}(i, src)
	}
	for range WiredSources {
		r := <-done
		results[r.i] = r.h
	}
	return AssembleHealth(results)
}
