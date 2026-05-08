//go:build windows
// +build windows

package inspect

import (
	"errors"
	"fmt"
	"log"
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

type uiaTraversalStrategy string

const (
	uiaTraversalRawTrueCondition uiaTraversalStrategy = "raw-true-condition"
	uiaTraversalControlView      uiaTraversalStrategy = "control-view"
	uiaTraversalContentView      uiaTraversalStrategy = "content-view"
)

type uiaTraversalOptions struct {
	Strategy uiaTraversalStrategy
}

func normalizeTraversalOptions(opts *uiaTraversalOptions) uiaTraversalOptions {
	if opts == nil {
		return uiaTraversalOptions{Strategy: uiaTraversalRawTrueCondition}
	}
	switch opts.Strategy {
	case uiaTraversalControlView, uiaTraversalContentView, uiaTraversalRawTrueCondition:
		return *opts
	default:
		return uiaTraversalOptions{Strategy: uiaTraversalRawTrueCondition}
	}
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

var (
	invokePatternCall          = uiaInvokePatternInvoke
	selectPatternCall          = uiaSelectionItemPatternSelect
	setValuePatternCall        = uiaValuePatternSetValue
	doDefaultActionPatternCall = uiaLegacyIAccessiblePatternDoDefaultAction
	togglePatternCall          = uiaTogglePatternToggle
	expandPatternCall          = uiaExpandCollapsePatternExpand
	collapsePatternCall        = uiaExpandCollapsePatternCollapse
)

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
	setOptStr := func(name string, prop int32, dst **string) {
		v, err := uiaGetCurrentPropertyValue(el.NativePtr, prop)
		if err != nil {
			markPropertyErr(el, name, err)
			return
		}
		r := decodeVariant(v)
		el.PropertyState[name] = r.Status
		s := strings.TrimSpace(r.S)
		if s == "" {
			*dst = nil
			return
		}
		*dst = &s
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
	setOptStr("Value", 30045, &el.Element.Value)
	setStr("AutomationId", uiaPropertyAutomationID, &el.Element.AutomationID)
	setStr("ClassName", uiaPropertyClassName, &el.Element.ClassName)
	setOptStr("HelpText", uiaPropertyHelpText, &el.Element.HelpText)
	setOptStr("AccessKey", uiaPropertyAccessKey, &el.Element.AccessKey)
	setOptStr("AcceleratorKey", uiaPropertyAccelerator, &el.Element.AcceleratorKey)
	setBool("HasKeyboardFocus", uiaPropertyHasFocus, &el.Element.HasKeyboardFocus)
	setBool("IsKeyboardFocusable", uiaPropertyIsFocusable, &el.Element.IsKeyboardFocusable)
	setOptStr("ItemType", uiaPropertyItemType, &el.Element.ItemType)
	setInt("ProcessId", uiaPropertyProcessID, &el.Element.ProcessID)
	setBool("IsEnabled", uiaPropertyIsEnabled, &el.Element.IsEnabled)
	setBool("IsPassword", uiaPropertyIsPassword, &el.Element.IsPassword)
	setBool("IsOffscreen", uiaPropertyIsOffscreen, &el.Element.IsOffscreen)
	setStr("FrameworkId", uiaPropertyFrameworkID, &el.Element.FrameworkID)
	setBool("IsRequiredForForm", uiaPropertyIsRequired, &el.Element.IsRequiredForForm)
	setOptStr("ItemStatus", uiaPropertyItemStatus, &el.Element.ItemStatus)
	setOptStr("LabeledBy", uiaPropertyLabeledBy, &el.Element.LabeledBy)
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
	log.Printf("inspect.uia.native.find_children checkpoint=\"FindAll array length\" length=%d", n)
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
	log.Printf("inspect.uia.com_client.children parent_key=%s parent_runtime_id=%s parent_native_ptr_present=%t", el.Key, el.RuntimeID, el.NativePtr != 0)
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
	return c.executePatternAction(el, "Invoke", "Invoke", func(ptr uintptr) error { return invokePatternCall(ptr) })
}
func (c *nativeUIAComClient) Select(el *uiaBridgeElement) error {
	return c.executePatternAction(el, "Select", "SelectionItem", func(ptr uintptr) error { return selectPatternCall(ptr) })
}
func (c *nativeUIAComClient) SetValue(el *uiaBridgeElement, value string) error {
	if el == nil {
		return errUIANilElement
	}
	if strings.TrimSpace(el.Key) == "" {
		return &UIAElementStaleError{Op: "SetValue", Err: errors.New("element reference is stale")}
	}
	return c.executePatternAction(el, "SetValue", "Value", func(ptr uintptr) error { return setValuePatternCall(ptr, value) })
}
func (c *nativeUIAComClient) DoDefaultAction(el *uiaBridgeElement) error {
	return c.executePatternAction(el, "DoDefaultAction", "LegacyIAccessible", func(ptr uintptr) error { return doDefaultActionPatternCall(ptr) })
}
func (c *nativeUIAComClient) Toggle(el *uiaBridgeElement) error {
	return c.executePatternAction(el, "Toggle", "Toggle", func(ptr uintptr) error { return togglePatternCall(ptr) })
}
func (c *nativeUIAComClient) Expand(el *uiaBridgeElement) error {
	return c.executePatternAction(el, "Expand", "ExpandCollapse", func(ptr uintptr) error { return expandPatternCall(ptr) })
}
func (c *nativeUIAComClient) Collapse(el *uiaBridgeElement) error {
	return c.executePatternAction(el, "Collapse", "ExpandCollapse", func(ptr uintptr) error { return collapsePatternCall(ptr) })
}

func (c *nativeUIAComClient) executePatternAction(el *uiaBridgeElement, op, pattern string, call func(uintptr) error) error {
	if el == nil {
		return errUIANilElement
	}
	resolved, err := c.ElementByRuntimeID(el.Key)
	if err != nil {
		return err
	}
	for _, supported := range resolved.SupportedPatterns {
		if strings.EqualFold(strings.TrimSpace(supported), pattern) {
			var actionErr error
			err := c.worker.Do(op, func(*uiaWorkerState) error {
				if resolved.NativePtr == 0 {
					return &UIAElementStaleError{Op: op, Err: errors.New("element pointer is stale")}
				}
				actionErr = call(resolved.NativePtr)
				return nil
			})
			if err != nil {
				return err
			}
			if actionErr != nil {
				return actionErr
			}
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
