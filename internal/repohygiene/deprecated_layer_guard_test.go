package repohygiene

import (
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestNoNewImportsIntoDeprecatedInternalAppLayer(t *testing.T) {
	t.Helper()

	root := repoRoot(t)
	cmd := exec.Command("git", "ls-files", "*.go")
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git ls-files failed: %v\n%s", err, output)
	}

	allowedBridgePoints := map[string]bool{
		"internal/app/bootstrap.go": true,
		"internal/app/lifecycle.go": true,
		"internal/app/runtime.go":   true,
		"internal/app/doc.go":       true,
	}

	fset := token.NewFileSet()
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		rel := strings.TrimSpace(line)
		if rel == "" || allowedBridgePoints[rel] {
			continue
		}
		abs := filepath.Join(root, rel)
		src, err := os.ReadFile(abs)
		if err != nil {
			if os.IsNotExist(err) {
				// File deleted in working tree but not yet staged.
				continue
			}
			t.Fatalf("ReadFile(%q) error = %v", rel, err)
		}
		file, err := parser.ParseFile(fset, abs, src, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("ParseFile(%q) error = %v", rel, err)
		}
		for _, imp := range file.Imports {
			path := strings.Trim(imp.Path.Value, `"`)
			if path == "goahk/internal/app" {
				t.Fatalf("%s imports deprecated layer goahk/internal/app; route through goahk/internal/runtime instead", rel)
			}
		}
	}
}
