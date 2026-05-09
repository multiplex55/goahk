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
	isRecursiveMode func() bool
}

func NewViewerEventAdapter(controller *Controller, view ViewUpdater, ui UIThreadMarshaller) *ViewerEventAdapter {
	gates := inspect.GetUIAFeatureGates()
	depth := gates.MaxInitialDepth
	if gates.DisableAutoExpand {
		depth = 1
	}
	return &ViewerEventAdapter{controller: controller, view: view, ui: ui, autoExpandDepth: func() int { return depth }, isRecursiveMode: func() bool { return false }}
}

func (a *ViewerEventAdapter) OnWindowSelected(hwnd string, activate bool) {
	a.onWindowSelectionRequest(func() (WindowSelectionResult, error) {
		return a.controller.SelectWindowWithOptions(hwnd, SelectWindowOptions{
			Activate:        activate,
			ForceTreeReload: false,
			Reason:          "window list/table selection",
		})
	}, hwnd)
}

func (a *ViewerEventAdapter) OnRefreshTreeRequested(activate bool) {
	a.onWindowSelectionRequest(func() (WindowSelectionResult, error) {
		return a.controller.RefreshCurrentTree(activate)
	}, "current")
}

func (a *ViewerEventAdapter) onWindowSelectionRequest(load func() (WindowSelectionResult, error), target string) {
	a.view.SetBusy(true)
	a.view.SetStatus("retrying transient root resolution...")
	go func() {
		result, err := load()
		var expandResults []TreeExpandResult
		if err == nil {
			if a.isRecursiveMode != nil && a.isRecursiveMode() {
				g := inspect.GetUIAFeatureGates()
				expandResults = a.controller.PopulateTreeFromRoot(result.Root.Root.NodeID, TreePopulateOptions{
					MaxDepth:        g.MaxInitialDepth,
					MaxNodes:        g.MaxInitialNodes,
					BranchTimeout:   g.BranchTimeout,
					TotalTimeout:    g.TotalLoadTimeout,
					ContinueOnError: true,
				})
			} else {
				expandResults = a.controller.ExpandTreeDepthFromChildren(result.Children, a.autoExpandDepth()-1)
			}
		}
		if err != nil {
			log.Printf("uia.viewer on_window_selected_err target=%s err=%v", target, err)
			a.ui.Queue(func() {
				a.view.SetBusy(false)
				msg := formatFatal("InspectWindow", target, err)
				a.view.SetStatus(msg)
				if shouldShowSelectionErrorModal(err) {
					a.view.ShowFatal(msg)
				}
			})
			return
		}
		a.ui.Queue(func() {
			a.view.SetBusy(false)
			if result.NoOp {
				a.view.SetStatus("selected window unchanged; tree preserved")
				return
			}

			rootID := result.Root.Root.NodeID
			rootNode := result.Root.Root
			if result.Root.State.FallbackUsed && strings.TrimSpace(rootNode.DisplayLabel) != "" {
				rootNode.DisplayLabel += " [ACC/MSAA fallback]"
			}
			a.view.UpdateTreeRoot(rootNode)
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

			status := fmt.Sprintf("window loaded %s: properties=%d patterns=%d children=%d", formatStageTarget("GetTreeRoot", target), len(result.Details.Properties), len(result.Details.Patterns), len(result.Children))
			modeSummary := fmt.Sprintf(
				"requested=%s active=%s provider=%s backend=%s fallback=%t",
				result.Root.State.RequestedMode,
				result.Root.State.ActiveMode,
				result.Root.State.Provider,
				result.Root.State.Backend,
				result.Root.State.FallbackUsed,
			)
			warnings := make([]string, 0, 6)
			for _, warning := range result.RootRetryWarnings {
				warnings = append(warnings, formatWarning("GetTreeRoot", target, warning.Error()))
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
			isUIAOnly := result.Root.State.RequestedMode == inspect.InspectModeUIAOnly
			if len(warnings) > 0 {
				status = "loaded with traversal warnings: " + strings.Join(warnings, "; ")
			}
			if result.ChildLoadErr != nil {
				status = "UIA root loaded, but child traversal failed: " + result.ChildLoadErr.Error()
				if isUIAOnly {
					status = "ERROR " + status
				}
			}
			if result.DetailsErr != nil && result.Details.StatusText != "" {
				status = status + "; " + result.Details.StatusText
			}
			didFallbackSwitch := result.Root.State.RequestedMode != result.Root.State.ActiveMode
			parity := "parity preconditions satisfied"
			if result.Root.State.FallbackUsed || (result.Root.State.RequestedMode != inspect.InspectModeUIATree && result.Root.State.RequestedMode != inspect.InspectModeUIAOnly) || strings.ToLower(strings.TrimSpace(result.Root.State.Provider)) != "uia" || strings.ToLower(strings.TrimSpace(result.Root.State.Backend)) != "native-com" {
				parity = "parity preconditions not met"
			}
			status += "; parity: " + parity + " (" + modeSummary + ")"
			if result.Root.State.FallbackUsed && didFallbackSwitch {
				status = fmt.Sprintf(
					"DEGRADED TREE: requested %s but active %s; %s",
					result.Root.State.RequestedMode,
					result.Root.State.ActiveMode,
					status,
				)
				status = status + "; degrade reason: " + strings.TrimSpace(result.Root.State.DegradeReason)
			}
			switch strings.ToLower(strings.TrimSpace(result.Root.Source.Backend)) {
			case "native-com":
				status += "; source: native-com UIA success"
			case "native-msaa":
				status += "; source: native-msaa compatibility tree"
			case "hwnd":
				status += "; source: HWND compatibility tree"
			case "unavailable":
				status += "; source: fallback to HWND/ACC"
			}
			if isUIAOnly && result.ChildLoadErr != nil {
				status += "; UIA-only failure: child traversal failed"
			}
			if isUIAOnly && result.Root.State.ActiveMode != inspect.InspectModeUIAOnly {
				status += "; UIA-only failure: fallback disabled"
			}
			if shallowTreeWarning(result) != "" {
				status += "; warning: " + shallowTreeWarning(result)
			}
			branchFailures := 0
			for _, expanded := range expandResults {
				if expanded.Err != nil {
					branchFailures++
					errText := strings.ToLower(expanded.Err.Error())
					switch {
					case strings.Contains(errText, "budget reached"):
						status += "; budget reached"
					case strings.Contains(errText, "max depth reached"):
						status += "; max depth reached"
					case strings.Contains(errText, "timeout"):
						status += "; timeout reached"
					}
				}
			}
			if branchFailures > 0 {
				status += fmt.Sprintf("; warnings: %d branch failures", branchFailures)
			}

			a.view.SetStatus(status)
			log.Printf("uia.viewer ui_update_done target=%s root_node=%s properties=%d patterns=%d children=%d", target, rootID, len(result.Details.Properties), len(result.Details.Patterns), len(result.Children))
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
		resp, err := a.controller.RefreshNodeChildren(nodeID)
		a.ui.Queue(func() {
			a.view.SetBusy(false)
			if err != nil {
				msg := formatFatal("GetNodeChildren", nodeID, err)
				a.view.SetStatus(msg)
				return
			}
			a.view.UpdateNodeChildren(nodeID, resp.Children)
			a.view.SelectTreeNode(nodeID)
			a.view.SetStatus("node expanded " + formatStageTarget("GetNodeChildren", nodeID))
		})
	}()
}

func (a *ViewerEventAdapter) OnTreeSelected(nodeID string) {
	a.view.SetBusy(true)
	go func() {
		err := a.controller.RefreshSelection(nodeID)
		details, detailsErr := a.controller.RefreshSelectionDetails()
		a.ui.Queue(func() {
			a.view.SetBusy(false)
			if err != nil {
				msg := formatFatal("SelectNode", nodeID, err)
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
