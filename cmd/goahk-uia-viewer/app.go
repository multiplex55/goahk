package main

import (
	"context"

	"goahk/internal/inspect"
)

type ViewerApp struct {
	service inspect.Service
}

func NewViewerApp(service inspect.Service) *ViewerApp {
	return &ViewerApp{service: service}
}

func (a *ViewerApp) ListWindows(ctx context.Context, req inspect.ListWindowsRequest) (inspect.ListWindowsResponse, error) {
	return a.service.ListWindows(ctx, req)
}

func (a *ViewerApp) InspectWindow(ctx context.Context, req inspect.InspectWindowRequest) (inspect.InspectWindowResponse, error) {
	return a.service.InspectWindow(ctx, req)
}

func (a *ViewerApp) GetTreeRoot(ctx context.Context, req inspect.GetTreeRootRequest) (inspect.GetTreeRootResponse, error) {
	return a.service.GetTreeRoot(ctx, req)
}

func (a *ViewerApp) GetNodeChildren(ctx context.Context, req inspect.GetNodeChildrenRequest) (inspect.GetNodeChildrenResponse, error) {
	return a.service.GetNodeChildren(ctx, req)
}

func (a *ViewerApp) SelectNode(ctx context.Context, req inspect.SelectNodeRequest) (inspect.SelectNodeResponse, error) {
	return a.service.SelectNode(ctx, req)
}

func (a *ViewerApp) GetFocusedElement(ctx context.Context, req inspect.GetFocusedElementRequest) (inspect.GetFocusedElementResponse, error) {
	return a.service.GetFocusedElement(ctx, req)
}

func (a *ViewerApp) GetElementUnderCursor(ctx context.Context, req inspect.GetElementUnderCursorRequest) (inspect.GetElementUnderCursorResponse, error) {
	return a.service.GetElementUnderCursor(ctx, req)
}

func (a *ViewerApp) HighlightNode(ctx context.Context, req inspect.HighlightNodeRequest) (inspect.HighlightNodeResponse, error) {
	return a.service.HighlightNode(ctx, req)
}

func (a *ViewerApp) ClearHighlight(ctx context.Context, req inspect.ClearHighlightRequest) (inspect.ClearHighlightResponse, error) {
	return a.service.ClearHighlight(ctx, req)
}

func (a *ViewerApp) CopyBestSelector(ctx context.Context, req inspect.CopyBestSelectorRequest) (inspect.CopyBestSelectorResponse, error) {
	return a.service.CopyBestSelector(ctx, req)
}

func (a *ViewerApp) GetPatternActions(ctx context.Context, req inspect.GetPatternActionsRequest) (inspect.GetPatternActionsResponse, error) {
	return a.service.GetPatternActions(ctx, req)
}

func (a *ViewerApp) InvokePattern(ctx context.Context, req inspect.InvokePatternRequest) (inspect.InvokePatternResponse, error) {
	return a.service.InvokePattern(ctx, req)
}

func (a *ViewerApp) ToggleFollowCursor(ctx context.Context, req inspect.ToggleFollowCursorRequest) (inspect.ToggleFollowCursorResponse, error) {
	return a.service.ToggleFollowCursor(ctx, req)
}

func (a *ViewerApp) RefreshWindows(ctx context.Context, req inspect.RefreshWindowsRequest) (inspect.RefreshWindowsResponse, error) {
	return a.service.RefreshWindows(ctx, req)
}
