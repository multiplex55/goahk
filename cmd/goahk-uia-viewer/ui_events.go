package main

import (
	"errors"
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
	controller      *Controller
	view            ViewUpdater
	ui              UIThreadMarshaller
	autoExpandDepth func() int
}

func NewViewerEventAdapter(controller *Controller, view ViewUpdater, ui UIThreadMarshaller) *ViewerEventAdapter {
	return &ViewerEventAdapter{controller: controller, view: view, ui: ui, autoExpandDepth: func() int { return 3 }}
}

func (a *ViewerEventAdapter) OnWindowSelected(hwnd string, activate bool) {
	a.view.SetBusy(true)
	a.view.SetStatus("retrying transient root resolution...")
	go func() {
		result, err := a.controller.SelectWindow(hwnd, activate)
		var expandResults []TreeExpandResult
		if err == nil {
			expandResults = a.controller.ExpandTreeDepth(result.Root.Root.NodeID, a.autoExpandDepth())
		}
		if err != nil {
			log.Printf("uia.viewer on_window_selected_err hwnd=%s activate=%t err=%v", hwnd, activate, err)
			a.ui.Queue(func() {
				a.view.SetBusy(false)
				msg := formatFatal("InspectWindow", hwnd, err)
				a.view.SetStatus(msg)
				if shouldShowSelectionErrorModal(err) {
					a.view.ShowFatal(msg)
				}
			})
			return
		}
		a.ui.Queue(func() {
			a.view.SetBusy(false)

			rootID := result.Root.Root.NodeID
			a.view.UpdateTreeRoot(result.Root.Root)
			a.view.SelectTreeNode(rootID)
			hasDetails := result.DetailsErr == nil || result.Details.StatusText != ""
			if hasDetails {
				a.view.UpdateWindowDetails(result.Details)
				a.view.UpdateNodeDetails(result.Details)
			}

			if len(result.Children) > 0 {
				a.view.UpdateNodeChildren(rootID, result.Children)
				a.view.ExpandTreeNode(rootID)
			}
			for _, expanded := range expandResults {
				if expanded.Err == nil {
					a.view.UpdateNodeChildren(expanded.ParentID, expanded.Children)
					a.view.ExpandTreeNode(expanded.ParentID)
				}
			}

			status := fmt.Sprintf("window loaded %s: properties=%d patterns=%d children=%d", formatStageTarget("GetTreeRoot", hwnd), len(result.Details.Properties), len(result.Details.Patterns), len(result.Children))
			modeSummary := fmt.Sprintf("requested=%s active=%s provider=%s fallback=%t", result.Root.State.RequestedMode, result.Root.State.ActiveMode, result.Root.Source.Provider, result.Root.State.FallbackUsed)
			warnings := make([]string, 0, 6)
			for _, warning := range result.RootRetryWarnings {
				warnings = append(warnings, formatWarning("GetTreeRoot", hwnd, warning.Error()))
			}
			if result.DetailsErr != nil {
				warnings = append(warnings, formatWarning("GetNodeDetails", rootID, result.DetailsErr.Error()))
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
				status = "loaded with warning: " + strings.Join(warnings, "; ")
			}
			if result.DetailsErr != nil && result.Details.StatusText != "" {
				status = status + "; " + result.Details.StatusText
			}
			if result.Root.State.FallbackUsed {
				status = status + "; fallback mode active: degraded HWND/compatibility tree, selector parity may differ (" + modeSummary + ")"
			} else {
				status = status + "; mode: " + modeSummary
			}
			switch strings.ToLower(strings.TrimSpace(result.Root.Source.Backend)) {
			case "native-com":
				status += "; source: native-com UIA success"
			case "synthetic":
				status += "; source: synthetic UIA/compatibility tree"
			case "unavailable":
				status += "; source: fallback to HWND/ACC"
			}
			if result.Root.State.RequestedMode == inspect.InspectModeUIAOnly && result.Root.State.ActiveMode != inspect.InspectModeUIAOnly {
				status += "; UIA-only failure: fallback disabled"
			}
			if shallowTreeWarning(result) != "" {
				status += "; warning: " + shallowTreeWarning(result)
			}

			a.view.SetStatus(status)
			log.Printf("uia.viewer ui_update_done hwnd=%s root_node=%s properties=%d patterns=%d children=%d", hwnd, rootID, len(result.Details.Properties), len(result.Details.Patterns), len(result.Children))
		})
	}()
}

func shallowTreeWarning(result WindowSelectionResult) string {
	if len(result.Children) == 0 {
		return "tree appears shallow (no root children)"
	}
	if strings.Contains(strings.ToLower(result.Root.Root.DisplayLabel), "placeholder") {
		return "tree appears synthetic placeholder content"
	}
	if (result.Root.State.RequestedMode == inspect.InspectModeAuto || result.Root.State.RequestedMode == inspect.InspectModeUIATree || result.Root.State.RequestedMode == inspect.InspectModeUIAOnly) &&
		strings.ToLower(strings.TrimSpace(result.Root.Source.Backend)) != "native-com" {
		return "UIA mode resolved to non-native backend"
	}
	return ""
}

func shouldShowSelectionErrorModal(err error) bool {
	if err == nil {
		return false
	}
	return !errors.Is(err, inspect.ErrTransientFailure) && !isTransientInspectError(err)
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
				return
			}
			if detailsErr != nil {
				msg := formatFatal("GetNodeDetails", nodeID, detailsErr)
				a.view.SetStatus(msg)
				return
			}
			a.view.UpdateNodeDetails(details)
			a.view.SetStatus("node selected " + formatStageTarget("GetNodeDetails", nodeID))
		})
	}()
}
