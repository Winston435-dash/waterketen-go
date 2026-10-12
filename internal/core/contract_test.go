package core

import (
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestCoreDoesNotImportUI(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		parsed, err := parser.ParseFile(token.NewFileSet(), file, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, imp := range parsed.Imports {
			path, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				t.Fatal(err)
			}
			if strings.HasPrefix(path, "waterketen/internal/ui") || strings.HasPrefix(path, "waterketen/internal/app") || strings.Contains(path, "charmbracelet/lipgloss") || strings.Contains(path, "charmbracelet/bubbletea") || strings.Contains(path, "charmbracelet/bubbles") {
				t.Errorf("%s imports UI dependency %s", file, path)
			}
		}
	}
}
