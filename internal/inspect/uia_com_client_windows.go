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
	worker             *uiaCOMWorker
	elementsByKey      map[string]*uiaBridgeElement
	elementsByFallback map[string]*uiaBridgeElement
	mu                 sync.RWMutex
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
	patternProbeCall           = uiaGetCurrentPattern
)

func wrapNativeElementBorrowed(ptr uintptr, hwnd window.HWND, parentKey string, siblingIndex int) (*uiaBridgeElement, error) {
	comAddRef(ptr)
	el, err := wrapNativeElementOwned(ptr, hwnd, parentKey, siblingIndex, nil)
	if el != nil {
		el.OwnsNativePtr = false
	}
	return el, err
}

type uiaWrapOptions struct {
	PropertyLoadLevel uiaPropertyLoadLevel
	PopulatePatterns  bool
}

type uiaPropertyLoadLevel int

const (
	uiaPropertyLoadTree uiaPropertyLoadLevel = iota
	uiaPropertyLoadDetails
)

var uiaGetCurrentPropertyValueCall = uiaGetCurrentPropertyValue
var uiaElementRuntimeIDCall = uiaElementRuntimeID

func normalizeWrapOptions(opts *uiaWrapOptions) uiaWrapOptions {
	if opts == nil {
		return uiaWrapOptions{PropertyLoadLevel: uiaPropertyLoadDetails, PopulatePatterns: true}
	}
	g := GetUIAFeatureGates()
	resolved := uiaWrapOptions{
		PropertyLoadLevel: opts.PropertyLoadLevel,
		PopulatePatterns:  opts.PopulatePatterns,
	}
	if g.MinimalProperties {
		resolved.PropertyLoadLevel = uiaPropertyLoadTree
	}
	if g.DisablePatterns {
		resolved.PopulatePatterns = false
	}
	return resolved
}

func wrapNativeElementOwned(ptr uintptr, hwnd window.HWND, parentKey string, siblingIndex int, opts *uiaWrapOptions) (*uiaBridgeElement, error) {
	if ptr == 0 {
		return nil, &UIAComUnavailableError{Op: "WrapElement", Err: errors.New("nil COM element")}
	}
	resolvedOpts := normalizeWrapOptions(opts)
	rid, err := uiaElementRuntimeIDCall(ptr)
	if err != nil {
		return nil, err
	}
	key := canonicalUIAKey(rid, true)
	rawRuntimeID := strings.TrimSpace(rid)
	el := &uiaElement{RuntimeID: rawRuntimeID, HWND: hwnd.String()}
	b := &uiaBridgeElement{Key: key, RuntimeID: rawRuntimeID, AllowHWNDFallback: hwnd != 0, SupportedPatterns: nil, PropertyState: map[string]string{}, UnsupportedProperty: map[string]bool{}, NativePtr: ptr, Element: el, OwnsNativePtr: true}
	switch resolvedOpts.PropertyLoadLevel {
	case uiaPropertyLoadDetails:
		populateElementDetailsProperties(b, resolvedOpts.PopulatePatterns)
	default:
		populateElementTreeProperties(b)
	}
	if b.Key == "" {
		if parentKey != "" && siblingIndex >= 0 {
			b.Key = fallbackPathKey(parentKey, siblingIndex, b.Element)
		} else {
			b.Key = fmt.Sprintf("ptr:%x", ptr)
		}
	}
	if resolvedOpts.PopulatePatterns && resolvedOpts.PropertyLoadLevel != uiaPropertyLoadDetails {
		populateSupportedPatterns(b)
	}
	return b, nil
}

func populateElementTreeProperties(el *uiaBridgeElement) {
	if el == nil || el.NativePtr == 0 || el.Element == nil {
		return
	}
	setStr := func(name string, prop int32, dst *string) {
		v, err := uiaGetCurrentPropertyValueCall(el.NativePtr, prop)
		if err != nil {
			markPropertyErr(el, name, err)
			return
		}
		defer clearVariant(&v)
		r := decodeVariant(v)
		el.PropertyState[name] = r.Status
		if r.Status == propertyStatusUnsupported {
			el.UnsupportedProperty[name] = true
			return
		}
		*dst = strings.TrimSpace(r.S)
	}
	setInt := func(name string, prop int32, dst *int) {
		v, err := uiaGetCurrentPropertyValueCall(el.NativePtr, prop)
		if err != nil {
			markPropertyErr(el, name, err)
			return
		}
		defer clearVariant(&v)
		r := decodeVariant(v)
		el.PropertyState[name] = r.Status
		*dst = r.I
	}
	setControlType := func() {
		v, err := uiaGetCurrentPropertyValueCall(el.NativePtr, uiaPropertyControlType)
		if err != nil {
			markPropertyErr(el, "ControlType", err)
			return
		}
		defer clearVariant(&v)
		r := decodeVariant(v)
		el.PropertyState["ControlType"] = r.Status
		el.Element.ControlType = controlTypeNameForID(r.I)
	}
	setControlType()
	setStr("LocalizedControlType", uiaPropertyLocalizedCtl, &el.Element.LocalizedControlType)
	setStr("Name", uiaPropertyName, &el.Element.Name)
	setStr("ClassName", uiaPropertyClassName, &el.Element.ClassName)
	setInt("ProcessId", uiaPropertyProcessID, &el.Element.ProcessID)
	setInt("NativeWindowHandle", uiaPropertyNativeHWND, new(int))
	v, err := uiaGetCurrentPropertyValueCall(el.NativePtr, uiaPropertyBoundingRect)
	if err == nil {
		defer clearVariant(&v)
		r := decodeVariant(v)
		el.PropertyState["BoundingRectangle"] = r.Status
		el.Element.BoundingRect = r.Rect
	} else {
		markPropertyErr(el, "BoundingRectangle", err)
	}
}

func populateElementDetailsProperties(el *uiaBridgeElement, includePatterns bool) {
	populateElementTreeProperties(el)
	if el == nil || el.NativePtr == 0 || el.Element == nil {
		return
	}
	setStr := func(name string, prop int32, dst *string) {
		v, err := uiaGetCurrentPropertyValueCall(el.NativePtr, prop)
		if err != nil {
			markPropertyErr(el, name, err)
			return
		}
		defer clearVariant(&v)
		r := decodeVariant(v)
		el.PropertyState[name] = r.Status
		if r.Status == propertyStatusUnsupported {
			el.UnsupportedProperty[name] = true
			return
		}
		*dst = strings.TrimSpace(r.S)
	}
	setOptStr := func(name string, prop int32, dst **string) {
		v, err := uiaGetCurrentPropertyValueCall(el.NativePtr, prop)
		if err != nil {
			markPropertyErr(el, name, err)
			return
		}
		defer clearVariant(&v)
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
		v, err := uiaGetCurrentPropertyValueCall(el.NativePtr, prop)
		if err != nil {
			markPropertyErr(el, name, err)
			return
		}
		defer clearVariant(&v)
		r := decodeVariant(v)
		el.PropertyState[name] = r.Status
		*dst = r.B
	}
	setOptStr("Value", 30045, &el.Element.Value)
	setStr("AutomationId", uiaPropertyAutomationID, &el.Element.AutomationID)
	setOptStr("HelpText", uiaPropertyHelpText, &el.Element.HelpText)
	setOptStr("AccessKey", uiaPropertyAccessKey, &el.Element.AccessKey)
	setOptStr("AcceleratorKey", uiaPropertyAccelerator, &el.Element.AcceleratorKey)
	setBool("HasKeyboardFocus", uiaPropertyHasFocus, &el.Element.HasKeyboardFocus)
	setBool("IsKeyboardFocusable", uiaPropertyIsFocusable, &el.Element.IsKeyboardFocusable)
	setOptStr("ItemType", uiaPropertyItemType, &el.Element.ItemType)
	setBool("IsEnabled", uiaPropertyIsEnabled, &el.Element.IsEnabled)
	setBool("IsPassword", uiaPropertyIsPassword, &el.Element.IsPassword)
	setBool("IsOffscreen", uiaPropertyIsOffscreen, &el.Element.IsOffscreen)
	setBool("IsControlElement", uiaPropertyIsCtrlElem, &el.Element.IsControlElement)
	setBool("IsContentElement", uiaPropertyIsContent, &el.Element.IsContentElement)
	setStr("FrameworkId", uiaPropertyFrameworkID, &el.Element.FrameworkID)
	setBool("IsRequiredForForm", uiaPropertyIsRequired, &el.Element.IsRequiredForForm)
	setOptStr("ItemStatus", uiaPropertyItemStatus, &el.Element.ItemStatus)
	setOptStr("LabeledBy", uiaPropertyLabeledBy, &el.Element.LabeledBy)
	if includePatterns {
		populateSupportedPatterns(el)
	}
}

func controlTypeNameForID(id int) string {
	if n, ok := knownControlTypeIDs[id]; ok {
		return n
	}
	if id <= 0 {
		return ""
	}
	return fmt.Sprintf("ControlType(%d)", id)
}

var knownControlTypeIDs = map[int]string{
	50000: "Button", 50001: "Calendar", 50002: "CheckBox", 50003: "ComboBox", 50004: "Edit",
	50005: "Hyperlink", 50006: "Image", 50007: "ListItem", 50008: "List", 50009: "Menu",
	50010: "MenuBar", 50011: "MenuItem", 50012: "ProgressBar", 50013: "RadioButton", 50014: "ScrollBar",
	50015: "Slider", 50016: "Spinner", 50017: "StatusBar", 50018: "Tab", 50019: "TabItem",
	50020: "Text", 50021: "ToolBar", 50022: "ToolTip", 50023: "Tree", 50024: "TreeItem",
	50025: "Custom", 50026: "Group", 50027: "Thumb", 50028: "DataGrid", 50029: "DataItem",
	50030: "Document", 50031: "SplitButton", 50032: "Window", 50033: "Pane", 50034: "Header",
	50035: "HeaderItem", 50036: "Table", 50037: "TitleBar", 50038: "Separator", 50039: "SemanticZoom",
	50040: "AppBar",
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
		ok, err := patternProbeCall(el.NativePtr, d.id)
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
	return wrapNativeElementOwned(ptr, hwnd, "", -1, &uiaWrapOptions{PropertyLoadLevel: uiaPropertyLoadTree, PopulatePatterns: false})
}

func (nativeUIAAPI) FocusedElement(state *uiaWorkerState) (*uiaBridgeElement, error) {
	ptr, err := uiaGetFocusedElement(state.automation)
	if err != nil {
		return nil, err
	}
	return wrapNativeElementOwned(ptr, 0, "", -1, &uiaWrapOptions{PropertyLoadLevel: uiaPropertyLoadTree, PopulatePatterns: false})
}

func (nativeUIAAPI) ElementFromPoint(state *uiaWorkerState, x, y int) (*uiaBridgeElement, error) {
	ptr, err := uiaElementFromPoint(state.automation, x, y)
	if err != nil {
		return nil, err
	}
	return wrapNativeElementOwned(ptr, 0, "", -1, &uiaWrapOptions{PropertyLoadLevel: uiaPropertyLoadTree, PopulatePatterns: false})
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
	var diag error
	for i := int32(0); i < n; i++ {
		ptr, getErr := uiaArrayGet(arr, i)
		if getErr != nil {
			log.Printf("inspect.uia.native.find_children checkpoint=\"get child failed\" index=%d err=%v", i, getErr)
			diag = errors.Join(diag, fmt.Errorf("child[%d] get element: %w", i, getErr))
			continue
		}
		child, wrapErr := wrapNativeElementOwned(ptr, 0, parent.Key, int(i), &uiaWrapOptions{PropertyLoadLevel: uiaPropertyLoadTree, PopulatePatterns: false})
		if wrapErr != nil {
			log.Printf("inspect.uia.native.find_children checkpoint=\"wrap child failed\" index=%d err=%v", i, wrapErr)
			diag = errors.Join(diag, fmt.Errorf("child[%d] wrap: %w", i, wrapErr))
			continue
		}
		child.Element.ParentRef = parent.Key
		child.ParentKey = parent.Key
		out = append(out, child)
	}
	return out, diag
}

func (nativeUIAAPI) GetParent(state *uiaWorkerState, el *uiaBridgeElement) (*uiaBridgeElement, error) {
	if el == nil || el.NativePtr == 0 {
		return nil, &UIAElementStaleError{Op: "GetParentElement", Err: errors.New("element is stale")}
	}
	parent, err := uiaGetParentElement(state.treeWalker, el.NativePtr)
	if err != nil || parent == 0 {
		return nil, err
	}
	return wrapNativeElementOwned(parent, 0, "", -1, &uiaWrapOptions{PropertyLoadLevel: uiaPropertyLoadTree, PopulatePatterns: false})
}

func fallbackPathKey(parentKey string, siblingIndex int, el *uiaElement) string {
	parent := canonicalUIAKey(parentKey, false)
	if parent == "" {
		return ""
	}
	controlType := ""
	name := ""
	if el != nil {
		controlType = strings.TrimSpace(el.ControlType)
		if controlType == "" {
			controlType = strings.TrimSpace(el.LocalizedControlType)
		}
		name = strings.TrimSpace(el.Name)
	}
	return fmt.Sprintf("path:%s/%d/%s/%s", parent, siblingIndex, controlType, name)
}

func newNativeUIAComClient() (uiaAutomationClient, error) {
	worker, err := newUIACOMWorker()
	if err != nil {
		return nil, err
	}
	return &nativeUIAComClient{
		worker:             worker,
		elementsByKey:      map[string]*uiaBridgeElement{},
		elementsByFallback: map[string]*uiaBridgeElement{},
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
func (c *nativeUIAComClient) ElementByKey(key string) (*uiaBridgeElement, error) {
	id := canonicalUIAKey(key, false)
	if id == "" {
		return nil, &UIAElementStaleError{Op: "ElementByKey", Err: errors.New("key is stale or unavailable")}
	}
	c.mu.RLock()
	el, ok := c.elementsByKey[id]
	c.mu.RUnlock()
	if !ok || el == nil {
		if fallback, ok := c.elementsByFallback[id]; ok && fallback != nil {
			return cloneBridgeElement(fallback), nil
		}
		return nil, &UIAElementStaleError{Op: "ElementByKey", Err: fmt.Errorf("key %q is stale or unavailable", id)}
	}
	if el.NativePtr != 0 {
		if err := c.worker.Do("ElementByKey", func(*uiaWorkerState) error {
			populateElementDetailsProperties(el, true)
			return nil
		}); err != nil {
			return nil, err
		}
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
	resolved, err := c.ElementByKey(el.Key)
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
	return strings.Join(parts, ".")
}

func canonicalUIAKey(raw string, isRuntimeID bool) string {
	key := strings.TrimSpace(raw)
	if key == "" {
		return ""
	}
	for _, prefix := range []string{"rid:", "ptr:", "path:", "fallback:"} {
		if strings.HasPrefix(key, prefix) {
			return key
		}
	}
	if isRuntimeID {
		return "rid:" + key
	}
	return key
}

func (c *nativeUIAComClient) cacheBridgeElement(el *uiaBridgeElement) {
	if el == nil {
		return
	}
	id := canonicalUIAKey(el.Key, false)
	if id == "" && el.Element != nil {
		id = canonicalUIAKey(el.Element.RuntimeID, true)
	}
	if id == "" && el.ParentKey != "" && el.Element != nil && el.ParentKey != "" {
		siblingID := strings.TrimSpace(el.Element.RuntimeID)
		if siblingID == "" {
			siblingID = strings.TrimSpace(el.Element.Name)
		}
		controlType := strings.TrimSpace(el.Element.ControlType)
		if controlType == "" {
			controlType = strings.TrimSpace(el.Element.LocalizedControlType)
		}
		id = "path:" + canonicalUIAKey(el.ParentKey, false) + "/" + siblingID + "/" + controlType + "/" + strings.TrimSpace(el.Element.Name)
	}
	if id == "" {
		id = fmt.Sprintf("ptr:%x", el.NativePtr)
	}
	if el.Key == "" {
		el.Key = id
	}
	if el.Element != nil && el.Element.RuntimeID == "" {
		el.Element.RuntimeID = strings.TrimPrefix(id, "rid:")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if existing, ok := c.elementsByKey[id]; ok && existing != nil && existing.NativePtr != el.NativePtr {
		// Do not release here. This runs outside the UIA COM worker apartment and can crash.
		replacement := cloneBridgeElement(el)
		replacement.OwnsNativePtr = true
		c.elementsByKey[id] = replacement
		delete(c.elementsByFallback, id)
		return
	}
	c.elementsByKey[id] = cloneBridgeElement(el)
	c.elementsByKey[id].OwnsNativePtr = true
}

func (c *nativeUIAComClient) releaseCachedElements() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, el := range c.elementsByKey {
		if el != nil && el.OwnsNativePtr {
			comRelease(el.NativePtr)
		}
	}
	for _, el := range c.elementsByFallback {
		if el != nil && el.OwnsNativePtr {
			comRelease(el.NativePtr)
		}
	}
	c.elementsByKey = map[string]*uiaBridgeElement{}
	c.elementsByFallback = map[string]*uiaBridgeElement{}
}

func (c *nativeUIAComClient) Close() error {
	if c == nil {
		return nil
	}
	if c.worker != nil {
		_ = c.worker.Do("ReleaseCachedElements", func(*uiaWorkerState) error {
			c.releaseCachedElements()
			return nil
		})
		return c.worker.Close()
	}
	c.releaseCachedElements()
	return nil
}
