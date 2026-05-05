//go:build windows

package main

import (
	"fmt"

	"github.com/lxn/walk"
)

type viewerUI struct {
	controller                 *Controller
	events                     *ViewerEventAdapter
	dispatcher                 UIThreadMarshaller
	walkUIThread               *walk.MainWindow
	suppressTreeSelectionEvent bool
	mw                         *walk.MainWindow
	root                       *walk.Composite

	windowTable *walk.TableView
	windowModel *windowTableModel
	refreshBtn  *walk.PushButton
	visibleChk  *walk.CheckBox
	titleChk    *walk.CheckBox
	activateChk *walk.CheckBox

	filterLbl       *walk.Label
	filterEdit      *walk.LineEdit
	macroSidebarBtn *walk.PushButton

	infoTable       *walk.TableView
	infoModel       *infoTableModel
	propertiesTV    *walk.TableView
	propertiesModel *propertyTableModel
	patternsTree    *walk.TreeView
	patternModel    *patternTreeModel
	treeView        *walk.TreeView
	treeModel       *uiaTreeModel

	statusBar  *walk.StatusBar
	statusText *walk.StatusBarItem
}

func NewViewerWindow(controller *Controller) (viewerWindow, error) {
	logStartup("ui_window_new begin")
	if controller == nil {
		return nil, fmt.Errorf("controller is required")
	}

	ui := newViewerUI(controller)
	if err := ui.buildWindow(); err != nil {
		return nil, err
	}
	ui.attachEvents()
	ui.events = NewViewerEventAdapter(controller, ui, ui.dispatcher)
	logStartup("ui_window_new ready")
	return ui, nil
}

func (ui *viewerUI) Run() error {
	if ui == nil {
		return fmt.Errorf("viewer UI is nil")
	}
	if ui.mw == nil {
		return fmt.Errorf("viewer main window is nil")
	}

	ui.mw.SetVisible(true)
	ui.mw.Activate()
	logStartup("visible/activate")
	exitCode := ui.mw.Run()
	logStartup("run returned")
	if exitCode != 0 {
		return fmt.Errorf("viewer exited with status code %d", exitCode)
	}
	return nil
}
