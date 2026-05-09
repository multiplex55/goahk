//go:build windows
// +build windows

package inspect

import (
	"fmt"
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

const (
	uiaVTableIUIAutomationInvokePatternInvoke                     = 3
	uiaVTableIUIAutomationSelectionItemPatternSelect              = 3
	uiaVTableIUIAutomationValuePatternSetValue                    = 4
	uiaVTableIUIAutomationLegacyIAccessiblePatternDoDefaultAction = 18
	uiaVTableIUIAutomationTogglePatternToggle                     = 3
	uiaVTableIUIAutomationExpandCollapsePatternExpand             = 3
	uiaVTableIUIAutomationExpandCollapsePatternCollapse           = 4
)

func uiaInvokePatternInvoke(el uintptr) error {
	pattern, err := uiaGetCurrentPatternPtr(el, uiaPatternInvoke, "Invoke")
	if err != nil {
		return err
	}
	defer comRelease(pattern)
	vt := *(*uintptr)(unsafe.Pointer(pattern))
	logUIAComCallStart("Invoke", fmt.Sprintf("pattern_ptr=0x%x element_ptr=0x%x", pattern, el))
	hr, _, _ := syscall.SyscallN(comVTableMethod(vt, uiaVTableIUIAutomationInvokePatternInvoke), pattern)
	logUIAComCallEnd("Invoke", hr, fmt.Sprintf("pattern_ptr=0x%x", pattern))
	err = hresultErr("Invoke", hr)
	logUIAComCallError("Invoke", err, fmt.Sprintf("element_ptr=0x%x", el))
	return err
}

func uiaSelectionItemPatternSelect(el uintptr) error {
	pattern, err := uiaGetCurrentPatternPtr(el, uiaPatternSelectionItem, "Select")
	if err != nil {
		return err
	}
	defer comRelease(pattern)
	vt := *(*uintptr)(unsafe.Pointer(pattern))
	hr, _, _ := syscall.SyscallN(comVTableMethod(vt, uiaVTableIUIAutomationSelectionItemPatternSelect), pattern)
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
	hr, _, _ := syscall.SyscallN(comVTableMethod(vt, uiaVTableIUIAutomationValuePatternSetValue), pattern, uintptr(unsafe.Pointer(bstr)))
	return hresultErr("SetValue", hr)
}

func uiaLegacyIAccessiblePatternDoDefaultAction(el uintptr) error {
	pattern, err := uiaGetCurrentPatternPtr(el, uiaPatternLegacyIAccessible, "DoDefaultAction")
	if err != nil {
		return err
	}
	defer comRelease(pattern)
	vt := *(*uintptr)(unsafe.Pointer(pattern))
	hr, _, _ := syscall.SyscallN(comVTableMethod(vt, uiaVTableIUIAutomationLegacyIAccessiblePatternDoDefaultAction), pattern)
	return hresultErr("DoDefaultAction", hr)
}

func uiaTogglePatternToggle(el uintptr) error {
	pattern, err := uiaGetCurrentPatternPtr(el, uiaPatternToggle, "Toggle")
	if err != nil {
		return err
	}
	defer comRelease(pattern)
	vt := *(*uintptr)(unsafe.Pointer(pattern))
	hr, _, _ := syscall.SyscallN(comVTableMethod(vt, uiaVTableIUIAutomationTogglePatternToggle), pattern)
	return hresultErr("Toggle", hr)
}

func uiaExpandCollapsePatternExpand(el uintptr) error {
	pattern, err := uiaGetCurrentPatternPtr(el, uiaPatternExpandCollapse, "Expand")
	if err != nil {
		return err
	}
	defer comRelease(pattern)
	vt := *(*uintptr)(unsafe.Pointer(pattern))
	hr, _, _ := syscall.SyscallN(comVTableMethod(vt, uiaVTableIUIAutomationExpandCollapsePatternExpand), pattern)
	return hresultErr("Expand", hr)
}

func uiaExpandCollapsePatternCollapse(el uintptr) error {
	pattern, err := uiaGetCurrentPatternPtr(el, uiaPatternExpandCollapse, "Collapse")
	if err != nil {
		return err
	}
	defer comRelease(pattern)
	vt := *(*uintptr)(unsafe.Pointer(pattern))
	hr, _, _ := syscall.SyscallN(comVTableMethod(vt, uiaVTableIUIAutomationExpandCollapsePatternCollapse), pattern)
	return hresultErr("Collapse", hr)
}

func uiaGetCurrentPatternPtr(el uintptr, patternID int32, op string) (uintptr, error) {
	var p uintptr
	vt := *(*uintptr)(unsafe.Pointer(el))
	logUIAComCallStart(op, fmt.Sprintf("element_ptr=0x%x pattern_id=%d", el, patternID))
	hr, _, _ := syscall.SyscallN(comVTableMethod(vt, uiaVTableIUIAutomationElementGetCurrentPattern), el, uintptr(patternID), uintptr(unsafe.Pointer(&p)))
	logUIAComCallEnd(op, hr, fmt.Sprintf("element_ptr=0x%x pattern_id=%d pattern_ptr=0x%x", el, patternID, p))
	if err := hresultErr(op, hr); err != nil {
		logUIAComCallError(op, err, fmt.Sprintf("element_ptr=0x%x pattern_id=%d", el, patternID))
		return 0, err
	}
	if p == 0 {
		return 0, &UIAComUnavailableError{Op: op, Err: syscall.ENOTSUP}
	}
	return p, nil
}
