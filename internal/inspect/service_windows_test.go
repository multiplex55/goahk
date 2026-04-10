//go:build windows
// +build windows

package inspect

import (
	"context"
	"testing"
)

func TestWindowsProvider_InspectWindow_DoesNotRefreshCache(t *testing.T) {
	adapter := &fakeAdapter{
		root: &uiaElement{Ref: "root", Name: "Root"},
		kids: map[string][]*uiaElement{
			"root": {{Ref: "c1", ParentRef: "root", Name: "Child"}},
		},
	}
	provider := &windowsProvider{
		core:       newProviderCore(adapter),
		highlights: newHighlightController(newNativeHighlightOverlay()),
	}

	rootResp, err := provider.GetTreeRoot(context.Background(), GetTreeRootRequest{HWND: "0x1", Refresh: false})
	if err != nil {
		t.Fatalf("GetTreeRoot setup failed: %v", err)
	}
	if _, err := provider.GetNodeChildren(context.Background(), GetNodeChildrenRequest{NodeID: rootResp.Root.NodeID}); err != nil {
		t.Fatalf("GetNodeChildren initial load failed: %v", err)
	}

	if _, err := provider.InspectWindow(context.Background(), InspectWindowRequest{HWND: "0x1"}); err != nil {
		t.Fatalf("InspectWindow failed: %v", err)
	}
	if _, err := provider.GetNodeChildren(context.Background(), GetNodeChildrenRequest{NodeID: rootResp.Root.NodeID}); err != nil {
		t.Fatalf("GetNodeChildren after InspectWindow failed: %v", err)
	}
	if got := adapter.childrenCallCount["root"]; got != 1 {
		t.Fatalf("expected cached children after InspectWindow (refresh=false path), calls=%d", got)
	}
}

func TestWindowsProvider_GetTreeRoot_RefreshInvalidatesCache(t *testing.T) {
	adapter := &fakeAdapter{
		root: &uiaElement{Ref: "root", Name: "Root"},
		kids: map[string][]*uiaElement{
			"root": {{Ref: "c1", ParentRef: "root", Name: "Child"}},
		},
	}
	provider := &windowsProvider{
		core:       newProviderCore(adapter),
		highlights: newHighlightController(newNativeHighlightOverlay()),
	}

	rootResp, err := provider.GetTreeRoot(context.Background(), GetTreeRootRequest{HWND: "0x1", Refresh: false})
	if err != nil {
		t.Fatalf("GetTreeRoot setup failed: %v", err)
	}
	if _, err := provider.GetNodeChildren(context.Background(), GetNodeChildrenRequest{NodeID: rootResp.Root.NodeID}); err != nil {
		t.Fatalf("GetNodeChildren initial load failed: %v", err)
	}
	if _, err := provider.GetTreeRoot(context.Background(), GetTreeRootRequest{HWND: "0x1", Refresh: true}); err != nil {
		t.Fatalf("GetTreeRoot refresh failed: %v", err)
	}
	if _, err := provider.GetNodeChildren(context.Background(), GetNodeChildrenRequest{NodeID: rootResp.Root.NodeID}); err != nil {
		t.Fatalf("GetNodeChildren after refresh failed: %v", err)
	}
	if got := adapter.childrenCallCount["root"]; got != 2 {
		t.Fatalf("expected children to reload after refresh invalidates cache, calls=%d", got)
	}
}
