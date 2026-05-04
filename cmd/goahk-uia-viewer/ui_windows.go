//go:build windows

package main

import (
	"fmt"

	"github.com/lxn/walk"
)

type viewerUI struct {
	controller   *Controller
	events       *ViewerEventAdapter
	dispatcher   UIThreadMarshaller
	walkUIThread *walk.MainWindow
	mw           *walk.MainWindow

	windowTable *walk.TableView
	windowModel *windowTableModel
	refreshBtn  *walk.PushButton
	visibleChk  *walk.CheckBox
	titleChk    *walk.CheckBox
	activateChk *walk.CheckBox

	infoView        *walk.TextEdit
	propertiesTV    *walk.TableView
	propertiesModel *propertyTableModel
	patternsTree    *walk.TreeView
	treeView        *walk.TreeView
	treeModel       *uiaTreeModel

	statusBar  *walk.StatusBar
	statusText *walk.StatusBarItem
}

func NewViewerWindow(controller *Controller) (viewerWindow, error) {
	if controller == nil {
		return nil, fmt.Errorf("controller is required")
	}

	ui := newViewerUI(controller)
	if err := ui.buildWindow(); err != nil {
		return nil, err
	}
	ui.attachEvents()
	ui.events = NewViewerEventAdapter(controller, ui, ui.dispatcher)
	ui.initialRefresh()
	return ui, nil
}

func (ui *viewerUI) Run() error {
	ui.mw.Run()
	return nil
}
