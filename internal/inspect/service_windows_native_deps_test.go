//go:build windows
// +build windows

package inspect

import (
	"context"
	"errors"
	"testing"

	"goahk/internal/window"
)

type fakeNativeBridge struct {
	resolveRoot   func(window.HWND) (*uiaElement, error)
	elementByHWND func(window.HWND) (*uiaElement, error)
	parentHWND    func(window.HWND) (window.HWND, bool, error)
	childHWNDs    func(window.HWND) ([]window.HWND, error)
	focusedHWND   func() (window.HWND, error)
	cursorPos     func() (int, int, error)
	hwndFromPoint func(int, int) (window.HWND, error)
	invoke        func(window.HWND) error
	selectFn      func(window.HWND) error
	setValue      func(window.HWND, string) error
	defaultAction func(window.HWND) error
	toggle        func(window.HWND) error
	expand        func(window.HWND) error
	collapse      func(window.HWND) error
}

func (f fakeNativeBridge) ResolveRoot(h window.HWND) (*uiaElement, error) { return f.resolveRoot(h) }
func (f fakeNativeBridge) ElementByHWND(h window.HWND) (*uiaElement, error) {
	return f.elementByHWND(h)
}
func (f fakeNativeBridge) ParentHWND(h window.HWND) (window.HWND, bool, error) {
	return f.parentHWND(h)
}
func (f fakeNativeBridge) ChildHWNDs(h window.HWND) ([]window.HWND, error) { return f.childHWNDs(h) }
func (f fakeNativeBridge) FocusedHWND() (window.HWND, error)               { return f.focusedHWND() }
func (f fakeNativeBridge) CursorPosition() (int, int, error)               { return f.cursorPos() }
func (f fakeNativeBridge) HWNDFromPoint(x, y int) (window.HWND, error)     { return f.hwndFromPoint(x, y) }
func (f fakeNativeBridge) Invoke(h window.HWND) error                      { return f.invoke(h) }
func (f fakeNativeBridge) Select(h window.HWND) error                      { return f.selectFn(h) }
func (f fakeNativeBridge) SetValue(h window.HWND, v string) error          { return f.setValue(h, v) }
func (f fakeNativeBridge) DoDefaultAction(h window.HWND) error             { return f.defaultAction(h) }
func (f fakeNativeBridge) Toggle(h window.HWND) error                      { return f.toggle(h) }
func (f fakeNativeBridge) Expand(h window.HWND) error                      { return f.expand(h) }
func (f fakeNativeBridge) Collapse(h window.HWND) error                    { return f.collapse(h) }

func TestNativeUIADeps_Phase1LookupAndTraversal(t *testing.T) {
	mk := func(h window.HWND, p window.HWND) *uiaElement {
		return &uiaElement{Ref: makeElementRef(h), ParentRef: makeElementRef(p), Name: h.String()}
	}
	deps := &nativeUIADeps{bridge: fakeNativeBridge{
		resolveRoot:   func(h window.HWND) (*uiaElement, error) { return mk(h, 0), nil },
		elementByHWND: func(h window.HWND) (*uiaElement, error) { return mk(h, 0x1), nil },
		parentHWND:    func(window.HWND) (window.HWND, bool, error) { return 0x1, true, nil },
		childHWNDs:    func(window.HWND) ([]window.HWND, error) { return []window.HWND{0x2, 0x3}, nil },
		focusedHWND:   func() (window.HWND, error) { return 0x9, nil },
		cursorPos:     func() (int, int, error) { return 10, 20, nil },
		hwndFromPoint: func(int, int) (window.HWND, error) { return 0xA, nil },
		invoke:        func(window.HWND) error { return nil },
		selectFn:      func(window.HWND) error { return nil },
		setValue:      func(window.HWND, string) error { return nil },
		defaultAction: func(window.HWND) error { return nil },
		toggle:        func(window.HWND) error { return nil },
		expand:        func(window.HWND) error { return nil },
		collapse:      func(window.HWND) error { return nil },
	}}

	root, err := deps.ResolveWindowRoot(context.Background(), "0x1")
	if err != nil || root.Ref != "hwnd:0x1" {
		t.Fatalf("ResolveWindowRoot failed: %+v, %v", root, err)
	}
	children, err := deps.GetChildren(context.Background(), "hwnd:0x1")
	if err != nil || len(children) != 2 {
		t.Fatalf("GetChildren failed: len=%d err=%v", len(children), err)
	}
	parent, err := deps.GetParent(context.Background(), "hwnd:0x2")
	if err != nil || parent.Ref != "hwnd:0x1" {
		t.Fatalf("GetParent failed: %+v err=%v", parent, err)
	}
	count, ok, err := deps.GetChildCount(context.Background(), "hwnd:0x1")
	if err != nil || !ok || count != 2 {
		t.Fatalf("GetChildCount failed: count=%d ok=%v err=%v", count, ok, err)
	}
	focused, err := deps.GetFocusedElement(context.Background())
	if err != nil || focused.Ref == "" {
		t.Fatalf("GetFocusedElement failed: %+v err=%v", focused, err)
	}
	x, y, err := deps.GetCursorPosition(context.Background())
	if err != nil || x != 10 || y != 20 {
		t.Fatalf("GetCursorPosition failed: x=%d y=%d err=%v", x, y, err)
	}
	under, err := deps.ElementFromPoint(context.Background(), x, y)
	if err != nil || under.Ref != "hwnd:0xa" {
		t.Fatalf("ElementFromPoint failed: %+v err=%v", under, err)
	}
}

func TestNativeUIADeps_Phase2ActionDispatch(t *testing.T) {
	called := map[string]string{}
	deps := &nativeUIADeps{bridge: fakeNativeBridge{
		resolveRoot:   func(window.HWND) (*uiaElement, error) { return nil, nil },
		elementByHWND: func(window.HWND) (*uiaElement, error) { return nil, nil },
		parentHWND:    func(window.HWND) (window.HWND, bool, error) { return 0, false, nil },
		childHWNDs:    func(window.HWND) ([]window.HWND, error) { return nil, nil },
		focusedHWND:   func() (window.HWND, error) { return 0, nil },
		cursorPos:     func() (int, int, error) { return 0, 0, nil },
		hwndFromPoint: func(int, int) (window.HWND, error) { return 0, nil },
		invoke:        func(h window.HWND) error { called["invoke"] = h.String(); return nil },
		selectFn:      func(h window.HWND) error { called["select"] = h.String(); return nil },
		setValue:      func(h window.HWND, v string) error { called["setValue"] = h.String() + ":" + v; return nil },
		defaultAction: func(h window.HWND) error { called["default"] = h.String(); return nil },
		toggle:        func(h window.HWND) error { called["toggle"] = h.String(); return nil },
		expand:        func(h window.HWND) error { called["expand"] = h.String(); return nil },
		collapse:      func(h window.HWND) error { called["collapse"] = h.String(); return nil },
	}}

	ref := "hwnd:0x2a"
	_ = deps.Invoke(context.Background(), ref)
	_ = deps.Select(context.Background(), ref)
	_ = deps.SetValue(context.Background(), ref, "ok")
	_ = deps.DoDefaultAction(context.Background(), ref)
	_ = deps.Toggle(context.Background(), ref)
	_ = deps.Expand(context.Background(), ref)
	_ = deps.Collapse(context.Background(), ref)

	if len(called) != 7 || called["setValue"] != "0x2a:ok" {
		t.Fatalf("unexpected dispatch map: %+v", called)
	}
}

func TestNativeUIADeps_ErrorTransitionsAndInvalidRefs(t *testing.T) {
	deps := &nativeUIADeps{bridge: fakeNativeBridge{
		resolveRoot:   func(window.HWND) (*uiaElement, error) { return nil, errUIAElementNotAvailable },
		elementByHWND: func(window.HWND) (*uiaElement, error) { return nil, nil },
		parentHWND:    func(window.HWND) (window.HWND, bool, error) { return 0, false, nil },
		childHWNDs:    func(window.HWND) ([]window.HWND, error) { return nil, errors.New("boom") },
		focusedHWND:   func() (window.HWND, error) { return 0, nil },
		cursorPos:     func() (int, int, error) { return 0, 0, nil },
		hwndFromPoint: func(int, int) (window.HWND, error) { return 0, nil },
		invoke:        func(window.HWND) error { return nil },
		selectFn:      func(window.HWND) error { return nil },
		setValue:      func(window.HWND, string) error { return nil },
		defaultAction: func(window.HWND) error { return nil },
		toggle:        func(window.HWND) error { return nil },
		expand:        func(window.HWND) error { return nil },
		collapse:      func(window.HWND) error { return nil },
	}}

	if _, err := deps.ResolveWindowRoot(context.Background(), "bad"); !errors.Is(err, errUIANilElement) {
		t.Fatalf("expected nil element parse failure, got %v", err)
	}
	if _, err := deps.GetElementByRef(context.Background(), "bad"); !errors.Is(err, errUIANilElement) {
		t.Fatalf("expected invalid ref -> nil element, got %v", err)
	}
	if _, err := deps.GetFocusedElement(context.Background()); !errors.Is(err, errUIAElementNotAvailable) {
		t.Fatalf("expected empty focus -> stale, got %v", err)
	}
	if _, _, err := deps.GetChildCount(context.Background(), "hwnd:0x1"); err == nil {
		t.Fatalf("expected child traversal failure")
	}
}
