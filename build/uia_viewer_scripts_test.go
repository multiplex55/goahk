package build

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUIAViewerScriptsExist(t *testing.T) {
	t.Parallel()

	for _, rel := range []string{
		"build.bat",
		"dev-uia-viewer.bat",
		"build-uia-viewer.bat",
		"check-no-source-binaries.bat",
		"dev-uia-viewer.sh",
		"build-uia-viewer.sh",
	} {
		rel := rel
		t.Run(rel, func(t *testing.T) {
			t.Parallel()
			if _, err := os.Stat(filepath.Join(".", rel)); err != nil {
				t.Fatalf("expected script %q to exist: %v", rel, err)
			}
		})
	}
}

func TestViewerDirectoryHasNoFrontendArtifacts(t *testing.T) {
	t.Parallel()

	viewerDir := filepath.Join("..", "cmd", "goahk-uia-viewer")

	blockedExactPaths := []string{
		"frontend",
		"wails.json",
		"package.json",
		"package-lock.json",
		"pnpm-lock.yaml",
		"yarn.lock",
		"vite.config.js",
		"vite.config.ts",
	}

	for _, rel := range blockedExactPaths {
		blocked := filepath.Join(viewerDir, rel)
		if _, err := os.Stat(blocked); err == nil {
			t.Fatalf("unexpected stale UI frontend artifact found: %s", blocked)
		} else if !os.IsNotExist(err) {
			t.Fatalf("stat %s: %v", blocked, err)
		}
	}

	blockedDirs := map[string]struct{}{
		"node_modules": {},
		"dist":         {},
		".vite":        {},
	}

	err := filepath.WalkDir(viewerDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		if _, found := blockedDirs[d.Name()]; found {
			t.Fatalf("unexpected stale UI frontend directory found under viewer: %s", path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk viewer directory: %v", err)
	}
}

func TestUIAViewerScriptsContainExpectedCommands(t *testing.T) {
	t.Parallel()

	cases := []struct {
		path  string
		wants []string
	}{
		{path: "dev-uia-viewer.bat", wants: []string{"go run ./cmd/goahk-uia-viewer"}},
		{
			path: "build-uia-viewer.bat",
			wants: []string{
				`set "DIST_DIR=%ROOT%\dist\goahk-uia-viewer"`,
				`go build -trimpath -v -o "%DIST_DIR%\goahk-uia-viewer.exe" ./cmd/goahk-uia-viewer`,
				`go run github.com/akavel/rsrc@v0.10.2 -manifest "%MANIFEST%" -o "%MANIFEST_SYSO%"`,
			},
		},
		{path: "dev-uia-viewer.sh", wants: []string{"go run ./cmd/goahk-uia-viewer"}},
		{
			path: "build-uia-viewer.sh",
			wants: []string{
				`DIST_DIR="${ROOT}/dist/goahk-uia-viewer"`,
				`go build -o "${DIST_DIR}/goahk-uia-viewer.exe" ./cmd/goahk-uia-viewer`,
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.path, func(t *testing.T) {
			t.Parallel()
			body, err := os.ReadFile(filepath.Join(".", tc.path))
			if err != nil {
				t.Fatalf("ReadFile(%s) error = %v", tc.path, err)
			}
			text := string(body)
			for _, want := range tc.wants {
				if !strings.Contains(text, want) {
					t.Fatalf("expected %s to contain %q", tc.path, want)
				}
			}
		})
	}
}
