package hotkey

import (
	"go/parser"
	"go/token"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestHotkeyPackage_DoesNotImportSyntheticInputInternals(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	dir := filepath.Dir(file)
	matches, err := filepath.Glob(filepath.Join(dir, "*.go"))
	if err != nil {
		t.Fatalf("glob err: %v", err)
	}
	fset := token.NewFileSet()
	for _, path := range matches {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		parsed, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		for _, imp := range parsed.Imports {
			pkg := strings.Trim(imp.Path.Value, "\"")
			if pkg == "goahk/internal/input" {
				pos := fset.PositionFor(imp.Pos(), true)
				t.Fatalf("unexpected input import in trigger path: %s:%d", pos.Filename, pos.Line)
			}
		}
	}
}
