//go:build windows
// +build windows

package inspect

import "testing"

func TestBuildTags_WindowsProviderSelected(t *testing.T) {
	p, ok := newWindowsProvider().(*windowsProvider)
	if !ok {
		t.Fatalf("expected windowsProvider implementation")
	}
	if _, ok := p.core.adapter.(*windowsUIAAdapter); !ok {
		t.Fatalf("expected concrete windows UIA adapter, got %T", p.core.adapter)
	}
}
