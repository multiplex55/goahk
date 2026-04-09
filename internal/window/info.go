package window

import (
	"encoding/json"
	"fmt"
)

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
	Rect   *Rect
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

func (i Info) Bounds() (Rect, bool) {
	if i.Rect == nil {
		return Rect{}, false
	}
	return *i.Rect, true
}

func (i Info) Position() (int, int, bool) {
	bounds, ok := i.Bounds()
	if !ok {
		return 0, 0, false
	}
	return bounds.Left, bounds.Top, true
}

func (i Info) Size() (int, int, bool) {
	bounds, ok := i.Bounds()
	if !ok {
		return 0, 0, false
	}
	return bounds.Width(), bounds.Height(), true
}

func (i Info) MarshalJSON() ([]byte, error) {
	type infoPayload struct {
		HWND   HWND  `json:"HWND"`
		Title  string `json:"Title"`
		Class  string `json:"Class"`
		PID    uint32 `json:"PID"`
		Exe    string `json:"Exe"`
		Active bool   `json:"Active"`
		Rect   *Rect  `json:"Rect,omitempty"`
	}
	return json.Marshal(infoPayload{
		HWND:   i.HWND,
		Title:  i.Title,
		Class:  i.Class,
		PID:    i.PID,
		Exe:    i.Exe,
		Active: i.Active,
		Rect:   i.Rect,
	})
}
