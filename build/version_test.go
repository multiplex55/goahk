package build

import (
	"testing"
	"time"
)

func TestNormalizeSemver(t *testing.T) {
	got, err := NormalizeSemver("1.2.3")
	if err != nil || got != "v1.2.3" {
		t.Fatalf("NormalizeSemver() = (%q, %v)", got, err)
	}
	if _, err := NormalizeSemver("1.2"); err == nil {
		t.Fatal("expected invalid semver error")
	}
}

func TestStampVersion(t *testing.T) {
	got, err := StampVersion("v1.2.3", "abcdef123", time.Date(2026, 4, 7, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("StampVersion() error = %v", err)
	}
	if got != "v1.2.3+20260407.abcdef1" {
		t.Fatalf("StampVersion() = %q", got)
	}
}
