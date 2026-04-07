//go:build windows

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
	user32                         = syscall.NewLazyDLL("user32.dll")
	kernel32                       = syscall.NewLazyDLL("kernel32.dll")
	procEnumWindows                = user32.NewProc("EnumWindows")
	procGetForegroundWindow        = user32.NewProc("GetForegroundWindow")
	procSetForegroundWindow        = user32.NewProc("SetForegroundWindow")
	procIsWindowVisible            = user32.NewProc("IsWindowVisible")
	procGetWindowTextW             = user32.NewProc("GetWindowTextW")
	procGetWindowTextLengthW       = user32.NewProc("GetWindowTextLengthW")
	procGetClassNameW              = user32.NewProc("GetClassNameW")
	procGetWindowThreadProcessID   = user32.NewProc("GetWindowThreadProcessId")
	procOpenProcess                = kernel32.NewProc("OpenProcess")
	procCloseHandle                = kernel32.NewProc("CloseHandle")
	procQueryFullProcessImageNameW = kernel32.NewProc("QueryFullProcessImageNameW")
	errEnumAborted                 = errors.New("window enumeration aborted")
)

type OSProvider struct{}

func NewOSProvider() *OSProvider { return &OSProvider{} }

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
		return nil, errEnumAborted
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
	exe, err := getProcessExeBaseName(pid)
	if err != nil {
		exe = ""
	}
	return Info{
		HWND:   hwnd,
		Title:  title,
		Class:  className,
		PID:    pid,
		Exe:    exe,
		Active: hwnd == active,
	}, nil
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

func getProcessExeBaseName(pid uint32) (string, error) {
	h, _, err := procOpenProcess.Call(processQueryLimitedInformation, 0, uintptr(pid))
	if h == 0 {
		if err != syscall.Errno(0) {
			return "", fmt.Errorf("OpenProcess(%d): %w", pid, err)
		}
		return "", fmt.Errorf("OpenProcess(%d): failed", pid)
	}
	defer procCloseHandle.Call(h)

	buf := make([]uint16, syscall.MAX_PATH)
	size := uint32(len(buf))
	ret, _, qErr := procQueryFullProcessImageNameW.Call(h, 0, uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&size)))
	if ret == 0 {
		if qErr != syscall.Errno(0) {
			return "", fmt.Errorf("QueryFullProcessImageNameW(%d): %w", pid, qErr)
		}
		return "", fmt.Errorf("QueryFullProcessImageNameW(%d): failed", pid)
	}
	full := syscall.UTF16ToString(buf[:size])
	return filepath.Base(full), nil
}
