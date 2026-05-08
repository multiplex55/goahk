//go:build windows
// +build windows

package inspect

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"unsafe"

	"goahk/internal/window"
	"golang.org/x/sys/windows"
)

type accBridgeElement struct {
	Key           string
	ParentKey     string
	RuntimeID     string
	HWND          string
	Name          string
	Role          string
	Value         *string
	ClassName     string
	Framework     string
	Rect          *Rect
	ChildCount    int
	Description   *string
	State         *string
	DefaultAction *string
	NativeObj     uintptr
	NativeChildID int32
	PathSegment   string
}

type nativeACCBridge interface {
	ObjectFromWindow(window.HWND) (*accBridgeElement, error)
	ObjectFromPoint(x, y int) (*accBridgeElement, error)
	ObjectByKey(key string) (*accBridgeElement, error)
	Parent(*accBridgeElement) (*accBridgeElement, error)
	Children(*accBridgeElement) ([]*accBridgeElement, error)
	CursorPosition() (int, int, error)
}

type nativeACCDeps struct {
	bridge nativeACCBridge

	mu           sync.RWMutex
	sessionID    string
	refToElement map[string]*accBridgeElement
	keyToRef     map[string]string
	nextID       uint64
}

var accSessionCounter atomic.Uint64

func newACCSessionID() string {
	return strconv.FormatUint(accSessionCounter.Add(1), 36)
}

func newNativeACCDeps() windowsUIADeps {
	return &nativeACCDeps{
		bridge:       newWin32ACCBridge(),
		sessionID:    newACCSessionID(),
		refToElement: map[string]*accBridgeElement{},
		keyToRef:     map[string]string{},
	}
}

func (d *nativeACCDeps) ResolveWindowRoot(_ context.Context, hwnd string) (*uiaElement, error) {
	target, err := parseHWND(hwnd)
	if err != nil {
		return nil, errUIANilElement
	}
	el, err := d.bridge.ObjectFromWindow(target)
	if err != nil {
		return nil, err
	}
	return d.registerElement(el), nil
}

func (d *nativeACCDeps) GetFocusedElement(context.Context) (*uiaElement, error) {
	return nil, ErrProviderActionUnsupported
}

func (d *nativeACCDeps) GetCursorPosition(context.Context) (int, int, error) {
	return d.bridge.CursorPosition()
}

func (d *nativeACCDeps) ElementFromPoint(_ context.Context, x, y int) (*uiaElement, error) {
	el, err := d.bridge.ObjectFromPoint(x, y)
	if err != nil {
		return nil, err
	}
	return d.registerElement(el), nil
}

func (d *nativeACCDeps) GetElementByRef(_ context.Context, ref string) (*uiaElement, error) {
	cached, err := d.lookupByRef(ref)
	if err != nil {
		return nil, errUIANilElement
	}
	latest, err := d.bridge.ObjectByKey(cached.Key)
	if err != nil {
		return nil, err
	}
	if latest == nil {
		return nil, errUIAElementNotAvailable
	}
	mapped := d.registerElement(latest)
	path := d.pathForCopyDisplay(mapped.Ref)
	if strings.TrimSpace(path) != "" {
		mapped.Status = &path
	}
	return mapped, nil
}

func (d *nativeACCDeps) GetParent(_ context.Context, ref string) (*uiaElement, error) {
	cached, err := d.lookupByRef(ref)
	if err != nil {
		return nil, errUIANilElement
	}
	parent, err := d.bridge.Parent(cached)
	if err != nil {
		return nil, err
	}
	if parent == nil {
		return nil, errUIAElementNotAvailable
	}
	return d.registerElement(parent), nil
}

func (d *nativeACCDeps) GetChildren(_ context.Context, ref string) ([]*uiaElement, error) {
	cached, err := d.lookupByRef(ref)
	if err != nil {
		return nil, errUIANilElement
	}
	children, err := d.bridge.Children(cached)
	if err != nil {
		return nil, err
	}
	out := make([]*uiaElement, 0, len(children))
	for _, child := range children {
		if child == nil {
			continue
		}
		if strings.TrimSpace(child.ParentKey) == "" {
			child.ParentKey = strings.TrimSpace(cached.Key)
		}
		out = append(out, d.registerElement(child))
	}
	return out, nil
}

func (d *nativeACCDeps) GetChildCount(ctx context.Context, ref string) (int, bool, error) {
	children, err := d.GetChildren(ctx, ref)
	if err != nil {
		return 0, false, err
	}
	return len(children), true, nil
}

func (d *nativeACCDeps) Invoke(context.Context, string) error { return ErrProviderActionUnsupported }
func (d *nativeACCDeps) Select(context.Context, string) error { return ErrProviderActionUnsupported }
func (d *nativeACCDeps) SetValue(context.Context, string, string) error {
	return ErrProviderActionUnsupported
}
func (d *nativeACCDeps) DoDefaultAction(context.Context, string) error {
	return ErrProviderActionUnsupported
}
func (d *nativeACCDeps) Toggle(context.Context, string) error   { return ErrProviderActionUnsupported }
func (d *nativeACCDeps) Expand(context.Context, string) error   { return ErrProviderActionUnsupported }
func (d *nativeACCDeps) Collapse(context.Context, string) error { return ErrProviderActionUnsupported }

func (d *nativeACCDeps) lookupByRef(ref string) (*accBridgeElement, error) {
	parsed, err := parseNodeRef(ref)
	if err != nil || parsed.Provider != nodeRefProviderACC || parsed.Session != d.sessionID {
		return nil, ErrInvalidNodeRef
	}
	d.mu.RLock()
	defer d.mu.RUnlock()
	el, ok := d.refToElement[ref]
	if !ok {
		return nil, &NodeRefNotFoundError{Provider: nodeRefProviderACC, Ref: ref}
	}
	copy := *el
	return &copy, nil
}

func (d *nativeACCDeps) registerElement(el *accBridgeElement) *uiaElement {
	if el == nil {
		return nil
	}
	key := strings.TrimSpace(el.Key)
	if key == "" {
		key = "hwnd:" + strings.TrimSpace(el.HWND) + ":rid:" + strings.TrimSpace(el.RuntimeID) + ":child:" + strconv.Itoa(int(el.NativeChildID))
	}
	var ref string
	d.mu.Lock()
	if existing, ok := d.keyToRef[key]; ok {
		ref = existing
	} else {
		d.nextID++
		ref = makeACCNodeRef(d.sessionID, strconv.FormatUint(d.nextID, 36))
		d.keyToRef[key] = ref
	}
	stored := *el
	d.refToElement[ref] = &stored
	d.mu.Unlock()

	mapped := &uiaElement{
		Ref:                  ref,
		RuntimeID:            strings.TrimSpace(el.RuntimeID),
		HWND:                 strings.TrimSpace(el.HWND),
		Name:                 strings.TrimSpace(el.Name),
		ControlType:          normalizeControlType(el.Role, el.Role),
		LocalizedControlType: strings.ToLower(strings.TrimSpace(el.Role)),
		ClassName:            strings.TrimSpace(el.ClassName),
		FrameworkID:          strings.TrimSpace(el.Framework),
		HelpText:             el.Description,
		ItemStatus:           el.State,
		Value:                el.Value,
		BoundingRect:         toUIARect(el.Rect),
		IsEnabled:            true,
		IsControlElement:     true,
		IsContentElement:     true,
		UnsupportedProps: map[string]bool{
			"AutomationId":        true,
			"HelpText":            true,
			"AccessKey":           true,
			"AcceleratorKey":      true,
			"HasKeyboardFocus":    true,
			"IsKeyboardFocusable": true,
			"ItemType":            true,
			"ItemStatus":          true,
			"IsPassword":          true,
			"IsOffscreen":         true,
			"IsRequiredForForm":   true,
			"LabeledBy":           true,
		},
		SupportedPatterns: []string{"LegacyIAccessible"},
	}
	if strings.TrimSpace(el.ParentKey) != "" {
		if parentRef := d.refForKey(strings.TrimSpace(el.ParentKey)); parentRef != "" {
			mapped.ParentRef = parentRef
		}
	}
	if seg := strings.TrimSpace(el.PathSegment); seg != "" {
		if mapped.Status == nil || strings.TrimSpace(*mapped.Status) == "" {
			mapped.Status = &seg
		}
	}
	return mapped
}
func derefString(v *string) string {
	if v == nil {
		return ""
	}
	return strings.TrimSpace(*v)
}

func (d *nativeACCDeps) refForKey(key string) string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.keyToRef[key]
}

func (d *nativeACCDeps) pathForCopyDisplay(ref string) string {
	parts := make([]string, 0, 8)
	current := ref
	for i := 0; i < 32 && current != ""; i++ {
		el, err := d.lookupByRef(current)
		if err != nil {
			break
		}
		label := strings.TrimSpace(el.Name)
		if label == "" {
			label = strings.TrimSpace(el.Role)
		}
		if label == "" {
			label = strings.TrimSpace(el.Key)
		}
		parts = append(parts, label)
		if strings.TrimSpace(el.ParentKey) == "" {
			break
		}
		current = d.refForKey(strings.TrimSpace(el.ParentKey))
	}
	for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
		parts[i], parts[j] = parts[j], parts[i]
	}
	return strings.Join(parts, " > ")
}

func toUIARect(r *Rect) *uiaRect {
	if r == nil {
		return nil
	}
	return &uiaRect{Left: r.Left, Top: r.Top, Width: r.Width, Height: r.Height}
}

type win32ACCBridge struct{}

func newWin32ACCBridge() nativeACCBridge { return win32ACCBridge{} }

func (win32ACCBridge) ObjectFromWindow(hwnd window.HWND) (*accBridgeElement, error) {
	acc, childID, err := accObjectFromWindow(hwnd)
	if err != nil {
		return nil, err
	}
	return buildACCElement(acc, childID, hwnd)
}
func (win32ACCBridge) ObjectFromPoint(x, y int) (*accBridgeElement, error) {
	acc, childID, hwnd, err := accObjectFromPoint(x, y)
	if err != nil {
		return nil, err
	}
	return buildACCElement(acc, childID, hwnd)
}
func (win32ACCBridge) ObjectByKey(string) (*accBridgeElement, error) {
	return nil, errUIAElementNotAvailable
}
func (win32ACCBridge) Parent(el *accBridgeElement) (*accBridgeElement, error) {
	if el == nil || el.NativeObj == 0 {
		return nil, errUIAElementNotAvailable
	}
	parent, err := accParent(el.NativeObj)
	if err != nil || parent == 0 {
		return nil, err
	}
	return buildACCElement(parent, 0, 0)
}
func (win32ACCBridge) Children(*accBridgeElement) ([]*accBridgeElement, error) {
	return nil, nil
}
func (win32ACCBridge) CursorPosition() (int, int, error) { return currentCursorPos() }

var (
	oleaccDLL                      = windows.NewLazySystemDLL("oleacc.dll")
	procAccessibleObjectFromWindow = oleaccDLL.NewProc("AccessibleObjectFromWindow")
	procAccessibleObjectFromPoint  = oleaccDLL.NewProc("AccessibleObjectFromPoint")
	procWindowFromAccessibleObject = oleaccDLL.NewProc("WindowFromAccessibleObject")
)

var iidIAccessible = windows.GUID{Data1: 0x618736e0, Data2: 0x3c3d, Data3: 0x11cf, Data4: [8]byte{0x81, 0x0c, 0x00, 0xaa, 0x00, 0x38, 0x9b, 0x71}}

const objidClient = ^uintptr(3) // -4

type winPointLong struct{ X, Y int32 }

func accObjectFromWindow(hwnd window.HWND) (uintptr, int32, error) {
	var acc uintptr
	hr, _, _ := procAccessibleObjectFromWindow.Call(uintptr(hwnd), objidClient, uintptr(unsafe.Pointer(&iidIAccessible)), uintptr(unsafe.Pointer(&acc)))
	if windows.Handle(hr) != 0 {
		return 0, 0, fmt.Errorf("inspect: AccessibleObjectFromWindow failed: %#x", hr)
	}
	return acc, 0, nil
}
func accObjectFromPoint(x, y int) (uintptr, int32, window.HWND, error) {
	var acc uintptr
	var child variant
	pt := winPointLong{X: int32(x), Y: int32(y)}
	hr, _, _ := procAccessibleObjectFromPoint.Call(*(*uintptr)(unsafe.Pointer(&pt)), uintptr(unsafe.Pointer(&acc)), uintptr(unsafe.Pointer(&child)))
	if windows.Handle(hr) != 0 {
		return 0, 0, 0, fmt.Errorf("inspect: AccessibleObjectFromPoint failed: %#x", hr)
	}
	hwnd, _ := windowFromAccessibleObject(acc)
	return acc, int32(child.LVal), hwnd, nil
}
func windowFromAccessibleObject(acc uintptr) (window.HWND, error) {
	var hwnd uintptr
	hr, _, _ := procWindowFromAccessibleObject.Call(acc, uintptr(unsafe.Pointer(&hwnd)))
	if windows.Handle(hr) != 0 {
		return 0, fmt.Errorf("inspect: WindowFromAccessibleObject failed: %#x", hr)
	}
	return window.HWND(hwnd), nil
}
func buildACCElement(acc uintptr, childID int32, hwnd window.HWND) (*accBridgeElement, error) {
	if acc == 0 {
		return nil, errUIAElementNotAvailable
	}
	name, _ := accPropString(acc, 10, childID)
	role, _ := accPropRole(acc, childID)
	value, _ := accPropString(acc, 11, childID)
	desc, _ := accPropString(acc, 12, childID)
	state, _ := accPropState(acc, childID)
	defAction, _ := accPropString(acc, 13, childID)
	rect, _ := accLocation(acc, childID)
	parentObj, _ := accParent(acc)
	parentKey := ""
	if parentObj != 0 {
		parentKey = fmt.Sprintf("acc:%#x:%d", parentObj, 0)
	}
	return &accBridgeElement{
		Key: fmt.Sprintf("acc:%#x:%d", acc, childID), ParentKey: parentKey, RuntimeID: strconv.FormatUint(uint64(acc), 16), HWND: hwnd.String(),
		Name: derefString(name), Role: role, Value: value, Rect: rect, Description: desc, State: state, DefaultAction: defAction, NativeObj: acc, NativeChildID: childID, PathSegment: strings.TrimSpace(derefString(name)),
	}, nil
}

type variant struct {
	VT, W1, W2 uint16
	LVal       int32
	_          int32
}

func accPropString(acc uintptr, vtIndex uintptr, childID int32) (*string, error) { return nil, nil }
func accPropRole(acc uintptr, childID int32) (string, error)                     { return "", nil }
func accPropState(acc uintptr, childID int32) (*string, error)                   { return nil, nil }
func accLocation(acc uintptr, childID int32) (*Rect, error)                      { return nil, nil }
func accParent(acc uintptr) (uintptr, error)                                     { return 0, nil }
