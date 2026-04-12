//go:build !windows
// +build !windows

package inspect

import "context"

type unsupportedProvider struct{}

func newWindowsProvider() WindowsProvider {
	return unsupportedProvider{}
}

func (unsupportedProvider) ListWindows(context.Context, ListWindowsRequest) (ListWindowsResponse, error) {
	return ListWindowsResponse{}, ErrProviderActionUnsupported
}

func (unsupportedProvider) InspectWindow(context.Context, InspectWindowRequest) (InspectWindowResponse, error) {
	return InspectWindowResponse{}, ErrProviderActionUnsupported
}

func (unsupportedProvider) GetTreeRoot(context.Context, GetTreeRootRequest) (GetTreeRootResponse, error) {
	return GetTreeRootResponse{}, ErrProviderActionUnsupported
}

func (unsupportedProvider) GetNodeChildren(context.Context, GetNodeChildrenRequest) (GetNodeChildrenResponse, error) {
	return GetNodeChildrenResponse{}, ErrProviderActionUnsupported
}

func (unsupportedProvider) SelectNode(context.Context, SelectNodeRequest) (SelectNodeResponse, error) {
	return SelectNodeResponse{}, ErrProviderActionUnsupported
}

func (unsupportedProvider) GetNodeDetails(context.Context, GetNodeDetailsRequest) (GetNodeDetailsResponse, error) {
	return GetNodeDetailsResponse{}, ErrProviderActionUnsupported
}

func (unsupportedProvider) GetFocusedElement(context.Context, GetFocusedElementRequest) (GetFocusedElementResponse, error) {
	return GetFocusedElementResponse{}, ErrProviderActionUnsupported
}

func (unsupportedProvider) GetElementUnderCursor(context.Context, GetElementUnderCursorRequest) (GetElementUnderCursorResponse, error) {
	return GetElementUnderCursorResponse{}, ErrProviderActionUnsupported
}

func (unsupportedProvider) HighlightNode(context.Context, HighlightNodeRequest) (HighlightNodeResponse, error) {
	return HighlightNodeResponse{}, ErrProviderActionUnsupported
}

func (unsupportedProvider) ClearHighlight(context.Context, ClearHighlightRequest) (ClearHighlightResponse, error) {
	return ClearHighlightResponse{}, ErrProviderActionUnsupported
}

func (unsupportedProvider) CopyBestSelector(context.Context, CopyBestSelectorRequest) (CopyBestSelectorResponse, error) {
	return CopyBestSelectorResponse{}, ErrProviderActionUnsupported
}

func (unsupportedProvider) GetPatternActions(context.Context, GetPatternActionsRequest) (GetPatternActionsResponse, error) {
	return GetPatternActionsResponse{}, ErrProviderActionUnsupported
}

func (unsupportedProvider) InvokePattern(context.Context, InvokePatternRequest) (InvokePatternResponse, error) {
	return InvokePatternResponse{}, ErrProviderActionUnsupported
}

func (unsupportedProvider) ActivateWindow(context.Context, ActivateWindowRequest) (ActivateWindowResponse, error) {
	return ActivateWindowResponse{}, ErrProviderActionUnsupported
}

func (unsupportedProvider) ToggleFollowCursor(context.Context, ToggleFollowCursorRequest) (ToggleFollowCursorResponse, error) {
	return ToggleFollowCursorResponse{}, ErrProviderActionUnsupported
}
func (unsupportedProvider) PauseFollowCursor(context.Context, PauseFollowCursorRequest) (PauseFollowCursorResponse, error) {
	return PauseFollowCursorResponse{}, ErrProviderActionUnsupported
}
func (unsupportedProvider) ResumeFollowCursor(context.Context, ResumeFollowCursorRequest) (ResumeFollowCursorResponse, error) {
	return ResumeFollowCursorResponse{}, ErrProviderActionUnsupported
}
func (unsupportedProvider) LockFollowCursor(context.Context, LockFollowCursorRequest) (LockFollowCursorResponse, error) {
	return LockFollowCursorResponse{}, ErrProviderActionUnsupported
}
func (unsupportedProvider) UnlockFollowCursor(context.Context, UnlockFollowCursorRequest) (UnlockFollowCursorResponse, error) {
	return UnlockFollowCursorResponse{}, ErrProviderActionUnsupported
}

func (unsupportedProvider) RefreshWindows(context.Context, RefreshWindowsRequest) (RefreshWindowsResponse, error) {
	return RefreshWindowsResponse{}, ErrProviderActionUnsupported
}
func (unsupportedProvider) RefreshTreeRoot(context.Context, RefreshTreeRootRequest) (RefreshTreeRootResponse, error) {
	return RefreshTreeRootResponse{}, ErrProviderActionUnsupported
}
func (unsupportedProvider) RefreshNodeChildren(context.Context, RefreshNodeChildrenRequest) (RefreshNodeChildrenResponse, error) {
	return RefreshNodeChildrenResponse{}, ErrProviderActionUnsupported
}
func (unsupportedProvider) RefreshNodeDetails(context.Context, RefreshNodeDetailsRequest) (RefreshNodeDetailsResponse, error) {
	return RefreshNodeDetailsResponse{}, ErrProviderActionUnsupported
}
func (unsupportedProvider) GetDiagnostics(context.Context, GetDiagnosticsRequest) (GetDiagnosticsResponse, error) {
	return GetDiagnosticsResponse{}, ErrProviderActionUnsupported
}
