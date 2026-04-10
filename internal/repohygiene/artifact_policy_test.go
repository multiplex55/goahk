package repohygiene

import (
	"reflect"
	"testing"
)

func TestBlockedBinaryFiles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                  string
		files                 []string
		allowReleaseArtifacts bool
		want                  []string
	}{
		{
			name:  "filters out non-binary files",
			files: []string{"README.md", "build/build.bat", "internal/foo.txt"},
			want:  nil,
		},
		{
			name:  "blocks binaries when release artifacts are disabled",
			files: []string{"dist/releases/goahk.exe", "bin/helper.dll", "pkg/lib.so", "lib/lib.dylib"},
			want:  []string{"dist/releases/goahk.exe", "bin/helper.dll", "pkg/lib.so", "lib/lib.dylib"},
		},
		{
			name:                  "allows release binaries under approved path when enabled",
			allowReleaseArtifacts: true,
			files:                 []string{"dist/releases/goahk.exe", "dist/releases/plugin.dll", "bin/helper.dll"},
			want:                  []string{"bin/helper.dll"},
		},
		{
			name:                  "path and extension checks are case-insensitive",
			allowReleaseArtifacts: true,
			files:                 []string{"DIST/RELEASES/GOAHK.EXE", "DIST/releaSES/helper.DLL", "DIST/other/tool.EXE"},
			want:                  []string{"DIST/other/tool.EXE"},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := blockedBinaryFiles(tc.files, tc.allowReleaseArtifacts)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("blockedBinaryFiles() = %v, want %v", got, tc.want)
			}
		})
	}
}
