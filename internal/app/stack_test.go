package app

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestUIStackUsesCharmV2(t *testing.T) {
	for _, root := range []string{".", "../ui"} {
		err := filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() || !strings.HasSuffix(path, ".go") {
				return nil
			}
			parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
			if err != nil {
				return err
			}
			for _, imp := range parsed.Imports {
				name, err := strconv.Unquote(imp.Path.Value)
				if err != nil {
					return err
				}
				for _, module := range []string{"bubbletea", "bubbles", "lipgloss", "glamour"} {
					if strings.HasPrefix(name, "github.com/charmbracelet/"+module) {
						t.Errorf("%s imports legacy Charm module %s", path, name)
					}
				}
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}
