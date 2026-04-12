//go:build windows
// +build windows

package inspect

import "testing"

func TestBuildTags_WindowsProviderSelected(t *testing.T) {
	p, ok := newWindowsProvider().(*windowsProvider)
	if !ok {
		t.Fatalf("expected windowsProvider implementation")
	}
	if _, ok := p.uiaCore.adapter.(*windowsUIAAdapter); !ok {
		t.Fatalf("expected concrete windows UIA adapter, got %T", p.uiaCore.adapter)
	}
	if _, ok := p.accCore.adapter.(*windowsUIAAdapter); !ok {
		t.Fatalf("expected concrete ACC adapter, got %T", p.accCore.adapter)
	}
}
