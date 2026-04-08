package examples_test

import (
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestBasicScriptExampleBuilds(t *testing.T) {
	t.Parallel()
	buildPackage(t, "./basic-script")
}

func TestStandaloneScriptExampleBuilds(t *testing.T) {
	t.Parallel()
	buildPackage(t, "./standalone-script")
}

func TestListOpenAppsExampleBuilds(t *testing.T) {
	t.Parallel()
	buildPackage(t, "./list-open-apps")
}

func TestCustomCallbackExampleBuilds(t *testing.T) {
	t.Parallel()
	buildPackage(t, "./custom-callback")
}

func TestClipboardTransformPasteExampleBuilds(t *testing.T) {
	t.Parallel()
	buildPackage(t, "./clipboard-transform-paste")
}

func TestWindowAwareScriptExampleBuilds(t *testing.T) {
	t.Parallel()
	buildPackage(t, "./window-aware-script")
}

func TestMixedActionsExampleBuilds(t *testing.T) {
	t.Parallel()
	buildPackage(t, "./mixed-actions")
}

func TestWindowsTargetedCompilationRegressions(t *testing.T) {
	t.Parallel()

	targets := []struct {
		name string
		pkg  string
	}{
		{name: "example/basic-script", pkg: "./basic-script"},
		{name: "goahk/goahk", pkg: "../goahk"},
		{name: "internal/clipboard", pkg: "../internal/clipboard"},
		{name: "internal/services/messagebox", pkg: "../internal/services/messagebox"},
	}

	for _, target := range targets {
		target := target
		t.Run(target.name, func(t *testing.T) {
			t.Parallel()
			assertWindowsCompile(t, target.pkg)
		})
	}
}

func buildPackage(t *testing.T, pkg string) {
	t.Helper()
	outPath := filepath.Join(t.TempDir(), strings.TrimPrefix(filepath.Base(pkg), "./")+".exe")
	cmd := exec.Command("go", "build", "-o", outPath, pkg)
	cmd.Dir = "."
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build %s failed: %v\n%s", pkg, err, out)
	}
}

func assertWindowsCompile(t *testing.T, pkg string) {
	t.Helper()

	args := []string{"test", "-run", "^$", "-count=1", pkg}
	if runtime.GOOS != "windows" {
		outPath := filepath.Join(t.TempDir(), strings.ReplaceAll(strings.TrimPrefix(pkg, "../"), "/", "-")+".test.exe")
		args = []string{"test", "-c", "-o", outPath, pkg}
	}

	cmd := exec.Command("go", args...)
	cmd.Dir = "."
	if runtime.GOOS != "windows" {
		cmd.Env = append(cmd.Environ(), "GOOS=windows", "GOARCH=amd64", "CGO_ENABLED=0")
	}

	out, err := cmd.CombinedOutput()
	t.Logf("windows compile probe: go %s", strings.Join(args, " "))
	if len(out) > 0 {
		t.Logf("go output for %s:\n%s", pkg, out)
	}
	if err != nil {
		t.Fatalf("windows-targeted compile check failed for %s: %v", pkg, err)
	}
}
