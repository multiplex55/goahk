//go:build windows

package main

import "github.com/lxn/walk"

type viewerUI struct {
	mw *walk.MainWindow

	windowTable *walk.TableView
	refreshBtn  *walk.PushButton
	visibleChk  *walk.CheckBox
	titleChk    *walk.CheckBox
	activateChk *walk.CheckBox

	infoView     *walk.TextEdit
	propertiesTV *walk.TableView
	patternsTV   *walk.TableView
	treeView     *walk.TreeView

	statusBar  *walk.StatusBar
	statusText *walk.StatusBarItem
}
