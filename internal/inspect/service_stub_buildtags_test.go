//go:build !windows
// +build !windows

package inspect

import "testing"

func TestBuildTags_StubProviderSelected(t *testing.T) {
	if _, ok := newWindowsProvider().(unsupportedProvider); !ok {
		t.Fatalf("expected unsupportedProvider implementation")
	}
}
