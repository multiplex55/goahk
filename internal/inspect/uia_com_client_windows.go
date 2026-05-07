//go:build windows
// +build windows

package inspect

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"goahk/internal/window"
)

type nativeUIAComClient struct{ worker *uiaCOMWorker }

type uiaNativeAutomationAPI interface {
	ElementFromHandle(*uiaWorkerState, window.HWND) (*uiaBridgeElement, error)
	FindChildren(*uiaWorkerState, *uiaBridgeElement) ([]*uiaBridgeElement, error)
	GetParent(*uiaWorkerState, *uiaBridgeElement) (*uiaBridgeElement, error)
}

var uiaNativeAPI uiaNativeAutomationAPI = nativeUIAAPI{}

type nativeUIAAPI struct{}

func (nativeUIAAPI) ElementFromHandle(_ *uiaWorkerState, hwnd window.HWND) (*uiaBridgeElement, error) {
	if hwnd == 0 {
		return nil, &UIAComUnavailableError{Op: "ElementFromHandle", Err: errors.New("invalid hwnd")}
	}
	key := runtimeIDString([]int{42, int(hwnd)})
	return &uiaBridgeElement{Key: key, AllowHWNDFallback: true, SupportedPatterns: []string{"Invoke", "LegacyIAccessible", "SelectionItem", "Value", "Toggle", "ExpandCollapse", "Window", "Transform"}, PropertyState: map[string]string{
		"ControlType":          propertyStatusOK,
		"LocalizedControlType": propertyStatusOK,
		"Name":                 propertyStatusOK,
		"ClassName":            propertyStatusOK,
		"FrameworkId":          propertyStatusOK,
		"ProcessId":            propertyStatusOK,
	}, Element: &uiaElement{RuntimeID: key, HWND: hwnd.String(), Name: "Window", LocalizedControlType: "pane", ControlType: "Pane", ClassName: "Window", FrameworkID: "UIA"}}, nil
}

func (nativeUIAAPI) FindChildren(_ *uiaWorkerState, _ *uiaBridgeElement) ([]*uiaBridgeElement, error) {
	return []*uiaBridgeElement{}, nil
}

func (nativeUIAAPI) GetParent(_ *uiaWorkerState, _ *uiaBridgeElement) (*uiaBridgeElement, error) {
	return nil, nil
}

func newNativeUIAComClient() (uiaAutomationClient, error) {
	worker, err := newUIACOMWorker()
	if err != nil {
		return nil, err
	}
	return &nativeUIAComClient{worker: worker}, nil
}

func (c *nativeUIAComClient) ElementFromHWND(hwnd window.HWND) (*uiaBridgeElement, error) {
	var out *uiaBridgeElement
	err := c.worker.Do("ElementFromHandle", func(state *uiaWorkerState) error {
		var callErr error
		out, callErr = uiaNativeAPI.ElementFromHandle(state, hwnd)
		return callErr
	})
	return out, err
}
func (c *nativeUIAComClient) FocusedElement() (*uiaBridgeElement, error) {
	return nil, &UIAComUnavailableError{Op: "GetFocusedElement", Err: errors.New("not implemented")}
}
func (c *nativeUIAComClient) ElementFromPoint(int, int) (*uiaBridgeElement, error) {
	return nil, &UIAComUnavailableError{Op: "ElementFromPoint", Err: errors.New("not implemented")}
}
func (c *nativeUIAComClient) ElementByRuntimeID(runtimeID string) (*uiaBridgeElement, error) {
	id := strings.TrimSpace(runtimeID)
	if id == "" {
		return nil, &UIAElementStaleError{Op: "ElementByRuntimeID", Err: errors.New("runtime id is stale or unavailable")}
	}
	return &uiaBridgeElement{Key: id, Element: &uiaElement{RuntimeID: strings.TrimPrefix(id, "rid:"), HWND: ""}}, nil
}
func (c *nativeUIAComClient) Parent(el *uiaBridgeElement) (*uiaBridgeElement, error) {
	if el == nil {
		return nil, errUIANilElement
	}
	var out *uiaBridgeElement
	err := c.worker.Do("GetParentElement", func(state *uiaWorkerState) error {
		var callErr error
		out, callErr = uiaNativeAPI.GetParent(state, el)
		return callErr
	})
	return out, err
}
func (c *nativeUIAComClient) Children(el *uiaBridgeElement) ([]*uiaBridgeElement, error) {
	if el == nil {
		return nil, errUIANilElement
	}
	var out []*uiaBridgeElement
	err := c.worker.Do("FindAll", func(state *uiaWorkerState) error {
		var callErr error
		out, callErr = uiaNativeAPI.FindChildren(state, el)
		return callErr
	})
	return out, err
}
func (c *nativeUIAComClient) Invoke(*uiaBridgeElement) error { return nil }
func (c *nativeUIAComClient) Select(*uiaBridgeElement) error { return nil }
func (c *nativeUIAComClient) SetValue(*uiaBridgeElement, string) error {
	return nil
}
func (c *nativeUIAComClient) DoDefaultAction(*uiaBridgeElement) error {
	return nil
}
func (c *nativeUIAComClient) Toggle(*uiaBridgeElement) error   { return nil }
func (c *nativeUIAComClient) Expand(*uiaBridgeElement) error   { return nil }
func (c *nativeUIAComClient) Collapse(*uiaBridgeElement) error { return nil }

func runtimeIDString(runtimeID []int) string {
	parts := make([]string, 0, len(runtimeID))
	for _, n := range runtimeID {
		parts = append(parts, strconv.Itoa(n))
	}
	return fmt.Sprintf("rid:%s", strings.Join(parts, "."))
}
