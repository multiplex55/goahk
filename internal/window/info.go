package window

import "fmt"

// HWND is an opaque native window handle wrapper.
type HWND uintptr

func (h HWND) String() string {
	return fmt.Sprintf("0x%X", uintptr(h))
}

// Info describes a top-level window.
type Info struct {
	HWND   HWND
	Title  string
	Class  string
	PID    uint32
	Exe    string
	Active bool
}
