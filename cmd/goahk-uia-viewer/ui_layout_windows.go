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
	if err := ui.root.SetLayout(walk.NewVBoxLayout()); err != nil {
		return fmt.Errorf("set root layout: %w", err)
	}

	splitter, err := walk.NewHSplitter(ui.root)
	if err != nil {
		return fmt.Errorf("create horizontal splitter: %w", err)
	}

	if err := ui.buildLeftPane(splitter); err != nil {
		return err
	}
	if err := ui.buildMiddlePane(splitter); err != nil {
		return err
	}
	if err := ui.buildRightPane(splitter); err != nil {
		return err
	}
	splitter.SetFixed(splitter.ContainerBase.Children().At(1), false)

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

func (ui *viewerUI) buildLeftPane(parent walk.Container) error {
	var err error
	leftPane, err := walk.NewComposite(parent)
	if err != nil {
		return fmt.Errorf("create left pane: %w", err)
	}
	if err := leftPane.SetLayout(walk.NewVBoxLayout()); err != nil {
		return fmt.Errorf("set left pane layout: %w", err)
	}

	if _, err = walk.NewLabel(leftPane); err != nil {
		return err
	}
	leftPane.Children().At(0).(*walk.Label).SetText("Windows and Controls")
	if ui.windowTable, err = walk.NewTableView(leftPane); err != nil {
		return err
	}
	ui.windowTable.SetColumnsOrderable(true)
	ui.windowTable.SetGridlines(true)
	ui.windowTable.SetAlternatingRowBG(true)
	ui.windowTable.SetMultiSelection(false)
	ui.windowTable.SetHeaderHidden(false)
	if err := ui.windowTable.Columns().Add(walk.NewTableViewColumn()); err != nil {
		return err
	}
	col := ui.windowTable.Columns().At(0)
	col.SetName("Title")
	col.SetTitle("Title")
	col.SetWidth(260)
	if err := ui.windowTable.Columns().Add(walk.NewTableViewColumn()); err != nil {
		return err
	}
	col = ui.windowTable.Columns().At(1)
	col.SetName("Process")
	col.SetTitle("Process")
	col.SetWidth(120)
	if err := ui.windowTable.Columns().Add(walk.NewTableViewColumn()); err != nil {
		return err
	}
	col = ui.windowTable.Columns().At(2)
	col.SetName("ID")
	col.SetTitle("ID")
	col.SetWidth(120)

	toolbar, err := walk.NewComposite(leftPane)
	if err != nil {
		return fmt.Errorf("create left toolbar: %w", err)
	}
	toolLayout := walk.NewHBoxLayout()
	toolLayout.SetMargins(walk.Margins{0, 0, 0, 0})
	toolLayout.SetSpacing(6)
	if err := toolbar.SetLayout(toolLayout); err != nil {
		return fmt.Errorf("set left toolbar layout: %w", err)
	}

	if ui.refreshBtn, err = walk.NewPushButton(toolbar); err != nil {
		return err
	}
	ui.refreshBtn.SetText("Refresh list")
	if ui.modeCombo, err = walk.NewComboBox(toolbar); err != nil {
		return err
	}
	ui.modeCombo.SetModel([]string{"Auto/UIA + fallback", "ACC/MSAA (forced)"})
	ui.modeCombo.SetCurrentIndex(0)
	if ui.visibleChk, err = walk.NewCheckBox(toolbar); err != nil {
		return err
	}
	ui.visibleChk.SetText("Visible")
	ui.visibleChk.SetChecked(true)
	if ui.titleChk, err = walk.NewCheckBox(toolbar); err != nil {
		return err
	}
	ui.titleChk.SetText("Title")
	ui.titleChk.SetChecked(true)
	if ui.activateChk, err = walk.NewCheckBox(toolbar); err != nil {
		return err
	}
	ui.activateChk.SetText("Activate")

	return nil
}

func (ui *viewerUI) buildMiddlePane(parent walk.Container) error {
	var err error
	middlePane, err := walk.NewComposite(parent)
	if err != nil {
		return fmt.Errorf("create middle pane: %w", err)
	}
	midLayout := walk.NewVBoxLayout()
	midLayout.SetSpacing(4)
	if err := middlePane.SetLayout(midLayout); err != nil {
		return fmt.Errorf("set middle pane layout: %w", err)
	}

	if _, err = walk.NewLabel(middlePane); err != nil {
		return err
	}
	middlePane.Children().At(0).(*walk.Label).SetText("Window Info")
	if ui.infoTable, err = walk.NewTableView(middlePane); err != nil {
		return err
	}
	ui.infoTable.SetColumnsOrderable(true)
	ui.infoTable.SetGridlines(true)
	ui.infoTable.SetAlternatingRowBG(true)
	if err := ui.infoTable.Columns().Add(walk.NewTableViewColumn()); err != nil {
		return err
	}
	infoCol := ui.infoTable.Columns().At(0)
	infoCol.SetName("Property")
	infoCol.SetTitle("Property")
	infoCol.SetWidth(150)
	if err := ui.infoTable.Columns().Add(walk.NewTableViewColumn()); err != nil {
		return err
	}
	infoCol = ui.infoTable.Columns().At(1)
	infoCol.SetName("Value")
	infoCol.SetTitle("Value")
	infoCol.SetWidth(260)
	ui.infoTable.SetMinMaxSize(walk.Size{Width: 0, Height: 140}, walk.Size{})

	if _, err = walk.NewLabel(middlePane); err != nil {
		return err
	}
	middlePane.Children().At(2).(*walk.Label).SetText("Properties")
	if ui.propertiesTV, err = walk.NewTableView(middlePane); err != nil {
		return err
	}
	ui.propertiesTV.SetColumnsOrderable(true)
	ui.propertiesTV.SetGridlines(true)
	ui.propertiesTV.SetAlternatingRowBG(true)
	ui.propertiesTV.SetHeaderHidden(false)
	if err := ui.propertiesTV.Columns().Add(walk.NewTableViewColumn()); err != nil {
		return err
	}
	propCol := ui.propertiesTV.Columns().At(0)
	propCol.SetName("PropertyId")
	propCol.SetTitle("PropertyId")
	propCol.SetWidth(190)
	if err := ui.propertiesTV.Columns().Add(walk.NewTableViewColumn()); err != nil {
		return err
	}
	propCol = ui.propertiesTV.Columns().At(1)
	propCol.SetName("Value")
	propCol.SetTitle("Value")
	propCol.SetWidth(230)
	if err := ui.propertiesTV.Columns().Add(walk.NewTableViewColumn()); err != nil {
		return err
	}
	propCol = ui.propertiesTV.Columns().At(2)
	propCol.SetName("Status")
	propCol.SetTitle("Status")
	propCol.SetWidth(100)

	if _, err = walk.NewLabel(middlePane); err != nil {
		return err
	}
	middlePane.Children().At(4).(*walk.Label).SetText("Patterns")
	if ui.patternsTree, err = walk.NewTreeView(middlePane); err != nil {
		return err
	}
	ui.patternsTree.SetMinMaxSize(walk.Size{Width: 0, Height: 120}, walk.Size{})

	midLayout.SetStretchFactor(ui.infoTable, 0)
	midLayout.SetStretchFactor(ui.propertiesTV, 1)
	midLayout.SetStretchFactor(ui.patternsTree, 0)
	return nil
}

func (ui *viewerUI) buildRightPane(parent walk.Container) error {
	var err error
	rightPane, err := walk.NewComposite(parent)
	if err != nil {
		return fmt.Errorf("create right pane: %w", err)
	}
	rightLayout := walk.NewVBoxLayout()
	if err := rightPane.SetLayout(rightLayout); err != nil {
		return fmt.Errorf("set right pane layout: %w", err)
	}
	if _, err = walk.NewLabel(rightPane); err != nil {
		return err
	}
	rightPane.Children().At(0).(*walk.Label).SetText("UIA Tree")
	if ui.treeView, err = walk.NewTreeView(rightPane); err != nil {
		return err
	}
	rightLayout.SetStretchFactor(ui.treeView, 1)

	filterRow, err := walk.NewComposite(rightPane)
	if err != nil {
		return fmt.Errorf("create right filter row: %w", err)
	}
	filterLayout := walk.NewHBoxLayout()
	filterLayout.SetMargins(walk.Margins{0, 0, 0, 0})
	filterLayout.SetSpacing(6)
	if err := filterRow.SetLayout(filterLayout); err != nil {
		return fmt.Errorf("set right filter row layout: %w", err)
	}
	if ui.filterLbl, err = walk.NewLabel(filterRow); err != nil {
		return err
	}
	ui.filterLbl.SetText("Filter:")
	if ui.filterEdit, err = walk.NewLineEdit(filterRow); err != nil {
		return err
	}
	if ui.macroSidebarBtn, err = walk.NewPushButton(filterRow); err != nil {
		return err
	}
	ui.macroSidebarBtn.SetText("Show macro sidebar =>")
	ui.macroSidebarBtn.SetEnabled(false)
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
	ui.statusText.SetText("Ready. Click Refresh to enumerate windows.")
	return nil
}

func firstLine(s string) string {
	out := strings.TrimSpace(s)
	if idx := strings.IndexByte(out, '\n'); idx >= 0 {
		return out[:idx]
	}
	return out
}
