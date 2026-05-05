//go:build windows

package main

import "github.com/lxn/walk"

func (ui *viewerUI) setStatus(text string) {
	if ui.statusText != nil {
		ui.statusText.SetText(text)
	}
	if ui.statusBar != nil {
		ui.statusBar.Invalidate()
	}
}

func (ui *viewerUI) ShowFatal(message string) {
	if ui == nil || ui.mw == nil {
		return
	}
	walk.MsgBox(ui.mw, "goahk UIA Viewer Error", message, walk.MsgBoxIconError|walk.MsgBoxOK)
}
