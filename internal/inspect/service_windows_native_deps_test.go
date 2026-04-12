//go:build windows
// +build windows

package inspect

import (
	"context"
	"errors"
	"strings"
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
	parentCalls   int
	childCalls    int
}

func (f *fakeNativeBridge) ResolveRoot(h window.HWND) (*uiaElement, error) { return f.resolveRoot(h) }
func (f *fakeNativeBridge) ElementByHWND(h window.HWND) (*uiaElement, error) {
	return f.elementByHWND(h)
}
func (f *fakeNativeBridge) ParentHWND(h window.HWND) (window.HWND, bool, error) {
	f.parentCalls++
	return f.parentHWND(h)
}
func (f *fakeNativeBridge) ChildHWNDs(h window.HWND) ([]window.HWND, error) {
	f.childCalls++
	return f.childHWNDs(h)
}
func (f *fakeNativeBridge) FocusedHWND() (window.HWND, error)           { return f.focusedHWND() }
func (f *fakeNativeBridge) CursorPosition() (int, int, error)           { return f.cursorPos() }
func (f *fakeNativeBridge) HWNDFromPoint(x, y int) (window.HWND, error) { return f.hwndFromPoint(x, y) }
func (f *fakeNativeBridge) Invoke(h window.HWND) error                  { return f.invoke(h) }
func (f *fakeNativeBridge) Select(h window.HWND) error                  { return f.selectFn(h) }
func (f *fakeNativeBridge) SetValue(h window.HWND, v string) error      { return f.setValue(h, v) }
func (f *fakeNativeBridge) DoDefaultAction(h window.HWND) error         { return f.defaultAction(h) }
func (f *fakeNativeBridge) Toggle(h window.HWND) error                  { return f.toggle(h) }
func (f *fakeNativeBridge) Expand(h window.HWND) error                  { return f.expand(h) }
func (f *fakeNativeBridge) Collapse(h window.HWND) error                { return f.collapse(h) }

func TestNativeUIADeps_LookupAndActions(t *testing.T) {
	mk := func(h window.HWND) *uiaElement {
		return &uiaElement{HWND: h.String(), RuntimeID: h.String(), Name: h.String()}
	}
	deps := &nativeUIADeps{bridge: &fakeNativeBridge{
		resolveRoot:   func(h window.HWND) (*uiaElement, error) { return mk(h), nil },
		elementByHWND: func(h window.HWND) (*uiaElement, error) { return mk(h), nil },
		parentHWND:    func(window.HWND) (window.HWND, bool, error) { return 0, false, nil },
		childHWNDs:    func(window.HWND) ([]window.HWND, error) { return nil, nil },
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
	if err != nil || !strings.HasPrefix(root.Ref, "uia:") {
		t.Fatalf("ResolveWindowRoot failed: %+v, %v", root, err)
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
	if err != nil || !strings.HasPrefix(under.Ref, "uia:") {
		t.Fatalf("ElementFromPoint failed: %+v err=%v", under, err)
	}
}

func TestNativeUIADeps_TreeMethodsAreGuarded(t *testing.T) {
	bridge := &fakeNativeBridge{
		resolveRoot:   func(window.HWND) (*uiaElement, error) { return nil, nil },
		elementByHWND: func(window.HWND) (*uiaElement, error) { return nil, nil },
		parentHWND:    func(window.HWND) (window.HWND, bool, error) { return 0x1, true, nil },
		childHWNDs:    func(window.HWND) ([]window.HWND, error) { return []window.HWND{0x2}, nil },
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
	}
	deps := &nativeUIADeps{bridge: bridge}

	if _, err := deps.GetParent(context.Background(), "uia:sess:2"); !errors.Is(err, ErrProviderActionUnsupported) {
		t.Fatalf("expected guarded parent lookup, got %v", err)
	}
	if _, err := deps.GetChildren(context.Background(), "uia:sess:1"); !errors.Is(err, ErrProviderActionUnsupported) {
		t.Fatalf("expected guarded child traversal, got %v", err)
	}
	if _, _, err := deps.GetChildCount(context.Background(), "uia:sess:1"); !errors.Is(err, ErrProviderActionUnsupported) {
		t.Fatalf("expected guarded child count, got %v", err)
	}
	if bridge.parentCalls != 0 || bridge.childCalls != 0 {
		t.Fatalf("window traversal deps should never be called in UIA mode, parent=%d child=%d", bridge.parentCalls, bridge.childCalls)
	}
}

func TestNativeUIADeps_ActionDispatchAndInvalidRefs(t *testing.T) {
	called := map[string]string{}
	deps := &nativeUIADeps{bridge: &fakeNativeBridge{
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

	resolved, err := deps.ResolveWindowRoot(context.Background(), "0x2a")
	if err != nil {
		t.Fatalf("ResolveWindowRoot failed: %v", err)
	}
	ref := resolved.Ref
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
	if _, err := deps.ResolveWindowRoot(context.Background(), "bad"); !errors.Is(err, errUIANilElement) {
		t.Fatalf("expected nil element parse failure, got %v", err)
	}
	if _, err := deps.GetElementByRef(context.Background(), "bad"); !errors.Is(err, errUIANilElement) {
		t.Fatalf("expected invalid ref -> nil element, got %v", err)
	}
}

func TestNativeUIADeps_CacheLifecycle(t *testing.T) {
	sameElement := &uiaElement{HWND: "0x44", RuntimeID: "runtime-44", Name: "save"}
	deps := &nativeUIADeps{
		bridge: &fakeNativeBridge{
			resolveRoot:   func(window.HWND) (*uiaElement, error) { return cloneUIAElement(sameElement), nil },
			elementByHWND: func(window.HWND) (*uiaElement, error) { return cloneUIAElement(sameElement), nil },
			parentHWND:    func(window.HWND) (window.HWND, bool, error) { return 0, false, nil },
			childHWNDs:    func(window.HWND) ([]window.HWND, error) { return nil, nil },
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
		},
		sessionID:     "sess-a",
		refToElement:  map[string]*uiaElement{},
		keyToRefCache: map[string]string{},
	}
	first, err := deps.ResolveWindowRoot(context.Background(), "0x44")
	if err != nil {
		t.Fatalf("ResolveWindowRoot: %v", err)
	}
	second, err := deps.GetElementByRef(context.Background(), first.Ref)
	if err != nil {
		t.Fatalf("GetElementByRef: %v", err)
	}
	if first.Ref != second.Ref {
		t.Fatalf("expected stable opaque ref across session, got %q and %q", first.Ref, second.Ref)
	}
	_, err = deps.lookupByRef(makeUIANodeRef("sess-a", "missing"))
	var notFound *NodeRefNotFoundError
	if !errors.As(err, &notFound) {
		t.Fatalf("expected typed not found error, got %v", err)
	}
}

func TestNativeUIADeps_RejectsWindowRefs(t *testing.T) {
	deps := &nativeUIADeps{
		bridge:        &fakeNativeBridge{},
		sessionID:     "sess-a",
		refToElement:  map[string]*uiaElement{},
		keyToRefCache: map[string]string{},
	}
	if _, err := deps.GetElementByRef(context.Background(), "win:0x2a"); !errors.Is(err, errUIANilElement) {
		t.Fatalf("expected window ref rejection in UIA mode, got %v", err)
	}
}
