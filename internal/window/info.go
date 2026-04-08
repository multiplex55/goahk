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

// Rect describes a window rectangle in screen coordinates.
type Rect struct {
	Left   int
	Top    int
	Right  int
	Bottom int
}

func (r Rect) Width() int {
	return r.Right - r.Left
}

func (r Rect) Height() int {
	return r.Bottom - r.Top
}
