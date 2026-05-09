//go:build windows

package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/lxn/walk"
	"goahk/internal/inspect"
)

func inspectModeFromComboIndex(idx int) inspect.InspectMode {
	if idx == 1 {
		return inspect.InspectModeUIAOnly
	}
	if idx == 2 {
		return inspect.InspectModeWindowTree
	}
	if idx == 3 {
		return inspect.InspectModeHWNDTree
	}
	return inspect.InspectModeAuto
}

func allowFallbackFromComboIndex(idx int) bool {
	return idx == 0
}

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
		details, detailsErr := ui.controller.RefreshSelectionDetails()
		ui.dispatcher.Queue(func() {
			ui.setLoading(false)
			ui.applyPostInvokeDetailsOnly(action, err, details, detailsErr)
		})
	}()
}

func (ui *viewerUI) applyPostInvokeDetailsOnly(action string, invokeErr error, details inspect.GetNodeDetailsResponse, detailsErr error) {
	if invokeErr != nil {
		ui.setStatus("action failed: " + invokeErr.Error())
		return
	}
	if detailsErr != nil {
		ui.setStatus("action succeeded, refresh failed: " + detailsErr.Error())
		return
	}
	// Post-invoke non-refresh path is intentionally non-mutating for the tree model:
	// no UpdateTreeRoot, UpdateNodeChildren, or ExpandTreeNode calls here.
	ui.UpdateWindowDetails(details)
	ui.UpdateNodeDetails(details)
	ui.setStatus(fmt.Sprintf("action completed: %s (patterns=%d)", callableActionLabel(action), len(details.Patterns)))
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

func (ui *viewerUI) preserveExpansionEnabled() bool {
	if ui == nil || ui.preserveExpandChk == nil {
		return true
	}
	return ui.preserveExpandChk.Checked()
}

func (ui *viewerUI) applyFilterTransition(filterText string) {
	if ui == nil || ui.treeModel == nil {
		return
	}
	next := strings.TrimSpace(filterText)
	prev := strings.TrimSpace(ui.lastFilterText)
	wasFiltered := prev != ""
	isFiltered := next != ""
	if !wasFiltered && isFiltered {
		ui.preFilterExpansion = ui.treeModel.SnapshotExpansion()
		ui.preFilterExpansion.SelectedID = ui.currentSelectedTreeNodeID()
		return
	}
	if wasFiltered && isFiltered {
		ui.preFilterExpansion = ui.treeModel.SnapshotExpansion()
		ui.expandFilterMatchAncestors()
		return
	}
	if wasFiltered && !isFiltered {
		if ui.preFilterExpansion != nil {
			ui.treeModel.RestoreExpansion(ui.preFilterExpansion)
			ui.restoreVisualExpansion(ui.preFilterExpansion)
		}
		ui.preFilterExpansion = nil
	}
	ui.lastFilterText = next
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
			if ui.suppressWindowSelectionEvent {
				return
			}
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
			if ui.suppressTreeExpandEvent {
				return
			}
			if node, ok := item.(*uiaTreeNode); ok && ui.events != nil {
				if ui.treeView.Expanded(node) {
					ui.treeModel.MarkExpanded(node.NodeID)
				} else {
					ui.treeModel.MarkCollapsed(node.NodeID)
				}
				ui.events.OnTreeExpanded(node.NodeID, ui.treeModel.AreChildrenLoaded(node.NodeID))
			}
		})
		ui.treeView.CurrentItemChanged().Attach(func() {
			if ui.suppressTreeSelectionEvent {
				return
			}
			if node, ok := ui.treeView.CurrentItem().(*uiaTreeNode); ok && ui.events != nil {
				ui.events.OnTreeSelected(node.NodeID)
			}
		})
	}
	if ui.refreshBtn != nil {
		ui.refreshBtn.Clicked().Attach(func() { ui.queueRefreshWindowListFromCurrentFilters() })
	}
	if ui.refreshTreeBtn != nil {
		ui.refreshTreeBtn.Clicked().Attach(func() {
			var snapshot *TreeExpansionSnapshot
			if ui.preserveExpansionEnabled() && ui.treeModel != nil {
				snapshot = ui.treeModel.SnapshotExpansion()
				snapshot.SelectedID = ui.currentSelectedTreeNodeID()
			}
			if ui.events != nil {
				ui.events.OnRefreshTreeRequested(ui.activateOnSelect())
			}
			if snapshot != nil {
				ui.preFilterExpansion = snapshot
			}
		})
	}
	if ui.modeCombo != nil {
		ui.controller.SetMode(inspectModeFromComboIndex(ui.modeCombo.CurrentIndex()))
		ui.controller.SetAllowFallback(allowFallbackFromComboIndex(ui.modeCombo.CurrentIndex()))
		ui.modeCombo.CurrentIndexChanged().Attach(func() {
			idx := ui.modeCombo.CurrentIndex()
			mode := inspectModeFromComboIndex(idx)
			ui.controller.SetMode(mode)
			ui.controller.SetAllowFallback(allowFallbackFromComboIndex(idx))
			ui.setStatus("inspect mode set: " + string(mode))
		})
	}
	if ui.visibleChk != nil {
		ui.visibleChk.CheckedChanged().Attach(func() { ui.queueRefreshWindowListFromCurrentFilters() })
	}
	if ui.titleChk != nil {
		ui.titleChk.CheckedChanged().Attach(func() { ui.queueRefreshWindowListFromCurrentFilters() })
	}
	if ui.filterEdit != nil {
		ui.filterEdit.TextChanged().Attach(func() {
			ui.applyFilterTransition(ui.filterEdit.Text())
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

func windowRowIndexByID(rows []windowTableRow, hwnd string) int {
	target := strings.ToLower(strings.TrimSpace(hwnd))
	if target == "" {
		return -1
	}
	for i, row := range rows {
		if strings.ToLower(strings.TrimSpace(row.ID)) == target {
			return i
		}
	}
	return -1
}

func (ui *viewerUI) initialRefresh() {
	ui.SetBusy(true)
	go func() {
		visible, title := ui.defaultRefreshArgs()
		resp, err := ui.controller.RefreshWindowList("", visible, title)
		ui.dispatcher.Queue(func() {
			ui.setLoading(false)
			if err != nil {
				ui.setStatus("ERROR " + formatStageTarget("RefreshWindows", "window-table") + ": " + err.Error())
				return
			}
			rows := mapWindowTableRows(resp.Windows, true)
			if ui.windowModel != nil {
				selectedHWND := ""
				if ui.windowTable != nil {
					if current, ok := ui.windowModel.WindowAt(ui.windowTable.CurrentIndex()); ok {
						selectedHWND = current.ID
					}
				}
				ui.suppressWindowSelectionEvent = true
				defer func() { ui.suppressWindowSelectionEvent = false }()
				ui.windowModel.SetRows(rows)
				if ui.windowTable != nil {
					if idx := windowRowIndexByID(rows, selectedHWND); idx >= 0 {
						ui.windowTable.SetCurrentIndex(idx)
					}
				}
			}
			ui.setStatus(fmt.Sprintf("loaded %d windows %s", len(rows), formatStageTarget("RefreshWindows", "window-table")))
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
	log.Printf("uia.viewer render_node_details properties=%d patterns=%d", len(details.Properties), len(details.Patterns))
	if ui.infoModel != nil {
		ui.infoModel.SetRows(mapWindowInfoRows(&details))
	}
	if ui.propertiesModel != nil {
		ui.propertiesModel.SetRows(mapPropertyRowsFromDetails(details))
	}
	if ui.patternModel != nil {
		ui.patternModel.SetRoots(mapPatternTree(details.Patterns))
	}
}
func (ui *viewerUI) UpdateTreeRoot(root inspect.TreeNodeDTO) {
	log.Printf("uia.viewer render_tree_root node=%s", root.NodeID)
	if ui == nil || ui.treeModel == nil {
		return
	}

	ui.treeModel.SetRoot(root)
	if ui.preFilterExpansion != nil && ui.preserveExpansionEnabled() {
		ui.treeModel.RestoreExpansion(ui.preFilterExpansion)
	}
	if ui.treeView == nil {
		return
	}

	if item, ok := ui.treeModel.ItemByID(root.NodeID); ok {
		ui.suppressTreeSelectionEvent = true
		defer func() { ui.suppressTreeSelectionEvent = false }()
		ui.treeView.SetCurrentItem(item)
	}
	if ui.preFilterExpansion != nil && ui.preserveExpansionEnabled() {
		ui.restoreVisualExpansion(ui.preFilterExpansion)
	}
	ui.treeView.Invalidate()
}
func (ui *viewerUI) UpdateNodeChildren(nodeID string, children []inspect.TreeNodeDTO) {
	log.Printf("uia.viewer render_node_children node=%s children=%d", nodeID, len(children))
	if ui.treeModel != nil {
		ui.treeModel.SetChildren(nodeID, children)
	}
}

func (ui *viewerUI) UpdateTreeBatch(results []TreeExpandResult) {
	if ui == nil || ui.treeModel == nil || len(results) == 0 {
		return
	}
	withTreeRedrawSuspended(ui, func() {
		for _, result := range results {
			if result.Err != nil {
				continue
			}
			ui.treeModel.SetChildren(result.ParentID, result.Children)
		}
		for _, result := range results {
			if result.Err != nil {
				continue
			}
			if ui.ShouldAutoExpand(result.ParentID) {
				ui.expandTreeNode(result.ParentID, false)
			}
		}
	})
}
func (ui *viewerUI) ExpandTreeNode(nodeID string) {
	ui.expandTreeNode(nodeID, false)
}

func (ui *viewerUI) ShouldAutoExpand(nodeID string) bool {
	if ui == nil || ui.treeModel == nil {
		return true
	}
	return ui.treeModel.ShouldAutoExpand(nodeID)
}

func (ui *viewerUI) expandTreeNode(nodeID string, userInitiated bool) {
	if ui == nil || ui.treeModel == nil {
		return
	}
	if !userInitiated && !ui.treeModel.ShouldAutoExpand(nodeID) {
		return
	}
	node, ok := ui.treeModel.ItemByID(nodeID)
	if !ok {
		return
	}
	if userInitiated {
		ui.treeModel.MarkExpanded(nodeID)
	} else {
		ui.treeModel.SetExpanded(nodeID, true)
	}
	if ui.treeView != nil {
		ui.suppressTreeExpandEvent = true
		defer func() { ui.suppressTreeExpandEvent = false }()
		if err := ui.treeView.SetExpanded(node, true); err != nil {
			log.Printf("uia.viewer expand_tree_node node=%s err=%v", nodeID, err)
		}
		ui.treeView.Invalidate()
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
	ui.suppressTreeSelectionEvent = true
	defer func() { ui.suppressTreeSelectionEvent = false }()
	ui.treeView.SetCurrentItem(item)
}

func (ui *viewerUI) currentSelectedTreeNodeID() string {
	if ui == nil || ui.treeView == nil {
		return ""
	}
	if node, ok := ui.treeView.CurrentItem().(*uiaTreeNode); ok && node != nil {
		return node.NodeID
	}
	return ""
}

func (ui *viewerUI) restoreVisualExpansion(snapshot *TreeExpansionSnapshot) {
	if ui == nil || ui.treeView == nil || ui.treeModel == nil || snapshot == nil {
		return
	}
	ui.suppressTreeExpandEvent = true
	defer func() { ui.suppressTreeExpandEvent = false }()
	for _, id := range snapshot.ExpandedIDs {
		node, ok := ui.treeModel.ItemByID(id)
		if !ok {
			continue
		}
		_ = ui.treeView.SetExpanded(node, true)
	}
	if snapshot.SelectedID != "" {
		if node, ok := ui.treeModel.ItemByID(snapshot.SelectedID); ok {
			ui.suppressTreeSelectionEvent = true
			ui.treeView.SetCurrentItem(node)
			ui.suppressTreeSelectionEvent = false
		}
	}
}

func (ui *viewerUI) expandFilterMatchAncestors() {
	// Placeholder for full filtered projection behavior; keep user-visible expansion continuity.
	if ui.preFilterExpansion != nil {
		ui.restoreVisualExpansion(ui.preFilterExpansion)
	}
}
