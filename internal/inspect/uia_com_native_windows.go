//go:build windows
// +build windows

package inspect

import (
	"errors"
	"fmt"
	"syscall"
	"unsafe"

	"goahk/internal/window"
)

const (
	uiaTreeScopeChildren    = 0x2
	uiaPropertyRuntimeID    = 30000
	uiaEElementNotAvailable = 0x80040201
	uiaEElementNotEnabled   = 0x80040200
	eAccessDenied           = 0x80070005
	coEObjNotConnected      = 0x800401FD
	rpcEServerUnavailable   = 0x800706BA
)

func comRelease(ptr uintptr) {
	if ptr == 0 {
		return
	}
	vt := *(*uintptr)(unsafe.Pointer(ptr))
	_, _, _ = syscall.SyscallN(*(*uintptr)(unsafe.Pointer(vt + 2*unsafe.Sizeof(uintptr(0)))), ptr)
}

func comAddRef(ptr uintptr) {
	if ptr == 0 {
		return
	}
	vt := *(*uintptr)(unsafe.Pointer(ptr))
	_, _, _ = syscall.SyscallN(*(*uintptr)(unsafe.Pointer(vt + 1*unsafe.Sizeof(uintptr(0)))), ptr)
}

func hresultErr(op string, hr uintptr) error {
	if int32(hr) >= 0 {
		return nil
	}
	code := uint32(hr)
	switch code {
	case uiaEElementNotAvailable, coEObjNotConnected:
		return &UIAElementStaleError{Op: op, Err: fmt.Errorf("element is stale or unavailable (hr=0x%x)", code)}
	case eAccessDenied:
		return &UIAComUnavailableError{Op: op, Err: fmt.Errorf("access denied; run the target app with matching privileges (hr=0x%x)", code)}
	case rpcEServerUnavailable:
		return &UIAComUnavailableError{Op: op, Err: fmt.Errorf("UI Automation server unavailable (hr=0x%x)", code)}
	default:
		return &UIAComUnavailableError{Op: op, Err: fmt.Errorf("COM call failed (hr=0x%x)", code)}
	}
}

func uiaCreateTrueCondition(automation uintptr) (uintptr, error) {
	var cond uintptr
	vt := *(*uintptr)(unsafe.Pointer(automation))
	hr, _, _ := syscall.SyscallN(*(*uintptr)(unsafe.Pointer(vt + 23*unsafe.Sizeof(uintptr(0)))), automation, uintptr(unsafe.Pointer(&cond)))
	return cond, hresultErr("CreateTrueCondition", hr)
}

func uiaGetRawViewWalker(automation uintptr) (uintptr, error) {
	var walker uintptr
	vt := *(*uintptr)(unsafe.Pointer(automation))
	hr, _, _ := syscall.SyscallN(*(*uintptr)(unsafe.Pointer(vt + 16*unsafe.Sizeof(uintptr(0)))), automation, uintptr(unsafe.Pointer(&walker)))
	return walker, hresultErr("get_RawViewWalker", hr)
}

func uiaElementFromHandle(automation uintptr, hwnd window.HWND) (uintptr, error) {
	var el uintptr
	vt := *(*uintptr)(unsafe.Pointer(automation))
	hr, _, _ := syscall.SyscallN(*(*uintptr)(unsafe.Pointer(vt + 6*unsafe.Sizeof(uintptr(0)))), automation, uintptr(hwnd), uintptr(unsafe.Pointer(&el)))
	return el, hresultErr("ElementFromHandle", hr)
}

func uiaGetFocusedElement(automation uintptr) (uintptr, error) {
	var el uintptr
	vt := *(*uintptr)(unsafe.Pointer(automation))
	hr, _, _ := syscall.SyscallN(*(*uintptr)(unsafe.Pointer(vt + 8*unsafe.Sizeof(uintptr(0)))), automation, uintptr(unsafe.Pointer(&el)))
	return el, hresultErr("GetFocusedElement", hr)
}

func uiaElementFromPoint(automation uintptr, x, y int) (uintptr, error) {
	var el uintptr
	pt := struct{ X, Y int32 }{X: int32(x), Y: int32(y)}
	vt := *(*uintptr)(unsafe.Pointer(automation))
	hr, _, _ := syscall.SyscallN(*(*uintptr)(unsafe.Pointer(vt + 7*unsafe.Sizeof(uintptr(0)))), automation, *(*uintptr)(unsafe.Pointer(&pt)), uintptr(unsafe.Pointer(&el)))
	return el, hresultErr("ElementFromPoint", hr)
}

func uiaFindAllChildren(el, trueCond uintptr) (uintptr, error) {
	var arr uintptr
	vt := *(*uintptr)(unsafe.Pointer(el))
	hr, _, _ := syscall.SyscallN(*(*uintptr)(unsafe.Pointer(vt + 6*unsafe.Sizeof(uintptr(0)))), el, uintptr(uiaTreeScopeChildren), trueCond, uintptr(unsafe.Pointer(&arr)))
	return arr, hresultErr("FindAll", hr)
}

func uiaArrayLength(arr uintptr) (int32, error) {
	var ln int32
	vt := *(*uintptr)(unsafe.Pointer(arr))
	hr, _, _ := syscall.SyscallN(*(*uintptr)(unsafe.Pointer(vt + 3*unsafe.Sizeof(uintptr(0)))), arr, uintptr(unsafe.Pointer(&ln)))
	return ln, hresultErr("IUIAutomationElementArray.Length", hr)
}

func uiaArrayGet(arr uintptr, idx int32) (uintptr, error) {
	var el uintptr
	vt := *(*uintptr)(unsafe.Pointer(arr))
	hr, _, _ := syscall.SyscallN(*(*uintptr)(unsafe.Pointer(vt + 4*unsafe.Sizeof(uintptr(0)))), arr, uintptr(idx), uintptr(unsafe.Pointer(&el)))
	return el, hresultErr("IUIAutomationElementArray.GetElement", hr)
}

func uiaGetParentElement(walker, el uintptr) (uintptr, error) {
	var parent uintptr
	vt := *(*uintptr)(unsafe.Pointer(walker))
	hr, _, _ := syscall.SyscallN(*(*uintptr)(unsafe.Pointer(vt + 3*unsafe.Sizeof(uintptr(0)))), walker, el, uintptr(unsafe.Pointer(&parent)))
	return parent, hresultErr("GetParentElement", hr)
}

func uiaElementRuntimeID(el uintptr) (string, error) {
	var arr uintptr
	vt := *(*uintptr)(unsafe.Pointer(el))
	hr, _, _ := syscall.SyscallN(*(*uintptr)(unsafe.Pointer(vt + 11*unsafe.Sizeof(uintptr(0)))), el, uintptr(uiaPropertyRuntimeID), uintptr(unsafe.Pointer(&arr)))
	if err := hresultErr("GetCurrentPropertyValue(RuntimeID)", hr); err != nil {
		var stale *UIAElementStaleError
		if errors.As(err, &stale) {
			return "", err
		}
		return "", nil
	}
	if arr == 0 {
		return "", nil
	}
	defer syscall.SyscallN(procSafeArrayDestroy.Addr(), arr)
	lb, ub := int32(0), int32(0)
	if hr, _, _ := syscall.SyscallN(procSafeArrayGetLBound.Addr(), arr, 1, uintptr(unsafe.Pointer(&lb))); int32(hr) < 0 {
		return "", hresultErr("SafeArrayGetLBound(RuntimeID)", hr)
	}
	if hr, _, _ := syscall.SyscallN(procSafeArrayGetUBound.Addr(), arr, 1, uintptr(unsafe.Pointer(&ub))); int32(hr) < 0 {
		return "", hresultErr("SafeArrayGetUBound(RuntimeID)", hr)
	}
	if ub < lb {
		return "", nil
	}
	runtimeID := make([]int, 0, ub-lb+1)
	for i := lb; i <= ub; i++ {
		v := int32(0)
		iCopy := i
		hr, _, _ := syscall.SyscallN(procSafeArrayGetElement.Addr(), arr, uintptr(unsafe.Pointer(&iCopy)), uintptr(unsafe.Pointer(&v)))
		if int32(hr) < 0 {
			return "", hresultErr("SafeArrayGetElement(RuntimeID)", hr)
		}
		runtimeID = append(runtimeID, int(v))
	}
	if len(runtimeID) == 0 {
		return "", nil
	}
	return runtimeIDString(runtimeID), nil
}
