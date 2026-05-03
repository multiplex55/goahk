//go:build windows

package main

import "github.com/lxn/walk"

type walkUIThread struct{}

func (walkUIThread) Queue(fn func()) {
	walk.Async(fn)
}

func (ui *viewerUI) setLoading(loading bool) {
	if ui.refreshBtn != nil {
		ui.refreshBtn.SetEnabled(!loading)
	}
	if ui.windowTable != nil {
		ui.windowTable.SetEnabled(!loading)
	}
	if ui.treeView != nil {
		ui.treeView.SetEnabled(!loading)
	}
}
