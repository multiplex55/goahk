//go:build windows
// +build windows

package inspect

import (
	"context"
	"errors"
	"testing"

	"goahk/internal/window"
)

type fakeACCBridge struct {
	root     *accBridgeElement
	byKey    map[string]*accBridgeElement
	children map[string][]*accBridgeElement
	point    *accBridgeElement
}

func (f *fakeACCBridge) ObjectFromWindow(window.HWND) (*accBridgeElement, error) {
	if f.root == nil {
		return nil, errors.New("missing root")
	}
	copy := *f.root
	return &copy, nil
}
func (f *fakeACCBridge) ObjectFromPoint(int, int) (*accBridgeElement, error) {
	if f.point == nil {
		return nil, errors.New("missing point")
	}
	copy := *f.point
	return &copy, nil
}
func (f *fakeACCBridge) ObjectByKey(key string) (*accBridgeElement, error) {
	if el, ok := f.byKey[key]; ok {
		copy := *el
		return &copy, nil
	}
	return nil, errors.New("missing key")
}
func (f *fakeACCBridge) Parent(el *accBridgeElement) (*accBridgeElement, error) {
	if el == nil || el.ParentKey == "" {
		return nil, nil
	}
	return f.ObjectByKey(el.ParentKey)
}
func (f *fakeACCBridge) Children(el *accBridgeElement) ([]*accBridgeElement, error) {
	if el == nil {
		return nil, nil
	}
	items := f.children[el.Key]
	out := make([]*accBridgeElement, 0, len(items))
	for _, item := range items {
		copy := *item
		out = append(out, &copy)
	}
	return out, nil
}
func (f *fakeACCBridge) CursorPosition() (int, int, error) { return 11, 22, nil }

func TestWindowsProvider_ModeSwitchUsesDistinctBackends(t *testing.T) {
	uia := &fakeAdapter{root: &uiaElement{Ref: "uia-root", RuntimeID: "1", Name: "UIA Root"}}
	acc := &fakeAdapter{root: &uiaElement{Ref: makeACCNodeRef("s", "1"), RuntimeID: "1", Name: "ACC Root"}}
	provider := newWindowsProviderWithModeAdapters(newUIAAdapter(uia), newUIAAdapter(acc), newUIAAdapter(acc), &fakeWindowAdapter{}).(*windowsProvider)

	uiaRoot, err := provider.GetTreeRoot(context.Background(), GetTreeRootRequest{HWND: "0x1", Mode: InspectModeUIATree})
	if err != nil {
		t.Fatalf("uia root: %v", err)
	}
	if uiaRoot.Source.Provider != "uia" || uiaRoot.Source.Source != "uia" {
		t.Fatalf("expected UIA source metadata, got %+v", uiaRoot.Source)
	}
	if _, err := provider.GetTreeRoot(context.Background(), GetTreeRootRequest{HWND: "0x1", Mode: InspectModeWindowTree}); err != nil {
		t.Fatalf("acc root: %v", err)
	}
	if uia.resolveRootCalls != 1 || acc.resolveRootCalls != 1 {
		t.Fatalf("expected one resolve per backend, got uia=%d acc=%d", uia.resolveRootCalls, acc.resolveRootCalls)
	}
	if provider.uiaCore == provider.accCore {
		t.Fatalf("expected dedicated provider cores for UIA and ACC")
	}
}

func TestACCNodeRefParsingAndCrossProviderRejection(t *testing.T) {
	deps := &nativeACCDeps{
		bridge:       &fakeACCBridge{},
		sessionID:    "sess-acc",
		refToElement: map[string]*accBridgeElement{},
		keyToRef:     map[string]string{},
	}
	accRef := makeACCNodeRef("sess-acc", "1")
	deps.refToElement[accRef] = &accBridgeElement{Key: "k1", Name: "child"}

	if _, err := deps.lookupByRef(accRef); err != nil {
		t.Fatalf("expected acc ref to resolve, got %v", err)
	}
	if _, err := deps.lookupByRef(makeUIANodeRef("sess-acc", "1")); !errors.Is(err, ErrInvalidNodeRef) {
		t.Fatalf("expected UIA ref to be rejected by ACC provider, got %v", err)
	}
	if _, err := deps.lookupByRef(makeACCNodeRef("other", "1")); !errors.Is(err, ErrInvalidNodeRef) {
		t.Fatalf("expected ACC ref from another session to be rejected, got %v", err)
	}
}

func TestACCTraversalAndPropertyMapping(t *testing.T) {
	rootRect := &Rect{Left: 10, Top: 20, Width: 300, Height: 200}
	childRect := &Rect{Left: 40, Top: 50, Width: 60, Height: 30}
	value := "OK"
	bridge := &fakeACCBridge{
		root: &accBridgeElement{Key: "root", RuntimeID: "root", HWND: "0x1", Name: "Window", Role: "Window", ClassName: "MainWnd", Framework: "MSAA", Rect: rootRect},
		byKey: map[string]*accBridgeElement{
			"root":  {Key: "root", RuntimeID: "root", HWND: "0x1", Name: "Window", Role: "Window", ClassName: "MainWnd", Framework: "MSAA", Rect: rootRect},
			"child": {Key: "child", ParentKey: "root", RuntimeID: "child", HWND: "0x1", Name: "OK", Role: "PushButton", ClassName: "Button", Framework: "MSAA", Rect: childRect, Value: &value},
		},
		children: map[string][]*accBridgeElement{
			"root": {
				{Key: "child", ParentKey: "root", RuntimeID: "child", HWND: "0x1", Name: "OK", Role: "PushButton", ClassName: "Button", Framework: "MSAA", Rect: childRect, Value: &value, NativeObj: 0x200, NativeChildID: 0, PathSegment: "PushButton[1]"},
				{Key: "childid:1", ParentKey: "root", RuntimeID: "childid:1", HWND: "0x1", Name: "", Role: "StaticText", ClassName: "Static", Framework: "MSAA", Rect: &Rect{Left: 41, Top: 90, Width: 80, Height: 20}, NativeObj: 0x200, NativeChildID: 1, PathSegment: "StaticText[1]"},
			},
		},
		point: &accBridgeElement{Key: "child", ParentKey: "root", RuntimeID: "child", HWND: "0x1", Name: "OK", Role: "PushButton", ClassName: "Button", Framework: "MSAA", Rect: childRect, Value: &value, NativeObj: 0x200, NativeChildID: 0, PathSegment: "PushButton[1]"},
	}
	deps := &nativeACCDeps{bridge: bridge, sessionID: "sess", refToElement: map[string]*accBridgeElement{}, keyToRef: map[string]string{}}
	provider := newWindowsProviderWithModeAdapters(newUIAAdapter(&fakeAdapter{root: &uiaElement{Ref: "uia", RuntimeID: "1", Name: "UIA"}}), newUIAAdapter(deps), newUIAAdapter(deps), &fakeWindowAdapter{}).(*windowsProvider)

	rootResp, err := provider.GetTreeRoot(context.Background(), GetTreeRootRequest{HWND: "0x1", Mode: InspectModeWindowTree})
	if err != nil {
		t.Fatalf("GetTreeRoot: %v", err)
	}
	childrenResp, err := provider.GetNodeChildren(context.Background(), GetNodeChildrenRequest{NodeID: rootResp.Root.NodeID})
	if err != nil {
		t.Fatalf("GetNodeChildren: %v", err)
	}
	if len(childrenResp.Children) != 2 {
		t.Fatalf("expected two children (object + childID), got %d", len(childrenResp.Children))
	}
	if got := childrenResp.Children[0].NodeID; len(got) < len("node:acc:") || got[:len("node:acc:")] != "node:acc:" {
		t.Fatalf("expected ACC node namespace, got %q", got)
	}

	details, err := provider.GetNodeDetails(context.Background(), GetNodeDetailsRequest{NodeID: childrenResp.Children[0].NodeID})
	if err != nil {
		t.Fatalf("GetNodeDetails: %v", err)
	}
	if details.Source.Provider != "acc" {
		t.Fatalf("expected ACC source metadata, got %+v", details.Source)
	}
	if details.Element.Bounds == nil || details.Element.Bounds.Width != 60 {
		t.Fatalf("expected bounds from ACC location, got %+v", details.Element.Bounds)
	}
	if details.ACCPath == "" {
		t.Fatalf("expected generated ACC path")
	}
	details2, err := provider.GetNodeDetails(context.Background(), GetNodeDetailsRequest{NodeID: childrenResp.Children[1].NodeID})
	if err != nil {
		t.Fatalf("GetNodeDetails second: %v", err)
	}
	if details2.ACCPath == "" || details2.ACCPath == details.ACCPath {
		t.Fatalf("expected stable copyable distinct path, got first=%q second=%q", details.ACCPath, details2.ACCPath)
	}
}

func TestACCDetailsPropertyStatus_EmptyVsUnsupported(t *testing.T) {
	bridge := &fakeACCBridge{
		root:  &accBridgeElement{Key: "root", RuntimeID: "root", HWND: "0x1", Name: "Window", Role: "Window"},
		byKey: map[string]*accBridgeElement{"root": {Key: "root", RuntimeID: "root", HWND: "0x1", Name: "Window", Role: "Window"}},
	}
	deps := &nativeACCDeps{bridge: bridge, sessionID: "sess", refToElement: map[string]*accBridgeElement{}, keyToRef: map[string]string{}}
	provider := newWindowsProviderWithModeAdapters(newUIAAdapter(deps), newUIAAdapter(deps), newUIAAdapter(deps), &fakeWindowAdapter{}).(*windowsProvider)
	rootResp, _ := provider.GetTreeRoot(context.Background(), GetTreeRootRequest{HWND: "0x1", Mode: InspectModeWindowTree})
	details, err := provider.GetNodeDetails(context.Background(), GetNodeDetailsRequest{NodeID: rootResp.Root.NodeID})
	if err != nil {
		t.Fatalf("details err: %v", err)
	}
	status := map[string]string{}
	for _, p := range details.Properties {
		status[p.Name] = p.Status
	}
	if status["Value"] != "empty" || status["HelpText"] != "unsupported" {
		t.Fatalf("expected Value=empty and HelpText=unsupported, got %#v", status)
	}
}
