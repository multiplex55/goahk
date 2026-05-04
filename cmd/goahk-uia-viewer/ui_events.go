package main

import "goahk/internal/inspect"

type UIThreadMarshaller interface{ Queue(func()) }

type ViewUpdater interface {
	SetBusy(bool)
	SetStatus(string)
	UpdateWindowDetails(inspect.GetNodeDetailsResponse)
	UpdateNodeDetails(inspect.GetNodeDetailsResponse)
	UpdateTreeRoot(inspect.TreeNodeDTO)
	UpdateNodeChildren(string, []inspect.TreeNodeDTO)
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
		err := a.controller.SelectWindow(hwnd, activate)
		if err != nil {
			a.ui.Queue(func() {
				a.view.SetBusy(false)
				a.view.SetStatus("select window failed: " + err.Error())
			})
			return
		}
		root, err := a.controller.LoadTreeRoot()
		if err != nil {
			a.ui.Queue(func() {
				a.view.SetBusy(false)
				a.view.SetStatus("load tree root failed: " + err.Error())
			})
			return
		}
		details, err := a.controller.RefreshSelectedNodeDetails()
		a.ui.Queue(func() {
			a.view.SetBusy(false)
			a.view.UpdateTreeRoot(root.Root)
			if err != nil {
				a.view.SetStatus("load root details failed: " + err.Error())
				return
			}
			a.view.UpdateWindowDetails(details)
			a.view.SetStatus("window loaded")
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
