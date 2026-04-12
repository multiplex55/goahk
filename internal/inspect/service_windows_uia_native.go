//go:build windows
// +build windows

package inspect

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

func newNativeUIADeps() windowsUIADeps {
	return &nativeUIADeps{
		bridge:       newWin32UIABridge(),
		sessionID:    newUIASessionID(),
		refToElement: map[string]*cachedBridgeElement{},
		keyToRef:     map[string]string{},
	}
}

var uiaSessionCounter atomic.Uint64

func newUIASessionID() string {
	return strconv.FormatUint(uiaSessionCounter.Add(1), 36)
}

type cachedBridgeElement struct {
	bridge *uiaBridgeElement
	elem   *uiaElement
}

type nativeUIADeps struct {
	bridge nativeUIABridge

	mu           sync.RWMutex
	sessionID    string
	refToElement map[string]*cachedBridgeElement
	keyToRef     map[string]string
	nextID       uint64
}

func (d *nativeUIADeps) ResolveWindowRoot(_ context.Context, hwnd string) (*uiaElement, error) {
	target, err := parseHWND(hwnd)
	if err != nil {
		return nil, errUIANilElement
	}
	be, err := d.bridge.ResolveRoot(target)
	if err != nil {
		return nil, err
	}
	return d.registerBridgeElement(be), nil
}

func (d *nativeUIADeps) GetFocusedElement(_ context.Context) (*uiaElement, error) {
	be, err := d.bridge.FocusedElement()
	if err != nil {
		return nil, err
	}
	return d.registerBridgeElement(be), nil
}

func (d *nativeUIADeps) GetCursorPosition(_ context.Context) (int, int, error) {
	return d.bridge.CursorPosition()
}

func (d *nativeUIADeps) ElementFromPoint(_ context.Context, x, y int) (*uiaElement, error) {
	be, err := d.bridge.ElementFromPoint(x, y)
	if err != nil {
		return nil, err
	}
	return d.registerBridgeElement(be), nil
}

func (d *nativeUIADeps) GetElementByRef(_ context.Context, ref string) (*uiaElement, error) {
	cached, err := d.lookupByRef(ref)
	if err != nil {
		return nil, errUIANilElement
	}
	latest, err := d.bridge.ElementByKey(cached.bridge.Key)
	if err != nil {
		if retry := d.tryRefreshAfterStale(cached.bridge, err); retry != nil {
			latest, err = retry, nil
		}
	}
	if err != nil {
		return nil, err
	}
	return d.registerBridgeElement(latest), nil
}

func (d *nativeUIADeps) GetParent(_ context.Context, ref string) (*uiaElement, error) {
	cached, err := d.lookupByRef(ref)
	if err != nil {
		return nil, errUIANilElement
	}
	parent, err := d.bridge.Parent(cloneBridgeElement(cached.bridge))
	if err != nil {
		return nil, err
	}
	if parent == nil || parent.Element == nil {
		return nil, errUIAElementNotAvailable
	}
	return d.registerBridgeElement(parent), nil
}

func (d *nativeUIADeps) GetChildren(_ context.Context, ref string) ([]*uiaElement, error) {
	cached, err := d.lookupByRef(ref)
	if err != nil {
		return nil, errUIANilElement
	}
	children, err := d.bridge.Children(cloneBridgeElement(cached.bridge))
	if err != nil {
		return nil, err
	}
	out := make([]*uiaElement, 0, len(children))
	for _, child := range children {
		if child == nil || child.Element == nil {
			continue
		}
		registered := d.registerBridgeElement(child)
		registered.ParentRef = ref
		out = append(out, registered)
	}
	return out, nil
}

func (d *nativeUIADeps) GetChildCount(ctx context.Context, ref string) (int, bool, error) {
	children, err := d.GetChildren(ctx, ref)
	if err != nil {
		return 0, false, err
	}
	return len(children), true, nil
}

func (d *nativeUIADeps) Invoke(_ context.Context, ref string) error {
	return d.withBridgeElement(ref, d.bridge.Invoke)
}
func (d *nativeUIADeps) Select(_ context.Context, ref string) error {
	return d.withBridgeElement(ref, d.bridge.Select)
}
func (d *nativeUIADeps) SetValue(_ context.Context, ref, value string) error {
	return d.withBridgeElement(ref, func(el *uiaBridgeElement) error { return d.bridge.SetValue(el, value) })
}
func (d *nativeUIADeps) DoDefaultAction(_ context.Context, ref string) error {
	return d.withBridgeElement(ref, d.bridge.DoDefaultAction)
}
func (d *nativeUIADeps) Toggle(_ context.Context, ref string) error {
	return d.withBridgeElement(ref, d.bridge.Toggle)
}
func (d *nativeUIADeps) Expand(_ context.Context, ref string) error {
	return d.withBridgeElement(ref, d.bridge.Expand)
}
func (d *nativeUIADeps) Collapse(_ context.Context, ref string) error {
	return d.withBridgeElement(ref, d.bridge.Collapse)
}

func (d *nativeUIADeps) withBridgeElement(ref string, fn func(*uiaBridgeElement) error) error {
	cached, err := d.lookupByRef(ref)
	if err != nil {
		return errUIANilElement
	}
	return fn(cloneBridgeElement(cached.bridge))
}

func (d *nativeUIADeps) lookupByRef(ref string) (*cachedBridgeElement, error) {
	parsed, err := parseNodeRef(ref)
	if err != nil || parsed.Provider != nodeRefProviderUIA || parsed.Session != d.sessionID {
		return nil, ErrInvalidNodeRef
	}
	d.mu.RLock()
	defer d.mu.RUnlock()
	entry, ok := d.refToElement[ref]
	if !ok {
		return nil, &NodeRefNotFoundError{Provider: nodeRefProviderUIA, Ref: ref}
	}
	return &cachedBridgeElement{bridge: cloneBridgeElement(entry.bridge), elem: cloneUIAElement(entry.elem)}, nil
}

func (d *nativeUIADeps) registerBridgeElement(be *uiaBridgeElement) *uiaElement {
	if be == nil || be.Element == nil {
		return nil
	}
	el := cloneUIAElement(be.Element)
	if el.UnsupportedProps == nil {
		el.UnsupportedProps = map[string]bool{}
	}
	for k, v := range be.UnsupportedProperty {
		el.UnsupportedProps[k] = v
	}
	if el.PropertyStates == nil {
		el.PropertyStates = map[string]string{}
	}
	for k, status := range be.PropertyState {
		el.PropertyStates[k] = normalizePropertyStatus(status)
	}
	for k, unsupported := range be.UnsupportedProperty {
		if unsupported {
			el.PropertyStates[k] = propertyStatusUnsupported
			continue
		}
		if _, hasStatus := el.PropertyStates[k]; !hasStatus {
			el.PropertyStates[k] = propertyStatusEmpty
		}
	}
	if len(be.SupportedPatterns) > 0 {
		el.SupportedPatterns = append([]string(nil), be.SupportedPatterns...)
	}
	key := strings.TrimSpace(be.Key)
	if key == "" {
		key = d.cacheKeyForElement(el)
	}

	d.mu.Lock()
	defer d.mu.Unlock()
	if existingRef, ok := d.keyToRef[key]; ok {
		el.Ref = existingRef
		d.refToElement[existingRef] = &cachedBridgeElement{bridge: cloneBridgeElement(&uiaBridgeElement{Element: el, Key: key, AllowHWNDFallback: be.AllowHWNDFallback, SupportedPatterns: el.SupportedPatterns, UnsupportedProperty: el.UnsupportedProps, PropertyState: el.PropertyStates}), elem: cloneUIAElement(el)}
		return el
	}
	d.nextID++
	ref := makeUIANodeRef(d.sessionID, strconv.FormatUint(d.nextID, 36))
	el.Ref = ref
	storedBridge := cloneBridgeElement(&uiaBridgeElement{Element: cloneUIAElement(el), Key: key, AllowHWNDFallback: be.AllowHWNDFallback, SupportedPatterns: el.SupportedPatterns, UnsupportedProperty: el.UnsupportedProps, PropertyState: el.PropertyStates})
	d.keyToRef[key] = ref
	d.refToElement[ref] = &cachedBridgeElement{bridge: storedBridge, elem: cloneUIAElement(el)}
	return el
}

func (d *nativeUIADeps) tryRefreshAfterStale(el *uiaBridgeElement, err error) *uiaBridgeElement {
	var staleErr *UIAElementStaleError
	if !errors.As(err, &staleErr) || el == nil || !el.AllowHWNDFallback {
		return nil
	}
	hwnd, parseErr := parseHWND(el.Element.HWND)
	if parseErr != nil || hwnd == 0 {
		return nil
	}
	fresh, rootErr := d.bridge.ResolveRoot(hwnd)
	if rootErr != nil || fresh == nil {
		return nil
	}
	if fresh.Key == "" {
		return fresh
	}
	latest, latestErr := d.bridge.ElementByKey(fresh.Key)
	if latestErr != nil {
		return nil
	}
	return latest
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
	if el.SupportedPatterns != nil {
		cloned.SupportedPatterns = append([]string(nil), el.SupportedPatterns...)
	}
	if el.UnsupportedProps != nil {
		cloned.UnsupportedProps = map[string]bool{}
		for k, v := range el.UnsupportedProps {
			cloned.UnsupportedProps[k] = v
		}
	}
	if el.PropertyStates != nil {
		cloned.PropertyStates = map[string]string{}
		for k, v := range el.PropertyStates {
			cloned.PropertyStates[k] = v
		}
	}
	return &cloned
}

func cloneBridgeElement(el *uiaBridgeElement) *uiaBridgeElement {
	if el == nil {
		return nil
	}
	out := *el
	out.Element = cloneUIAElement(el.Element)
	if el.SupportedPatterns != nil {
		out.SupportedPatterns = append([]string(nil), el.SupportedPatterns...)
	}
	if el.UnsupportedProperty != nil {
		out.UnsupportedProperty = map[string]bool{}
		for k, v := range el.UnsupportedProperty {
			out.UnsupportedProperty[k] = v
		}
	}
	if el.PropertyState != nil {
		out.PropertyState = map[string]string{}
		for k, v := range el.PropertyState {
			out.PropertyState[k] = v
		}
	}
	return &out
}
