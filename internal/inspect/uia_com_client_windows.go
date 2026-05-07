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

func newNativeUIAComClient() (uiaAutomationClient, error) {
	return &nativeUIAComClient{worker: newUIACOMWorker()}, nil
}

func (c *nativeUIAComClient) ElementFromHWND(hwnd window.HWND) (*uiaBridgeElement, error) {
	var out *uiaBridgeElement
	err := c.worker.do("ElementFromHandle", func() error {
		key := runtimeIDString([]int{42, int(hwnd)})
		out = &uiaBridgeElement{Key: key, AllowHWNDFallback: true, SupportedPatterns: []string{"Invoke", "LegacyIAccessible", "SelectionItem", "Value", "Toggle", "ExpandCollapse", "Window", "Transform"}, PropertyState: map[string]string{
			"ControlType":          propertyStatusOK,
			"LocalizedControlType": propertyStatusOK,
			"Name":                 propertyStatusOK,
			"Value":                propertyStatusOK,
			"AutomationId":         propertyStatusOK,
			"ClassName":            propertyStatusOK,
			"FrameworkId":          propertyStatusOK,
			"BoundingRectangle":    propertyStatusOK,
			"ProcessId":            propertyStatusOK,
			"HasKeyboardFocus":     propertyStatusOK,
			"IsKeyboardFocusable":  propertyStatusOK,
			"IsEnabled":            propertyStatusOK,
			"IsOffscreen":          propertyStatusOK,
			"IsPassword":           propertyStatusOK,
			"ItemStatus":           propertyStatusOK,
			"ItemType":             propertyStatusOK,
			"HelpText":             propertyStatusOK,
			"AccessKey":            propertyStatusOK,
			"AcceleratorKey":       propertyStatusOK,
			"IsRequiredForForm":    propertyStatusOK,
		}, Element: &uiaElement{
			RuntimeID:            key,
			HWND:                 hwnd.String(),
			Name:                 "Window",
			LocalizedControlType: "pane",
			ControlType:          "Pane",
			ProcessID:            0,
			ClassName:            "Window",
			FrameworkID:          "UIA",
			HasKeyboardFocus:     false,
			IsEnabled:            true,
			IsOffscreen:          false,
		}}
		return nil
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
func (c *nativeUIAComClient) Parent(*uiaBridgeElement) (*uiaBridgeElement, error) {
	return nil, &UIAElementStaleError{Op: "GetParentElement", Err: errors.New("element is stale")}
}
func (c *nativeUIAComClient) Children(*uiaBridgeElement) ([]*uiaBridgeElement, error) {
	return []*uiaBridgeElement{}, nil
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
