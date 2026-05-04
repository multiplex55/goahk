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
			ui.SetBusy(false)
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

func (ui *viewerUI) attachEvents() {
	ui.walkUIThread = ui.mw
	ui.dispatcher = &walkUIThread{mw: ui.walkUIThread}
	ui.windowModel = newWindowTableModel()
	ui.propertiesModel = newPropertyTableModel()
	ui.treeModel = newUIATreeModel()
	ui.patternByLabel = map[string]string{}
	ui.nodeByLabel = map[string]string{}
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
	if ui.propertiesTV != nil {
		ui.propertiesTV.SetModel(ui.propertiesModel)
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
		ui.patternsTree.ItemActivated().Attach(func() {
			if item := ui.patternsTree.CurrentItem(); item != nil {
				action := ui.patternByLabel[item.Text()]
				if strings.TrimSpace(action) != "" {
					ui.executePatternAction(action)
				}
			}
		})
	}
	if ui.treeView != nil {
		ui.treeView.ExpandedChanged().Attach(func(item walk.TreeItem) {
			if item != nil {
				nodeID := ui.nodeByLabel[item.Text()]
				if nodeID != "" && ui.events != nil {
					ui.events.OnTreeExpanded(nodeID, ui.treeModel.AreChildrenLoaded(nodeID))
				}
			}
		})
		ui.treeView.CurrentItemChanged().Attach(func() {
			if item := ui.treeView.CurrentItem(); item != nil {
				nodeID := ui.nodeByLabel[item.Text()]
				if nodeID != "" && ui.events != nil {
					ui.events.OnTreeSelected(nodeID)
				}
			}
		})
	}
	ui.refreshBtn.Clicked().Attach(func() { ui.initialRefresh() })
	if ui.statusBar != nil {
		ui.statusBar.MouseDown().Attach(func(_, _ int, _ walk.MouseButton) {
			update := ui.controller.OnStatusInteractionUpdate()
			ui.setStatus(update.Text)
		})
	}
}

func (ui *viewerUI) initialRefresh() {
	ui.SetBusy(true)
	go func() {
		visible, title := ui.defaultRefreshArgs()
		resp, err := ui.controller.RefreshWindows("", visible, title)
		ui.dispatcher.Queue(func() {
			ui.SetBusy(false)
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
	ui.dispatcher.Queue(func() {
		if ui.refreshBtn != nil {
			ui.refreshBtn.SetEnabled(!loading)
		}
		if ui.windowTable != nil {
			ui.windowTable.SetEnabled(!loading)
		}
		if ui.treeView != nil {
			ui.treeView.SetEnabled(!loading)
		}
	})
}

func (ui *viewerUI) SetBusy(b bool)     { ui.setLoading(b) }
func (ui *viewerUI) SetStatus(s string) { ui.setStatus(s) }
func (ui *viewerUI) UpdateWindowDetails(details inspect.GetNodeDetailsResponse) {
	ui.UpdateNodeDetails(details)
}
func (ui *viewerUI) UpdateNodeDetails(details inspect.GetNodeDetailsResponse) {
	if ui.infoView != nil {
		ui.infoView.SetText(formatSelectedInfo(&details))
	}
	if ui.propertiesModel != nil {
		ui.propertiesModel.SetRows(mapPropertyTableRows(details.Properties))
	}
	if ui.patternsTree != nil {
		ui.patternByLabel = map[string]string{}
		for _, group := range mapPatternTree(details.Patterns) {
			for _, child := range group.Children {
				ui.patternByLabel[child.Label] = child.ActionID
			}
		}
	}
}
func (ui *viewerUI) UpdateTreeRoot(root inspect.TreeNodeDTO) {
	if ui.treeModel != nil {
		ui.treeModel.SetRoot(root)
	}
	if ui.treeView != nil {
		ui.nodeByLabel = map[string]string{ui.treeModel.Label(root): root.NodeID}
	}
}
func (ui *viewerUI) UpdateNodeChildren(nodeID string, children []inspect.TreeNodeDTO) {
	if ui.treeModel != nil {
		ui.treeModel.SetChildren(nodeID, children)
		for _, child := range children {
			ui.nodeByLabel[ui.treeModel.Label(child)] = child.NodeID
		}
	}
}

func formatSelectedInfo(details *inspect.GetNodeDetailsResponse) string {
	if details == nil {
		return "No selection"
	}
	lines := []string{
		"Window:",
		"  Title: " + fallback(details.WindowInfo.Title),
		"  Process: " + fallback(details.WindowInfo.Process),
		"  PID: " + intOrNA(details.WindowInfo.PID),
		"  HWND: " + fallback(details.WindowInfo.HWND),
		"  Class: " + fallback(details.WindowInfo.Class),
		"",
		"Element:",
		"  NodeID: " + fallback(details.Element.NodeID),
		"  Name: " + fallback(details.Element.Name),
		"  ControlType: " + fallback(details.Element.ControlType),
		"  LocalizedControlType: " + fallback(details.Element.LocalizedControlType),
	}
	if strings.TrimSpace(details.ACCPath) != "" {
		lines = append(lines, "", "ACC Path: "+details.ACCPath)
	}
	return strings.Join(lines, "\n")
}

func fallback(v string) string {
	if strings.TrimSpace(v) == "" {
		return "N/A"
	}
	return v
}

func intOrNA(v int) string {
	if v <= 0 {
		return "N/A"
	}
	return fmt.Sprintf("%d", v)
}
