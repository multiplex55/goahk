package repohygiene

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUIAViewerActiveSourcesDoNotReferenceLegacyArtifacts(t *testing.T) {
	t.Parallel()

	root := repoRoot(t)

	include := []string{
		"README.md",
		"docs/BUILD.md",
		"docs/uia-viewer.md",
		"docs/testing-uia-viewer.md",
		"cmd/goahk-uia-viewer",
		"internal/inspect",
	}

	excludeContains := []string{
		"docs/changelog/",
		"docs/adr/",
		"testdata/",
	}

	legacyMarkers := []string{
		"ViewerApp",
		"WailsBoundMethods",
		"wails build",
		"frontend/src",
		"cmd/goahk-uia-viewer/frontend",
		"migration placeholder",
		"not wired",
	}

	for _, target := range include {
		abs := filepath.Join(root, target)
		info, err := os.Stat(abs)
		if err != nil {
			t.Fatalf("Stat(%q): %v", target, err)
		}

		if info.IsDir() {
			err = filepath.WalkDir(abs, func(path string, d os.DirEntry, walkErr error) error {
				if walkErr != nil {
					return walkErr
				}
				if d.IsDir() {
					return nil
				}
				rel, err := filepath.Rel(root, path)
				if err != nil {
					return err
				}
				if !scanableSourcePath(rel) || containsAny(rel, excludeContains) {
					return nil
				}
				assertNoLegacyMarkers(t, root, rel, legacyMarkers)
				return nil
			})
			if err != nil {
				t.Fatalf("WalkDir(%q): %v", target, err)
			}
			continue
		}

		if !scanableSourcePath(target) || containsAny(target, excludeContains) {
			continue
		}
		assertNoLegacyMarkers(t, root, target, legacyMarkers)
	}
}

func assertNoLegacyMarkers(t *testing.T, root, rel string, markers []string) {
	t.Helper()

	b, err := os.ReadFile(filepath.Join(root, rel))
	if err != nil {
		t.Fatalf("ReadFile(%q): %v", rel, err)
	}
	text := strings.ToLower(string(b))
	for _, marker := range markers {
		if strings.Contains(text, strings.ToLower(marker)) {
			t.Fatalf("%s contains legacy UIA viewer marker %q", rel, marker)
		}
	}
}

func scanableSourcePath(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".go", ".md", ".txt", ".ps1", ".bat", ".yml", ".yaml":
		return true
	default:
		return false
	}
}

func containsAny(path string, patterns []string) bool {
	norm := filepath.ToSlash(strings.ToLower(path))
	for _, p := range patterns {
		if strings.Contains(norm, filepath.ToSlash(strings.ToLower(p))) {
			return true
		}
	}
	return false
}
