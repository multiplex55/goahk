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
	FocusedElement(*uiaWorkerState) (*uiaBridgeElement, error)
	ElementFromPoint(*uiaWorkerState, int, int) (*uiaBridgeElement, error)
	FindChildren(*uiaWorkerState, *uiaBridgeElement) ([]*uiaBridgeElement, error)
	GetParent(*uiaWorkerState, *uiaBridgeElement) (*uiaBridgeElement, error)
}

var uiaNativeAPI uiaNativeAutomationAPI = nativeUIAAPI{}

type nativeUIAAPI struct{}

func wrapNativeElement(ptr uintptr, hwnd window.HWND) (*uiaBridgeElement, error) {
	if ptr == 0 {
		return nil, &UIAComUnavailableError{Op: "WrapElement", Err: errors.New("nil COM element")}
	}
	key := fmt.Sprintf("ptr:%x", ptr)
	el := &uiaElement{RuntimeID: key, HWND: hwnd.String()}
	return &uiaBridgeElement{Key: key, AllowHWNDFallback: hwnd != 0, SupportedPatterns: detectSupportedPatterns(el), PropertyState: map[string]string{}, UnsupportedProperty: map[string]bool{}, NativePtr: ptr, Element: el}, nil
}

func (nativeUIAAPI) ElementFromHandle(state *uiaWorkerState, hwnd window.HWND) (*uiaBridgeElement, error) {
	if hwnd == 0 {
		return nil, &UIAComUnavailableError{Op: "ElementFromHandle", Err: errors.New("invalid hwnd")}
	}
	ptr, err := uiaElementFromHandle(state.automation, hwnd)
	if err != nil {
		return nil, err
	}
	return wrapNativeElement(ptr, hwnd)
}

func (nativeUIAAPI) FocusedElement(state *uiaWorkerState) (*uiaBridgeElement, error) {
	ptr, err := uiaGetFocusedElement(state.automation)
	if err != nil {
		return nil, err
	}
	return wrapNativeElement(ptr, 0)
}

func (nativeUIAAPI) ElementFromPoint(state *uiaWorkerState, x, y int) (*uiaBridgeElement, error) {
	ptr, err := uiaElementFromPoint(state.automation, x, y)
	if err != nil {
		return nil, err
	}
	return wrapNativeElement(ptr, 0)
}

func detectSupportedPatterns(el *uiaElement) []string {
	if el == nil {
		return nil
	}
	patterns := []string{"Invoke", "LegacyIAccessible", "SelectionItem", "Toggle", "ExpandCollapse", "Window", "Transform"}
	if !el.IsPassword {
		patterns = append(patterns, "Value")
	}
	if strings.EqualFold(strings.TrimSpace(el.ControlType), "Window") {
		patterns = append(patterns, "Window")
	}
	seen := map[string]struct{}{}
	out := make([]string, 0, len(patterns))
	for _, p := range patterns {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	return out
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

func (nativeUIAAPI) FindChildren(state *uiaWorkerState, parent *uiaBridgeElement) ([]*uiaBridgeElement, error) {
	if parent == nil || parent.NativePtr == 0 {
		return nil, &UIAElementStaleError{Op: "FindAll", Err: errors.New("parent element is stale")}
	}
	arr, err := uiaFindAllChildren(parent.NativePtr, state.trueCond)
	if err != nil {
		return nil, err
	}
	defer comRelease(arr)
	n, err := uiaArrayLength(arr)
	if err != nil {
		return nil, err
	}
	out := make([]*uiaBridgeElement, 0, n)
	for i := int32(0); i < n; i++ {
		ptr, getErr := uiaArrayGet(arr, i)
		if getErr != nil {
			return nil, getErr
		}
		child, wrapErr := wrapNativeElement(ptr, 0)
		if wrapErr != nil {
			return nil, wrapErr
		}
		child.Element.ParentRef = parent.Key
		out = append(out, child)
	}
	return out, nil
}

func (nativeUIAAPI) GetParent(state *uiaWorkerState, el *uiaBridgeElement) (*uiaBridgeElement, error) {
	if el == nil || el.NativePtr == 0 {
		return nil, &UIAElementStaleError{Op: "GetParentElement", Err: errors.New("element is stale")}
	}
	parent, err := uiaGetParentElement(state.treeWalker, el.NativePtr)
	if err != nil || parent == 0 {
		return nil, err
	}
	return wrapNativeElement(parent, 0)
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
	var out *uiaBridgeElement
	err := c.worker.Do("GetFocusedElement", func(state *uiaWorkerState) error { var e error; out, e = uiaNativeAPI.FocusedElement(state); return e })
	c.cacheBridgeElement(out)
	return out, err
}
func (c *nativeUIAComClient) ElementFromPoint(x, y int) (*uiaBridgeElement, error) {
	var out *uiaBridgeElement
	err := c.worker.Do("ElementFromPoint", func(state *uiaWorkerState) error {
		var e error
		out, e = uiaNativeAPI.ElementFromPoint(state, x, y)
		return e
	})
	c.cacheBridgeElement(out)
	return out, err
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
func (c *nativeUIAComClient) Invoke(el *uiaBridgeElement) error {
	return c.requireActionable(el, "Invoke", "Invoke")
}
func (c *nativeUIAComClient) Select(el *uiaBridgeElement) error {
	return c.requireActionable(el, "Select", "SelectionItem")
}
func (c *nativeUIAComClient) SetValue(el *uiaBridgeElement, _ string) error {
	return c.requireActionable(el, "SetValue", "Value")
}
func (c *nativeUIAComClient) DoDefaultAction(el *uiaBridgeElement) error {
	return c.requireActionable(el, "DoDefaultAction", "LegacyIAccessible")
}
func (c *nativeUIAComClient) Toggle(el *uiaBridgeElement) error {
	return c.requireActionable(el, "Toggle", "Toggle")
}
func (c *nativeUIAComClient) Expand(el *uiaBridgeElement) error {
	return c.requireActionable(el, "Expand", "ExpandCollapse")
}
func (c *nativeUIAComClient) Collapse(el *uiaBridgeElement) error {
	return c.requireActionable(el, "Collapse", "ExpandCollapse")
}

func (c *nativeUIAComClient) requireActionable(el *uiaBridgeElement, op, pattern string) error {
	if el == nil {
		return errUIANilElement
	}
	for _, supported := range el.SupportedPatterns {
		if strings.EqualFold(strings.TrimSpace(supported), pattern) {
			return nil
		}
	}
	return &UIAComUnavailableError{Op: op, Err: fmt.Errorf("pattern %s is not supported", pattern)}
}

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
