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
		{
			path: "dev-uia-viewer.bat",
			wants: []string{
				`set "APP_DIR=%ROOT%\cmd\goahk-uia-viewer"`,
				`set "DIST_DIR=%ROOT%\dist\goahk-uia-viewer"`,
				"wails dev",
			},
		},
		{
			path: "build-uia-viewer.bat",
			wants: []string{
				`set "APP_DIR=%ROOT%\cmd\goahk-uia-viewer"`,
				`set "DIST_DIR=%ROOT%\dist\goahk-uia-viewer"`,
				"wails build -clean -o goahk-uia-viewer",
				`robocopy "build\bin" "%DIST_DIR%" /E`,
			},
		},
		{
			path: "dev-uia-viewer.sh",
			wants: []string{
				`APP_DIR="${ROOT}/cmd/goahk-uia-viewer"`,
				`DIST_DIR="${ROOT}/dist/goahk-uia-viewer"`,
				"wails dev",
			},
		},
		{
			path: "build-uia-viewer.sh",
			wants: []string{
				`APP_DIR="${ROOT}/cmd/goahk-uia-viewer"`,
				`DIST_DIR="${ROOT}/dist/goahk-uia-viewer"`,
				"wails build -clean -o goahk-uia-viewer",
				`cp -a "${APP_DIR}/build/bin/." "${DIST_DIR}/"`,
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
