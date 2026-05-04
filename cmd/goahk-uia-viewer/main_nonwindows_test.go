//go:build !windows

package main

import (
	"errors"
	"testing"
)

func TestRunNonWindows_ReturnsUnsupportedPlatformError(t *testing.T) {
	err := runNonWindows()
	if !errors.Is(err, errUnsupportedPlatform) {
		t.Fatalf("runNonWindows error = %v, want %v", err, errUnsupportedPlatform)
	}
}
