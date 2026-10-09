package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestDiscoverMarkdownExportsSharedExplanations(t *testing.T) {
	cmd := discoverCmd()
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetArgs([]string{"--markdown", "--data-root", t.TempDir()})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"# Waterketen ontdekken", "## Vragen", "## Bronnen", "## Lagen", "## Begrippen", "BOB — binnenonderkant buis", "Waterbedrijven_Nederland"} {
		if !strings.Contains(output.String(), want) {
			t.Errorf("documentation misses %q", want)
		}
	}
}

func TestDiscoverRejectsAmbiguousOrUnknownSelection(t *testing.T) {
	for _, args := range [][]string{{"--vraag", "drinkwater", "--laag", "riool"}, {"--vraag", "niet-bekend"}, {"--overzicht", "niet-bekend"}} {
		cmd := discoverCmd()
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})
		cmd.SetArgs(append(args, "--data-root", t.TempDir()))
		if err := cmd.Execute(); err == nil {
			t.Errorf("accepted invalid selection %v", args)
		}
	}
}

func TestDiscoverQuestionKeepsOfficialNamesAndExplainsLimits(t *testing.T) {
	cmd := discoverCmd()
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetArgs([]string{"--vraag", "drinkwater", "--data-root", t.TempDir()})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Wie levert hier drinkwater?", "Waterbedrijven_Nederland", "geen leidingen", "Vewin", "onbekend"} {
		if !strings.Contains(output.String(), want) {
			t.Errorf("uitleg mist %q:\n%s", want, output.String())
		}
	}
	if strings.Contains(output.String(), "https://") {
		t.Error("API-adressen horen in de technische verdieping, niet in de korte uitleg")
	}
}
