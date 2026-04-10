//go:build windows
// +build windows

package inspect

import "testing"

func TestBuildTags_WindowsProviderSelected(t *testing.T) {
	if _, ok := newWindowsProvider().(*windowsProvider); !ok {
		t.Fatalf("expected windowsProvider implementation")
	}
}
