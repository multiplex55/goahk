//go:build windows
// +build windows

package inspect

import (
	"context"
	"errors"
	"testing"

	"goahk/internal/window"
)

type fakeWindowTreeBridge struct {
	elements map[window.HWND]*uiaElement
	parents  map[window.HWND]window.HWND
	children map[window.HWND][]window.HWND
}

func (f fakeWindowTreeBridge) ResolveRoot(hwnd window.HWND) (*uiaElement, error) {
	if el, ok := f.elements[hwnd]; ok {
		return el, nil
	}
	return nil, errUIAElementNotAvailable
}
func (f fakeWindowTreeBridge) ElementByHWND(hwnd window.HWND) (*uiaElement, error) {
	if el, ok := f.elements[hwnd]; ok {
		return el, nil
	}
	return nil, nil
}
func (f fakeWindowTreeBridge) ParentHWND(hwnd window.HWND) (window.HWND, bool, error) {
	p, ok := f.parents[hwnd]
	return p, ok, nil
}
func (f fakeWindowTreeBridge) ChildHWNDs(hwnd window.HWND) ([]window.HWND, error) {
	return f.children[hwnd], nil
}

func TestNativeWindowTreeDeps_ParentAndChildrenTraversal(t *testing.T) {
	deps := &nativeWindowTreeDeps{bridge: fakeWindowTreeBridge{
		elements: map[window.HWND]*uiaElement{
			0x1: {Ref: "hwnd:0x1", RuntimeID: "1", Name: "root"},
			0x2: {Ref: "hwnd:0x2", RuntimeID: "2", Name: "left", ParentRef: "hwnd:0x1"},
			0x3: {Ref: "hwnd:0x3", RuntimeID: "3", Name: "right", ParentRef: "hwnd:0x1"},
		},
		parents:  map[window.HWND]window.HWND{0x2: 0x1, 0x3: 0x1},
		children: map[window.HWND][]window.HWND{0x1: {0x2, 0x3}},
	}}

	kids, err := deps.GetChildren(context.Background(), "hwnd:0x1")
	if err != nil {
		t.Fatalf("GetChildren failed: %v", err)
	}
	if len(kids) != 2 || kids[0].Ref != "hwnd:0x2" || kids[1].Ref != "hwnd:0x3" {
		t.Fatalf("unexpected children: %+v", kids)
	}
	parent, err := deps.GetParent(context.Background(), "hwnd:0x2")
	if err != nil || parent.Ref != "hwnd:0x1" {
		t.Fatalf("GetParent failed: %+v err=%v", parent, err)
	}
	count, ok, err := deps.GetChildCount(context.Background(), "hwnd:0x1")
	if err != nil || !ok || count != 2 {
		t.Fatalf("GetChildCount failed: count=%d ok=%v err=%v", count, ok, err)
	}
}

func TestNativeWindowTreeDeps_UIAOnlyMethodsGuarded(t *testing.T) {
	deps := &nativeWindowTreeDeps{bridge: fakeWindowTreeBridge{}}
	if _, err := deps.GetFocusedElement(context.Background()); !errors.Is(err, ErrProviderActionUnsupported) {
		t.Fatalf("expected focused lookup to be unsupported, got %v", err)
	}
	if _, _, err := deps.GetCursorPosition(context.Background()); !errors.Is(err, ErrProviderActionUnsupported) {
		t.Fatalf("expected cursor position to be unsupported, got %v", err)
	}
	if _, err := deps.ElementFromPoint(context.Background(), 1, 2); !errors.Is(err, ErrProviderActionUnsupported) {
		t.Fatalf("expected element-from-point to be unsupported, got %v", err)
	}
}
