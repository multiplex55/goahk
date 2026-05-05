//go:build windows

package main

import (
	"fmt"
	"strings"

	"github.com/lxn/walk"
	"goahk/internal/inspect"
)

func (ui *viewerUI) executePatternAction(action string) {
	action = normalizePatternActionName(action)
	ui.SetBusy(true)
	go func() {
		var err error
		cancelled := false
		if action == "setValue" {
			_, accepted, invokeErr := ui.controller.InvokeSetValue()
			err = invokeErr
			cancelled = !accepted && invokeErr == nil
		} else {
			_, err = ui.controller.InvokePatternForSelection(action)
		}
		if cancelled {
			ui.dispatcher.Queue(func() {
				ui.SetBusy(false)
				ui.setStatus("action cancelled: " + callableActionLabel(action))
			})
			return
		}
		details, detailsErr := ui.controller.RefreshSelectedNodeDetails()
		ui.dispatcher.Queue(func() {
			ui.setLoading(false)
			if err != nil {
				ui.setStatus("action failed: " + err.Error())
				return
			}
			if detailsErr != nil {
				ui.setStatus("action succeeded, refresh failed: " + detailsErr.Error())
				return
			}
			ui.UpdateNodeDetails(details)
			ui.setStatus("action completed: " + callableActionLabel(action))
		})
	}()
}

// walkUIThread enforces the UI threading rule for the viewer:
// service/controller work runs off-thread, while all widget mutation runs on the Walk thread.
type walkUIThread struct{ mw *walk.MainWindow }

func (m *walkUIThread) Queue(fn func()) {
	if m == nil || m.mw == nil || fn == nil {
		return
	}
	m.mw.Synchronize(fn)
}

func (ui *viewerUI) queueRefreshWindowListFromCurrentFilters() {
	if ui == nil {
		return
	}
	ui.initialRefresh()
}

func (ui *viewerUI) filterNotImplementedStatus() string {
	return "Tree filtering not implemented yet"
}

func (ui *viewerUI) attachEvents() {
	ui.walkUIThread = ui.mw
	ui.dispatcher = &walkUIThread{mw: ui.walkUIThread}
	ui.windowModel = newWindowTableModel()
	ui.infoModel = newInfoTableModel()
	ui.propertiesModel = newPropertyTableModel()
	ui.treeModel = newUIATreeModel()
	ui.patternModel = newPatternTreeModel()
	if ui.windowTable != nil {
		ui.windowTable.SetModel(ui.windowModel)
		ui.windowTable.CurrentIndexChanged().Attach(func() {
			idx := ui.windowTable.CurrentIndex()
			row, ok := ui.windowModel.WindowAt(idx)
			if !ok {
				return
			}
			if ui.events != nil {
				ui.events.OnWindowSelected(row.ID, ui.activateOnSelect())
			}
		})
	}
	if ui.infoTable != nil {
		ui.infoTable.SetModel(ui.infoModel)
	}
	if ui.propertiesTV != nil {
		ui.propertiesTV.SetModel(ui.propertiesModel)
		ui.attachPropertyContextMenu()
		ui.propertiesTV.ItemActivated().Attach(func() {
			idx := ui.propertiesTV.CurrentIndex()
			row, ok := ui.propertiesModel.RowAt(idx)
			if !ok {
				return
			}
			ui.controller.CopyProperty(row.Value)
			ui.setStatus("copied property: " + row.Name)
		})
	}
	if ui.activateChk != nil {
		ui.activateChk.SetChecked(true)
	}
	if ui.patternsTree != nil {
		ui.patternsTree.SetModel(ui.patternModel)
		ui.attachPatternContextMenu()
		ui.patternsTree.ItemActivated().Attach(func() {
			if node, ok := ui.patternsTree.CurrentItem().(*patternTreeNode); ok {
				if action, ok := patternActionForNode(node); ok {
					ui.executePatternAction(action)
				}
			}
		})
	}
	if ui.treeView != nil {
		ui.treeView.SetModel(ui.treeModel)
		ui.treeView.ExpandedChanged().Attach(func(item walk.TreeItem) {
			if node, ok := item.(*uiaTreeNode); ok && ui.events != nil {
				ui.events.OnTreeExpanded(node.NodeID, ui.treeModel.AreChildrenLoaded(node.NodeID))
			}
		})
		ui.treeView.CurrentItemChanged().Attach(func() {
			if node, ok := ui.treeView.CurrentItem().(*uiaTreeNode); ok && ui.events != nil {
				ui.events.OnTreeSelected(node.NodeID)
			}
		})
	}
	if ui.refreshBtn != nil {
		ui.refreshBtn.Clicked().Attach(func() { ui.queueRefreshWindowListFromCurrentFilters() })
	}
	if ui.visibleChk != nil {
		ui.visibleChk.CheckedChanged().Attach(func() { ui.queueRefreshWindowListFromCurrentFilters() })
	}
	if ui.titleChk != nil {
		ui.titleChk.CheckedChanged().Attach(func() { ui.queueRefreshWindowListFromCurrentFilters() })
	}
	if ui.filterEdit != nil {
		ui.filterEdit.TextChanged().Attach(func() {
			ui.setStatus(ui.filterNotImplementedStatus())
		})
	}
	if ui.macroSidebarBtn != nil {
		ui.macroSidebarBtn.Clicked().Attach(func() {
			ui.setStatus("Macro sidebar is not implemented yet")
		})
	}
	if ui.statusBar != nil {
		ui.statusBar.MouseDown().Attach(func(_, _ int, _ walk.MouseButton) {
			update := ui.controller.OnStatusInteractionUpdate()
			ui.setStatus(update.Text)
		})
	}
}

func patternActionForNode(node *patternTreeNode) (string, bool) {
	if node == nil || !node.IsActionableLeaf() {
		return "", false
	}
	action := strings.TrimSpace(string(node.ActionID()))
	if action == "" {
		return "", false
	}
	return action, true
}

func patternNodeCopyText(node *patternTreeNode) string {
	if node == nil {
		return ""
	}
	return strings.TrimSpace(node.Text())
}

func (ui *viewerUI) attachPatternContextMenu() {
	if ui.patternsTree == nil {
		return
	}
	menu, _ := walk.NewMenu()
	copyAction := walk.NewAction()
	copyAction.SetText("Copy Pattern/Action Text")
	copyAction.Triggered().Attach(func() {
		node, ok := ui.patternsTree.CurrentItem().(*patternTreeNode)
		if !ok {
			return
		}
		text := patternNodeCopyText(node)
		if text == "" {
			return
		}
		ui.controller.CopyProperty(text)
		ui.setStatus("copied pattern/action text: " + text)
	})
	invokeAction := walk.NewAction()
	invokeAction.SetText("Invoke Action")
	invokeAction.Triggered().Attach(func() {
		node, ok := ui.patternsTree.CurrentItem().(*patternTreeNode)
		if !ok {
			return
		}
		action, ok := patternActionForNode(node)
		if !ok {
			return
		}
		ui.executePatternAction(action)
	})
	_ = menu.Actions().Add(copyAction)
	_ = menu.Actions().Add(invokeAction)
	ui.patternsTree.SetContextMenu(menu)
	ui.patternsTree.MouseDown().Attach(func(x, y int, button walk.MouseButton) {
		if button != walk.RightButton {
			return
		}
		item := ui.patternsTree.ItemAt(x, y)
		node, _ := item.(*patternTreeNode)
		if node != nil {
			ui.patternsTree.SetCurrentItem(node)
		}
		_, canInvoke := patternActionForNode(node)
		invokeAction.SetEnabled(canInvoke)
	})
}

func (ui *viewerUI) attachPropertyContextMenu() {
	if ui.propertiesTV == nil {
		return
	}
	menu, _ := walk.NewMenu()
	addCopy := func(title string, action func(row propertyTableRow) (copied, status string, ok bool)) {
		a := walk.NewAction()
		a.SetText(title)
		a.Triggered().Attach(func() {
			row, ok := ui.currentPropertyRow()
			if !ok {
				return
			}
			copied, status, ok := action(row)
			if !ok {
				return
			}
			ui.controller.CopyProperty(copied)
			ui.setStatus(status)
		})
		_ = menu.Actions().Add(a)
	}
	addCopy("Copy Value", func(row propertyTableRow) (string, string, bool) {
		return propertyContextCopyValue(row, true), "copied value: " + row.Name, true
	})
	addCopy("Copy Property Name", func(row propertyTableRow) (string, string, bool) {
		return row.Name, "copied property name: " + row.Name, true
	})
	addCopy("Copy Row", func(row propertyTableRow) (string, string, bool) {
		return row.Name + ": " + propertyContextCopyValue(row, true), "copied row: " + row.Name, true
	})
	ui.propertiesTV.SetContextMenu(menu)
	ui.propertiesTV.MouseDown().Attach(func(x, y int, button walk.MouseButton) {
		if button != walk.RightButton {
			return
		}
		idx := ui.propertiesTV.IndexAt(x, y)
		if idx >= 0 {
			ui.propertiesTV.SetCurrentIndex(idx)
		}
	})
}

func (ui *viewerUI) currentPropertyRow() (propertyTableRow, bool) {
	if ui.propertiesModel == nil || ui.propertiesTV == nil {
		return propertyTableRow{}, false
	}
	return ui.propertiesModel.RowAt(ui.propertiesTV.CurrentIndex())
}

func propertyContextCopyValue(row propertyTableRow, preferAHKControlType bool) string {
	if preferAHKControlType && row.Name == "ControlType" {
		if i := strings.Index(row.Value, "("); i >= 0 && strings.HasSuffix(strings.TrimSpace(row.Value), ")") {
			inside := strings.TrimSpace(strings.TrimSuffix(row.Value[i+1:], ")"))
			if inside != "" {
				return inside
			}
		}
	}
	return row.Value
}

func (ui *viewerUI) initialRefresh() {
	ui.SetBusy(true)
	go func() {
		visible, title := ui.defaultRefreshArgs()
		resp, err := ui.controller.RefreshWindows("", visible, title)
		ui.dispatcher.Queue(func() {
			ui.setLoading(false)
			if err != nil {
				ui.setStatus("refresh failed: " + err.Error())
				return
			}
			rows := mapWindowTableRows(resp.Windows, true)
			if ui.windowModel != nil {
				ui.windowModel.SetRows(rows)
			}
			ui.setStatus(fmt.Sprintf("loaded %d windows", len(rows)))
		})
	}()
}

func (ui *viewerUI) setLoading(loading bool) {
	if ui.refreshBtn != nil {
		ui.refreshBtn.SetEnabled(!loading)
	}
	if ui.windowTable != nil {
		ui.windowTable.SetEnabled(!loading)
	}
	if ui.treeView != nil {
		ui.treeView.SetEnabled(!loading)
	}
}

func (ui *viewerUI) SetBusy(b bool)     { ui.setLoading(b) }
func (ui *viewerUI) SetStatus(s string) { ui.setStatus(s) }
func (ui *viewerUI) UpdateWindowDetails(details inspect.GetNodeDetailsResponse) {
	ui.UpdateNodeDetails(details)
}
func (ui *viewerUI) UpdateNodeDetails(details inspect.GetNodeDetailsResponse) {
	if ui.infoModel != nil {
		ui.infoModel.SetRows(mapWindowInfoRows(&details))
	}
	if ui.propertiesModel != nil {
		ui.propertiesModel.SetRows(mapPropertyTableRows(details.Properties))
	}
	if ui.patternModel != nil {
		ui.patternModel.SetRoots(mapPatternTree(details.Patterns))
	}
}
func (ui *viewerUI) UpdateTreeRoot(root inspect.TreeNodeDTO) {
	if ui == nil || ui.treeModel == nil {
		return
	}

	ui.treeModel.SetRoot(root)
	if ui.treeView == nil {
		return
	}

	ui.treeView.SetModel(ui.treeModel)
	if item, ok := ui.treeModel.ItemByID(root.NodeID); ok {
		ui.treeView.SetCurrentItem(item)
	}
	ui.treeView.Invalidate()
}
func (ui *viewerUI) UpdateNodeChildren(nodeID string, children []inspect.TreeNodeDTO) {
	if ui.treeModel != nil {
		ui.treeModel.SetChildren(nodeID, children)
	}
}
func (ui *viewerUI) ExpandTreeNode(nodeID string) {
	if ui.treeModel != nil {
		ui.treeModel.SetExpanded(nodeID, true)
	}
}
func (ui *viewerUI) SelectTreeNode(nodeID string) {
	if ui == nil || ui.treeView == nil || ui.treeModel == nil {
		return
	}
	item, ok := ui.treeModel.ItemByID(nodeID)
	if !ok {
		return
	}
	ui.treeView.SetCurrentItem(item)
}
