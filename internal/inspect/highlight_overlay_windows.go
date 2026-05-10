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

const (
	wmDestroy = 0x0002
	wmPaint   = 0x000F
)

const (
	swHide = 0
	swShow = 5
)

const (
	swpNoActivate = 0x0010
	swpShowWindow = 0x0040
)

const (
	psSolid     = 0
	hollowBrush = 5
)

type overlayWinRect struct {
	Left   int32
	Top    int32
	Right  int32
	Bottom int32
}

type paintStruct struct {
	hdc         uintptr
	fErase      int32
	rcPaint     overlayWinRect
	fRestore    int32
	fIncUpdate  int32
	rgbReserved [32]byte
}

type wndClassEx struct {
	cbSize        uint32
	style         uint32
	lpfnWndProc   uintptr
	cbClsExtra    int32
	cbWndExtra    int32
	hInstance     uintptr
	hIcon         uintptr
	hCursor       uintptr
	hbrBackground uintptr
	lpszMenuName  *uint16
	lpszClassName *uint16
	hIconSm       uintptr
}

var (
	overlayUser32           = syscall.NewLazyDLL("user32.dll")
	overlayGDI32            = syscall.NewLazyDLL("gdi32.dll")
	procGetSystemMetricsHUD = overlayUser32.NewProc("GetSystemMetrics")
	procRegisterClassExW    = overlayUser32.NewProc("RegisterClassExW")
	procCreateWindowExW     = overlayUser32.NewProc("CreateWindowExW")
	procDefWindowProcW      = overlayUser32.NewProc("DefWindowProcW")
	procDestroyWindow       = overlayUser32.NewProc("DestroyWindow")
	procShowWindow          = overlayUser32.NewProc("ShowWindow")
	procSetWindowPos        = overlayUser32.NewProc("SetWindowPos")
	procInvalidateRect      = overlayUser32.NewProc("InvalidateRect")
	procBeginPaint          = overlayUser32.NewProc("BeginPaint")
	procEndPaint            = overlayUser32.NewProc("EndPaint")
	procGetStockObject      = overlayGDI32.NewProc("GetStockObject")
	procSelectObject        = overlayGDI32.NewProc("SelectObject")
	procCreatePen           = overlayGDI32.NewProc("CreatePen")
	procDeleteObject        = overlayGDI32.NewProc("DeleteObject")
	procRectangle           = overlayGDI32.NewProc("Rectangle")
	overlayClassName        = syscall.StringToUTF16Ptr("goahk.inspect.highlight.overlay")
	overlayWindowProc       = syscall.NewCallback(overlayWndProc)
)

type nativeHighlightOverlay struct {
	mu   sync.Mutex
	hwnd uintptr
	rect Rect
}

func newNativeHighlightOverlay() highlightOverlay { return &nativeHighlightOverlay{} }

func (o *nativeHighlightOverlay) Show(ctx context.Context, rect Rect) error {
	screen, err := o.ScreenBounds(ctx)
	if err != nil {
		return err
	}
	normalized, ok := normalizeHighlightRect(&rect, false, screen)
	if !ok {
		return o.Clear(ctx)
	}

	o.mu.Lock()
	defer o.mu.Unlock()

	hwnd, err := o.ensureWindowLocked()
	if err != nil {
		return err
	}

	ret, _, setPosErr := procSetWindowPos.Call(
		hwnd,
		uintptr(^uintptr(1)+1), // HWND_TOPMOST
		uintptr(int32(normalized.Left)),
		uintptr(int32(normalized.Top)),
		uintptr(int32(normalized.Width)),
		uintptr(int32(normalized.Height)),
		swpNoActivate|swpShowWindow,
	)
	if ret == 0 {
		if setPosErr != syscall.Errno(0) {
			return fmt.Errorf("SetWindowPos: %w", setPosErr)
		}
		return fmt.Errorf("SetWindowPos: failed")
	}
	procShowWindow.Call(hwnd, swShow)
	procInvalidateRect.Call(hwnd, 0, 1)
	o.rect = normalized
	return nil
}

func (o *nativeHighlightOverlay) Clear(context.Context) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.hwnd != 0 {
		procShowWindow.Call(o.hwnd, swHide)
	}
	o.rect = Rect{}
	return nil
}

const (
	wsPopup         = 0x80000000
	wsExTopMost     = 0x00000008
	wsExToolWindow  = 0x00000080
	wsExTransparent = 0x00000020
	wsExLayered     = 0x00080000
	wsExNoActivate  = 0x08000000
)

func overlayWindowStyle() uintptr {
	return wsPopup
}

func overlayWindowExStyle() uintptr {
	return wsExTopMost | wsExToolWindow | wsExTransparent | wsExLayered | wsExNoActivate
}

func overlayPaintUsesBorderOnly() bool {
	return true
}

func (o *nativeHighlightOverlay) ScreenBounds(context.Context) (*Rect, error) {
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

func (o *nativeHighlightOverlay) ensureWindowLocked() (uintptr, error) {
	if o.hwnd != 0 {
		return o.hwnd, nil
	}
	wc := wndClassEx{cbSize: uint32(unsafe.Sizeof(wndClassEx{})), lpfnWndProc: overlayWindowProc, lpszClassName: overlayClassName}
	ret, _, regErr := procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))
	if ret == 0 && regErr != syscall.Errno(1410) { // already exists
		if regErr != syscall.Errno(0) {
			return 0, fmt.Errorf("RegisterClassExW: %w", regErr)
		}
		return 0, fmt.Errorf("RegisterClassExW: failed")
	}
	hwnd, _, createErr := procCreateWindowExW.Call(
		overlayWindowExStyle(),
		uintptr(unsafe.Pointer(overlayClassName)),
		0,
		overlayWindowStyle(),
		0, 0, 0, 0,
		0, 0, 0, 0,
	)
	if hwnd == 0 {
		if createErr != syscall.Errno(0) {
			return 0, fmt.Errorf("CreateWindowExW: %w", createErr)
		}
		return 0, fmt.Errorf("CreateWindowExW: failed")
	}
	o.hwnd = hwnd
	return hwnd, nil
}

func overlayWndProc(hwnd, msg, wParam, lParam uintptr) uintptr {
	switch msg {
	case wmPaint:
		var ps paintStruct
		hdc, _, _ := procBeginPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
		if hdc != 0 {
			pen, _, _ := procCreatePen.Call(psSolid, 3, uintptr(0x00D77800))
			if pen != 0 {
				oldPen, _, _ := procSelectObject.Call(hdc, pen)
				brush, _, _ := procGetStockObject.Call(hollowBrush)
				var oldBrush uintptr
				if brush != 0 {
					oldBrush, _, _ = procSelectObject.Call(hdc, brush)
				}
				procRectangle.Call(hdc, 0, 0, uintptr(ps.rcPaint.Right-ps.rcPaint.Left), uintptr(ps.rcPaint.Bottom-ps.rcPaint.Top))
				if oldBrush != 0 {
					procSelectObject.Call(hdc, oldBrush)
				}
				procSelectObject.Call(hdc, oldPen)
				procDeleteObject.Call(pen)
			}
			procEndPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
		}
		return 0
	case wmDestroy:
		return 0
	default:
		ret, _, _ := procDefWindowProcW.Call(hwnd, msg, wParam, lParam)
		return ret
	}
}
