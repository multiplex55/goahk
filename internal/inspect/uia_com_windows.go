//go:build windows
// +build windows

package inspect

import (
	"errors"
	"fmt"
	"strings"
	"unsafe"

	"goahk/internal/window"
)

type uiaBridgeElement struct {
	Element             *uiaElement
	Key                 string
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
}

type win32UIAComBridge struct {
	client uiaAutomationClient
}

func newWin32UIABridge() nativeUIABridge {
	return &win32UIAComBridge{client: newUnavailableUIAClient()}
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
func (b *win32UIAComBridge) Invoke(*uiaBridgeElement) error    { return ErrProviderActionUnsupported }
func (b *win32UIAComBridge) Select(*uiaBridgeElement) error    { return ErrProviderActionUnsupported }
func (b *win32UIAComBridge) SetValue(*uiaBridgeElement, string) error {
	return ErrProviderActionUnsupported
}
func (b *win32UIAComBridge) DoDefaultAction(*uiaBridgeElement) error {
	return ErrProviderActionUnsupported
}
func (b *win32UIAComBridge) Toggle(*uiaBridgeElement) error   { return ErrProviderActionUnsupported }
func (b *win32UIAComBridge) Expand(*uiaBridgeElement) error   { return ErrProviderActionUnsupported }
func (b *win32UIAComBridge) Collapse(*uiaBridgeElement) error { return ErrProviderActionUnsupported }

type unavailableUIAClient struct{}

func newUnavailableUIAClient() uiaAutomationClient { return unavailableUIAClient{} }

func (unavailableUIAClient) ElementFromHWND(window.HWND) (*uiaBridgeElement, error) {
	return nil, &UIAComUnavailableError{Op: "ElementFromHandle", Err: errors.New("UI Automation COM bridge is not initialized")}
}
func (unavailableUIAClient) FocusedElement() (*uiaBridgeElement, error) {
	return nil, &UIAComUnavailableError{Op: "GetFocusedElement", Err: errors.New("UI Automation COM bridge is not initialized")}
}
func (unavailableUIAClient) ElementFromPoint(int, int) (*uiaBridgeElement, error) {
	return nil, &UIAComUnavailableError{Op: "ElementFromPoint", Err: errors.New("UI Automation COM bridge is not initialized")}
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

func currentCursorPos() (int, int, error) {
	pt := winPoint{}
	ok, _, err := procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
	if ok == 0 {
		return 0, 0, fmt.Errorf("GetCursorPos: %w", err)
	}
	return int(pt.X), int(pt.Y), nil
}
