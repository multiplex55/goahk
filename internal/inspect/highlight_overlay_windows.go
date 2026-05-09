//go:build windows
// +build windows

package inspect

import (
	"context"
	"fmt"
	"sync"
	"syscall"
	"unsafe"
)

const (
	smXVIRTUALSCREEN  = 76
	smYVIRTUALSCREEN  = 77
	smCXVIRTUALSCREEN = 78
	smCYVIRTUALSCREEN = 79
)

var (
	overlayUser32                  = syscall.NewLazyDLL("user32.dll")
	overlayKernel32                = syscall.NewLazyDLL("kernel32.dll")
	procGetSystemMetricsHUD        = overlayUser32.NewProc("GetSystemMetrics")
	procCreateWindowExW            = overlayUser32.NewProc("CreateWindowExW")
	procDestroyWindow              = overlayUser32.NewProc("DestroyWindow")
	procShowWindow                 = overlayUser32.NewProc("ShowWindow")
	procSetWindowPos               = overlayUser32.NewProc("SetWindowPos")
	procSetLayeredWindowAttributes = overlayUser32.NewProc("SetLayeredWindowAttributes")
	procRegisterClassExW           = overlayUser32.NewProc("RegisterClassExW")
	procGetModuleHandleW           = overlayKernel32.NewProc("GetModuleHandleW")
)

type nativeHighlightOverlay struct {
	mu      sync.Mutex
	borders [4]syscall.Handle
	visible bool
}

func newNativeHighlightOverlay() highlightOverlay { return &nativeHighlightOverlay{} }

const (
	wsPopup         = 0x80000000
	wsExTopMost     = 0x00000008
	wsExToolWindow  = 0x00000080
	wsExTransparent = 0x00000020
	wsExLayered     = 0x00080000
	wsExNoActivate  = 0x08000000
	swpNoActivate   = 0x0010
	swpShowWindow   = 0x0040
	lwaColorKey     = 0x00000001
	swHide          = 0
)

var (
	overlayClassName      = syscall.StringToUTF16Ptr("goahk.HighlightOverlay.Border")
	overlayClassOnce      sync.Once
	overlayClassRegisterE error
)

type wndClassEx struct {
	cbSize        uint32
	style         uint32
	lpfnWndProc   uintptr
	cbClsExtra    int32
	cbWndExtra    int32
	hInstance     syscall.Handle
	hIcon         syscall.Handle
	hCursor       syscall.Handle
	hbrBackground syscall.Handle
	lpszMenuName  *uint16
	lpszClassName *uint16
	hIconSm       syscall.Handle
}

func (o *nativeHighlightOverlay) Show(ctx context.Context, r Rect) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if r.Width <= 0 || r.Height <= 0 {
		return fmt.Errorf("invalid rect size width=%d height=%d", r.Width, r.Height)
	}
	if r.Width < 2 || r.Height < 2 {
		return fmt.Errorf("rect too small width=%d height=%d", r.Width, r.Height)
	}
	if _, err := o.ScreenBounds(ctx); err != nil {
		return err
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	if err := ensureOverlayClass(); err != nil {
		return err
	}
	if err := o.ensureWindowsLocked(); err != nil {
		o.clearLocked()
		return err
	}
	const d = int32(4)
	l, t, w, h := int32(r.Left), int32(r.Top), int32(r.Width), int32(r.Height)
	top := [4]int32{l, t, w, d}
	right := [4]int32{l + w - d, t, d, h}
	bottom := [4]int32{l, t + h - d, w, d}
	left := [4]int32{l, t, d, h}
	strips := [4][4]int32{top, right, bottom, left}
	for i, hwnd := range o.borders {
		s := strips[i]
		if err := setPos(hwnd, s[0], s[1], s[2], s[3]); err != nil {
			return err
		}
	}
	o.visible = true
	return nil
}

func (o *nativeHighlightOverlay) Clear(context.Context) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.clearLocked()
	return nil
}

func (o *nativeHighlightOverlay) clearLocked() {
	for i, hwnd := range o.borders {
		if hwnd == 0 {
			continue
		}
		_, _, _ = procShowWindow.Call(uintptr(hwnd), swHide)
		_, _, _ = procDestroyWindow.Call(uintptr(hwnd))
		o.borders[i] = 0
	}
	o.visible = false
}

func (o *nativeHighlightOverlay) ensureWindowsLocked() error {
	for i := range o.borders {
		if o.borders[i] != 0 {
			continue
		}
		hwnd, err := createOverlayWindow()
		if err != nil {
			return err
		}
		o.borders[i] = hwnd
	}
	return nil
}

func createOverlayWindow() (syscall.Handle, error) {
	hwnd, _, err := procCreateWindowExW.Call(
		overlayWindowExStyle(),
		uintptr(unsafe.Pointer(overlayClassName)),
		0,
		overlayWindowStyle(),
		0, 0, 1, 1,
		0, 0, 0, 0,
	)
	if hwnd == 0 {
		return 0, fmt.Errorf("CreateWindowExW: %w", err)
	}
	if _, _, err := procSetLayeredWindowAttributes.Call(hwnd, 0x000000FF, 0, lwaColorKey); err != syscall.Errno(0) {
		_, _, _ = procDestroyWindow.Call(hwnd)
		return 0, fmt.Errorf("SetLayeredWindowAttributes: %w", err)
	}
	return syscall.Handle(hwnd), nil
}

func setPos(hwnd syscall.Handle, x, y, w, h int32) error {
	r, _, err := procSetWindowPos.Call(uintptr(hwnd), ^uintptr(0), uintptr(x), uintptr(y), uintptr(w), uintptr(h), swpNoActivate|swpShowWindow)
	if r == 0 {
		return fmt.Errorf("SetWindowPos: %w", err)
	}
	return nil
}

func ensureOverlayClass() error {
	overlayClassOnce.Do(func() {
		inst, _, err := procGetModuleHandleW.Call(0)
		if inst == 0 {
			overlayClassRegisterE = fmt.Errorf("GetModuleHandleW: %w", err)
			return
		}
		wc := wndClassEx{cbSize: uint32(unsafe.Sizeof(wndClassEx{})), lpfnWndProc: syscall.NewCallback(func(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
			ret, _, _ := overlayUser32.NewProc("DefWindowProcW").Call(hwnd, uintptr(msg), wParam, lParam)
			return ret
		}), hInstance: syscall.Handle(inst), hbrBackground: 0, lpszClassName: overlayClassName}
		atom, _, err := procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))
		if atom == 0 && err != syscall.Errno(1410) {
			overlayClassRegisterE = fmt.Errorf("RegisterClassExW: %w", err)
		}
	})
	return overlayClassRegisterE
}

func overlayWindowStyle() uintptr { return wsPopup }
func overlayWindowExStyle() uintptr {
	return wsExTopMost | wsExToolWindow | wsExTransparent | wsExLayered | wsExNoActivate
}
func overlayPaintUsesBorderOnly() bool { return true }

func (nativeHighlightOverlay) ScreenBounds(context.Context) (*Rect, error) {
	left, _, err := procGetSystemMetricsHUD.Call(smXVIRTUALSCREEN)
	if err != syscall.Errno(0) {
		return nil, fmt.Errorf("GetSystemMetrics(SM_XVIRTUALSCREEN): %w", err)
	}
	top, _, err := procGetSystemMetricsHUD.Call(smYVIRTUALSCREEN)
	if err != syscall.Errno(0) {
		return nil, fmt.Errorf("GetSystemMetrics(SM_YVIRTUALSCREEN): %w", err)
	}
	width, _, err := procGetSystemMetricsHUD.Call(smCXVIRTUALSCREEN)
	if err != syscall.Errno(0) {
		return nil, fmt.Errorf("GetSystemMetrics(SM_CXVIRTUALSCREEN): %w", err)
	}
	height, _, err := procGetSystemMetricsHUD.Call(smCYVIRTUALSCREEN)
	if err != syscall.Errno(0) {
		return nil, fmt.Errorf("GetSystemMetrics(SM_CYVIRTUALSCREEN): %w", err)
	}
	if width <= 0 || height <= 0 {
		return nil, nil
	}
	return &Rect{Left: int(int32(left)), Top: int(int32(top)), Width: int(int32(width)), Height: int(int32(height))}, nil
}
