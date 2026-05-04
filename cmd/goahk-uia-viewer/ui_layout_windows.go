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
	if err := ui.mw.SetLayout(walk.NewVBoxLayout()); err != nil {
		return fmt.Errorf("set main window layout: %w", err)
	}
	root, err := walk.NewComposite(ui.mw)
	if err != nil {
		return fmt.Errorf("create root composite: %w", err)
	}
	ui.root = root
	if err := ui.root.SetLayout(walk.NewGridLayout()); err != nil {
		return fmt.Errorf("set root layout: %w", err)
	}

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
	visibleOnly = true
	titleOnly = true
	if ui.visibleChk != nil {
		visibleOnly = ui.visibleChk.Checked()
	}
	if ui.titleChk != nil {
		titleOnly = ui.titleChk.Checked()
	}
	return visibleOnly, titleOnly
}

func (ui *viewerUI) activateOnSelect() bool {
	if ui.activateChk == nil {
		return true
	}
	return ui.activateChk.Checked()
}

func (ui *viewerUI) buildLeftPane() error {
	var err error
	parent := walk.Container(ui.root)
	if parent == nil {
		parent = ui.mw
	}
	if ui.refreshBtn, err = walk.NewPushButton(parent); err != nil {
		return err
	}
	ui.refreshBtn.SetText("Refresh")
	ui.refreshBtn.SetBounds(walk.Rectangle{X: 8, Y: 8, Width: 80, Height: 28})
	if ui.visibleChk, err = walk.NewCheckBox(parent); err != nil {
		return err
	}
	ui.visibleChk.SetText("Visible")
	ui.visibleChk.SetChecked(true)
	ui.visibleChk.SetBounds(walk.Rectangle{X: 96, Y: 10, Width: 80, Height: 24})
	if ui.titleChk, err = walk.NewCheckBox(parent); err != nil {
		return err
	}
	ui.titleChk.SetText("Title")
	ui.titleChk.SetChecked(true)
	ui.titleChk.SetBounds(walk.Rectangle{X: 180, Y: 10, Width: 80, Height: 24})
	if ui.activateChk, err = walk.NewCheckBox(parent); err != nil {
		return err
	}
	ui.activateChk.SetText("Activate")
	ui.activateChk.SetBounds(walk.Rectangle{X: 264, Y: 10, Width: 90, Height: 24})
	if ui.windowTable, err = walk.NewTableView(parent); err != nil {
		return err
	}
	ui.windowTable.SetBounds(walk.Rectangle{X: 8, Y: 44, Width: 450, Height: 760})
	return nil
}
func (ui *viewerUI) buildMiddlePane() error {
	var err error
	parent := walk.Container(ui.root)
	if parent == nil {
		parent = ui.mw
	}
	if ui.infoView, err = walk.NewTextEdit(parent); err != nil {
		return err
	}
	ui.infoView.SetReadOnly(true)
	ui.infoView.SetBounds(walk.Rectangle{X: 470, Y: 8, Width: 470, Height: 220})
	if ui.propertiesTV, err = walk.NewTableView(parent); err != nil {
		return err
	}
	ui.propertiesTV.SetBounds(walk.Rectangle{X: 470, Y: 236, Width: 470, Height: 280})
	if ui.patternsTree, err = walk.NewTreeView(parent); err != nil {
		return err
	}
	ui.patternsTree.SetBounds(walk.Rectangle{X: 470, Y: 524, Width: 470, Height: 280})
	return nil
}
func (ui *viewerUI) buildRightPane() error {
	var err error
	parent := walk.Container(ui.root)
	if parent == nil {
		parent = ui.mw
	}
	if ui.treeView, err = walk.NewTreeView(parent); err != nil {
		return err
	}
	ui.treeView.SetBounds(walk.Rectangle{X: 952, Y: 8, Width: 430, Height: 796})
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
