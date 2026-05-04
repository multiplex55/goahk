//go:build windows

package main

import (
	"fmt"
	"strings"

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

	if err := ui.buildLeftPane(); err != nil {
		return err
	}
	if err := ui.buildMiddlePane(); err != nil {
		return err
	}
	if err := ui.buildRightPane(); err != nil {
		return err
	}

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

func (ui *viewerUI) buildLeftPane() error {
	var err error
	if ui.refreshBtn, err = walk.NewPushButton(ui.mw); err != nil {
		return err
	}
	ui.refreshBtn.SetText("Refresh")
	if ui.visibleChk, err = walk.NewCheckBox(ui.mw); err != nil {
		return err
	}
	ui.visibleChk.SetText("Visible")
	ui.visibleChk.SetChecked(true)
	if ui.titleChk, err = walk.NewCheckBox(ui.mw); err != nil {
		return err
	}
	ui.titleChk.SetText("Title")
	ui.titleChk.SetChecked(true)
	if ui.activateChk, err = walk.NewCheckBox(ui.mw); err != nil {
		return err
	}
	ui.activateChk.SetText("Activate")
	if ui.windowTable, err = walk.NewTableView(ui.mw); err != nil {
		return err
	}
	return nil
}
func (ui *viewerUI) buildMiddlePane() error {
	var err error
	if ui.infoView, err = walk.NewTextEdit(ui.mw); err != nil {
		return err
	}
	ui.infoView.SetReadOnly(true)
	if ui.propertiesTV, err = walk.NewTableView(ui.mw); err != nil {
		return err
	}
	if ui.patternsTree, err = walk.NewTreeView(ui.mw); err != nil {
		return err
	}
	return nil
}
func (ui *viewerUI) buildRightPane() error {
	var err error
	if ui.treeView, err = walk.NewTreeView(ui.mw); err != nil {
		return err
	}
	return nil
}

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

func firstLine(s string) string {
	out := strings.TrimSpace(s)
	if idx := strings.IndexByte(out, '\n'); idx >= 0 {
		return out[:idx]
	}
	return out
}
