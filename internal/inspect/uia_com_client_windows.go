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
	elementsByFallback  map[string]*uiaBridgeElement
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
	rid, err := uiaElementRuntimeID(ptr)
	if err != nil {
		return nil, err
	}
	key := canonicalRuntimeID(rid)
	if key == "" {
		key = fmt.Sprintf("ptr:%x", ptr)
	}
	comAddRef(ptr)
	el := &uiaElement{RuntimeID: strings.TrimPrefix(key, "rid:"), HWND: hwnd.String()}
	b := &uiaBridgeElement{Key: key, RuntimeID: key, AllowHWNDFallback: hwnd != 0, SupportedPatterns: nil, PropertyState: map[string]string{}, UnsupportedProperty: map[string]bool{}, NativePtr: ptr, Element: el}
	populateElementProperties(b)
	populateSupportedPatterns(b)
	return b, nil
}

func populateElementProperties(el *uiaBridgeElement) {
	if el == nil || el.NativePtr == 0 || el.Element == nil {
		return
	}
	setStr := func(name string, prop int32, dst *string) {
		v, err := uiaGetCurrentPropertyValue(el.NativePtr, prop)
		if err != nil {
			markPropertyErr(el, name, err)
			return
		}
		r := decodeVariant(v)
		el.PropertyState[name] = r.Status
		if r.Status == propertyStatusUnsupported {
			el.UnsupportedProperty[name] = true
			return
		}
		*dst = strings.TrimSpace(r.S)
	}
	setBool := func(name string, prop int32, dst *bool) {
		v, err := uiaGetCurrentPropertyValue(el.NativePtr, prop)
		if err != nil {
			markPropertyErr(el, name, err)
			return
		}
		r := decodeVariant(v)
		el.PropertyState[name] = r.Status
		*dst = r.B
	}
	setInt := func(name string, prop int32, dst *int) {
		v, err := uiaGetCurrentPropertyValue(el.NativePtr, prop)
		if err != nil {
			markPropertyErr(el, name, err)
			return
		}
		r := decodeVariant(v)
		el.PropertyState[name] = r.Status
		*dst = r.I
	}
	setStr("ControlType", uiaPropertyControlType, &el.Element.ControlType)
	setStr("LocalizedControlType", uiaPropertyLocalizedCtl, &el.Element.LocalizedControlType)
	setStr("Name", uiaPropertyName, &el.Element.Name)
	setStr("Value", 30045, strPtrAssign(&el.Element.Value))
	setStr("AutomationId", uiaPropertyAutomationID, &el.Element.AutomationID)
	setStr("ClassName", uiaPropertyClassName, &el.Element.ClassName)
	setStr("HelpText", uiaPropertyHelpText, strPtrAssign(&el.Element.HelpText))
	setStr("AccessKey", uiaPropertyAccessKey, strPtrAssign(&el.Element.AccessKey))
	setStr("AcceleratorKey", uiaPropertyAccelerator, strPtrAssign(&el.Element.AcceleratorKey))
	setBool("HasKeyboardFocus", uiaPropertyHasFocus, &el.Element.HasKeyboardFocus)
	setBool("IsKeyboardFocusable", uiaPropertyIsFocusable, &el.Element.IsKeyboardFocusable)
	setStr("ItemType", uiaPropertyItemType, strPtrAssign(&el.Element.ItemType))
	setInt("ProcessId", uiaPropertyProcessID, &el.Element.ProcessID)
	setBool("IsEnabled", uiaPropertyIsEnabled, &el.Element.IsEnabled)
	setBool("IsPassword", uiaPropertyIsPassword, &el.Element.IsPassword)
	setBool("IsOffscreen", uiaPropertyIsOffscreen, &el.Element.IsOffscreen)
	setStr("FrameworkId", uiaPropertyFrameworkID, &el.Element.FrameworkID)
	setBool("IsRequiredForForm", uiaPropertyIsRequired, &el.Element.IsRequiredForForm)
	setStr("ItemStatus", uiaPropertyItemStatus, strPtrAssign(&el.Element.ItemStatus))
	setStr("LabeledBy", uiaPropertyLabeledBy, strPtrAssign(&el.Element.LabeledBy))
}

func markPropertyErr(el *uiaBridgeElement, name string, err error) {
	var stale *UIAElementStaleError
	if errors.As(err, &stale) {
		el.PropertyState[name] = propertyStatusStale
	} else {
		el.PropertyState[name] = propertyStatusUnavailable
	}
}

func populateSupportedPatterns(el *uiaBridgeElement) {
	defs := []struct {
		name string
		id   int32
	}{{"Invoke", 10000}, {"SelectionItem", 10010}, {"Value", 10002}, {"Toggle", 10015}, {"ExpandCollapse", 10005}, {"Window", 10009}, {"Transform", 10016}, {"Text", 10014}, {"Selection", 10001}, {"Scroll", 10004}, {"RangeValue", 10003}, {"Grid", 10006}, {"Table", 10012}, {"LegacyIAccessible", 10018}}
	for _, d := range defs {
		ok, err := uiaGetCurrentPattern(el.NativePtr, d.id)
		if err != nil {
			continue
		}
		if ok {
			el.SupportedPatterns = append(el.SupportedPatterns, d.name)
		}
	}
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
		child.ParentKey = parent.Key
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
		elementsByFallback:  map[string]*uiaBridgeElement{},
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
		if fallback, ok := c.elementsByFallback[id]; ok && fallback != nil {
			return cloneBridgeElement(fallback), nil
		}
		return nil, &UIAElementStaleError{Op: "ElementByRuntimeID", Err: fmt.Errorf("runtime id %q is stale or unavailable", id)}
	}
	return cloneBridgeElement(el), nil
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
	if id == "" && el.ParentKey != "" && el.Element != nil {
		id = canonicalRuntimeID(el.ParentKey) + "/idx:" + strings.TrimSpace(el.Element.Name)
	}
	if id == "" {
		id = fmt.Sprintf("best:%x", el.NativePtr)
	}
	if el.Key == "" {
		el.Key = id
	}
	if el.Element != nil && el.Element.RuntimeID == "" {
		el.Element.RuntimeID = strings.TrimPrefix(id, "rid:")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if existing, ok := c.elementsByRuntimeID[id]; ok && existing != nil && existing.NativePtr != el.NativePtr {
		c.elementsByFallback[id] = cloneBridgeElement(el)
		return
	}
	c.elementsByRuntimeID[id] = cloneBridgeElement(el)
}

func (c *nativeUIAComClient) releaseCachedElements() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, el := range c.elementsByRuntimeID {
		if el != nil {
			comRelease(el.NativePtr)
		}
	}
	for _, el := range c.elementsByFallback {
		if el != nil {
			comRelease(el.NativePtr)
		}
	}
	c.elementsByRuntimeID = map[string]*uiaBridgeElement{}
	c.elementsByFallback = map[string]*uiaBridgeElement{}
}
