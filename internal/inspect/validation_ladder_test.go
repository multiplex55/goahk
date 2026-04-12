//go:build windows
// +build windows

package inspect

import (
	"context"
	"testing"
)

func TestValidationLadder_A_to_F(t *testing.T) {
	t.Run("A root resolution", func(t *testing.T) {
		deps := &fakeAdapter{root: &uiaElement{Ref: "root", RuntimeID: "1", Name: "Root"}}
		provider := newWindowsProviderWithDeps(newUIAAdapter(deps), &fakeWindowAdapter{}).(*windowsProvider)
		resp, err := provider.GetTreeRoot(context.Background(), GetTreeRootRequest{HWND: "0x1", Mode: InspectModeUIATree})
		if err != nil || resp.Root.NodeID == "" {
			t.Fatalf("root resolution failed: root=%+v err=%v", resp.Root, err)
		}
	})

	t.Run("B child expansion", func(t *testing.T) {
		deps := &fakeAdapter{
			root: &uiaElement{Ref: "root", RuntimeID: "1", Name: "Root"},
			kids: map[string][]*uiaElement{"root": {{Ref: "child", RuntimeID: "2", ParentRef: "root", Name: "Child"}}},
		}
		provider := newWindowsProviderWithDeps(newUIAAdapter(deps), &fakeWindowAdapter{}).(*windowsProvider)
		root, _ := provider.GetTreeRoot(context.Background(), GetTreeRootRequest{HWND: "0x1"})
		children, err := provider.GetNodeChildren(context.Background(), GetNodeChildrenRequest{NodeID: root.Root.NodeID})
		if err != nil || len(children.Children) != 1 {
			t.Fatalf("child expansion failed: children=%+v err=%v", children.Children, err)
		}
	})

	t.Run("C properties parity contract", func(t *testing.T) {
		deps := &fakeAdapter{
			root:  &uiaElement{Ref: "root", RuntimeID: "1", Name: "Root", ControlType: "Window", BoundingRect: &uiaRect{Left: 0, Top: 0, Width: 100, Height: 100}},
			byRef: map[string]*uiaElement{"root": {Ref: "root", RuntimeID: "1", Name: "Root", ControlType: "Window", BoundingRect: &uiaRect{Left: 0, Top: 0, Width: 100, Height: 100}}},
		}
		provider := newWindowsProviderWithDeps(newUIAAdapter(deps), &fakeWindowAdapter{}).(*windowsProvider)
		root, _ := provider.GetTreeRoot(context.Background(), GetTreeRootRequest{HWND: "0x1"})
		details, err := provider.GetNodeDetails(context.Background(), GetNodeDetailsRequest{NodeID: root.Root.NodeID})
		if err != nil || details.Element.NodeID != root.Root.NodeID || len(details.Properties) == 0 {
			t.Fatalf("properties parity failed: details=%+v err=%v", details, err)
		}
	})

	t.Run("D pattern availability/action wiring", func(t *testing.T) {
		deps := &fakeAdapter{
			root:  &uiaElement{Ref: "root", RuntimeID: "1", Name: "Root", SupportedPatterns: []string{"Invoke"}},
			byRef: map[string]*uiaElement{"root": {Ref: "root", RuntimeID: "1", Name: "Root", SupportedPatterns: []string{"Invoke"}}},
		}
		provider := newWindowsProviderWithDeps(newUIAAdapter(deps), &fakeWindowAdapter{}).(*windowsProvider)
		root, _ := provider.GetTreeRoot(context.Background(), GetTreeRootRequest{HWND: "0x1"})
		actions, err := provider.GetPatternActions(context.Background(), GetPatternActionsRequest{NodeID: root.Root.NodeID})
		if err != nil || len(actions.Actions) == 0 {
			t.Fatalf("pattern action discovery failed: actions=%+v err=%v", actions.Actions, err)
		}
		if _, err := provider.InvokePattern(context.Background(), InvokePatternRequest{NodeID: root.Root.NodeID, Action: "invoke"}); err != nil {
			t.Fatalf("invoke wiring failed: %v", err)
		}
	})

	t.Run("E follow-cursor selection updates", func(t *testing.T) {
		deps := &fakeAdapter{
			root:  &uiaElement{Ref: "root", RuntimeID: "1", Name: "Root"},
			under: &uiaElement{Ref: "cursor", RuntimeID: "2", Name: "Cursor"},
			byRef: map[string]*uiaElement{"cursor": {Ref: "cursor", RuntimeID: "2", Name: "Cursor"}},
		}
		provider := newWindowsProviderWithDeps(newUIAAdapter(deps), &fakeWindowAdapter{}).(*windowsProvider)
		_, _ = provider.GetTreeRoot(context.Background(), GetTreeRootRequest{HWND: "0x1"})
		_, _ = provider.ToggleFollowCursor(context.Background(), ToggleFollowCursorRequest{Enabled: true})
		resp, err := provider.GetElementUnderCursor(context.Background(), GetElementUnderCursorRequest{})
		if err != nil || resp.Element.NodeID == "" {
			t.Fatalf("follow-cursor update failed: element=%+v err=%v", resp.Element, err)
		}
	})

	t.Run("F ACC mode alternate-tree behavior", func(t *testing.T) {
		uia := &fakeAdapter{root: nil}
		acc := &fakeAdapter{root: &uiaElement{Ref: "acc-root", RuntimeID: "10", Name: "ACC Root"}}
		provider := newWindowsProviderWithModeAdapters(newUIAAdapter(uia), newUIAAdapter(acc), &fakeWindowAdapter{}).(*windowsProvider)
		resp, err := provider.GetTreeRoot(context.Background(), GetTreeRootRequest{HWND: "0x1", Mode: InspectModeWindowTree})
		if err != nil || resp.State.ActiveMode != InspectModeWindowTree || resp.Root.Name != "ACC Root" {
			t.Fatalf("ACC mode behavior failed: root=%+v state=%+v err=%v", resp.Root, resp.State, err)
		}
	})
}
