//go:build windows
// +build windows

package inspect

import (
	"errors"
	"fmt"
	"log"
	"math"
	"syscall"
	"unsafe"

	"goahk/internal/window"
	"golang.org/x/sys/windows"
)

const (
	uiaTreeScopeChildren    = 0x2
	uiaPropertyRuntimeID    = 30000
	uiaPropertyBoundingRect = 30001
	uiaPropertyProcessID    = 30002
	uiaPropertyControlType  = 30003
	uiaPropertyLocalizedCtl = 30004
	uiaPropertyName         = 30005
	uiaPropertyAccelerator  = 30006
	uiaPropertyAccessKey    = 30007
	uiaPropertyHasFocus     = 30008
	uiaPropertyIsFocusable  = 30009
	uiaPropertyIsEnabled    = 30010
	uiaPropertyAutomationID = 30011
	uiaPropertyClassName    = 30012
	uiaPropertyHelpText     = 30013
	uiaPropertyIsCtrlElem   = 30016
	uiaPropertyIsContent    = 30017
	uiaPropertyIsPassword   = 30019
	uiaPropertyNativeHWND   = 30020
	uiaPropertyItemType     = 30021
	uiaPropertyIsOffscreen  = 30022
	uiaPropertyOrientation  = 30023
	uiaPropertyFrameworkID  = 30024
	uiaPropertyIsRequired   = 30025
	uiaPropertyItemStatus   = 30026
	uiaPropertyLabeledBy    = 30018
	uiaEElementNotAvailable = 0x80040201
	uiaEElementNotEnabled   = 0x80040200
	eAccessDenied           = 0x80070005
	coEObjNotConnected      = 0x800401FD
	rpcEServerUnavailable   = 0x800706BA
)

const (
	vtEmpty   = 0
	vtI4      = 3
	vtR8      = 5
	vtBool    = 11
	vtBSTR    = 8
	vtUnknown = 13
)

const (
	comVTableQueryInterface = 0
	comVTableAddRef         = 1
	comVTableRelease        = 2
)

const (
	uiaVTableIUIAutomationElementFromHandle   = 6
	uiaVTableIUIAutomationElementFromPoint    = 7
	uiaVTableIUIAutomationGetFocusedElement   = 8
	uiaVTableIUIAutomationCreateTrueCondition = 21
	uiaVTableIUIAutomationGetRawViewWalker    = 16
)

const (
	uiaVTableIUIAutomationElementFindAll                 = 6
	uiaVTableIUIAutomationElementGetCurrentRuntimeID     = 9
	uiaVTableIUIAutomationElementGetCurrentPropertyValue = 10
	uiaVTableIUIAutomationElementGetCurrentPattern       = 11
)

const (
	uiaVTableIUIAutomationElementArrayLength     = 3
	uiaVTableIUIAutomationElementArrayGetElement = 4
)

const (
	uiaVTableIUIAutomationTreeWalkerGetParentElement = 3
)

func comVTableMethod(vt uintptr, index uintptr) uintptr {
	return *(*uintptr)(unsafe.Pointer(vt + index*unsafe.Sizeof(uintptr(0))))
}

type comVariant struct {
	VT         uint16
	WReserved1 uint16
	WReserved2 uint16
	WReserved3 uint16
	Val        int64
	Val2       int64
}

type uiaPropRead struct {
	Status string
	I      int
	B      bool
	S      string
	Rect   *uiaRect
}

func comRelease(ptr uintptr) {
	if ptr == 0 {
		return
	}
	vt := *(*uintptr)(unsafe.Pointer(ptr))
	_, _, _ = syscall.SyscallN(comVTableMethod(vt, comVTableRelease), ptr)
}

func comAddRef(ptr uintptr) {
	if ptr == 0 {
		return
	}
	vt := *(*uintptr)(unsafe.Pointer(ptr))
	_, _, _ = syscall.SyscallN(comVTableMethod(vt, comVTableAddRef), ptr)
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
	hr, _, _ := syscall.SyscallN(comVTableMethod(vt, uiaVTableIUIAutomationCreateTrueCondition), automation, uintptr(unsafe.Pointer(&cond)))
	return cond, hresultErr("CreateTrueCondition", hr)
}

func uiaGetRawViewWalker(automation uintptr) (uintptr, error) {
	var walker uintptr
	vt := *(*uintptr)(unsafe.Pointer(automation))
	hr, _, _ := syscall.SyscallN(comVTableMethod(vt, uiaVTableIUIAutomationGetRawViewWalker), automation, uintptr(unsafe.Pointer(&walker)))
	return walker, hresultErr("get_RawViewWalker", hr)
}

func uiaElementFromHandle(automation uintptr, hwnd window.HWND) (uintptr, error) {
	var el uintptr
	vt := *(*uintptr)(unsafe.Pointer(automation))
	hr, _, _ := syscall.SyscallN(comVTableMethod(vt, uiaVTableIUIAutomationElementFromHandle), automation, uintptr(hwnd), uintptr(unsafe.Pointer(&el)))
	return el, hresultErr("ElementFromHandle", hr)
}

func uiaGetFocusedElement(automation uintptr) (uintptr, error) {
	var el uintptr
	vt := *(*uintptr)(unsafe.Pointer(automation))
	hr, _, _ := syscall.SyscallN(comVTableMethod(vt, uiaVTableIUIAutomationGetFocusedElement), automation, uintptr(unsafe.Pointer(&el)))
	return el, hresultErr("GetFocusedElement", hr)
}

func uiaElementFromPoint(automation uintptr, x, y int) (uintptr, error) {
	var el uintptr
	pt := struct{ X, Y int32 }{X: int32(x), Y: int32(y)}
	vt := *(*uintptr)(unsafe.Pointer(automation))
	hr, _, _ := syscall.SyscallN(comVTableMethod(vt, uiaVTableIUIAutomationElementFromPoint), automation, *(*uintptr)(unsafe.Pointer(&pt)), uintptr(unsafe.Pointer(&el)))
	return el, hresultErr("ElementFromPoint", hr)
}

func uiaFindAllChildren(el, trueCond uintptr) (uintptr, error) {
	log.Printf("inspect.uia.native.find_children checkpoint=\"FindAll started\" parent_ptr=0x%x true_condition_ptr=0x%x", el, trueCond)
	var arr uintptr
	vt := *(*uintptr)(unsafe.Pointer(el))
	hr, _, _ := syscall.SyscallN(comVTableMethod(vt, uiaVTableIUIAutomationElementFindAll), el, uintptr(uiaTreeScopeChildren), trueCond, uintptr(unsafe.Pointer(&arr)))
	log.Printf("inspect.uia.native.find_children checkpoint=\"FindAll returned\" array_ptr=0x%x", arr)
	return arr, hresultErr("FindAll", hr)
}

func uiaArrayLength(arr uintptr) (int32, error) {
	var ln int32
	vt := *(*uintptr)(unsafe.Pointer(arr))
	hr, _, _ := syscall.SyscallN(comVTableMethod(vt, uiaVTableIUIAutomationElementArrayLength), arr, uintptr(unsafe.Pointer(&ln)))
	return ln, hresultErr("IUIAutomationElementArray.Length", hr)
}

func uiaArrayGet(arr uintptr, idx int32) (uintptr, error) {
	var el uintptr
	vt := *(*uintptr)(unsafe.Pointer(arr))
	hr, _, _ := syscall.SyscallN(comVTableMethod(vt, uiaVTableIUIAutomationElementArrayGetElement), arr, uintptr(idx), uintptr(unsafe.Pointer(&el)))
	return el, hresultErr("IUIAutomationElementArray.GetElement", hr)
}

func uiaGetParentElement(walker, el uintptr) (uintptr, error) {
	var parent uintptr
	vt := *(*uintptr)(unsafe.Pointer(walker))
	hr, _, _ := syscall.SyscallN(comVTableMethod(vt, uiaVTableIUIAutomationTreeWalkerGetParentElement), walker, el, uintptr(unsafe.Pointer(&parent)))
	return parent, hresultErr("GetParentElement", hr)
}

func uiaElementRuntimeID(el uintptr) (string, error) {
	var arr uintptr
	vt := *(*uintptr)(unsafe.Pointer(el))
	hr, _, _ := syscall.SyscallN(comVTableMethod(vt, uiaVTableIUIAutomationElementGetCurrentRuntimeID), el, uintptr(unsafe.Pointer(&arr)))
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

func uiaGetCurrentPropertyValue(el uintptr, propertyID int32) (comVariant, error) {
	var v comVariant
	vt := *(*uintptr)(unsafe.Pointer(el))
	hr, _, _ := syscall.SyscallN(comVTableMethod(vt, uiaVTableIUIAutomationElementGetCurrentPropertyValue), el, uintptr(propertyID), uintptr(unsafe.Pointer(&v)))
	return v, hresultErr("GetCurrentPropertyValue", hr)
}

func uiaGetCurrentPattern(el uintptr, patternID int32) (bool, error) {
	var p uintptr
	vt := *(*uintptr)(unsafe.Pointer(el))
	hr, _, _ := syscall.SyscallN(comVTableMethod(vt, uiaVTableIUIAutomationElementGetCurrentPattern), el, uintptr(patternID), uintptr(unsafe.Pointer(&p)))
	if err := hresultErr("GetCurrentPattern", hr); err != nil {
		var stale *UIAElementStaleError
		if errors.As(err, &stale) {
			return false, err
		}
		return false, nil
	}
	if p != 0 {
		comRelease(p)
		return true, nil
	}
	return false, nil
}

func decodeVariant(v comVariant) uiaPropRead {
	switch v.VT {
	case vtEmpty:
		return uiaPropRead{Status: propertyStatusEmpty}
	case vtBSTR:
		if v.Val == 0 {
			return uiaPropRead{Status: propertyStatusEmpty}
		}
		s := windows.UTF16PtrToString((*uint16)(unsafe.Pointer(uintptr(v.Val))))
		if s == "" {
			return uiaPropRead{Status: propertyStatusEmpty}
		}
		return uiaPropRead{Status: propertyStatusOK, S: s}
	case vtI4:
		i := int(int32(v.Val))
		if i == 0 {
			return uiaPropRead{Status: propertyStatusEmpty}
		}
		return uiaPropRead{Status: propertyStatusOK, I: i}
	case vtBool:
		return uiaPropRead{Status: propertyStatusOK, B: int16(v.Val) != 0}
	case vtR8:
		f := math.Float64frombits(uint64(v.Val))
		return uiaPropRead{Status: propertyStatusOK, I: int(f)}
	case vtUnknown:
		if v.Val == 0 {
			return uiaPropRead{Status: propertyStatusEmpty}
		}
		return uiaPropRead{Status: propertyStatusOK, S: fmt.Sprintf("0x%x", uintptr(v.Val))}
	default:
		return uiaPropRead{Status: propertyStatusUnsupported}
	}
}
