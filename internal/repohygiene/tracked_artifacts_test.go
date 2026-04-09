package repohygiene

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var forbiddenBinaryExtensions = []string{".exe", ".dll", ".so", ".dylib"}

func TestNoTrackedBinaryArtifacts(t *testing.T) {
	t.Helper()

	cmd := exec.Command("git", "ls-files")
	cmd.Dir = repoRoot(t)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git ls-files failed: %v\n%s", err, output)
	}

	lines := bytes.Split(output, []byte{'\n'})
	var trackedForbidden []string
	for _, line := range lines {
		file := strings.TrimSpace(string(line))
		if file == "" {
			continue
		}
		ext := strings.ToLower(filepath.Ext(file))
		for _, forbidden := range forbiddenBinaryExtensions {
			if ext == forbidden {
				trackedForbidden = append(trackedForbidden, file)
				break
			}
		}
	}

	if len(trackedForbidden) > 0 {
		t.Fatalf("tracked binary artifacts are not allowed: %s", strings.Join(trackedForbidden, ", "))
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()

	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git rev-parse failed: %v\n%s", err, output)
	}

	return strings.TrimSpace(string(output))
}
