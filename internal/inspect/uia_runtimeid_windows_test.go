//go:build windows
// +build windows

package inspect

import "testing"

func TestRuntimeIDString(t *testing.T) {
	got := runtimeIDString([]int{42, 12345, 67890})
	if got != "42.12345.67890" {
		t.Fatalf("unexpected runtime id: %s", got)
	}
}
