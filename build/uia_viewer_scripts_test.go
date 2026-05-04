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
		"dev-uia-viewer.bat",
		"build-uia-viewer.bat",
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
