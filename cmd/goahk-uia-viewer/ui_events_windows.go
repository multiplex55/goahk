//go:build windows

package main

import (
	"fmt"

	"github.com/lxn/walk"
	"goahk/internal/inspect"
)

type walkUIThread struct{ mw *walk.MainWindow }

func (m walkUIThread) Queue(fn func()) {
	if m.mw == nil {
		return
	}
	m.mw.Synchronize(fn)
}

func (ui *viewerUI) attachEvents() {
	ui.dispatcher = walkUIThread{mw: ui.mw}
	ui.refreshBtn.Clicked().Attach(func() { ui.initialRefresh() })
}

func (ui *viewerUI) initialRefresh() {
	ui.SetBusy(true)
	go func() {
		visible, title := ui.defaultRefreshArgs()
		resp, err := ui.controller.RefreshWindows("", visible, title)
		ui.dispatcher.Queue(func() {
			ui.SetBusy(false)
			if err != nil {
				ui.setStatus("refresh failed: " + err.Error())
				return
			}
			rows := mapWindowTableRows(resp.Windows, true)
			ui.setStatus(fmt.Sprintf("loaded %d windows", len(rows)))
		})
	}()
}

func (ui *viewerUI) setLoading(loading bool) {
	ui.dispatcher.Queue(func() {
		if ui.refreshBtn != nil {
			ui.refreshBtn.SetEnabled(!loading)
		}
		if ui.windowTable != nil {
			ui.windowTable.SetEnabled(!loading)
		}
		if ui.treeView != nil {
			ui.treeView.SetEnabled(!loading)
		}
	})
}

func (ui *viewerUI) SetBusy(b bool) { ui.setLoading(b) }
func (ui *viewerUI) SetStatus(s string) { ui.setStatus(s) }
func (ui *viewerUI) UpdateWindowDetails(inspect.GetNodeDetailsResponse) {}
func (ui *viewerUI) UpdateNodeDetails(inspect.GetNodeDetailsResponse)   {}
func (ui *viewerUI) UpdateTreeRoot(inspect.TreeNodeDTO)                 {}
func (ui *viewerUI) UpdateNodeChildren(string, []inspect.TreeNodeDTO)   {}
