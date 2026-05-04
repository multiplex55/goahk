//go:build windows

package main

type walkUIThread struct{}

func (walkUIThread) Queue(fn func()) {
	go fn()
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
