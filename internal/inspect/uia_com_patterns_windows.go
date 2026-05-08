//go:build windows
// +build windows

package inspect

import (
	"syscall"
	"unsafe"
)

const (
	uiaPatternInvoke            = 10000
	uiaPatternSelectionItem     = 10010
	uiaPatternValue             = 10002
	uiaPatternLegacyIAccessible = 10018
	uiaPatternToggle            = 10015
	uiaPatternExpandCollapse    = 10005
)

func uiaInvokePatternInvoke(el uintptr) error {
	pattern, err := uiaGetCurrentPatternPtr(el, uiaPatternInvoke, "Invoke")
	if err != nil {
		return err
	}
	defer comRelease(pattern)
	vt := *(*uintptr)(unsafe.Pointer(pattern))
	hr, _, _ := syscall.SyscallN(*(*uintptr)(unsafe.Pointer(vt + 3*unsafe.Sizeof(uintptr(0)))), pattern)
	return hresultErr("Invoke", hr)
}

func uiaSelectionItemPatternSelect(el uintptr) error {
	pattern, err := uiaGetCurrentPatternPtr(el, uiaPatternSelectionItem, "Select")
	if err != nil {
		return err
	}
	defer comRelease(pattern)
	vt := *(*uintptr)(unsafe.Pointer(pattern))
	hr, _, _ := syscall.SyscallN(*(*uintptr)(unsafe.Pointer(vt + 3*unsafe.Sizeof(uintptr(0)))), pattern)
	return hresultErr("Select", hr)
}

func uiaValuePatternSetValue(el uintptr, value string) error {
	pattern, err := uiaGetCurrentPatternPtr(el, uiaPatternValue, "SetValue")
	if err != nil {
		return err
	}
	defer comRelease(pattern)
	bstr, convErr := syscall.UTF16PtrFromString(value)
	if convErr != nil {
		return &UIAComUnavailableError{Op: "SetValue", Err: convErr}
	}
	vt := *(*uintptr)(unsafe.Pointer(pattern))
	hr, _, _ := syscall.SyscallN(*(*uintptr)(unsafe.Pointer(vt + 4*unsafe.Sizeof(uintptr(0)))), pattern, uintptr(unsafe.Pointer(bstr)))
	return hresultErr("SetValue", hr)
}

func uiaLegacyIAccessiblePatternDoDefaultAction(el uintptr) error {
	pattern, err := uiaGetCurrentPatternPtr(el, uiaPatternLegacyIAccessible, "DoDefaultAction")
	if err != nil {
		return err
	}
	defer comRelease(pattern)
	vt := *(*uintptr)(unsafe.Pointer(pattern))
	hr, _, _ := syscall.SyscallN(*(*uintptr)(unsafe.Pointer(vt + 18*unsafe.Sizeof(uintptr(0)))), pattern)
	return hresultErr("DoDefaultAction", hr)
}

func uiaTogglePatternToggle(el uintptr) error {
	pattern, err := uiaGetCurrentPatternPtr(el, uiaPatternToggle, "Toggle")
	if err != nil {
		return err
	}
	defer comRelease(pattern)
	vt := *(*uintptr)(unsafe.Pointer(pattern))
	hr, _, _ := syscall.SyscallN(*(*uintptr)(unsafe.Pointer(vt + 3*unsafe.Sizeof(uintptr(0)))), pattern)
	return hresultErr("Toggle", hr)
}

func uiaExpandCollapsePatternExpand(el uintptr) error {
	pattern, err := uiaGetCurrentPatternPtr(el, uiaPatternExpandCollapse, "Expand")
	if err != nil {
		return err
	}
	defer comRelease(pattern)
	vt := *(*uintptr)(unsafe.Pointer(pattern))
	hr, _, _ := syscall.SyscallN(*(*uintptr)(unsafe.Pointer(vt + 3*unsafe.Sizeof(uintptr(0)))), pattern)
	return hresultErr("Expand", hr)
}

func uiaExpandCollapsePatternCollapse(el uintptr) error {
	pattern, err := uiaGetCurrentPatternPtr(el, uiaPatternExpandCollapse, "Collapse")
	if err != nil {
		return err
	}
	defer comRelease(pattern)
	vt := *(*uintptr)(unsafe.Pointer(pattern))
	hr, _, _ := syscall.SyscallN(*(*uintptr)(unsafe.Pointer(vt + 4*unsafe.Sizeof(uintptr(0)))), pattern)
	return hresultErr("Collapse", hr)
}

func uiaGetCurrentPatternPtr(el uintptr, patternID int32, op string) (uintptr, error) {
	var p uintptr
	vt := *(*uintptr)(unsafe.Pointer(el))
	hr, _, _ := syscall.SyscallN(*(*uintptr)(unsafe.Pointer(vt + 12*unsafe.Sizeof(uintptr(0)))), el, uintptr(patternID), uintptr(unsafe.Pointer(&p)))
	if err := hresultErr(op, hr); err != nil {
		return 0, err
	}
	if p == 0 {
		return 0, &UIAComUnavailableError{Op: op, Err: syscall.ENOTSUP}
	}
	return p, nil
}
