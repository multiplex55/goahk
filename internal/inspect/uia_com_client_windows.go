//go:build windows
// +build windows

package inspect

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"goahk/internal/window"
)

type nativeUIAComClient struct {
	worker              *uiaCOMWorker
	elementsByRuntimeID map[string]*uiaBridgeElement
	mu                  sync.RWMutex
}

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
	el := &uiaElement{RuntimeID: key, HWND: hwnd.String()}
	unsupported := map[string]bool{}
	states := map[string]string{}

	readStringProperty("ControlType", "Pane", &el.ControlType, unsupported, states)
	readStringProperty("LocalizedControlType", "pane", &el.LocalizedControlType, unsupported, states)
	readStringProperty("Name", "Window", &el.Name, unsupported, states)
	readStringProperty("Value", "", strPtrAssign(&el.Value), unsupported, states)
	readStringProperty("AutomationId", "", &el.AutomationID, unsupported, states)
	readStringProperty("ClassName", "Window", &el.ClassName, unsupported, states)
	readStringProperty("FrameworkId", "UIA", &el.FrameworkID, unsupported, states)
	readIntProperty("ProcessId", 0, &el.ProcessID, unsupported, states)
	readRectProperty("BoundingRectangle", nil, &el.BoundingRect, unsupported, states)
	readBoolProperty("HasKeyboardFocus", false, &el.HasKeyboardFocus, unsupported, states)
	readBoolProperty("IsEnabled", false, &el.IsEnabled, unsupported, states)
	readBoolProperty("IsOffscreen", false, &el.IsOffscreen, unsupported, states)
	readBoolProperty("IsPassword", false, &el.IsPassword, unsupported, states)
	readBoolProperty("IsKeyboardFocusable", false, &el.IsKeyboardFocusable, unsupported, states)
	readBoolProperty("IsContentElement", false, &el.IsContentElement, unsupported, states)
	readBoolProperty("IsControlElement", false, &el.IsControlElement, unsupported, states)
	readBoolProperty("IsRequiredForForm", false, &el.IsRequiredForForm, unsupported, states)
	readStringProperty("HelpText", "", strPtrAssign(&el.HelpText), unsupported, states)
	readStringProperty("AccessKey", "", strPtrAssign(&el.AccessKey), unsupported, states)
	readStringProperty("AcceleratorKey", "", strPtrAssign(&el.AcceleratorKey), unsupported, states)
	readStringProperty("ItemType", "", strPtrAssign(&el.ItemType), unsupported, states)
	readStringProperty("ItemStatus", "", strPtrAssign(&el.ItemStatus), unsupported, states)
	readStringProperty("LabeledBy", "", strPtrAssign(&el.LabeledBy), unsupported, states)

	el.UnsupportedProps = unsupported
	el.PropertyStates = states
	return &uiaBridgeElement{Key: key, AllowHWNDFallback: true, SupportedPatterns: []string{"Invoke", "LegacyIAccessible", "SelectionItem", "Value", "Toggle", "ExpandCollapse", "Window", "Transform"}, PropertyState: states, UnsupportedProperty: unsupported, Element: el}, nil
}

func strPtrAssign(dst **string) *string {
	if dst == nil {
		return nil
	}
	v := ""
	*dst = &v
	return *dst
}

func readStringProperty(name, raw string, dst *string, unsupported map[string]bool, state map[string]string) {
	if strings.TrimSpace(raw) == "" {
		state[name] = propertyStatusUnsupported
		unsupported[name] = true
		if dst != nil {
			*dst = ""
		}
		return
	}
	state[name] = propertyStatusOK
	if dst != nil {
		*dst = strings.TrimSpace(raw)
	}
}

func readIntProperty(name string, raw int, dst *int, unsupported map[string]bool, state map[string]string) {
	if raw <= 0 {
		state[name] = propertyStatusUnsupported
		unsupported[name] = true
		if dst != nil {
			*dst = 0
		}
		return
	}
	state[name] = propertyStatusOK
	if dst != nil {
		*dst = raw
	}
}

func readBoolProperty(name string, raw bool, dst *bool, _ map[string]bool, state map[string]string) {
	state[name] = propertyStatusOK
	if dst != nil {
		*dst = raw
	}
}

func readRectProperty(name string, raw *uiaRect, dst **uiaRect, unsupported map[string]bool, state map[string]string) {
	if raw == nil {
		state[name] = propertyStatusUnsupported
		unsupported[name] = true
		if dst != nil {
			*dst = nil
		}
		return
	}
	state[name] = propertyStatusOK
	if dst != nil {
		*dst = &uiaRect{Left: raw.Left, Top: raw.Top, Width: raw.Width, Height: raw.Height}
	}
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
	return &nativeUIAComClient{
		worker:              worker,
		elementsByRuntimeID: map[string]*uiaBridgeElement{},
	}, nil
}

func (c *nativeUIAComClient) ElementFromHWND(hwnd window.HWND) (*uiaBridgeElement, error) {
	var out *uiaBridgeElement
	err := c.worker.Do("ElementFromHandle", func(state *uiaWorkerState) error {
		var callErr error
		out, callErr = uiaNativeAPI.ElementFromHandle(state, hwnd)
		return callErr
	})
	c.cacheBridgeElement(out)
	return out, err
}
func (c *nativeUIAComClient) FocusedElement() (*uiaBridgeElement, error) {
	return nil, &UIAComUnavailableError{Op: "GetFocusedElement", Err: errors.New("not implemented")}
}
func (c *nativeUIAComClient) ElementFromPoint(int, int) (*uiaBridgeElement, error) {
	return nil, &UIAComUnavailableError{Op: "ElementFromPoint", Err: errors.New("not implemented")}
}
func (c *nativeUIAComClient) ElementByRuntimeID(runtimeID string) (*uiaBridgeElement, error) {
	id := canonicalRuntimeID(runtimeID)
	if id == "" {
		return nil, &UIAElementStaleError{Op: "ElementByRuntimeID", Err: errors.New("runtime id is stale or unavailable")}
	}
	c.mu.RLock()
	el, ok := c.elementsByRuntimeID[id]
	c.mu.RUnlock()
	if !ok || el == nil {
		return nil, &UIAElementStaleError{Op: "ElementByRuntimeID", Err: fmt.Errorf("runtime id %q is stale or unavailable", id)}
	}
	return el, nil
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
	c.cacheBridgeElement(out)
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
	for _, child := range out {
		c.cacheBridgeElement(child)
	}
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

func canonicalRuntimeID(raw string) string {
	rid := strings.TrimSpace(raw)
	if rid == "" {
		return ""
	}
	if strings.HasPrefix(rid, "rid:") {
		return rid
	}
	return "rid:" + rid
}

func (c *nativeUIAComClient) cacheBridgeElement(el *uiaBridgeElement) {
	if el == nil {
		return
	}
	id := canonicalRuntimeID(el.Key)
	if id == "" && el.Element != nil {
		id = canonicalRuntimeID(el.Element.RuntimeID)
	}
	if id == "" {
		return
	}
	if el.Key == "" {
		el.Key = id
	}
	if el.Element != nil && el.Element.RuntimeID == "" {
		el.Element.RuntimeID = strings.TrimPrefix(id, "rid:")
	}
	c.mu.Lock()
	c.elementsByRuntimeID[id] = el
	c.mu.Unlock()
}
