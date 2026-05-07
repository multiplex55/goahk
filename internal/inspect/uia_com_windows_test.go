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
func (fakeUIAClient) ElementByRuntimeID(string) (*uiaBridgeElement, error)    { return nil, nil }
func (fakeUIAClient) Parent(*uiaBridgeElement) (*uiaBridgeElement, error)     { return nil, nil }
func (fakeUIAClient) Children(*uiaBridgeElement) ([]*uiaBridgeElement, error) { return nil, nil }

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
	findChildren      func(*uiaWorkerState, *uiaBridgeElement) ([]*uiaBridgeElement, error)
	getParent         func(*uiaWorkerState, *uiaBridgeElement) (*uiaBridgeElement, error)
}

func (f fakeUiaNativeAPI) ElementFromHandle(s *uiaWorkerState, h window.HWND) (*uiaBridgeElement, error) {
	return f.elementFromHandle(s, h)
}
func (f fakeUiaNativeAPI) FindChildren(s *uiaWorkerState, el *uiaBridgeElement) ([]*uiaBridgeElement, error) {
	return f.findChildren(s, el)
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
		getParent: func(_ *uiaWorkerState, _ *uiaBridgeElement) (*uiaBridgeElement, error) { return nil, nil },
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
