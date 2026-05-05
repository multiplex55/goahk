package main

import (
	"fmt"
	"log"
	"strings"

	"goahk/internal/inspect"
)

type UIThreadMarshaller interface{ Queue(func()) }

type ViewUpdater interface {
	SetBusy(bool)
	SetStatus(string)
	ShowFatal(string)
	UpdateWindowDetails(inspect.GetNodeDetailsResponse)
	UpdateNodeDetails(inspect.GetNodeDetailsResponse)
	UpdateTreeRoot(inspect.TreeNodeDTO)
	UpdateNodeChildren(string, []inspect.TreeNodeDTO)
	ExpandTreeNode(string)
	SelectTreeNode(string)
}

func formatStageTarget(stage, target string) string {
	return fmt.Sprintf("%s [%s]", stage, target)
}

func formatFatal(stage, target string, err error) string {
	return fmt.Sprintf("ERROR %s: %s", formatStageTarget(stage, target), err.Error())
}

func formatWarning(stage, target string, warning string) string {
	return fmt.Sprintf("WARNING %s: %s", formatStageTarget(stage, target), warning)
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
				msg := formatFatal("InspectWindow", hwnd, err)
				a.view.SetStatus(msg)
				a.view.ShowFatal(msg)
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

			status := fmt.Sprintf("window loaded %s: properties=%d patterns=%d children=%d", formatStageTarget("GetTreeRoot", hwnd), len(result.Details.Properties), len(result.Details.Patterns), len(result.Children))
			warnings := make([]string, 0, 6)
			for _, warning := range result.RetryWarnings {
				warnings = append(warnings, formatWarning("GetTreeRoot", hwnd, warning))
			}
			if result.ChildLoadErr != nil {
				warnings = append(warnings, formatWarning("GetNodeChildren", rootID, result.ChildLoadErr.Error()))
			}
			if result.SelectErr != nil {
				warnings = append(warnings, formatWarning("SelectNode", rootID, result.SelectErr.Error()))
			}
			if result.HighlightErr != nil {
				warnings = append(warnings, formatWarning("HighlightNode", rootID, result.HighlightErr.Error()))
			}
			if len(warnings) > 0 {
				status = strings.Join(warnings, "; ")
			}

			a.view.SetStatus(status)
			log.Printf("uia.viewer ui_update_done hwnd=%s root_node=%s properties=%d patterns=%d children=%d", hwnd, rootID, len(result.Details.Properties), len(result.Details.Patterns), len(result.Children))
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
				msg := formatFatal("GetTreeRoot", nodeID, err)
				a.view.SetStatus(msg)
				a.view.ShowFatal(msg)
				return
			}
			a.view.UpdateNodeChildren(nodeID, resp.Children)
			a.view.SelectTreeNode(nodeID)
			a.view.SetStatus("node expanded " + formatStageTarget("GetTreeRoot", nodeID))
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
				msg := formatFatal("InspectWindow", nodeID, err)
				a.view.SetStatus(msg)
				a.view.ShowFatal(msg)
				return
			}
			if detailsErr != nil {
				msg := formatFatal("GetNodeDetails", nodeID, detailsErr)
				a.view.SetStatus(msg)
				a.view.ShowFatal(msg)
				return
			}
			a.view.UpdateNodeDetails(details)
			a.view.SetStatus("node selected " + formatStageTarget("GetNodeDetails", nodeID))
		})
	}()
}
