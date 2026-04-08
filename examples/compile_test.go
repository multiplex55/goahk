package examples_test

import (
	"os/exec"
	"path/filepath"
	"testing"
)

func TestBasicScriptExampleBuilds(t *testing.T) {
	t.Parallel()

	outPath := filepath.Join(t.TempDir(), "basic-script.exe")
	cmd := exec.Command("go", "build", "-o", outPath, "./basic-script")
	cmd.Dir = "."
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build ./basic-script failed: %v\n%s", err, out)
	}
}
