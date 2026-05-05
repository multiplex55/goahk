package main

import (
	"fmt"
	"strings"

	"goahk/internal/inspect"
)

type UIThreadMarshaller interface{ Queue(func()) }

type ViewUpdater interface {
	SetBusy(bool)
	SetStatus(string)
	UpdateWindowDetails(inspect.GetNodeDetailsResponse)
	UpdateNodeDetails(inspect.GetNodeDetailsResponse)
	UpdateTreeRoot(inspect.TreeNodeDTO)
	UpdateNodeChildren(string, []inspect.TreeNodeDTO)
	ExpandTreeNode(string)
	SelectTreeNode(string)
}

type ViewerEventAdapter struct {
	controller *Controller
	view       ViewUpdater
	ui         UIThreadMarshaller
}

func NewViewerEventAdapter(controller *Controller, view ViewUpdater, ui UIThreadMarshaller) *ViewerEventAdapter {
	return &ViewerEventAdapter{controller: controller, view: view, ui: ui}
}

func (a *ViewerEventAdapter) OnWindowSelected(hwnd string, activate bool) {
	a.view.SetBusy(true)
	go func() {
		result, err := a.controller.SelectWindow(hwnd, activate)
		if err != nil {
			a.ui.Queue(func() {
				a.view.SetBusy(false)
				a.view.SetStatus("select window failed: select window: " + err.Error())
			})
			return
		}
		a.ui.Queue(func() {
			a.view.SetBusy(false)

			rootID := result.Root.Root.NodeID
			a.view.UpdateTreeRoot(result.Root.Root)
			a.view.SelectTreeNode(rootID)
			a.view.UpdateWindowDetails(result.Details)
			a.view.UpdateNodeDetails(result.Details)

			if len(result.Children) > 0 {
				a.view.UpdateNodeChildren(rootID, result.Children)
				a.view.ExpandTreeNode(rootID)
			}

			status := fmt.Sprintf("window loaded: properties=%d patterns=%d children=%d", len(result.Details.Properties), len(result.Details.Patterns), len(result.Children))
			warnings := make([]string, 0, 2)
			if result.ChildLoadErr != nil {
				warnings = append(warnings, "children: "+result.ChildLoadErr.Error())
			}
			if result.HighlightErr != nil {
				warnings = append(warnings, "highlight: "+result.HighlightErr.Error())
			}
			if len(warnings) > 0 {
				status += " (warning: " + strings.Join(warnings, "; ") + ")"
			}
			a.view.SetStatus(status)
		})
	}()
}

func (a *ViewerEventAdapter) OnTreeExpanded(nodeID string, loaded bool) {
	if loaded {
		return
	}
	a.view.SetBusy(true)
	go func() {
		resp, err := a.controller.ExpandNode(nodeID)
		a.ui.Queue(func() {
			a.view.SetBusy(false)
			if err != nil {
				a.view.SetStatus("expand node failed: " + err.Error())
				return
			}
			a.view.UpdateNodeChildren(nodeID, resp.Children)
			a.view.SelectTreeNode(nodeID)
			a.view.SetStatus("node expanded")
		})
	}()
}

func (a *ViewerEventAdapter) OnTreeSelected(nodeID string) {
	a.view.SetBusy(true)
	go func() {
		err := a.controller.SelectNode(nodeID)
		details, detailsErr := a.controller.RefreshSelectedNodeDetails()
		a.ui.Queue(func() {
			a.view.SetBusy(false)
			if err != nil {
				a.view.SetStatus("select node failed: " + err.Error())
				return
			}
			if detailsErr != nil {
				a.view.SetStatus("node details failed: " + detailsErr.Error())
				return
			}
			a.view.UpdateNodeDetails(details)
			a.view.SetStatus("node selected")
		})
	}()
}
