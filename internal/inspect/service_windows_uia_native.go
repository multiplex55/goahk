//go:build windows
// +build windows

package inspect

import (
	"context"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"unsafe"

	"goahk/internal/window"
)

type nativeUIABridge interface {
	ResolveRoot(window.HWND) (*uiaElement, error)
	ElementByHWND(window.HWND) (*uiaElement, error)
	ParentHWND(window.HWND) (window.HWND, bool, error)
	ChildHWNDs(window.HWND) ([]window.HWND, error)
	FocusedHWND() (window.HWND, error)
	CursorPosition() (int, int, error)
	HWNDFromPoint(x, y int) (window.HWND, error)
	Invoke(window.HWND) error
	Select(window.HWND) error
	SetValue(window.HWND, string) error
	DoDefaultAction(window.HWND) error
	Toggle(window.HWND) error
	Expand(window.HWND) error
	Collapse(window.HWND) error
}

func newNativeUIADeps() windowsUIADeps {
	return &nativeUIADeps{
		bridge:        newWin32UIABridge(),
		sessionID:     newUIASessionID(),
		refToElement:  map[string]*uiaElement{},
		keyToRefCache: map[string]string{},
	}
}

var uiaSessionCounter atomic.Uint64

func newUIASessionID() string {
	return strconv.FormatUint(uiaSessionCounter.Add(1), 36)
}

type nativeUIADeps struct {
	bridge nativeUIABridge

	mu            sync.RWMutex
	sessionID     string
	refToElement  map[string]*uiaElement
	keyToRefCache map[string]string
	nextID        uint64
}

func (d *nativeUIADeps) ResolveWindowRoot(_ context.Context, hwnd string) (*uiaElement, error) {
	target, err := parseHWND(hwnd)
	if err != nil {
		return nil, errUIANilElement
	}
	el, err := d.bridge.ResolveRoot(target)
	if err != nil {
		return nil, err
	}
	if el == nil {
		return nil, errUIANilElement
	}
	return d.registerElement(el), nil
}

func (d *nativeUIADeps) GetElementByRef(_ context.Context, ref string) (*uiaElement, error) {
	el, err := d.lookupByRef(ref)
	if err != nil {
		return nil, errUIANilElement
	}
	hwnd, err := parseHWND(el.HWND)
	if err != nil {
		return nil, errUIANilElement
	}
	latest, err := d.bridge.ElementByHWND(hwnd)
	if err != nil {
		return nil, err
	}
	if latest == nil {
		return nil, errUIAElementNotAvailable
	}
	return d.registerElement(latest), nil
}

func (d *nativeUIADeps) GetChildren(context.Context, string) ([]*uiaElement, error) {
	return nil, ErrProviderActionUnsupported
}

func (d *nativeUIADeps) GetParent(context.Context, string) (*uiaElement, error) {
	return nil, ErrProviderActionUnsupported
}

func (d *nativeUIADeps) GetChildCount(context.Context, string) (int, bool, error) {
	return 0, false, ErrProviderActionUnsupported
}

func (d *nativeUIADeps) GetFocusedElement(_ context.Context) (*uiaElement, error) {
	hwnd, err := d.bridge.FocusedHWND()
	if err != nil {
		return nil, err
	}
	if hwnd == 0 {
		return nil, errUIAElementNotAvailable
	}
	el, err := d.bridge.ElementByHWND(hwnd)
	if err != nil {
		return nil, err
	}
	return d.registerElement(el), nil
}

func (d *nativeUIADeps) GetCursorPosition(_ context.Context) (int, int, error) {
	return d.bridge.CursorPosition()
}

func (d *nativeUIADeps) ElementFromPoint(_ context.Context, x, y int) (*uiaElement, error) {
	hwnd, err := d.bridge.HWNDFromPoint(x, y)
	if err != nil {
		return nil, err
	}
	if hwnd == 0 {
		return nil, errUIAElementNotAvailable
	}
	el, err := d.bridge.ElementByHWND(hwnd)
	if err != nil {
		return nil, err
	}
	return d.registerElement(el), nil
}

func (d *nativeUIADeps) Invoke(_ context.Context, ref string) error {
	return d.withHWND(ref, d.bridge.Invoke)
}
func (d *nativeUIADeps) Select(_ context.Context, ref string) error {
	return d.withHWND(ref, d.bridge.Select)
}
func (d *nativeUIADeps) SetValue(_ context.Context, ref, value string) error {
	hwnd, err := d.resolveHWND(ref)
	if err != nil {
		return errUIANilElement
	}
	return d.bridge.SetValue(hwnd, value)
}
func (d *nativeUIADeps) DoDefaultAction(_ context.Context, ref string) error {
	return d.withHWND(ref, d.bridge.DoDefaultAction)
}
func (d *nativeUIADeps) Toggle(_ context.Context, ref string) error {
	return d.withHWND(ref, d.bridge.Toggle)
}
func (d *nativeUIADeps) Expand(_ context.Context, ref string) error {
	return d.withHWND(ref, d.bridge.Expand)
}
func (d *nativeUIADeps) Collapse(_ context.Context, ref string) error {
	return d.withHWND(ref, d.bridge.Collapse)
}

func (d *nativeUIADeps) withHWND(ref string, fn func(window.HWND) error) error {
	hwnd, err := d.resolveHWND(ref)
	if err != nil {
		return errUIANilElement
	}
	return fn(hwnd)
}

func (d *nativeUIADeps) resolveHWND(ref string) (window.HWND, error) {
	el, err := d.lookupByRef(ref)
	if err != nil {
		return 0, err
	}
	return parseHWND(el.HWND)
}

func (d *nativeUIADeps) lookupByRef(ref string) (*uiaElement, error) {
	parsed, err := parseNodeRef(ref)
	if err != nil || parsed.Provider != nodeRefProviderUIA || parsed.Session != d.sessionID {
		return nil, ErrInvalidNodeRef
	}
	d.mu.RLock()
	defer d.mu.RUnlock()
	el, ok := d.refToElement[ref]
	if !ok {
		return nil, &NodeRefNotFoundError{Provider: nodeRefProviderUIA, Ref: ref}
	}
	return cloneUIAElement(el), nil
}

func (d *nativeUIADeps) registerElement(el *uiaElement) *uiaElement {
	if el == nil {
		return nil
	}
	key := d.cacheKeyForElement(el)
	d.mu.Lock()
	defer d.mu.Unlock()
	if existingRef, ok := d.keyToRefCache[key]; ok {
		cloned := cloneUIAElement(el)
		cloned.Ref = existingRef
		d.refToElement[existingRef] = cloneUIAElement(cloned)
		return cloned
	}
	d.nextID++
	ref := makeUIANodeRef(d.sessionID, strconv.FormatUint(d.nextID, 36))
	cloned := cloneUIAElement(el)
	cloned.Ref = ref
	d.keyToRefCache[key] = ref
	d.refToElement[ref] = cloneUIAElement(cloned)
	return cloned
}

func (d *nativeUIADeps) cacheKeyForElement(el *uiaElement) string {
	if rid := strings.TrimSpace(el.RuntimeID); rid != "" {
		return "rid:" + rid
	}
	return "hwnd:" + strings.TrimSpace(el.HWND)
}

func cloneUIAElement(el *uiaElement) *uiaElement {
	if el == nil {
		return nil
	}
	cloned := *el
	return &cloned
}

type win32UIABridge struct{}

func newWin32UIABridge() nativeUIABridge { return win32UIABridge{} }

func (win32UIABridge) ResolveRoot(hwnd window.HWND) (*uiaElement, error)   { return describeHWND(hwnd) }
func (win32UIABridge) ElementByHWND(hwnd window.HWND) (*uiaElement, error) { return describeHWND(hwnd) }
func (win32UIABridge) ParentHWND(hwnd window.HWND) (window.HWND, bool, error) {
	p, _, _ := procGetParent.Call(uintptr(hwnd))
	if p == 0 {
		return 0, false, nil
	}
	return window.HWND(p), true, nil
}
func (win32UIABridge) ChildHWNDs(hwnd window.HWND) ([]window.HWND, error) {
	first, _, _ := procGetWindow.Call(uintptr(hwnd), uintptr(gwChild))
	out := []window.HWND{}
	for cur := first; cur != 0; {
		out = append(out, window.HWND(cur))
		next, _, _ := procGetWindow.Call(cur, uintptr(gwHwndNext))
		cur = next
	}
	return out, nil
}
func (win32UIABridge) FocusedHWND() (window.HWND, error) {
	h, _, _ := procGetForegroundWindow.Call()
	return window.HWND(h), nil
}
func (win32UIABridge) CursorPosition() (int, int, error) {
	var pt winPoint
	ok, _, err := procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
	if ok == 0 {
		return 0, 0, err
	}
	return int(pt.X), int(pt.Y), nil
}
func (win32UIABridge) HWNDFromPoint(x, y int) (window.HWND, error) {
	packed := uintptr(uint32(y))<<16 | uintptr(uint32(x)&0xFFFF)
	h, _, _ := procWindowFromPoint.Call(packed)
	return window.HWND(h), nil
}
func (win32UIABridge) Invoke(window.HWND) error           { return ErrProviderActionUnsupported }
func (win32UIABridge) Select(window.HWND) error           { return ErrProviderActionUnsupported }
func (win32UIABridge) SetValue(window.HWND, string) error { return ErrProviderActionUnsupported }
func (win32UIABridge) DoDefaultAction(window.HWND) error  { return ErrProviderActionUnsupported }
func (win32UIABridge) Toggle(window.HWND) error           { return ErrProviderActionUnsupported }
func (win32UIABridge) Expand(window.HWND) error           { return ErrProviderActionUnsupported }
func (win32UIABridge) Collapse(window.HWND) error         { return ErrProviderActionUnsupported }
