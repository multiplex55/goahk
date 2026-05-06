//go:build windows
// +build windows

package inspect

import (
	"errors"

	"goahk/internal/window"
)

type nativeUIAComClient struct{}

func newNativeUIAComClient() (uiaAutomationClient, error) {
	return nil, errors.New("UI Automation COM client initialization is not yet available in this build")
}

func (c *nativeUIAComClient) ElementFromHWND(window.HWND) (*uiaBridgeElement, error) {
	return nil, &UIAComUnavailableError{Op: "ElementFromHandle", Err: errors.New("UI Automation COM client is unavailable")}
}
func (c *nativeUIAComClient) FocusedElement() (*uiaBridgeElement, error) {
	return nil, &UIAComUnavailableError{Op: "GetFocusedElement", Err: errors.New("UI Automation COM client is unavailable")}
}
func (c *nativeUIAComClient) ElementFromPoint(int, int) (*uiaBridgeElement, error) {
	return nil, &UIAComUnavailableError{Op: "ElementFromPoint", Err: errors.New("UI Automation COM client is unavailable")}
}
func (c *nativeUIAComClient) ElementByRuntimeID(string) (*uiaBridgeElement, error) {
	return nil, &UIAElementStaleError{Op: "ElementByRuntimeID", Err: errors.New("runtime id is stale or unavailable")}
}
func (c *nativeUIAComClient) Parent(*uiaBridgeElement) (*uiaBridgeElement, error) {
	return nil, &UIAElementStaleError{Op: "GetParentElement", Err: errors.New("element is stale")}
}
func (c *nativeUIAComClient) Children(*uiaBridgeElement) ([]*uiaBridgeElement, error) {
	return nil, &UIAElementStaleError{Op: "GetChildren", Err: errors.New("element is stale")}
}
