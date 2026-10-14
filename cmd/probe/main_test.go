package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"testing"

	"waterketen/internal/testutil"
)

func TestProbeChecksFeatureIterationError(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "main.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	parents := make(map[ast.Node]ast.Node)
	var stack []ast.Node
	ast.Inspect(file, func(node ast.Node) bool {
		if node == nil {
			stack = stack[:len(stack)-1]
			return false
		}
		if len(stack) > 0 {
			parents[node] = stack[len(stack)-1]
		}
		stack = append(stack, node)
		return true
	})

	found := false
	ast.Inspect(file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || selector.Sel.Name != "Each" {
			return true
		}
		receiver, ok := selector.X.(*ast.Ident)
		if !ok || receiver.Name != "tbl" {
			return true
		}
		found = true
		assignment, ok := parents[call].(*ast.AssignStmt)
		if !ok {
			t.Error("tbl.Each discards its geometry-decoding error")
			return true
		}
		guard, ok := parents[assignment].(*ast.IfStmt)
		if !ok || guard.Init != assignment {
			t.Error("tbl.Each error is not checked immediately")
			return true
		}
		errChecked := false
		ast.Inspect(guard.Cond, func(node ast.Node) bool {
			if ident, ok := node.(*ast.Ident); ok && ident.Name == "err" {
				errChecked = true
			}
			return true
		})
		if !errChecked {
			t.Error("tbl.Each error guard does not inspect err")
		}
		return true
	})
	if !found {
		t.Fatal("probe no longer iterates feature rows")
	}
}

func TestReadFeaturesRejectsMalformedGeometryAfterValidRow(t *testing.T) {
	path := filepath.Join(t.TempDir(), "features.gpkg")
	db := testutil.GeoPackage(t, path, `
		INSERT INTO gpkg_geometry_columns VALUES('features','geom',28992);
		CREATE TABLE features (fid INTEGER PRIMARY KEY, geom BLOB);
	`)
	if _, err := db.Exec(`INSERT INTO features VALUES (1, ?), (2, ?)`, testutil.Point(1, 2), []byte("bad")); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	_, features, err := readFeatures(path, "riool")
	if err == nil {
		t.Fatal("malformed feature geometry was reported as a complete scan")
	}
	if len(features) != 0 {
		t.Fatalf("partial features escaped after failed scan: %d", len(features))
	}
}
