package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestViewerDependsOnInspectBoundaryOnly(t *testing.T) {
	forbidden := []string{
		"goahk/internal/uia",
		"goahk/internal/window",
		"goahk/internal/input",
		"goahk/internal/clipboard",
	}

	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("glob failed: %v", err)
	}
	fset := token.NewFileSet()
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		node, err := parser.ParseFile(fset, file, src, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parse %s: %v", file, err)
		}
		for _, imp := range node.Imports {
			path := strings.Trim(imp.Path.Value, "\"")
			for _, ban := range forbidden {
				if path == ban {
					t.Fatalf("%s imports forbidden backend package %q; use internal/inspect service boundary", file, path)
				}
			}
		}
	}

	_ = ast.Inspect
}
