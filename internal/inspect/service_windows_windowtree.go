//go:build windows
// +build windows

package inspect

import (
	"context"
	"strconv"
	"unsafe"

	"goahk/internal/window"
	"golang.org/x/sys/windows"
)

type windowTreeBridge interface {
	ResolveRoot(window.HWND) (*uiaElement, error)
	ElementByHWND(window.HWND) (*uiaElement, error)
	ParentHWND(window.HWND) (window.HWND, bool, error)
	ChildHWNDs(window.HWND) ([]window.HWND, error)
}

type nativeWindowTreeDeps struct{ bridge windowTreeBridge }

func newNativeWindowTreeDeps() windowsUIADeps {
	return &nativeWindowTreeDeps{bridge: newWin32WindowTreeBridge()}
}

func newWindowTreeAdapter(deps windowsUIADeps) uiaAdapter {
	if deps == nil {
		deps = &unsupportedUIADeps{}
	}
	return newUIAAdapter(deps)
}

func (d *nativeWindowTreeDeps) ResolveWindowRoot(_ context.Context, hwnd string) (*uiaElement, error) {
	target, err := parseHWND(hwnd)
	if err != nil {
		return nil, errUIANilElement
	}
	return d.bridge.ResolveRoot(target)
}

func (d *nativeWindowTreeDeps) GetElementByRef(_ context.Context, ref string) (*uiaElement, error) {
	hwnd, err := parseWindowTreeRef(ref)
	if err != nil {
		return nil, errUIANilElement
	}
	el, err := d.bridge.ElementByHWND(hwnd)
	if err != nil {
		return nil, err
	}
	if el == nil {
		return nil, errUIAElementNotAvailable
	}
	return el, nil
}

func (d *nativeWindowTreeDeps) GetParent(_ context.Context, ref string) (*uiaElement, error) {
	hwnd, err := parseWindowTreeRef(ref)
	if err != nil {
		return nil, errUIANilElement
	}
	parent, ok, err := d.bridge.ParentHWND(hwnd)
	if err != nil {
		return nil, err
	}
	if !ok || parent == 0 {
		return nil, errUIAElementNotAvailable
	}
	return d.bridge.ElementByHWND(parent)
}

func (d *nativeWindowTreeDeps) GetChildren(_ context.Context, ref string) ([]*uiaElement, error) {
	hwnd, err := parseWindowTreeRef(ref)
	if err != nil {
		return nil, errUIANilElement
	}
	children, err := d.bridge.ChildHWNDs(hwnd)
	if err != nil {
		return nil, err
	}
	out := make([]*uiaElement, 0, len(children))
	for _, child := range children {
		el, err := d.bridge.ElementByHWND(child)
		if err != nil {
			return nil, err
		}
		if el == nil {
			continue
		}
		out = append(out, el)
	}
	return out, nil
}

func (d *nativeWindowTreeDeps) GetChildCount(ctx context.Context, ref string) (int, bool, error) {
	children, err := d.GetChildren(ctx, ref)
	if err != nil {
		return 0, false, err
	}
	return len(children), true, nil
}

func (d *nativeWindowTreeDeps) GetFocusedElement(context.Context) (*uiaElement, error) {
	return nil, ErrProviderActionUnsupported
}

func (d *nativeWindowTreeDeps) GetCursorPosition(context.Context) (int, int, error) {
	return 0, 0, ErrProviderActionUnsupported
}

func (d *nativeWindowTreeDeps) ElementFromPoint(context.Context, int, int) (*uiaElement, error) {
	return nil, ErrProviderActionUnsupported
}

func (d *nativeWindowTreeDeps) Invoke(context.Context, string) error {
	return ErrProviderActionUnsupported
}
func (d *nativeWindowTreeDeps) Select(context.Context, string) error {
	return ErrProviderActionUnsupported
}
func (d *nativeWindowTreeDeps) SetValue(context.Context, string, string) error {
	return ErrProviderActionUnsupported
}
func (d *nativeWindowTreeDeps) DoDefaultAction(context.Context, string) error {
	return ErrProviderActionUnsupported
}
func (d *nativeWindowTreeDeps) Toggle(context.Context, string) error {
	return ErrProviderActionUnsupported
}
func (d *nativeWindowTreeDeps) Expand(context.Context, string) error {
	return ErrProviderActionUnsupported
}
func (d *nativeWindowTreeDeps) Collapse(context.Context, string) error {
	return ErrProviderActionUnsupported
}

type win32WindowTreeBridge struct{}

func newWin32WindowTreeBridge() windowTreeBridge { return win32WindowTreeBridge{} }

func (win32WindowTreeBridge) ResolveRoot(hwnd window.HWND) (*uiaElement, error) {
	return describeHWND(hwnd)
}
func (win32WindowTreeBridge) ElementByHWND(hwnd window.HWND) (*uiaElement, error) {
	return describeHWND(hwnd)
}
func (win32WindowTreeBridge) ParentHWND(hwnd window.HWND) (window.HWND, bool, error) {
	p, _, _ := procGetParent.Call(uintptr(hwnd))
	if p == 0 {
		return 0, false, nil
	}
	return window.HWND(p), true, nil
}
func (win32WindowTreeBridge) ChildHWNDs(hwnd window.HWND) ([]window.HWND, error) {
	first, _, _ := procGetWindow.Call(uintptr(hwnd), uintptr(gwChild))
	out := []window.HWND{}
	for cur := first; cur != 0; {
		out = append(out, window.HWND(cur))
		next, _, _ := procGetWindow.Call(cur, uintptr(gwHwndNext))
		cur = next
	}
	return out, nil
}

var (
	user32DLL                = windows.NewLazySystemDLL("user32.dll")
	procGetParent            = user32DLL.NewProc("GetParent")
	procGetWindow            = user32DLL.NewProc("GetWindow")
	procGetForegroundWindow  = user32DLL.NewProc("GetForegroundWindow")
	procGetCursorPos         = user32DLL.NewProc("GetCursorPos")
	procWindowFromPoint      = user32DLL.NewProc("WindowFromPoint")
	procGetWindowTextLengthW = user32DLL.NewProc("GetWindowTextLengthW")
	procGetWindowTextW       = user32DLL.NewProc("GetWindowTextW")
	procGetClassNameW        = user32DLL.NewProc("GetClassNameW")
)

const (
	gwChild    = 5
	gwHwndNext = 2
)

type winPoint struct{ X, Y int32 }

func describeHWND(hwnd window.HWND) (*uiaElement, error) {
	if hwnd == 0 {
		return nil, errUIAElementNotAvailable
	}
	title, _ := getWindowText(hwnd)
	className, _ := getClassName(hwnd)
	parent, _, _ := win32WindowTreeBridge{}.ParentHWND(hwnd)
	return &uiaElement{
		Ref:         makeWindowNodeRef(hwnd.String()),
		RuntimeID:   strconv.FormatUint(uint64(hwnd), 10),
		HWND:        hwnd.String(),
		ParentRef:   makeWindowNodeRef(parent.String()),
		Name:        title,
		ControlType: "Window",
		ClassName:   className,
		FrameworkID: "Win32",
		IsEnabled:   true,
		UnsupportedProps: map[string]bool{
			"LocalizedControlType": true,
			"Value":                true,
			"AutomationId":         true,
			"BoundingRectangle":    true,
			"HelpText":             true,
			"AccessKey":            true,
			"AcceleratorKey":       true,
			"HasKeyboardFocus":     true,
			"IsKeyboardFocusable":  true,
			"ItemType":             true,
			"ItemStatus":           true,
			"ProcessId":            true,
			"IsPassword":           true,
			"IsOffscreen":          true,
			"IsRequiredForForm":    true,
			"LabeledBy":            true,
		},
		SupportedPatterns: []string{},
	}, nil
}

func parseWindowTreeRef(ref string) (window.HWND, error) {
	parsed, err := parseNodeRef(ref)
	if err != nil || parsed.Provider != nodeRefProviderWin {
		return 0, ErrInvalidNodeRef
	}
	return parseHWND(parsed.ID)
}

func getWindowText(hwnd window.HWND) (string, error) {
	l, _, _ := procGetWindowTextLengthW.Call(uintptr(hwnd))
	buf := make([]uint16, l+1)
	_, _, err := procGetWindowTextW.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	if err != windows.ERROR_SUCCESS && err != nil {
		return "", err
	}
	return windows.UTF16ToString(buf), nil
}

func getClassName(hwnd window.HWND) (string, error) {
	buf := make([]uint16, 256)
	ret, _, err := procGetClassNameW.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	if ret == 0 {
		if err != windows.ERROR_SUCCESS && err != nil {
			return "", err
		}
		return "", nil
	}
	return windows.UTF16ToString(buf[:ret]), nil
}
