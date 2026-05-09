//go:build windows

package main

import (
	"syscall"

	"github.com/lxn/walk"
)

const (
	wmSetRedraw = 0x000B
)

var (
	user32SendMessageW = syscall.NewLazyDLL("user32.dll").NewProc("SendMessageW")
)

func withTreeRedrawSuspended(ui *viewerUI, fn func()) {
	if ui == nil || fn == nil {
		return
	}
	if ui.treeView == nil {
		fn()
		return
	}

	type suspender interface{ SetSuspended(bool) }
	if s, ok := any(ui.treeView).(suspender); ok {
		s.SetSuspended(true)
		defer s.SetSuspended(false)
		fn()
		ui.treeView.Invalidate()
		return
	}

	hwnd := uintptr(ui.treeView.Handle())
	if hwnd == 0 {
		fn()
		return
	}
	sendSetRedraw(hwnd, 0)
	defer func() {
		sendSetRedraw(hwnd, 1)
		ui.treeView.Invalidate()
	}()
	fn()
}

func sendSetRedraw(hwnd uintptr, redraw uintptr) {
	if hwnd == 0 {
		return
	}
	_, _, _ = user32SendMessageW.Call(hwnd, wmSetRedraw, redraw, 0)
}
