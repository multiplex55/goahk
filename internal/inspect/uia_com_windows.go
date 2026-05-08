//go:build windows
// +build windows

package inspect

import (
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"unsafe"

	"goahk/internal/window"
)

type uiaBridgeElement struct {
	Element             *uiaElement
	Key                 string
	NativePtr           uintptr
	AllowHWNDFallback   bool
	SupportedPatterns   []string
	UnsupportedProperty map[string]bool
	PropertyState       map[string]string
}

type nativeUIABridge interface {
	ResolveRoot(window.HWND) (*uiaBridgeElement, error)
	FocusedElement() (*uiaBridgeElement, error)
	ElementFromPoint(x, y int) (*uiaBridgeElement, error)
	ElementByKey(key string) (*uiaBridgeElement, error)
	Parent(*uiaBridgeElement) (*uiaBridgeElement, error)
	Children(*uiaBridgeElement) ([]*uiaBridgeElement, error)
	CursorPosition() (int, int, error)
	Invoke(*uiaBridgeElement) error
	Select(*uiaBridgeElement) error
	SetValue(*uiaBridgeElement, string) error
	DoDefaultAction(*uiaBridgeElement) error
	Toggle(*uiaBridgeElement) error
	Expand(*uiaBridgeElement) error
	Collapse(*uiaBridgeElement) error
}

type uiaAutomationClient interface {
	ElementFromHWND(window.HWND) (*uiaBridgeElement, error)
	FocusedElement() (*uiaBridgeElement, error)
	ElementFromPoint(x, y int) (*uiaBridgeElement, error)
	ElementByRuntimeID(runtimeID string) (*uiaBridgeElement, error)
	Parent(*uiaBridgeElement) (*uiaBridgeElement, error)
	Children(*uiaBridgeElement) ([]*uiaBridgeElement, error)
	Invoke(*uiaBridgeElement) error
	Select(*uiaBridgeElement) error
	SetValue(*uiaBridgeElement, string) error
	DoDefaultAction(*uiaBridgeElement) error
	Toggle(*uiaBridgeElement) error
	Expand(*uiaBridgeElement) error
	Collapse(*uiaBridgeElement) error
}

type win32UIAComBridge struct {
	client  uiaAutomationClient
	initErr error
}

var (
	newUIAComClient   = newNativeUIAComClient
	uiaNativeCOMReady atomic.Bool
)

func newWin32UIABridge() nativeUIABridge {
	client, err := newUIAComClient()
	if err != nil {
		uiaNativeCOMReady.Store(false)
		return &win32UIAComBridge{client: newUnavailableUIAClient(err), initErr: err}
	}
	uiaNativeCOMReady.Store(true)
	return &win32UIAComBridge{client: client}
}

func (b *win32UIAComBridge) ResolveRoot(hwnd window.HWND) (*uiaBridgeElement, error) {
	return b.client.ElementFromHWND(hwnd)
}

func (b *win32UIAComBridge) FocusedElement() (*uiaBridgeElement, error) {
	return b.client.FocusedElement()
}

func (b *win32UIAComBridge) ElementFromPoint(x, y int) (*uiaBridgeElement, error) {
	return b.client.ElementFromPoint(x, y)
}

func (b *win32UIAComBridge) ElementByKey(key string) (*uiaBridgeElement, error) {
	if strings.TrimSpace(key) == "" {
		return nil, errUIANilElement
	}
	return b.client.ElementByRuntimeID(key)
}

func (b *win32UIAComBridge) Parent(el *uiaBridgeElement) (*uiaBridgeElement, error) {
	if el == nil {
		return nil, errUIANilElement
	}
	return b.client.Parent(el)
}

func (b *win32UIAComBridge) Children(el *uiaBridgeElement) ([]*uiaBridgeElement, error) {
	if el == nil {
		return nil, errUIANilElement
	}
	return b.client.Children(el)
}

func (b *win32UIAComBridge) CursorPosition() (int, int, error) { return currentCursorPos() }
func (b *win32UIAComBridge) Invoke(el *uiaBridgeElement) error { return b.client.Invoke(el) }
func (b *win32UIAComBridge) Select(el *uiaBridgeElement) error { return b.client.Select(el) }
func (b *win32UIAComBridge) SetValue(el *uiaBridgeElement, value string) error {
	return b.client.SetValue(el, value)
}
func (b *win32UIAComBridge) DoDefaultAction(el *uiaBridgeElement) error {
	return b.client.DoDefaultAction(el)
}
func (b *win32UIAComBridge) Toggle(el *uiaBridgeElement) error   { return b.client.Toggle(el) }
func (b *win32UIAComBridge) Expand(el *uiaBridgeElement) error   { return b.client.Expand(el) }
func (b *win32UIAComBridge) Collapse(el *uiaBridgeElement) error { return b.client.Collapse(el) }

func newUnavailableUIAClient(initErr error) uiaAutomationClient {
	return unavailableUIAClient{initErr: initErr}
}

type unavailableUIAClient struct{ initErr error }

func (c unavailableUIAClient) wrap(op string) error {
	err := c.initErr
	if err == nil {
		err = errors.New("UI Automation COM bridge is not initialized")
	}
	return &UIAComUnavailableError{Op: op, Err: err}
}

func (c unavailableUIAClient) ElementFromHWND(window.HWND) (*uiaBridgeElement, error) {
	return nil, c.wrap("ElementFromHandle")
}
func (c unavailableUIAClient) FocusedElement() (*uiaBridgeElement, error) {
	return nil, c.wrap("GetFocusedElement")
}
func (c unavailableUIAClient) ElementFromPoint(int, int) (*uiaBridgeElement, error) {
	return nil, c.wrap("ElementFromPoint")
}
func (unavailableUIAClient) ElementByRuntimeID(string) (*uiaBridgeElement, error) {
	return nil, &UIAElementStaleError{Op: "ElementByRuntimeID", Err: errors.New("runtime id is stale or unavailable")}
}
func (unavailableUIAClient) Parent(*uiaBridgeElement) (*uiaBridgeElement, error) {
	return nil, &UIAElementStaleError{Op: "GetParentElement", Err: errors.New("element is stale")}
}
func (unavailableUIAClient) Children(*uiaBridgeElement) ([]*uiaBridgeElement, error) {
	return nil, &UIAElementStaleError{Op: "GetChildren", Err: errors.New("element is stale")}
}
func (unavailableUIAClient) Invoke(*uiaBridgeElement) error { return ErrProviderActionUnsupported }
func (unavailableUIAClient) Select(*uiaBridgeElement) error { return ErrProviderActionUnsupported }
func (unavailableUIAClient) SetValue(*uiaBridgeElement, string) error {
	return ErrProviderActionUnsupported
}
func (unavailableUIAClient) DoDefaultAction(*uiaBridgeElement) error {
	return ErrProviderActionUnsupported
}
func (unavailableUIAClient) Toggle(*uiaBridgeElement) error   { return ErrProviderActionUnsupported }
func (unavailableUIAClient) Expand(*uiaBridgeElement) error   { return ErrProviderActionUnsupported }
func (unavailableUIAClient) Collapse(*uiaBridgeElement) error { return ErrProviderActionUnsupported }

func currentCursorPos() (int, int, error) {
	pt := winPoint{}
	ok, _, err := procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
	if ok == 0 {
		return 0, 0, fmt.Errorf("GetCursorPos: %w", err)
	}
	return int(pt.X), int(pt.Y), nil
}
