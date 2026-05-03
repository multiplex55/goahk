//go:build !windows

package main

import (
	"bytes"
	"testing"
)

func TestRunNonWindows_PrintsWindowsOnlyMessage(t *testing.T) {
	var out bytes.Buffer
	if err := runNonWindows(&out); err != nil {
		t.Fatalf("runNonWindows returned error: %v", err)
	}

	want := windowsOnlyMessage + "\n"
	if out.String() != want {
		t.Fatalf("unexpected output: got %q want %q", out.String(), want)
	}
}
