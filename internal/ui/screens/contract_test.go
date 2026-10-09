package screens

import (
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// TestScreensDoNotImportLipgloss enforces contract 1 structurally. A screen
// resolves every colour through ui/components (which owns the tokens and the
// single core→ink translation), so a screen is unable to drift into its own
// style or colour table. Without this, screens.go and kaart.go both grew raw
// lipgloss.NewStyle calls before a review caught them.
func TestScreensDoNotImportLipgloss(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		af, err := parser.ParseFile(fset, f, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("%s: %v", f, err)
		}
		for _, imp := range af.Imports {
			path, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				continue
			}
			if strings.Contains(path, "charmbracelet/lipgloss") {
				t.Errorf("%s importeert lipgloss (%q) — schermen gebruiken ui/components", f, path)
			}
		}
	}
}
