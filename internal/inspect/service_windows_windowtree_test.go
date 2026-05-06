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
			0x1: {Ref: "win:0x1", RuntimeID: "1", Name: "root"},
			0x2: {Ref: "win:0x2", RuntimeID: "2", Name: "left", ParentRef: "win:0x1"},
			0x3: {Ref: "win:0x3", RuntimeID: "3", Name: "right", ParentRef: "win:0x1"},
		},
		parents:  map[window.HWND]window.HWND{0x2: 0x1, 0x3: 0x1},
		children: map[window.HWND][]window.HWND{0x1: {0x2, 0x3}},
	}}

	kids, err := deps.GetChildren(context.Background(), "win:0x1")
	if err != nil {
		t.Fatalf("GetChildren failed: %v", err)
	}
	if len(kids) != 2 || kids[0].Ref != "win:0x2" || kids[1].Ref != "win:0x3" {
		t.Fatalf("unexpected children: %+v", kids)
	}
	parent, err := deps.GetParent(context.Background(), "win:0x2")
	if err != nil || parent.Ref != "win:0x1" {
		t.Fatalf("GetParent failed: %+v err=%v", parent, err)
	}
	count, ok, err := deps.GetChildCount(context.Background(), "win:0x1")
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
	if _, err := deps.GetElementByRef(context.Background(), "uia:sess:1"); !errors.Is(err, errUIANilElement) {
		t.Fatalf("expected UIA ref rejection in window tree mode, got %v", err)
	}
}

func TestNativeWindowTreeDeps_GetElementByRef_BoundingRectPresent(t *testing.T) {
	deps := &nativeWindowTreeDeps{bridge: fakeWindowTreeBridge{
		elements: map[window.HWND]*uiaElement{
			0x10: {
				Ref:          "win:0x10",
				RuntimeID:    "16",
				BoundingRect: &uiaRect{Left: 10, Top: 20, Width: 300, Height: 200},
				UnsupportedProps: map[string]bool{
					"BoundingRectangle": false,
				},
			},
		},
	}}

	el, err := deps.GetElementByRef(context.Background(), "win:0x10")
	if err != nil {
		t.Fatalf("GetElementByRef failed: %v", err)
	}
	if el.BoundingRect == nil {
		t.Fatalf("expected bounding rect to be populated")
	}
	if got := el.BoundingRect; got.Left != 10 || got.Top != 20 || got.Width != 300 || got.Height != 200 {
		t.Fatalf("unexpected bounding rect: %+v", got)
	}
	if el.UnsupportedProps["BoundingRectangle"] {
		t.Fatalf("expected BoundingRectangle to be supported when rect is present")
	}
}

func TestNativeWindowTreeDeps_GetElementByRef_BoundingRectAbsent(t *testing.T) {
	deps := &nativeWindowTreeDeps{bridge: fakeWindowTreeBridge{
		elements: map[window.HWND]*uiaElement{
			0x20: {
				Ref:       "win:0x20",
				RuntimeID: "32",
				UnsupportedProps: map[string]bool{
					"BoundingRectangle": true,
				},
			},
		},
	}}

	el, err := deps.GetElementByRef(context.Background(), "win:0x20")
	if err != nil {
		t.Fatalf("GetElementByRef failed: %v", err)
	}
	if el.BoundingRect != nil {
		t.Fatalf("expected bounding rect to be nil when unavailable, got %+v", el.BoundingRect)
	}
	if !el.UnsupportedProps["BoundingRectangle"] {
		t.Fatalf("expected BoundingRectangle to remain unsupported when rect is unavailable")
	}
}
