//go:build windows
// +build windows

package inspect

import (
	"errors"
	"fmt"
	"strings"
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
func (fakeUIAClient) Close() error                                            { return nil }

func TestWrapNativeElementOwnedAndBorrowed_Ownership(t *testing.T) {
	origRuntime := uiaElementRuntimeIDCall
	defer func() { uiaElementRuntimeIDCall = origRuntime }()
	uiaElementRuntimeIDCall = func(uintptr) (string, error) { return "rid:own", nil }
	owned, _ := wrapNativeElementOwned(0x1, 0, "", -1, &uiaWrapOptions{PropertyLoadLevel: uiaPropertyLoadTree, PopulatePatterns: false})
	borrowed, _ := wrapNativeElementBorrowed(0x2, 0, "", -1)
	if owned == nil || !owned.OwnsNativePtr {
		t.Fatalf("expected owned wrapper")
	}
	if borrowed == nil || borrowed.OwnsNativePtr {
		t.Fatalf("expected borrowed wrapper")
	}
}

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

func TestNativeUIAComClient_ChildrenReturnsPartialWithDiagnostic(t *testing.T) {
	orig := uiaNativeAPI
	defer func() { uiaNativeAPI = orig }()
	uiaNativeAPI = fakeUiaNativeAPI{
		elementFromHandle: func(_ *uiaWorkerState, _ window.HWND) (*uiaBridgeElement, error) { return nil, nil },
		findChildren: func(_ *uiaWorkerState, _ *uiaBridgeElement) ([]*uiaBridgeElement, error) {
			return []*uiaBridgeElement{{Key: "rid:ok", Element: &uiaElement{Name: "ok"}}}, errors.New("child[1]: wrap failed")
		},
		focusedElement:   func(_ *uiaWorkerState) (*uiaBridgeElement, error) { return nil, nil },
		elementFromPoint: func(_ *uiaWorkerState, _, _ int) (*uiaBridgeElement, error) { return nil, nil },
		getParent:        func(_ *uiaWorkerState, _ *uiaBridgeElement) (*uiaBridgeElement, error) { return nil, nil },
	}
	clientAny, err := newNativeUIAComClient()
	if err != nil {
		t.Fatal(err)
	}
	client := clientAny.(*nativeUIAComClient)
	t.Cleanup(func() { _ = client.worker.Close() })

	children, gotErr := client.Children(&uiaBridgeElement{Key: "rid:parent", NativePtr: 0x1, Element: &uiaElement{Name: "parent"}})
	if len(children) != 1 || children[0].Key != "rid:ok" {
		t.Fatalf("expected partial child success, got %+v", children)
	}
	if gotErr == nil || !strings.Contains(gotErr.Error(), "wrap failed") {
		t.Fatalf("expected diagnostic error, got %v", gotErr)
	}
}

func TestNativeUIAComClient_TreeWrappingSucceedsWhenPatternProbeFails(t *testing.T) {
	origNative := uiaNativeAPI
	origProbe := patternProbeCall
	defer func() {
		uiaNativeAPI = origNative
		patternProbeCall = origProbe
	}()

	root := &uiaBridgeElement{
		Key:     "rid:tree-root",
		Element: &uiaElement{Name: "Root"},
	}
	uiaNativeAPI = fakeUiaNativeAPI{
		elementFromHandle: func(_ *uiaWorkerState, _ window.HWND) (*uiaBridgeElement, error) { return root, nil },
		findChildren:      func(_ *uiaWorkerState, _ *uiaBridgeElement) ([]*uiaBridgeElement, error) { return nil, nil },
		focusedElement:    func(_ *uiaWorkerState) (*uiaBridgeElement, error) { return nil, nil },
		elementFromPoint:  func(_ *uiaWorkerState, _, _ int) (*uiaBridgeElement, error) { return nil, nil },
		getParent:         func(_ *uiaWorkerState, _ *uiaBridgeElement) (*uiaBridgeElement, error) { return nil, nil },
	}
	patternProbeCall = func(uintptr, int32) (bool, error) { return false, errors.New("probe failure") }

	clientAny, err := newNativeUIAComClient()
	if err != nil {
		t.Fatal(err)
	}
	client := clientAny.(*nativeUIAComClient)
	t.Cleanup(func() { _ = client.worker.Close() })

	gotRoot, err := client.ElementFromHWND(window.HWND(0x66))
	if err != nil || gotRoot == nil {
		t.Fatalf("tree load failed: root=%+v err=%v", gotRoot, err)
	}
	if len(gotRoot.SupportedPatterns) != 0 {
		t.Fatalf("expected tree load to skip pattern probing, got %+v", gotRoot.SupportedPatterns)
	}
	detailsEl, err := client.ElementByKey(gotRoot.Key)
	if err != nil {
		t.Fatalf("details lookup should tolerate probe failures, err=%v", err)
	}
	if len(detailsEl.SupportedPatterns) != 0 {
		t.Fatalf("expected no discovered patterns when probes fail, got %+v", detailsEl.SupportedPatterns)
	}
}

func TestWrapNativeElementOwned_TreeLoadSkipsDetailsOnlyProperties(t *testing.T) {
	origGetProp := uiaGetCurrentPropertyValueCall
	origRuntime := uiaElementRuntimeIDCall
	defer func() {
		uiaGetCurrentPropertyValueCall = origGetProp
		uiaElementRuntimeIDCall = origRuntime
	}()

	called := map[int32]int{}
	uiaElementRuntimeIDCall = func(uintptr) (string, error) { return "rid:test", nil }
	uiaGetCurrentPropertyValueCall = func(_ uintptr, prop int32) (comVariant, error) {
		called[prop]++
		return comVariant{}, fmt.Errorf("test")
	}

	_, _ = wrapNativeElementOwned(0x1, 0, "", -1, &uiaWrapOptions{PropertyLoadLevel: uiaPropertyLoadTree, PopulatePatterns: false})

	if called[uiaPropertyAutomationID] != 0 {
		t.Fatalf("expected tree load to skip AutomationId property call, got %d", called[uiaPropertyAutomationID])
	}
	if called[uiaPropertyHelpText] != 0 {
		t.Fatalf("expected tree load to skip HelpText property call, got %d", called[uiaPropertyHelpText])
	}
	if called[uiaPropertyControlType] == 0 || called[uiaPropertyName] == 0 || called[uiaPropertyClassName] == 0 {
		t.Fatalf("expected tree load essentials to be called, got controlType=%d name=%d class=%d", called[uiaPropertyControlType], called[uiaPropertyName], called[uiaPropertyClassName])
	}
}

func TestNativeUIAFindChildren_NilArrayReturnsStale(t *testing.T) {
	clientAny, err := newNativeUIAComClient()
	if err != nil {
		t.Fatal(err)
	}
	client := clientAny.(*nativeUIAComClient)
	t.Cleanup(func() { _ = client.worker.Close() })

	origFindAll := uiaFindAllChildrenCall
	origLen := uiaArrayLengthCall
	origGet := uiaArrayGetCall
	defer func() {
		uiaFindAllChildrenCall = origFindAll
		uiaArrayLengthCall = origLen
		uiaArrayGetCall = origGet
	}()

	uiaFindAllChildrenCall = func(uintptr, uintptr) (uintptr, error) { return 0, nil }
	lengthCalled := false
	uiaArrayLengthCall = func(uintptr) (int32, error) { lengthCalled = true; return 0, nil }
	uiaArrayGetCall = func(uintptr, int32) (uintptr, error) { t.Fatal("should not be called"); return 0, nil }

	_, err = uiaNativeAPI.FindChildren(&uiaWorkerState{trueCond: 1}, &uiaBridgeElement{NativePtr: 1, Key: "rid:p", Element: &uiaElement{RuntimeID: "rid:p"}})
	var stale *UIAElementStaleError
	if !errors.As(err, &stale) {
		t.Fatalf("expected stale error, got %T (%v)", err, err)
	}
	if lengthCalled {
		t.Fatal("uiaArrayLength should not be called for nil array")
	}
}

func TestFallbackPathKey_RuntimeIDMissingUsesSiblingIndex(t *testing.T) {
	key := fallbackPathKey("rid:1", 3, &uiaElement{LocalizedControlType: "button", Name: "OK"})
	if key != "path:rid:1/3/button/OK" {
		t.Fatalf("unexpected fallback key: %q", key)
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
		{"IUnknown.QueryInterface", comVTableQueryInterface, 0},
		{"IUnknown.AddRef", comVTableAddRef, 1},
		{"IUnknown.Release", comVTableRelease, 2},
		{"IUIAutomation.CreateTrueCondition", uiaVTableIUIAutomationCreateTrueCondition, 21},
		{"IUIAutomation.get_RawViewWalker", uiaVTableIUIAutomationGetRawViewWalker, 16},
		{"IUIAutomation.ElementFromHandle", uiaVTableIUIAutomationElementFromHandle, 6},
		{"IUIAutomation.ElementFromPoint", uiaVTableIUIAutomationElementFromPoint, 7},
		{"IUIAutomation.GetFocusedElement", uiaVTableIUIAutomationGetFocusedElement, 8},
		{"IUIAutomationElement.FindAll", uiaVTableIUIAutomationElementFindAll, 6},
		{"IUIAutomationElement.GetRuntimeID", uiaVTableIUIAutomationElementGetRuntimeID, 4},
		{"IUIAutomationElement.GetCurrentPropertyValue", uiaVTableIUIAutomationElementGetCurrentPropertyValue, 10},
		{"IUIAutomationElement.GetCurrentPattern", uiaVTableIUIAutomationElementGetCurrentPattern, 16},
		{"IUIAutomationElementArray.Length", uiaVTableIUIAutomationElementArrayLength, 3},
		{"IUIAutomationElementArray.GetElement", uiaVTableIUIAutomationElementArrayGetElement, 4},
		{"IUIAutomationTreeWalker.GetParentElement", uiaVTableIUIAutomationTreeWalkerGetParentElement, 3},
		{"IUIAutomationInvokePattern.Invoke", uiaVTableIUIAutomationInvokePatternInvoke, 3},
		{"IUIAutomationSelectionItemPattern.Select", uiaVTableIUIAutomationSelectionItemPatternSelect, 3},
		{"IUIAutomationValuePattern.SetValue", uiaVTableIUIAutomationValuePatternSetValue, 3},
		{"IUIAutomationLegacyIAccessiblePattern.Select", uiaVTableIUIAutomationLegacyIAccessiblePatternSelect, 3},
		{"IUIAutomationLegacyIAccessiblePattern.DoDefaultAction", uiaVTableIUIAutomationLegacyIAccessiblePatternDoDefaultAction, 4},
		{"IUIAutomationLegacyIAccessiblePattern.SetValue", uiaVTableIUIAutomationLegacyIAccessiblePatternSetValue, 5},
		{"IUIAutomationTogglePattern.Toggle", uiaVTableIUIAutomationTogglePatternToggle, 3},
		{"IUIAutomationExpandCollapsePattern.Expand", uiaVTableIUIAutomationExpandCollapsePatternExpand, 3},
		{"IUIAutomationExpandCollapsePattern.Collapse", uiaVTableIUIAutomationExpandCollapsePatternCollapse, 4},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Fatalf("%s constant mismatch: got %d want %d", tc.name, tc.got, tc.want)
			}
		})
	}
}

func TestWithBSTR_AllocFreeLifecycle(t *testing.T) {
	origAlloc := sysAllocStringCall
	origFree := sysFreeStringCall
	defer func() {
		sysAllocStringCall = origAlloc
		sysFreeStringCall = origFree
	}()

	var allocs, frees int
	var freed uintptr
	sysAllocStringCall = func(s *uint16) uintptr {
		if s == nil {
			t.Fatal("expected utf16 string pointer")
		}
		allocs++
		return 0x99
	}
	sysFreeStringCall = func(bstr uintptr) {
		frees++
		freed = bstr
	}

	wantErr := errors.New("invoke failed")
	err := withBSTR("abc", "SetValue", func(bstr uintptr) error {
		if bstr != 0x99 {
			t.Fatalf("unexpected bstr: got 0x%x", bstr)
		}
		return wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected callback error, got %v", err)
	}
	if allocs != 1 || frees != 1 {
		t.Fatalf("expected one alloc/free, got allocs=%d frees=%d", allocs, frees)
	}
	if freed != 0x99 {
		t.Fatalf("unexpected freed ptr: got 0x%x", freed)
	}
}

func TestUIAElementRuntimeID_NilPointer(t *testing.T) {
	_, err := uiaElementRuntimeID(0)
	if err == nil || !strings.Contains(err.Error(), "GetRuntimeId") {
		t.Fatalf("expected GetRuntimeId nil element error, got %v", err)
	}
}

func TestSafeArrayRuntimeIDInts_NilSafeArray(t *testing.T) {
	got, err := safeArrayRuntimeIDInts(0)
	if err == nil {
		t.Fatal("expected SafeArray error for nil pointer")
	}
	if got != nil {
		t.Fatalf("expected nil runtime id on nil safe array, got %v", got)
	}
}

func TestSafeArrayRuntimeIDIntsInBounds_ValidParse(t *testing.T) {
	values := map[int32]int32{0: 1, 1: 2, 2: 3}
	got, err := safeArrayRuntimeIDIntsInBounds(0, 2, func(i int32) (int32, error) {
		return values[i], nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 3 || got[0] != 1 || got[1] != 2 || got[2] != 3 {
		t.Fatalf("unexpected runtime id parse: %v", got)
	}
}

func TestWrapNativeElementOwned_RuntimeIDAssignment(t *testing.T) {
	if got := runtimeIDString([]int{1, 2, 3}); got != "1.2.3" {
		t.Fatalf("runtimeIDString returned %q", got)
	}
	if got := canonicalUIAKey("1.2.3", true); got != "rid:1.2.3" {
		t.Fatalf("canonicalUIAKey should prefix runtime id, got %q", got)
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

func TestNativeUIAComClient_ElementByKey_DetailPopulationRunsOnWorker(t *testing.T) {
	origGetProp := uiaGetCurrentPropertyValueCall
	origRuntime := uiaElementRuntimeIDCall
	origProbe := patternProbeCall
	origObserver := uiaWorkerJobObserver
	defer func() {
		uiaGetCurrentPropertyValueCall = origGetProp
		uiaElementRuntimeIDCall = origRuntime
		patternProbeCall = origProbe
		uiaWorkerJobObserver = origObserver
	}()

	clientAny, err := newNativeUIAComClient()
	if err != nil {
		t.Fatal(err)
	}
	client := clientAny.(*nativeUIAComClient)
	t.Cleanup(func() { _ = client.worker.Close() })

	client.cacheBridgeElement(&uiaBridgeElement{Key: "rid:200", NativePtr: 0x20, Element: &uiaElement{Name: "node"}})

	insideElementByKeyJob := false
	uiaWorkerJobObserver = func(op string) {
		insideElementByKeyJob = op == "ElementByKey"
	}
	uiaElementRuntimeIDCall = func(uintptr) (string, error) { return "rid:200", nil }
	uiaGetCurrentPropertyValueCall = func(_ uintptr, _ int32) (comVariant, error) {
		if !insideElementByKeyJob {
			t.Fatalf("expected property read on worker-dispatched ElementByKey job")
		}
		return comVariant{}, errors.New("test property failure")
	}
	patternProbeCall = func(uintptr, int32) (bool, error) {
		if !insideElementByKeyJob {
			t.Fatalf("expected pattern probe on worker-dispatched ElementByKey job")
		}
		return false, errors.New("test probe failure")
	}

	if _, err := client.ElementByKey("rid:200"); err != nil {
		t.Fatalf("ElementByKey should still return cached element, err=%v", err)
	}
}

func TestNativeUIAComClient_ElementByKey_DisablePatternsSkipsPatternProbe(t *testing.T) {
	orig := GetUIAFeatureGates()
	defer SetUIAFeatureGates(orig)
	SetUIAFeatureGates(UIAFeatureGates{DisablePatterns: true, MaxInitialDepth: orig.MaxInitialDepth, MaxInitialNodes: orig.MaxInitialNodes, BranchTimeout: orig.BranchTimeout, TotalLoadTimeout: orig.TotalLoadTimeout})

	origProbe := patternProbeCall
	defer func() { patternProbeCall = origProbe }()
	patternProbeCall = func(uintptr, int32) (bool, error) {
		t.Fatalf("pattern probing should be disabled")
		return false, nil
	}

	clientAny, err := newNativeUIAComClient()
	if err != nil {
		t.Fatal(err)
	}
	client := clientAny.(*nativeUIAComClient)
	t.Cleanup(func() { _ = client.worker.Close() })

	client.cacheBridgeElement(&uiaBridgeElement{Key: "rid:201", NativePtr: 0x21, Element: &uiaElement{Name: "node"}})
	got, err := client.ElementByKey("rid:201")
	if err != nil {
		t.Fatalf("ElementByKey failed with DisablePatterns enabled: %v", err)
	}
	if len(got.SupportedPatterns) != 0 {
		t.Fatalf("expected no patterns when DisablePatterns=true, got %+v", got.SupportedPatterns)
	}
}
