package repohygiene

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNoWailsModuleReferences(t *testing.T) {
	t.Helper()

	root := repoRoot(t)
	for _, rel := range []string{"go.mod", "go.sum"} {
		path := filepath.Join(root, rel)
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("ReadFile(%q) error = %v", path, err)
		}
		text := strings.ToLower(string(b))
		for _, forbidden := range []string{"wails", "github.com/wailsapp/wails"} {
			if strings.Contains(text, forbidden) {
				t.Fatalf("%s contains forbidden module reference %q", rel, forbidden)
			}
		}
	}
}
