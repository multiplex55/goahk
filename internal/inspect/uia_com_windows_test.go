//go:build windows
// +build windows

package inspect

import (
	"errors"
	"testing"

	"goahk/internal/window"
)

type fakeUIAClient struct{}

func (fakeUIAClient) ElementFromHWND(window.HWND) (*uiaBridgeElement, error)  { return nil, nil }
func (fakeUIAClient) FocusedElement() (*uiaBridgeElement, error)              { return nil, nil }
func (fakeUIAClient) ElementFromPoint(int, int) (*uiaBridgeElement, error)    { return nil, nil }
func (fakeUIAClient) ElementByKey(string) (*uiaBridgeElement, error)          { return nil, nil }
func (fakeUIAClient) Parent(*uiaBridgeElement) (*uiaBridgeElement, error)     { return nil, nil }
func (fakeUIAClient) Children(*uiaBridgeElement) ([]*uiaBridgeElement, error) { return nil, nil }
func (fakeUIAClient) Invoke(*uiaBridgeElement) error                          { return nil }
func (fakeUIAClient) Select(*uiaBridgeElement) error                          { return nil }
func (fakeUIAClient) SetValue(*uiaBridgeElement, string) error                { return nil }
func (fakeUIAClient) DoDefaultAction(*uiaBridgeElement) error                 { return nil }
func (fakeUIAClient) Toggle(*uiaBridgeElement) error                          { return nil }
func (fakeUIAClient) Expand(*uiaBridgeElement) error                          { return nil }
func (fakeUIAClient) Collapse(*uiaBridgeElement) error                        { return nil }

func TestNewWin32UIABridge_UsesNativeClientWhenAvailable(t *testing.T) {
	orig := newUIAComClient
	defer func() { newUIAComClient = orig }()
	newUIAComClient = func() (uiaAutomationClient, error) { return fakeUIAClient{}, nil }

	bridge := newWin32UIABridge().(*win32UIAComBridge)
	if bridge.initErr != nil {
		t.Fatalf("expected nil initErr, got %v", bridge.initErr)
	}
}

func TestNewWin32UIABridge_FallsBackWithInitError(t *testing.T) {
	orig := newUIAComClient
	defer func() { newUIAComClient = orig }()
	newUIAComClient = func() (uiaAutomationClient, error) { return nil, errors.New("com init failed") }

	bridge := newWin32UIABridge().(*win32UIAComBridge)
	if bridge.initErr == nil {
		t.Fatal("expected init error")
	}
	if _, err := bridge.ResolveRoot(0); err == nil || !errors.Is(err, bridge.initErr) {
		t.Fatalf("expected wrapped init error, got %v", err)
	}
}

type fakeUiaNativeAPI struct {
	elementFromHandle func(*uiaWorkerState, window.HWND) (*uiaBridgeElement, error)
	focusedElement    func(*uiaWorkerState) (*uiaBridgeElement, error)
	elementFromPoint  func(*uiaWorkerState, int, int) (*uiaBridgeElement, error)
	findChildren      func(*uiaWorkerState, *uiaBridgeElement) ([]*uiaBridgeElement, error)
	getParent         func(*uiaWorkerState, *uiaBridgeElement) (*uiaBridgeElement, error)
}

func (f fakeUiaNativeAPI) ElementFromHandle(s *uiaWorkerState, h window.HWND) (*uiaBridgeElement, error) {
	return f.elementFromHandle(s, h)
}
func (f fakeUiaNativeAPI) FindChildren(s *uiaWorkerState, el *uiaBridgeElement) ([]*uiaBridgeElement, error) {
	return f.findChildren(s, el)
}
func (f fakeUiaNativeAPI) FocusedElement(s *uiaWorkerState) (*uiaBridgeElement, error) {
	return f.focusedElement(s)
}
func (f fakeUiaNativeAPI) ElementFromPoint(s *uiaWorkerState, x, y int) (*uiaBridgeElement, error) {
	return f.elementFromPoint(s, x, y)
}
func (f fakeUiaNativeAPI) GetParent(s *uiaWorkerState, el *uiaBridgeElement) (*uiaBridgeElement, error) {
	return f.getParent(s, el)
}

func TestNativeUIAComClient_ElementFromHandleAndChildrenParentPaths(t *testing.T) {
	orig := uiaNativeAPI
	defer func() { uiaNativeAPI = orig }()
	root := &uiaBridgeElement{Key: "rid:1", Element: &uiaElement{RuntimeID: "rid:1", Name: "Root", LocalizedControlType: "window", ControlType: "Window", ClassName: "Wnd", FrameworkID: "UIA", ProcessID: 123}}
	uiaNativeAPI = fakeUiaNativeAPI{
		elementFromHandle: func(_ *uiaWorkerState, _ window.HWND) (*uiaBridgeElement, error) { return root, nil },
		findChildren: func(_ *uiaWorkerState, _ *uiaBridgeElement) ([]*uiaBridgeElement, error) {
			return []*uiaBridgeElement{{Key: "rid:2", Element: &uiaElement{Name: ""}}, {Key: "rid:3", Element: &uiaElement{Name: "child"}}}, nil
		},
		focusedElement: func(_ *uiaWorkerState) (*uiaBridgeElement, error) {
			return &uiaBridgeElement{Key: "rid:f", Element: &uiaElement{Name: "focus"}}, nil
		},
		elementFromPoint: func(_ *uiaWorkerState, _, _ int) (*uiaBridgeElement, error) {
			return &uiaBridgeElement{Key: "rid:pt", Element: &uiaElement{Name: "point"}}, nil
		},
		getParent: func(_ *uiaWorkerState, _ *uiaBridgeElement) (*uiaBridgeElement, error) {
			return &uiaBridgeElement{Key: "rid:p", Element: &uiaElement{Name: "parent"}}, nil
		},
	}
	client, err := newNativeUIAComClient()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.(*nativeUIAComClient).worker.Close() })

	gotRoot, err := client.ElementFromHWND(window.HWND(0x44))
	if err != nil || gotRoot.Element.ProcessID != 123 {
		t.Fatalf("root err=%v root=%+v", err, gotRoot)
	}
	children, err := client.Children(root)
	if err != nil || len(children) != 2 || children[0].Element.Name != "" {
		t.Fatalf("children err=%v kids=%+v", err, children)
	}
	parent, err := client.Parent(root)
	if err != nil || parent.Element.Name != "parent" {
		t.Fatalf("parent err=%v parent=%+v", err, parent)
	}
	if focus, err := client.FocusedElement(); err != nil || focus.Key != "rid:f" {
		t.Fatalf("focus err=%v focus=%+v", err, focus)
	}
	if under, err := client.ElementFromPoint(10, 20); err != nil || under.Key != "rid:pt" {
		t.Fatalf("point err=%v under=%+v", err, under)
	}
}

func TestNativeUIAComClient_ErrorPathsSurface(t *testing.T) {
	orig := uiaNativeAPI
	defer func() { uiaNativeAPI = orig }()
	uiaNativeAPI = fakeUiaNativeAPI{
		elementFromHandle: func(_ *uiaWorkerState, _ window.HWND) (*uiaBridgeElement, error) {
			return nil, errors.New("ElementFromHandle failure")
		},
		findChildren: func(_ *uiaWorkerState, _ *uiaBridgeElement) ([]*uiaBridgeElement, error) {
			return nil, errors.New("FindAll failure")
		},
		focusedElement:   func(_ *uiaWorkerState) (*uiaBridgeElement, error) { return nil, errors.New("focus failure") },
		elementFromPoint: func(_ *uiaWorkerState, _, _ int) (*uiaBridgeElement, error) { return nil, errors.New("point failure") },
		getParent:        func(_ *uiaWorkerState, _ *uiaBridgeElement) (*uiaBridgeElement, error) { return nil, nil },
	}
	client, err := newNativeUIAComClient()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.(*nativeUIAComClient).worker.Close() })
	if _, err := client.ElementFromHWND(window.HWND(0x55)); err == nil || err.Error() != "ElementFromHandle failure" {
		t.Fatalf("unexpected err %v", err)
	}
	if _, err := client.Children(&uiaBridgeElement{Key: "rid:1", Element: &uiaElement{RuntimeID: "rid:1"}}); err == nil || err.Error() != "FindAll failure" {
		t.Fatalf("unexpected err %v", err)
	}
}

func TestNativeUIAComClient_PatternActions_InvokeSupportedExecutesOnce(t *testing.T) {
	clientAny, err := newNativeUIAComClient()
	if err != nil {
		t.Fatal(err)
	}
	client := clientAny.(*nativeUIAComClient)
	t.Cleanup(func() { _ = client.worker.Close() })

	el := &uiaBridgeElement{Key: "rid:101", NativePtr: 0x1234, SupportedPatterns: []string{"Invoke"}, Element: &uiaElement{RuntimeID: "101"}}
	client.cacheBridgeElement(el)

	orig := invokePatternCall
	defer func() { invokePatternCall = orig }()
	calls := 0
	invokePatternCall = func(ptr uintptr) error {
		calls++
		if ptr != el.NativePtr {
			t.Fatalf("unexpected ptr: got %x want %x", ptr, el.NativePtr)
		}
		return nil
	}

	if err := client.Invoke(el); err != nil {
		t.Fatalf("Invoke returned error: %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 invoke call, got %d", calls)
	}
}

func TestNativeUIAComClient_PatternActions_UnsupportedPatternReturnsCapabilityError(t *testing.T) {
	clientAny, err := newNativeUIAComClient()
	if err != nil {
		t.Fatal(err)
	}
	client := clientAny.(*nativeUIAComClient)
	t.Cleanup(func() { _ = client.worker.Close() })

	el := &uiaBridgeElement{Key: "rid:102", NativePtr: 0x4321, SupportedPatterns: []string{"Invoke"}, Element: &uiaElement{RuntimeID: "102"}}
	client.cacheBridgeElement(el)

	err = client.Toggle(el)
	if err == nil {
		t.Fatal("expected unsupported pattern error")
	}
	var unavailable *UIAComUnavailableError
	if !errors.As(err, &unavailable) {
		t.Fatalf("expected UIAComUnavailableError, got %T (%v)", err, err)
	}
}

func TestNativeUIAComClient_PatternActions_StaleElementClassification(t *testing.T) {
	clientAny, err := newNativeUIAComClient()
	if err != nil {
		t.Fatal(err)
	}
	client := clientAny.(*nativeUIAComClient)
	t.Cleanup(func() { _ = client.worker.Close() })

	el := &uiaBridgeElement{Key: "rid:does-not-exist", NativePtr: 0x1111, SupportedPatterns: []string{"Value"}, Element: &uiaElement{RuntimeID: "does-not-exist"}}
	err = client.SetValue(el, "")
	if err == nil {
		t.Fatal("expected stale error")
	}
	var stale *UIAElementStaleError
	if !errors.As(err, &stale) {
		t.Fatalf("expected UIAElementStaleError, got %T (%v)", err, err)
	}
}

func TestNativeUIAComClient_ParentReconstructionReturnsDirectParent(t *testing.T) {
	orig := uiaNativeAPI
	defer func() { uiaNativeAPI = orig }()
	uiaNativeAPI = fakeUiaNativeAPI{
		elementFromHandle: func(_ *uiaWorkerState, _ window.HWND) (*uiaBridgeElement, error) {
			return &uiaBridgeElement{Key: "rid:root", Element: &uiaElement{Name: "root"}}, nil
		},
		findChildren:     func(_ *uiaWorkerState, _ *uiaBridgeElement) ([]*uiaBridgeElement, error) { return nil, nil },
		focusedElement:   func(_ *uiaWorkerState) (*uiaBridgeElement, error) { return nil, nil },
		elementFromPoint: func(_ *uiaWorkerState, _, _ int) (*uiaBridgeElement, error) { return nil, nil },
		getParent: func(_ *uiaWorkerState, _ *uiaBridgeElement) (*uiaBridgeElement, error) {
			return &uiaBridgeElement{Key: "rid:parent", Element: &uiaElement{Name: "parent"}}, nil
		},
	}
	client, err := newNativeUIAComClient()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.(*nativeUIAComClient).worker.Close() })
	p, err := client.Parent(&uiaBridgeElement{Key: "rid:child", Element: &uiaElement{Name: "child"}})
	if err != nil || p.Key != "rid:parent" {
		t.Fatalf("expected direct parent, got parent=%+v err=%v", p, err)
	}
}

func TestCanonicalUIAKey(t *testing.T) {
	cases := []struct {
		raw         string
		isRuntimeID bool
		want        string
	}{
		{raw: "rid:1.2", isRuntimeID: true, want: "rid:1.2"},
		{raw: "ptr:abc", isRuntimeID: true, want: "ptr:abc"},
		{raw: "path:root/0/Button/OK", isRuntimeID: false, want: "path:root/0/Button/OK"},
		{raw: "fallback:root", isRuntimeID: false, want: "fallback:root"},
		{raw: "1.2.3", isRuntimeID: true, want: "rid:1.2.3"},
		{raw: "1.2.3", isRuntimeID: false, want: "1.2.3"},
	}
	for _, tc := range cases {
		if got := canonicalUIAKey(tc.raw, tc.isRuntimeID); got != tc.want {
			t.Fatalf("canonicalUIAKey(%q, %t) = %q, want %q", tc.raw, tc.isRuntimeID, got, tc.want)
		}
	}
}

func TestUIAVTableIndexConstants(t *testing.T) {
	tests := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"IUIAutomation.CreateTrueCondition", uiaVTableIUIAutomationCreateTrueCondition, 21},
		{"IUIAutomation.get_RawViewWalker", uiaVTableIUIAutomationGetRawViewWalker, 16},
		{"IUIAutomation.ElementFromHandle", uiaVTableIUIAutomationElementFromHandle, 6},
		{"IUIAutomation.ElementFromPoint", uiaVTableIUIAutomationElementFromPoint, 7},
		{"IUIAutomation.GetFocusedElement", uiaVTableIUIAutomationGetFocusedElement, 8},
		{"IUIAutomationElement.FindAll", uiaVTableIUIAutomationElementFindAll, 6},
		{"IUIAutomationElement.GetCurrentRuntimeId", uiaVTableIUIAutomationElementGetCurrentRuntimeID, 9},
		{"IUIAutomationElement.GetCurrentPropertyValue", uiaVTableIUIAutomationElementGetCurrentPropertyValue, 10},
		{"IUIAutomationElement.GetCurrentPattern", uiaVTableIUIAutomationElementGetCurrentPattern, 11},
		{"IUIAutomationElementArray.Length", uiaVTableIUIAutomationElementArrayLength, 3},
		{"IUIAutomationElementArray.GetElement", uiaVTableIUIAutomationElementArrayGetElement, 4},
		{"IUIAutomationTreeWalker.GetParentElement", uiaVTableIUIAutomationTreeWalkerGetParentElement, 3},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Fatalf("%s constant mismatch: got %d want %d", tc.name, tc.got, tc.want)
			}
		})
	}
}

func TestNativeUIAComClient_ElementByKey_Lookup(t *testing.T) {
	clientAny, err := newNativeUIAComClient()
	if err != nil {
		t.Fatal(err)
	}
	client := clientAny.(*nativeUIAComClient)
	t.Cleanup(func() { _ = client.worker.Close() })

	client.cacheBridgeElement(&uiaBridgeElement{Key: "rid:100", NativePtr: 0x10, Element: &uiaElement{Name: "rid-node"}})
	client.cacheBridgeElement(&uiaBridgeElement{Key: "path:rid:100/0/Button/Save", NativePtr: 0x11, Element: &uiaElement{Name: "save"}})

	if _, err := client.ElementByKey("100"); err == nil {
		t.Fatalf("expected stale lookup for non-key runtime-id input")
	}
	if got, err := client.ElementByKey("rid:100"); err != nil || got.Element.Name != "rid-node" {
		t.Fatalf("expected rid lookup success, got=%+v err=%v", got, err)
	}
	if got, err := client.ElementByKey("path:rid:100/0/Button/Save"); err != nil || got.Element.Name != "save" {
		t.Fatalf("expected path lookup success, got=%+v err=%v", got, err)
	}
}
