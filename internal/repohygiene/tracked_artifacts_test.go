package repohygiene

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestNoTrackedBinaryArtifacts(t *testing.T) {
	t.Helper()

	cmd := exec.Command("git", "ls-files")
	cmd.Dir = repoRoot(t)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git ls-files failed: %v\n%s", err, output)
	}

	lines := bytes.Split(output, []byte{'\n'})
	files := make([]string, 0, len(lines))
	for _, line := range lines {
		file := strings.TrimSpace(string(line))
		if file != "" {
			files = append(files, file)
		}
	}

	allowReleaseArtifacts := strings.EqualFold(os.Getenv("GOAHK_ALLOW_RELEASE_ARTIFACTS"), "1") ||
		strings.EqualFold(os.Getenv("GOAHK_ALLOW_RELEASE_ARTIFACTS"), "true")
	blocked := blockedBinaryFiles(files, allowReleaseArtifacts)
	if len(blocked) > 0 {
		t.Fatalf("blocked tracked binaries are not allowed outside %s unless GOAHK_ALLOW_RELEASE_ARTIFACTS=1: %s", releaseArtifactPrefix, strings.Join(blocked, ", "))
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
