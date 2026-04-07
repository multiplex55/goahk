package build

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

var semverPattern = regexp.MustCompile(`^v?([0-9]+)\.([0-9]+)\.([0-9]+)(?:[-+][0-9A-Za-z.-]+)?$`)

func NormalizeSemver(v string) (string, error) {
	v = strings.TrimSpace(v)
	if !semverPattern.MatchString(v) {
		return "", fmt.Errorf("invalid semantic version %q", v)
	}
	if strings.HasPrefix(v, "v") {
		return v, nil
	}
	return "v" + v, nil
}

func StampVersion(base, commit string, t time.Time) (string, error) {
	normalized, err := NormalizeSemver(base)
	if err != nil {
		return "", err
	}
	sha := strings.ToLower(strings.TrimSpace(commit))
	if len(sha) > 7 {
		sha = sha[:7]
	}
	if sha == "" {
		sha = "unknown"
	}
	stamp := t.UTC().Format("20060102") + "." + sha
	if strings.Contains(normalized, "+") {
		return normalized + "." + stamp, nil
	}
	return normalized + "+" + stamp, nil
}
