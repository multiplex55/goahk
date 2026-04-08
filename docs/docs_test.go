package docs_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReferencedExamplePathsExist(t *testing.T) {
	t.Parallel()

	repoRoot := filepath.Join("..")
	paths := []string{
		"README.md",
		"docs/USAGE.md",
		"docs/architecture.md",
		"docs/adr/0001-code-first-primary.md",
		"examples/basic-script/main.go",
		"examples/custom-callback/main.go",
		"examples/mixed-actions/main.go",
		"examples/clipboard-transform-paste/main.go",
		"examples/window-aware-script/main.go",
		"cmd/goahk/main.go",
		"internal/config/adapter.go",
	}

	for _, p := range paths {
		p := p
		t.Run(p, func(t *testing.T) {
			t.Parallel()
			if _, err := os.Stat(filepath.Join(repoRoot, p)); err != nil {
				t.Fatalf("expected path %q to exist: %v", p, err)
			}
		})
	}
}

func TestPrimaryQuickStartUsesCodeFirstChain(t *testing.T) {
	t.Parallel()

	readme, err := os.ReadFile(filepath.Join("..", "README.md"))
	if err != nil {
		t.Fatalf("ReadFile(README.md) error = %v", err)
	}
	usage, err := os.ReadFile(filepath.Join("USAGE.md"))
	if err != nil {
		t.Fatalf("ReadFile(USAGE.md) error = %v", err)
	}

	for _, doc := range []struct {
		name string
		text string
	}{
		{name: "README.md", text: string(readme)},
		{name: "docs/USAGE.md", text: string(usage)},
	} {
		doc := doc
		t.Run(doc.name, func(t *testing.T) {
			t.Parallel()
			mustContain(t, doc.text, `"goahk/goahk"`)
			mustContain(t, doc.text, "goahk.NewApp()")
			mustContain(t, doc.text, `Bind("1", goahk.MessageBox("goahk", "You pressed 1"))`)
			mustContain(t, doc.text, `Run(context.Background())`)
		})
	}
}

func mustContain(t *testing.T, text, want string) {
	t.Helper()
	if !strings.Contains(text, want) {
		t.Fatalf("expected content to contain %q", want)
	}
}
