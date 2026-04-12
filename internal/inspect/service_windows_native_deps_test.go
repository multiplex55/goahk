//go:build windows
// +build windows

package inspect

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"goahk/internal/window"
)

type fakeNativeBridge struct {
	resolveRoot   func(window.HWND) (*uiaBridgeElement, error)
	focused       func() (*uiaBridgeElement, error)
	fromPoint     func(int, int) (*uiaBridgeElement, error)
	byKey         func(string) (*uiaBridgeElement, error)
	parent        func(*uiaBridgeElement) (*uiaBridgeElement, error)
	children      func(*uiaBridgeElement) ([]*uiaBridgeElement, error)
	cursorPos     func() (int, int, error)
	invoke        func(*uiaBridgeElement) error
	selectFn      func(*uiaBridgeElement) error
	setValue      func(*uiaBridgeElement, string) error
	defaultAction func(*uiaBridgeElement) error
	toggle        func(*uiaBridgeElement) error
	expand        func(*uiaBridgeElement) error
	collapse      func(*uiaBridgeElement) error
	resolveRootN  int
	focusedN      int
	fromPointN    int
	parentN       int
	childrenN     int
	elementByKeyN int
}

func (f *fakeNativeBridge) ResolveRoot(h window.HWND) (*uiaBridgeElement, error) {
	f.resolveRootN++
	return f.resolveRoot(h)
}
func (f *fakeNativeBridge) FocusedElement() (*uiaBridgeElement, error) {
	f.focusedN++
	return f.focused()
}
func (f *fakeNativeBridge) ElementFromPoint(x, y int) (*uiaBridgeElement, error) {
	f.fromPointN++
	return f.fromPoint(x, y)
}
func (f *fakeNativeBridge) ElementByKey(key string) (*uiaBridgeElement, error) {
	f.elementByKeyN++
	return f.byKey(key)
}
func (f *fakeNativeBridge) Parent(el *uiaBridgeElement) (*uiaBridgeElement, error) {
	f.parentN++
	return f.parent(el)
}
func (f *fakeNativeBridge) Children(el *uiaBridgeElement) ([]*uiaBridgeElement, error) {
	f.childrenN++
	return f.children(el)
}
func (f *fakeNativeBridge) CursorPosition() (int, int, error) { return f.cursorPos() }
func (f *fakeNativeBridge) Invoke(el *uiaBridgeElement) error { return f.invoke(el) }
func (f *fakeNativeBridge) Select(el *uiaBridgeElement) error { return f.selectFn(el) }
func (f *fakeNativeBridge) SetValue(el *uiaBridgeElement, v string) error {
	return f.setValue(el, v)
}
func (f *fakeNativeBridge) DoDefaultAction(el *uiaBridgeElement) error { return f.defaultAction(el) }
func (f *fakeNativeBridge) Toggle(el *uiaBridgeElement) error          { return f.toggle(el) }
func (f *fakeNativeBridge) Expand(el *uiaBridgeElement) error          { return f.expand(el) }
func (f *fakeNativeBridge) Collapse(el *uiaBridgeElement) error        { return f.collapse(el) }

func bridgeEl(key, hwnd, name string, patterns ...string) *uiaBridgeElement {
	return &uiaBridgeElement{Key: key, Element: &uiaElement{RuntimeID: key, HWND: hwnd, Name: name}, SupportedPatterns: patterns}
}

func newBridgeFixture() *fakeNativeBridge {
	return &fakeNativeBridge{
		resolveRoot: func(h window.HWND) (*uiaBridgeElement, error) {
			return bridgeEl(fmt.Sprintf("rid:%s", h.String()), h.String(), "root"), nil
		},
		focused:   func() (*uiaBridgeElement, error) { return bridgeEl("rid:focus", "0x9", "focus", "Invoke"), nil },
		fromPoint: func(_, _ int) (*uiaBridgeElement, error) { return bridgeEl("rid:point", "0xa", "under"), nil },
		byKey:     func(key string) (*uiaBridgeElement, error) { return bridgeEl(key, "0x2a", "node"), nil },
		parent: func(el *uiaBridgeElement) (*uiaBridgeElement, error) {
			return bridgeEl("rid:parent", "0x1", "parent"), nil
		},
		children: func(el *uiaBridgeElement) ([]*uiaBridgeElement, error) {
			return []*uiaBridgeElement{bridgeEl("rid:c1", "0x11", "c1"), bridgeEl("rid:c2", "0x12", "c2")}, nil
		},
		cursorPos:     func() (int, int, error) { return 10, 20, nil },
		invoke:        func(*uiaBridgeElement) error { return nil },
		selectFn:      func(*uiaBridgeElement) error { return nil },
		setValue:      func(*uiaBridgeElement, string) error { return nil },
		defaultAction: func(*uiaBridgeElement) error { return nil },
		toggle:        func(*uiaBridgeElement) error { return nil },
		expand:        func(*uiaBridgeElement) error { return nil },
		collapse:      func(*uiaBridgeElement) error { return nil },
	}
}

func TestNativeUIADeps_DispatchResolution(t *testing.T) {
	bridge := newBridgeFixture()
	deps := &nativeUIADeps{bridge: bridge, sessionID: "sess", refToElement: map[string]*cachedBridgeElement{}, keyToRef: map[string]string{}}

	if _, err := deps.ResolveWindowRoot(context.Background(), "0x1"); err != nil {
		t.Fatalf("ResolveWindowRoot: %v", err)
	}
	if _, err := deps.GetFocusedElement(context.Background()); err != nil {
		t.Fatalf("GetFocusedElement: %v", err)
	}
	if _, err := deps.ElementFromPoint(context.Background(), 5, 6); err != nil {
		t.Fatalf("ElementFromPoint: %v", err)
	}
	if bridge.resolveRootN != 1 || bridge.focusedN != 1 || bridge.fromPointN != 1 {
		t.Fatalf("dispatch counts mismatch root=%d focus=%d point=%d", bridge.resolveRootN, bridge.focusedN, bridge.fromPointN)
	}
}

func TestNativeUIADeps_ParentAndChildrenTraversal(t *testing.T) {
	bridge := newBridgeFixture()
	deps := &nativeUIADeps{bridge: bridge, sessionID: "sess", refToElement: map[string]*cachedBridgeElement{}, keyToRef: map[string]string{}}
	root, _ := deps.ResolveWindowRoot(context.Background(), "0x1")

	parent, err := deps.GetParent(context.Background(), root.Ref)
	if err != nil || parent.Name != "parent" {
		t.Fatalf("GetParent failed parent=%+v err=%v", parent, err)
	}
	children, err := deps.GetChildren(context.Background(), root.Ref)
	if err != nil || len(children) != 2 {
		t.Fatalf("GetChildren failed children=%+v err=%v", children, err)
	}
	if children[0].ParentRef != root.Ref {
		t.Fatalf("expected child parent ref=%q got=%q", root.Ref, children[0].ParentRef)
	}
	if bridge.parentN != 1 || bridge.childrenN != 1 {
		t.Fatalf("dispatch counts parent=%d children=%d", bridge.parentN, bridge.childrenN)
	}
}

func TestNativeUIADeps_StaleRetryRequiresFallbackMarker(t *testing.T) {
	root := bridgeEl("rid:stale", "0x2a", "stale")
	root.AllowHWNDFallback = true
	calls := 0
	bridge := newBridgeFixture()
	bridge.resolveRoot = func(h window.HWND) (*uiaBridgeElement, error) { return root, nil }
	bridge.byKey = func(key string) (*uiaBridgeElement, error) {
		calls++
		if calls == 1 {
			return nil, &UIAElementStaleError{Op: "ElementByRuntimeID", Err: errors.New("stale")}
		}
		return bridgeEl("rid:stale", "0x2a", "fresh"), nil
	}
	deps := &nativeUIADeps{bridge: bridge, sessionID: "sess", refToElement: map[string]*cachedBridgeElement{}, keyToRef: map[string]string{}}
	registered, _ := deps.ResolveWindowRoot(context.Background(), "0x2a")

	got, err := deps.GetElementByRef(context.Background(), registered.Ref)
	if err != nil {
		t.Fatalf("GetElementByRef retry failed: %v", err)
	}
	if got.Name != "fresh" || calls != 2 {
		t.Fatalf("expected retry refresh result, got=%+v calls=%d", got, calls)
	}

	// no fallback marker -> stale surfaces
	rootNoFallback := bridgeEl("rid:no-fallback", "0x2b", "stale")
	bridge.resolveRoot = func(h window.HWND) (*uiaBridgeElement, error) { return rootNoFallback, nil }
	calls = 0
	bridge.byKey = func(string) (*uiaBridgeElement, error) {
		calls++
		return nil, &UIAElementStaleError{Op: "ElementByRuntimeID", Err: errors.New("stale")}
	}
	registered, _ = deps.ResolveWindowRoot(context.Background(), "0x2b")
	if _, err := deps.GetElementByRef(context.Background(), registered.Ref); err == nil || !strings.Contains(err.Error(), "stale") {
		t.Fatalf("expected stale error without fallback marker, got %v", err)
	}
}

func TestNativeUIADeps_BridgeMappingIncludesPatternsAndUnsupportedProps(t *testing.T) {
	bridge := newBridgeFixture()
	bridge.resolveRoot = func(h window.HWND) (*uiaBridgeElement, error) {
		el := bridgeEl("rid:mapped", h.String(), "mapped", "Value", "Invoke")
		el.UnsupportedProperty = map[string]bool{"HelpText": true}
		return el, nil
	}
	deps := &nativeUIADeps{bridge: bridge, sessionID: "sess", refToElement: map[string]*cachedBridgeElement{}, keyToRef: map[string]string{}}
	got, err := deps.ResolveWindowRoot(context.Background(), "0x31")
	if err != nil {
		t.Fatalf("ResolveWindowRoot: %v", err)
	}
	if len(got.SupportedPatterns) != 2 || !got.UnsupportedProps["HelpText"] {
		t.Fatalf("mapping mismatch element=%+v", got)
	}
}

func TestNativeUIADeps_BridgePropertyAbsenceStateTranslation(t *testing.T) {
	bridge := newBridgeFixture()
	bridge.resolveRoot = func(h window.HWND) (*uiaBridgeElement, error) {
		el := bridgeEl("rid:states", h.String(), "states")
		el.UnsupportedProperty = map[string]bool{
			"HelpText":  true,  // unsupported by provider
			"ItemType":  false, // supported but empty
			"IsEnabled": false, // explicit false must not become unsupported
		}
		el.PropertyState = map[string]string{
			"Value": propertyStatusUnavailable,
			"Name":  propertyStatusStale,
		}
		el.Element.IsEnabled = false
		return el, nil
	}
	deps := &nativeUIADeps{bridge: bridge, sessionID: "sess", refToElement: map[string]*cachedBridgeElement{}, keyToRef: map[string]string{}}
	got, err := deps.ResolveWindowRoot(context.Background(), "0x44")
	if err != nil {
		t.Fatalf("ResolveWindowRoot: %v", err)
	}
	if got.PropertyStates["HelpText"] != propertyStatusUnsupported {
		t.Fatalf("expected unsupported status for HelpText, got %q", got.PropertyStates["HelpText"])
	}
	if got.PropertyStates["ItemType"] != propertyStatusEmpty {
		t.Fatalf("expected empty status for ItemType, got %q", got.PropertyStates["ItemType"])
	}
	if got.PropertyStates["Value"] != propertyStatusUnavailable {
		t.Fatalf("expected unavailable status for Value, got %q", got.PropertyStates["Value"])
	}
	if got.PropertyStates["Name"] != propertyStatusStale {
		t.Fatalf("expected stale status for Name, got %q", got.PropertyStates["Name"])
	}
	if got.UnsupportedProps["IsEnabled"] {
		t.Fatalf("explicit false must not be marked unsupported")
	}
}
