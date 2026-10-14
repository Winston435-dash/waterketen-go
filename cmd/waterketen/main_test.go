package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"waterketen/internal/core"
	"waterketen/internal/testutil"
)

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, io.ErrClosedPipe }

func assetFixture(t *testing.T, date string, broken bool) string {
	t.Helper()
	root := t.TempDir()
	path := filepath.Join(root, "layer.gpkg")
	schema := `INSERT INTO gpkg_geometry_columns VALUES('features','geom',28992);`
	if !broken {
		schema += `CREATE TABLE features(fid INTEGER PRIMARY KEY,geom BLOB,BOB_ACTUEE REAL,BOB_ACTU_1 REAL);`
	}
	db := testutil.GeoPackage(t, path, schema)
	if !broken {
		if _, err := db.Exec(`INSERT INTO features VALUES(1,?,-2,-4),(2,?,0,-2)`, testutil.Point(0, 0), testutil.Point(10, 10)); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Mkdir(filepath.Join(root, "metadata"), 0700); err != nil {
		t.Fatal(err)
	}
	meta, err := json.Marshal(core.AssetMeta{Area: "test", Layer: "riool", ProcessedPath: path, FetchedAt: date})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "metadata", "asset.json"), meta, 0600); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestStatusHandlesMissingAndInvalidDates(t *testing.T) {
	for _, date := range []string{"", "bad", "2026-10-13T10:00:00Z"} {
		cmd := statusCmd()
		var output bytes.Buffer
		cmd.SetOut(&output)
		cmd.SetArgs([]string{"--data-root", assetFixture(t, date, false)})
		if err := cmd.Execute(); err != nil {
			t.Fatal(err)
		}
		want := "datum onbekend"
		if date != "" && date != "bad" {
			want = "2026-10-13"
		}
		if !strings.Contains(output.String(), want) {
			t.Fatalf("date %q: %s", date, output.String())
		}
	}
}

func TestTextCommandsReturnOutputWriteErrors(t *testing.T) {
	root := assetFixture(t, "", false)
	cases := []struct {
		name string
		cmd  *cobra.Command
		args []string
	}{
		{"sources", sourcesCmd(), nil},
		{"status", statusCmd(), []string{"--data-root", root}},
		{"analyse", analyseCmd(), []string{"--gebied", "test", "--laag", "riool", "--data-root", root}},
		{"braille", brailleCmd(), []string{"--gebied", "test", "--laag", "riool", "--data-root", root}},
		{"discover", discoverCmd(), []string{"--data-root", root, "--overzicht", "vragen"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.cmd.SetOut(failingWriter{})
			tc.cmd.SetArgs(tc.args)
			if err := tc.cmd.Execute(); !errors.Is(err, io.ErrClosedPipe) {
				t.Fatalf("Execute() error = %v, want %v", err, io.ErrClosedPipe)
			}
		})
	}
}

func TestReadCommandsRejectUnexpectedArguments(t *testing.T) {
	root := assetFixture(t, "", false)
	cases := []struct {
		name string
		cmd  *cobra.Command
		args []string
	}{
		{"sources", sourcesCmd(), []string{"extra"}},
		{"status", statusCmd(), []string{"--data-root", root, "extra"}},
		{"analyse", analyseCmd(), []string{"--gebied", "test", "--laag", "riool", "--data-root", root, "extra"}},
		{"braille", brailleCmd(), []string{"--gebied", "test", "--laag", "riool", "--data-root", root, "extra"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.cmd.SetOut(&bytes.Buffer{})
			tc.cmd.SetErr(&bytes.Buffer{})
			tc.cmd.SetArgs(tc.args)
			if err := tc.cmd.Execute(); err == nil {
				t.Fatal("unexpected positional argument was accepted")
			}
		})
	}
}

func TestAnalyseUsesSharedSummaryAndPropagatesReadErrors(t *testing.T) {
	cmd := analyseCmd()
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetArgs([]string{"--gebied", "test", "--laag", "riool", "--json", "--data-root", assetFixture(t, "", false)})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	var result map[string]any
	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result["features"] != float64(2) || result["bob_measured"] != float64(2) {
		t.Fatalf("wrong summary: %v", result)
	}
	for _, cmd := range []*cobra.Command{analyseCmd(), brailleCmd()} {
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})
		cmd.SetArgs([]string{"--gebied", "test", "--laag", "riool", "--data-root", assetFixture(t, "", true)})
		if err := cmd.Execute(); err == nil {
			t.Fatal("unreadable layer reported success")
		}
	}
}

func TestSourcesJSONKeepsZeroCount(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, `{"count":0}`) }))
	defer server.Close()
	previous := core.WiredSources
	core.WiredSources = []core.Source{{ID: "test", URL: server.URL, ProbeKind: "arcgis"}}
	t.Cleanup(func() { core.WiredSources = previous })
	cmd := sourcesCmd()
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetArgs([]string{"--json", "--live"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	var result []map[string]any
	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if count, exists := result[0]["count"]; !exists || count != float64(0) {
		t.Fatalf("zero count omitted: %v", result[0])
	}
}
