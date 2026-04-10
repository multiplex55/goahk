package repohygiene

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var (
	goDirectivePattern  = regexp.MustCompile(`(?m)^go\s+([0-9]+\.[0-9]+(?:\.[0-9]+)?)\s*$`)
	stableTokenPattern  = regexp.MustCompile(`(?m)^\s*go-version:\s*['\"]?stable['\"]?\s*$`)
	workflowVersionLine = regexp.MustCompile(`(?m)^\s*go-version:\s*\$\{\{\s*matrix\.go-version\s*\}\}\s*$`)
)

func TestGoVersionPolicyConsistency(t *testing.T) {
	t.Helper()

	root := repoRoot(t)
	goModPath := filepath.Join(root, "go.mod")
	workflowPath := filepath.Join(root, ".github", "workflows", "go.yml")

	goModBytes, err := os.ReadFile(goModPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", goModPath, err)
	}
	workflowBytes, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", workflowPath, err)
	}

	goModText := string(goModBytes)
	workflowText := string(workflowBytes)

	matches := goDirectivePattern.FindStringSubmatch(goModText)
	if len(matches) != 2 {
		t.Fatalf("could not locate go directive in go.mod")
	}
	declared := matches[1]
	declaredMinor := majorMinor(declared)

	if stableTokenPattern.MatchString(workflowText) {
		t.Fatalf("workflow must not use ambiguous go-version: stable; pin explicit versions")
	}

	if !workflowVersionLine.MatchString(workflowText) {
		t.Fatalf("workflow must configure setup-go using matrix.go-version")
	}

	if !strings.Contains(workflowText, "'"+declaredMinor+".x'") {
		t.Fatalf("workflow matrix must include go.mod minimum version '%s.x'", declaredMinor)
	}
}

func majorMinor(v string) string {
	parts := strings.Split(v, ".")
	if len(parts) < 2 {
		return v
	}
	return parts[0] + "." + parts[1]
}
