//go:build windows

package main

import (
	"fmt"

	"github.com/lxn/walk"
)

type viewerUI struct {
	controller *Controller
	mw         *walk.MainWindow

	windowFilter *walk.LineEdit
	windowTable  *walk.TableView
	windowModel  *windowTableModel
	refreshBtn   *walk.PushButton
	visibleChk   *walk.CheckBox
	titleChk     *walk.CheckBox
	activateChk  *walk.CheckBox

	infoView        *walk.TextEdit
	propertiesTV    *walk.TableView
	propertiesModel *propertyTableModel
	patternsTV      *walk.TableView
	patternsModel   *patternTreeModel
	treeView        *walk.TreeView
	treeModel       *uiaTreeModel

	statusBar  *walk.StatusBar
	statusText *walk.StatusBarItem

	events     *ViewerEventAdapter
	dispatcher walkUIThread
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
	_, err := ui.mw.Run()
	return err
}
