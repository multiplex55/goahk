package docs_test

import (
	"fmt"
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
		"docs/examples/custom-callback-stop-safe.md",
		"docs/examples/window-inspect.md",
		"docs/examples/move-window.md",
		"docs/examples/mouse-move-and-click.md",
		"docs/examples/long-running-task-with-replace-policy.md",
		"docs/examples/emergency-stop.md",
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
			mustContain(t, doc.text, `goahk.MessageBox(`)
			mustContain(t, doc.text, `Run(context.Background())`)
		})
	}
}

func TestRuntimeReliabilityExampleDocsSections(t *testing.T) {
	t.Parallel()

	root := filepath.Join(".", "examples")
	docs := []string{
		"custom-callback-stop-safe.md",
		"window-inspect.md",
		"move-window.md",
		"mouse-move-and-click.md",
		"long-running-task-with-replace-policy.md",
		"emergency-stop.md",
	}
	for _, name := range docs {
		name := name
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			body, err := os.ReadFile(filepath.Join(root, name))
			if err != nil {
				t.Fatalf("ReadFile(%s) error = %v", name, err)
			}
			text := string(body)
			mustContain(t, text, "## Program snippet")
			mustContain(t, text, "## Expected runtime log sequence")
			mustContain(t, text, "## Cancellation behavior notes")
			mustContain(t, text, "dispatch_")
		})
	}
}

func TestReliabilityExamplesReferenceKnownCommands(t *testing.T) {
	t.Parallel()

	emergency, err := os.ReadFile(filepath.Join(".", "examples", "emergency-stop.md"))
	if err != nil {
		t.Fatalf("ReadFile(emergency-stop.md) error = %v", err)
	}
	mustContain(t, string(emergency), "hard_stop")
	mustContain(t, string(emergency), "Escape")
	mustContain(t, string(emergency), "Shift+Escape")

	replaceDoc, err := os.ReadFile(filepath.Join(".", "examples", "long-running-task-with-replace-policy.md"))
	if err != nil {
		t.Fatalf("ReadFile(long-running-task-with-replace-policy.md) error = %v", err)
	}
	mustContain(t, string(replaceDoc), `"concurrency": "replace"`)
	mustContain(t, string(replaceDoc), "policy_replace_cancel_running")
}

func mustContain(t *testing.T, text, want string) {
	t.Helper()
	if !strings.Contains(text, want) {
		t.Fatalf("expected content to contain %q", want)
	}
}

func mustContainAll(t *testing.T, text string, wants ...string) {
	t.Helper()
	for _, want := range wants {
		if !strings.Contains(text, want) {
			t.Fatalf("expected content to contain %q", want)
		}
	}
}

func TestArchitectureRuntimePlaneSectionsPresent(t *testing.T) {
	t.Parallel()

	body, err := os.ReadFile(filepath.Join(".", "architecture.md"))
	if err != nil {
		t.Fatalf("ReadFile(architecture.md) error = %v", err)
	}
	mustContainAll(t, string(body),
		"control plane",
		"work plane",
		"Callback lifecycle contract",
		"Policy semantics",
	)
}

func TestUsageReliabilityLinksStayCurrent(t *testing.T) {
	t.Parallel()
	usage, err := os.ReadFile(filepath.Join(".", "USAGE.md"))
	if err != nil {
		t.Fatalf("ReadFile(USAGE.md) error = %v", err)
	}
	for _, slug := range []string{
		"custom-callback-stop-safe",
		"window-inspect",
		"move-window",
		"mouse-move-and-click",
		"long-running-task-with-replace-policy",
		"emergency-stop",
	} {
		mustContain(t, string(usage), fmt.Sprintf("docs/examples/%s.md", slug))
	}
}
