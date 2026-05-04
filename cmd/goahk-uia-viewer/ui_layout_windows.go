//go:build windows

package main

import (
	"fmt"

	"github.com/lxn/walk"
)

func newViewerUI(controller *Controller) *viewerUI {
	return &viewerUI{controller: controller}
}

func (ui *viewerUI) buildWindow() error {
	mw, err := walk.NewMainWindow()
	if err != nil {
		return fmt.Errorf("create main window: %w", err)
	}
	ui.mw = mw
	ui.mw.SetTitle("goahk UIA Viewer")
	ui.mw.SetSize(walk.Size{Width: 1400, Height: 900})

	if ui.windowFilter, err = walk.NewLineEdit(ui.mw); err != nil { return err }
	if ui.refreshBtn, err = walk.NewPushButton(ui.mw); err != nil { return err }
	ui.refreshBtn.SetText("Refresh")
	if ui.visibleChk, err = walk.NewCheckBox(ui.mw); err != nil { return err }
	ui.visibleChk.SetText("Visible only")
	ui.visibleChk.SetChecked(true)
	if ui.titleChk, err = walk.NewCheckBox(ui.mw); err != nil { return err }
	ui.titleChk.SetText("Title only")
	ui.titleChk.SetChecked(true)
	if ui.activateChk, err = walk.NewCheckBox(ui.mw); err != nil { return err }
	ui.activateChk.SetText("Activate on select")

	if ui.windowTable, err = walk.NewTableView(ui.mw); err != nil { return err }
	if ui.infoView, err = walk.NewTextEdit(ui.mw); err != nil { return err }
	if ui.propertiesTV, err = walk.NewTableView(ui.mw); err != nil { return err }
	if ui.patternsTV, err = walk.NewTableView(ui.mw); err != nil { return err }
	if ui.treeView, err = walk.NewTreeView(ui.mw); err != nil { return err }

	if err := ui.buildStatusBar(); err != nil {
		return err
	}
	return nil
}

func (ui *viewerUI) defaultRefreshArgs() (visibleOnly, titleOnly bool) {
	if ui.visibleChk != nil {
		visibleOnly = ui.visibleChk.Checked()
	}
	if ui.titleChk != nil {
		titleOnly = ui.titleChk.Checked()
	}
	return visibleOnly, titleOnly
}

func (ui *viewerUI) buildLeftPane() error   { return nil }
func (ui *viewerUI) buildMiddlePane() error { return nil }
func (ui *viewerUI) buildRightPane() error  { return nil }

func (ui *viewerUI) buildStatusBar() error {
	sb, err := walk.NewStatusBar(ui.mw)
	if err != nil {
		return err
	}
	ui.statusBar = sb
	item := walk.NewStatusBarItem()
	if err := sb.Items().Add(item); err != nil {
		return err
	}
	ui.statusText = item
	ui.statusText.SetText("ready")
	return nil
}
