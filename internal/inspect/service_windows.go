//go:build windows
// +build windows

package inspect

import (
	"context"
)

type windowsProvider struct{}

func newWindowsProvider() WindowsProvider {
	return windowsProvider{}
}

func (windowsProvider) ListWindows(context.Context, ListWindowsRequest) (ListWindowsResponse, error) {
	return ListWindowsResponse{}, ErrProviderActionUnsupported
}

func (windowsProvider) InspectWindow(context.Context, InspectWindowRequest) (InspectWindowResponse, error) {
	return InspectWindowResponse{}, ErrProviderActionUnsupported
}

func (windowsProvider) GetTreeRoot(context.Context, GetTreeRootRequest) (GetTreeRootResponse, error) {
	return GetTreeRootResponse{}, ErrProviderActionUnsupported
}

func (windowsProvider) GetNodeChildren(context.Context, GetNodeChildrenRequest) (GetNodeChildrenResponse, error) {
	return GetNodeChildrenResponse{}, ErrProviderActionUnsupported
}

func (windowsProvider) SelectNode(context.Context, SelectNodeRequest) (SelectNodeResponse, error) {
	return SelectNodeResponse{}, ErrProviderActionUnsupported
}

func (windowsProvider) GetFocusedElement(context.Context, GetFocusedElementRequest) (GetFocusedElementResponse, error) {
	return GetFocusedElementResponse{}, ErrProviderActionUnsupported
}

func (windowsProvider) GetElementUnderCursor(context.Context, GetElementUnderCursorRequest) (GetElementUnderCursorResponse, error) {
	return GetElementUnderCursorResponse{}, ErrProviderActionUnsupported
}

func (windowsProvider) HighlightNode(context.Context, HighlightNodeRequest) (HighlightNodeResponse, error) {
	return HighlightNodeResponse{}, ErrProviderActionUnsupported
}

func (windowsProvider) ClearHighlight(context.Context, ClearHighlightRequest) (ClearHighlightResponse, error) {
	return ClearHighlightResponse{}, ErrProviderActionUnsupported
}

func (windowsProvider) CopyBestSelector(context.Context, CopyBestSelectorRequest) (CopyBestSelectorResponse, error) {
	return CopyBestSelectorResponse{}, ErrProviderActionUnsupported
}

func (windowsProvider) GetPatternActions(context.Context, GetPatternActionsRequest) (GetPatternActionsResponse, error) {
	return GetPatternActionsResponse{}, ErrProviderActionUnsupported
}

func (windowsProvider) InvokePattern(context.Context, InvokePatternRequest) (InvokePatternResponse, error) {
	return InvokePatternResponse{}, ErrProviderActionUnsupported
}

func (windowsProvider) ToggleFollowCursor(context.Context, ToggleFollowCursorRequest) (ToggleFollowCursorResponse, error) {
	return ToggleFollowCursorResponse{}, ErrProviderActionUnsupported
}

func (windowsProvider) RefreshWindows(context.Context, RefreshWindowsRequest) (RefreshWindowsResponse, error) {
	return RefreshWindowsResponse{}, ErrProviderActionUnsupported
}
