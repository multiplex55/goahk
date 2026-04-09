package runtime

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRuntimePackage_DoesNotDependOnConfigSchemaTypes(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("Glob() error = %v", err)
	}

	fset := token.NewFileSet()
	for _, path := range files {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		src, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("ReadFile(%q) error = %v", path, err)
		}
		if strings.Contains(string(src), "internal/config/schema") {
			t.Fatalf("%s imports internal/config/schema; runtime must stay schema-agnostic", path)
		}

		file, err := parser.ParseFile(fset, path, src, parser.ParseComments)
		if err != nil {
			t.Fatalf("ParseFile(%q) error = %v", path, err)
		}
		ast.Inspect(file, func(n ast.Node) bool {
			sel, ok := n.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			ident, ok := sel.X.(*ast.Ident)
			if !ok || ident.Name != "config" {
				return true
			}
			switch sel.Sel.Name {
			case "Config", "HotkeyBinding", "Step", "Flow", "FlowStep", "UIASelector":
				t.Fatalf("%s references config schema type config.%s; keep schema usage inside internal/config/adapter.go", path, sel.Sel.Name)
			}
			return true
		})
	}
}
