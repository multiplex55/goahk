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
		if action == "setValue" {
			_, err = ui.controller.InvokeSetValue()
		} else {
			_, err = ui.controller.InvokePatternForSelection(action)
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

type walkUIThread struct{ mw *walk.MainWindow }

func (m walkUIThread) Queue(fn func()) {
	if m.mw == nil {
		return
	}
	m.mw.Synchronize(fn)
}

func (ui *viewerUI) attachEvents() {
	ui.dispatcher = walkUIThread{mw: ui.mw}
	ui.windowModel = newWindowTableModel()
	if ui.windowTable != nil {
		ui.windowTable.SetModel(ui.windowModel)
		ui.windowTable.CurrentIndexChanged().Attach(func() {
			idx := ui.windowTable.CurrentIndex()
			row, ok := ui.windowModel.WindowAt(idx)
			if !ok {
				return
			}
			if ui.events != nil {
				ui.events.OnWindowSelected(row.ID)
			}
		})
	}
	if ui.activateChk != nil {
		ui.activateChk.SetChecked(true)
	}
	ui.refreshBtn.Clicked().Attach(func() { ui.initialRefresh() })
	if ui.statusBar != nil {
		ui.statusBar.MouseDown().Attach(func(_, _ int, _ walk.MouseButton) {
			ui.setStatus(ui.controller.OnStatusInteraction())
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
	if ui.infoView != nil {
		ui.infoView.SetText(formatSelectedInfo(&details))
	}
}
func (ui *viewerUI) UpdateNodeDetails(details inspect.GetNodeDetailsResponse) {
	if ui.infoView != nil {
		ui.infoView.SetText(formatSelectedInfo(&details))
	}
}
func (ui *viewerUI) UpdateTreeRoot(inspect.TreeNodeDTO)               {}
func (ui *viewerUI) UpdateNodeChildren(string, []inspect.TreeNodeDTO) {}

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
