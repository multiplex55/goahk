//go:build windows
// +build windows

package window

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"syscall"
	"unsafe"
)

const (
	processQueryLimitedInformation = 0x1000
)

var (
	windowUser32                   = syscall.NewLazyDLL("user32.dll")
	windowKernel32                 = syscall.NewLazyDLL("kernel32.dll")
	procEnumWindows                = windowUser32.NewProc("EnumWindows")
	procGetForegroundWindow        = windowUser32.NewProc("GetForegroundWindow")
	procSetForegroundWindow        = windowUser32.NewProc("SetForegroundWindow")
	procIsWindowVisible            = windowUser32.NewProc("IsWindowVisible")
	procIsIconic                   = windowUser32.NewProc("IsIconic")
	procIsZoomed                   = windowUser32.NewProc("IsZoomed")
	procGetWindowTextW             = windowUser32.NewProc("GetWindowTextW")
	procGetWindowTextLengthW       = windowUser32.NewProc("GetWindowTextLengthW")
	procGetClassNameW              = windowUser32.NewProc("GetClassNameW")
	procGetWindowThreadProcessID   = windowUser32.NewProc("GetWindowThreadProcessId")
	procGetWindowRect              = windowUser32.NewProc("GetWindowRect")
	procSetWindowPos               = windowUser32.NewProc("SetWindowPos")
	procShowWindow                 = windowUser32.NewProc("ShowWindow")
	procMonitorFromWindow          = windowUser32.NewProc("MonitorFromWindow")
	procGetMonitorInfoW            = windowUser32.NewProc("GetMonitorInfoW")
	procOpenProcess                = windowKernel32.NewProc("OpenProcess")
	procCloseHandle                = windowKernel32.NewProc("CloseHandle")
	procQueryFullProcessImageNameW = windowKernel32.NewProc("QueryFullProcessImageNameW")
	errWindowEnumAborted           = errors.New("window enumeration aborted")
)

const (
	swpNoZOrder             = 0x0004
	swpNoSize               = 0x0001
	swpNoMove               = 0x0002
	swMinimize              = 6
	swMaximize              = 3
	swRestore               = 9
	monitorDefaultToNearest = 0x00000002
)

type rect struct {
	Left   int32
	Top    int32
	Right  int32
	Bottom int32
}

type monitorInfo struct {
	CbSize    uint32
	RcMonitor rect
	RcWork    rect
	DwFlags   uint32
}

type OSProvider struct{}

func (p *OSProvider) EnumerateWindows(ctx context.Context) ([]Info, error) {
	activeHWND, _ := getForegroundWindow()
	out := make([]Info, 0, 32)
	cb := syscall.NewCallback(func(hwnd uintptr, _ uintptr) uintptr {
		select {
		case <-ctx.Done():
			return 0
		default:
		}
		visible, err := isWindowVisible(HWND(hwnd))
		if err != nil || !visible {
			return 1
		}
		info, err := p.readInfo(HWND(hwnd), activeHWND)
		if err != nil {
			return 1
		}
		out = append(out, info)
		return 1
	})
	ret, _, callErr := procEnumWindows.Call(cb, 0)
	if ret == 0 {
		if errors.Is(ctx.Err(), context.Canceled) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, ctx.Err()
		}
		if callErr != syscall.Errno(0) {
			return nil, fmt.Errorf("EnumWindows: %w", callErr)
		}
		return nil, errWindowEnumAborted
	}
	return out, nil
}

func (p *OSProvider) ActiveWindow(ctx context.Context) (Info, error) {
	select {
	case <-ctx.Done():
		return Info{}, ctx.Err()
	default:
	}
	hwnd, err := getForegroundWindow()
	if err != nil {
		return Info{}, err
	}
	return p.readInfo(hwnd, hwnd)
}

func (p *OSProvider) ActivateWindow(ctx context.Context, hwnd HWND) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	ret, _, err := procSetForegroundWindow.Call(uintptr(hwnd))
	if ret == 0 {
		if err != syscall.Errno(0) {
			return fmt.Errorf("SetForegroundWindow(%s): %w", hwnd, err)
		}
		return fmt.Errorf("SetForegroundWindow(%s): failed", hwnd)
	}
	return nil
}

func (p *OSProvider) WindowBounds(ctx context.Context, hwnd HWND) (Rect, error) {
	select {
	case <-ctx.Done():
		return Rect{}, ctx.Err()
	default:
	}
	bounds, err := getWindowBounds(hwnd)
	if err != nil {
		return Rect{}, err
	}
	return *bounds, nil
}

func (p *OSProvider) WorkAreaForWindow(ctx context.Context, hwnd HWND) (Rect, error) {
	select {
	case <-ctx.Done():
		return Rect{}, ctx.Err()
	default:
	}
	monitor, _, monitorErr := procMonitorFromWindow.Call(uintptr(hwnd), uintptr(monitorDefaultToNearest))
	if monitor == 0 {
		if monitorErr != syscall.Errno(0) {
			return Rect{}, fmt.Errorf("MonitorFromWindow(%s): %w", hwnd, monitorErr)
		}
		return Rect{}, fmt.Errorf("MonitorFromWindow(%s): failed", hwnd)
	}
	info := monitorInfo{CbSize: uint32(unsafe.Sizeof(monitorInfo{}))}
	ret, _, err := procGetMonitorInfoW.Call(monitor, uintptr(unsafe.Pointer(&info)))
	if ret == 0 {
		if err != syscall.Errno(0) {
			return Rect{}, fmt.Errorf("GetMonitorInfoW(%s): %w", hwnd, err)
		}
		return Rect{}, fmt.Errorf("GetMonitorInfoW(%s): failed", hwnd)
	}
	return Rect{
		Left:   int(info.RcWork.Left),
		Top:    int(info.RcWork.Top),
		Right:  int(info.RcWork.Right),
		Bottom: int(info.RcWork.Bottom),
	}, nil
}

func (p *OSProvider) MoveWindow(ctx context.Context, hwnd HWND, x, y int) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	ret, _, err := procSetWindowPos.Call(uintptr(hwnd), 0, uintptr(x), uintptr(y), 0, 0, swpNoSize|swpNoZOrder)
	if ret == 0 {
		if err != syscall.Errno(0) {
			return fmt.Errorf("SetWindowPos(move %s): %w", hwnd, err)
		}
		return fmt.Errorf("SetWindowPos(move %s): failed", hwnd)
	}
	return nil
}

func (p *OSProvider) ResizeWindow(ctx context.Context, hwnd HWND, width, height int) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	ret, _, err := procSetWindowPos.Call(uintptr(hwnd), 0, 0, 0, uintptr(width), uintptr(height), swpNoMove|swpNoZOrder)
	if ret == 0 {
		if err != syscall.Errno(0) {
			return fmt.Errorf("SetWindowPos(resize %s): %w", hwnd, err)
		}
		return fmt.Errorf("SetWindowPos(resize %s): failed", hwnd)
	}
	return nil
}

func (p *OSProvider) MinimizeWindow(ctx context.Context, hwnd HWND) error {
	return p.showWindow(ctx, hwnd, swMinimize, "minimize")
}

func (p *OSProvider) MaximizeWindow(ctx context.Context, hwnd HWND) error {
	return p.showWindow(ctx, hwnd, swMaximize, "maximize")
}

func (p *OSProvider) RestoreWindow(ctx context.Context, hwnd HWND) error {
	return p.showWindow(ctx, hwnd, swRestore, "restore")
}

func (p *OSProvider) showWindow(ctx context.Context, hwnd HWND, cmd int, action string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	ret, _, err := procShowWindow.Call(uintptr(hwnd), uintptr(cmd))
	if ret == 0 && err != syscall.Errno(0) {
		return fmt.Errorf("ShowWindow(%s %s): %w", action, hwnd, err)
	}
	return nil
}

func (p *OSProvider) readInfo(hwnd, active HWND) (Info, error) {
	title, err := getWindowText(hwnd)
	if err != nil {
		return Info{}, err
	}
	className, err := getClassName(hwnd)
	if err != nil {
		return Info{}, err
	}
	pid, err := getWindowPID(hwnd)
	if err != nil {
		return Info{}, err
	}
	path, pathStatus, pathErr := getProcessImagePath(pid)
	exe := ""
	if path != "" {
		exe = filepath.Base(path)
	}
	bounds, err := getWindowBounds(hwnd)
	if err != nil {
		bounds = nil
	}
	visible, _ := isWindowVisible(hwnd)
	minimized, _ := isWindowMinimized(hwnd)
	maximized, _ := isWindowMaximized(hwnd)
	return Info{
		HWND:              hwnd,
		Title:             title,
		Class:             className,
		PID:               pid,
		Exe:               exe,
		ProcessPath:       path,
		ProcessPathStatus: pathStatus,
		ProcessPathError:  pathErr,
		Active:            hwnd == active,
		Visible:           boolPtr(visible),
		Minimized:         boolPtr(minimized),
		Maximized:         boolPtr(maximized),
		Rect:              bounds,
	}, nil
}

func getWindowBounds(hwnd HWND) (*Rect, error) {
	var r rect
	ret, _, err := procGetWindowRect.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&r)))
	if ret == 0 {
		if err != syscall.Errno(0) {
			return nil, fmt.Errorf("GetWindowRect(%s): %w", hwnd, err)
		}
		return nil, fmt.Errorf("GetWindowRect(%s): failed", hwnd)
	}
	bounds := &Rect{Left: int(r.Left), Top: int(r.Top), Right: int(r.Right), Bottom: int(r.Bottom)}
	return bounds, nil
}

func getForegroundWindow() (HWND, error) {
	ret, _, err := procGetForegroundWindow.Call()
	if ret == 0 {
		if err != syscall.Errno(0) {
			return 0, fmt.Errorf("GetForegroundWindow: %w", err)
		}
		return 0, errors.New("GetForegroundWindow returned null")
	}
	return HWND(ret), nil
}

func isWindowVisible(hwnd HWND) (bool, error) {
	ret, _, err := procIsWindowVisible.Call(uintptr(hwnd))
	if ret == 0 {
		if err != syscall.Errno(0) {
			return false, fmt.Errorf("IsWindowVisible(%s): %w", hwnd, err)
		}
		return false, nil
	}
	return true, nil
}

func isWindowMinimized(hwnd HWND) (bool, error) {
	ret, _, err := procIsIconic.Call(uintptr(hwnd))
	if ret == 0 {
		if err != syscall.Errno(0) {
			return false, fmt.Errorf("IsIconic(%s): %w", hwnd, err)
		}
		return false, nil
	}
	return true, nil
}

func isWindowMaximized(hwnd HWND) (bool, error) {
	ret, _, err := procIsZoomed.Call(uintptr(hwnd))
	if ret == 0 {
		if err != syscall.Errno(0) {
			return false, fmt.Errorf("IsZoomed(%s): %w", hwnd, err)
		}
		return false, nil
	}
	return true, nil
}

func getWindowText(hwnd HWND) (string, error) {
	lengthRet, _, lengthErr := procGetWindowTextLengthW.Call(uintptr(hwnd))
	if lengthRet == 0 && lengthErr != syscall.Errno(0) {
		return "", fmt.Errorf("GetWindowTextLengthW(%s): %w", hwnd, lengthErr)
	}
	buf := make([]uint16, int(lengthRet)+1)
	ret, _, err := procGetWindowTextW.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	if ret == 0 && err != syscall.Errno(0) {
		return "", fmt.Errorf("GetWindowTextW(%s): %w", hwnd, err)
	}
	return syscall.UTF16ToString(buf), nil
}

func getClassName(hwnd HWND) (string, error) {
	buf := make([]uint16, 256)
	ret, _, err := procGetClassNameW.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	if ret == 0 {
		if err != syscall.Errno(0) {
			return "", fmt.Errorf("GetClassNameW(%s): %w", hwnd, err)
		}
		return "", nil
	}
	return syscall.UTF16ToString(buf[:ret]), nil
}

func getWindowPID(hwnd HWND) (uint32, error) {
	var pid uint32
	_, _, err := procGetWindowThreadProcessID.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&pid)))
	if pid == 0 {
		if err != syscall.Errno(0) {
			return 0, fmt.Errorf("GetWindowThreadProcessId(%s): %w", hwnd, err)
		}
		return 0, fmt.Errorf("GetWindowThreadProcessId(%s): missing pid", hwnd)
	}
	return pid, nil
}

func getProcessImagePath(pid uint32) (path string, status string, errText string) {
	h, _, err := procOpenProcess.Call(processQueryLimitedInformation, 0, uintptr(pid))
	if h == 0 {
		if err != syscall.Errno(0) {
			return "", "open_failed", fmt.Sprintf("OpenProcess(%d): %v", pid, err)
		}
		return "", "open_failed", fmt.Sprintf("OpenProcess(%d): failed", pid)
	}
	defer procCloseHandle.Call(h)

	buf := make([]uint16, syscall.MAX_PATH)
	size := uint32(len(buf))
	ret, _, qErr := procQueryFullProcessImageNameW.Call(h, 0, uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&size)))
	if ret == 0 {
		if qErr != syscall.Errno(0) {
			return "", "query_failed", fmt.Sprintf("QueryFullProcessImageNameW(%d): %v", pid, qErr)
		}
		return "", "query_failed", fmt.Sprintf("QueryFullProcessImageNameW(%d): failed", pid)
	}
	full := syscall.UTF16ToString(buf[:size])
	return full, "ok", ""
}

func boolPtr(v bool) *bool {
	return &v
}
