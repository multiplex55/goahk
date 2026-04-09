package examples_test

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var startMarkerPattern = regexp.MustCompile(`// snippet:start:([a-z0-9-]+)`)

func TestReadmeSnippetsMatchPrimaryExamples(t *testing.T) {
	t.Parallel()

	rootReadme, err := os.ReadFile(filepath.Join("..", "README.md"))
	if err != nil {
		t.Fatalf("read root README: %v", err)
	}

	readmeSnippets := extractMarkedSnippets(t, string(rootReadme), "README.md")

	sources := map[string]string{
		"basic-script-main":        "basic-script/main.go",
		"messagebox-and-exit-main": "messagebox-and-exit/main.go",
		"clipboard-helper-main":    "clipboard-helper/main.go",
		"window-aware-script-main": "window-aware-script/main.go",
	}

	for marker, relPath := range sources {
		sourcePath := filepath.Join(".", relPath)
		body, err := os.ReadFile(sourcePath)
		if err != nil {
			t.Fatalf("read source %s: %v", sourcePath, err)
		}
		sourceSnippets := extractMarkedSnippets(t, string(body), relPath)

		src, ok := sourceSnippets[marker]
		if !ok {
			t.Fatalf("missing snippet marker %q in source %s", marker, relPath)
		}
		readme, ok := readmeSnippets[marker]
		if !ok {
			t.Fatalf("missing snippet marker %q in README.md", marker)
		}

		if normalizeSnippet(src) != normalizeSnippet(readme) {
			t.Fatalf("snippet %q mismatch between README.md and %s", marker, relPath)
		}
	}
}

func extractMarkedSnippets(t *testing.T, content, sourceName string) map[string]string {
	t.Helper()

	content = strings.ReplaceAll(content, "\r\n", "\n")
	result := map[string]string{}
	matches := startMarkerPattern.FindAllStringSubmatchIndex(content, -1)

	for _, idx := range matches {
		marker := content[idx[2]:idx[3]]
		if _, exists := result[marker]; exists {
			t.Fatalf("duplicate snippet marker %q in %s", marker, sourceName)
		}

		startLineEnd := strings.Index(content[idx[0]:], "\n")
		if startLineEnd == -1 {
			t.Fatalf("invalid snippet marker line for %q in %s", marker, sourceName)
		}
		bodyStart := idx[0] + startLineEnd + 1

		endMarker := fmt.Sprintf("// snippet:end:%s", marker)
		bodyEnd := strings.Index(content[bodyStart:], endMarker)
		if bodyEnd == -1 {
			t.Fatalf("missing end marker %q in %s", endMarker, sourceName)
		}

		result[marker] = strings.TrimSuffix(content[bodyStart:bodyStart+bodyEnd], "\n")
	}
	return result
}

func normalizeSnippet(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	lines := strings.Split(strings.TrimSpace(s), "\n")
	for i := range lines {
		lines[i] = strings.TrimSpace(lines[i])
	}
	return strings.Join(lines, "\n")
}
