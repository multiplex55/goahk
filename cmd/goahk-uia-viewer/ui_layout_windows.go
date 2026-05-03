//go:build windows

package main

import "github.com/lxn/walk"

func newViewerUI() *viewerUI {
	ui := &viewerUI{}
	ui.visibleChk = new(walk.CheckBox)
	ui.titleChk = new(walk.CheckBox)
	ui.activateChk = new(walk.CheckBox)
	ui.visibleChk.SetChecked(true)
	ui.titleChk.SetChecked(true)
	ui.activateChk.SetChecked(true)
	return ui
}

func (ui *viewerUI) defaultRefreshArgs() (visibleOnly, titleOnly bool) {
	return true, true
}
