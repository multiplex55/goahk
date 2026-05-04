package repohygiene

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestViewerMigrationActiveFilesDoNotContainStaleMarkers(t *testing.T) {
	t.Parallel()

	type scopedFile struct {
		path   string
		banned []string
	}

	files := []scopedFile{
		{
			path: "docs/uia-viewer.md",
			banned: []string{
				"migration placeholder",
				"wails",
				"npm run",
				"npx",
				"vite",
				"webview",
				"viewerfacade",
				"viewer façade",
			},
		},
		{
			path: "docs/testing-uia-viewer.md",
			banned: []string{
				"migration placeholder",
				"react",
				"redux",
				"zustand",
				"store",
				"wails",
				"npm run",
				"vite",
				"webview",
				"viewerfacade",
				"viewer façade",
			},
		},
		{
			path: "docs/BUILD.md",
			banned: []string{
				"wails",
				"npm run",
				"npx",
				"yarn",
				"pnpm",
				"vite",
				"webview",
			},
		},
	}

	root := repoRoot(t)
	for _, tc := range files {
		tc := tc
		t.Run(tc.path, func(t *testing.T) {
			t.Parallel()
			b, err := os.ReadFile(filepath.Join(root, tc.path))
			if err != nil {
				t.Fatalf("ReadFile(%q) error = %v", tc.path, err)
			}
			text := strings.ToLower(string(b))
			for _, marker := range tc.banned {
				if strings.Contains(text, strings.ToLower(marker)) {
					t.Fatalf("%s contains stale viewer marker %q", tc.path, marker)
				}
			}
		})
	}
}
