//go:build windows

package main

import (
	"fmt"

	"github.com/lxn/walk"
)

type viewerUI struct {
	controller                   *Controller
	events                       *ViewerEventAdapter
	dispatcher                   UIThreadMarshaller
	walkUIThread                 *walk.MainWindow
	suppressTreeSelectionEvent   bool
	suppressTreeExpandEvent      bool
	suppressWindowSelectionEvent bool
	mw                           *walk.MainWindow
	root                         *walk.Composite

	windowTable       *walk.TableView
	windowModel       *windowTableModel
	refreshBtn        *walk.PushButton
	refreshTreeBtn    *walk.PushButton
	modeCombo         *walk.ComboBox
	autoExpandCombo   *walk.ComboBox
	preserveExpandChk *walk.CheckBox
	visibleChk        *walk.CheckBox
	titleChk          *walk.CheckBox
	activateChk       *walk.CheckBox

	filterLbl       *walk.Label
	filterEdit      *walk.LineEdit
	macroSidebarBtn *walk.PushButton

	infoTable          *walk.TableView
	infoModel          *infoTableModel
	propertiesTV       *walk.TableView
	propertiesModel    *propertyTableModel
	patternsTree       *walk.TreeView
	patternModel       *patternTreeModel
	treeView           *walk.TreeView
	treeModel          *uiaTreeModel
	preFilterExpansion *TreeExpansionSnapshot
	lastFilterText     string

	statusBar  *walk.StatusBar
	statusText *walk.StatusBarItem
}

func (ui *viewerUI) autoExpandDepth() int {
	if ui == nil || ui.autoExpandCombo == nil {
		return 2
	}
	switch ui.autoExpandCombo.CurrentIndex() {
	case 0:
		return 0
	case 1:
		return 1
	case 3:
		return 3
	default:
		return 2
	}
}

func (ui *viewerUI) isRecursiveExpandMode() bool {
	return ui != nil && ui.autoExpandCombo != nil && ui.autoExpandCombo.CurrentIndex() == 4
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
	ui.events.autoExpandDepth = ui.autoExpandDepth
	ui.events.isRecursiveMode = ui.isRecursiveExpandMode
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
