//go:build windows
// +build windows

package inspect

import (
	"context"
)

type windowsProvider struct {
	core *providerCore
}

func newWindowsProvider() WindowsProvider {
	return &windowsProvider{core: newProviderCore(newUnsupportedUIAAdapter())}
}

func (p *windowsProvider) ListWindows(context.Context, ListWindowsRequest) (ListWindowsResponse, error) {
	return ListWindowsResponse{}, ErrProviderActionUnsupported
}

func (p *windowsProvider) InspectWindow(ctx context.Context, req InspectWindowRequest) (InspectWindowResponse, error) {
	root, err := p.core.treeRoot(ctx, req.HWND)
	if err != nil {
		return InspectWindowResponse{}, err
	}
	return InspectWindowResponse{Window: WindowSummary{HWND: req.HWND}, RootNodeID: root.NodeID}, nil
}

func (p *windowsProvider) GetTreeRoot(ctx context.Context, req GetTreeRootRequest) (GetTreeRootResponse, error) {
	root, err := p.core.treeRoot(ctx, req.HWND)
	if err != nil {
		return GetTreeRootResponse{}, err
	}
	return GetTreeRootResponse{Root: root}, nil
}

func (p *windowsProvider) GetNodeChildren(ctx context.Context, req GetNodeChildrenRequest) (GetNodeChildrenResponse, error) {
	children, err := p.core.nodeChildren(ctx, req.NodeID)
	if err != nil {
		return GetNodeChildrenResponse{}, err
	}
	return GetNodeChildrenResponse{ParentNodeID: req.NodeID, Children: children}, nil
}

func (p *windowsProvider) SelectNode(ctx context.Context, req SelectNodeRequest) (SelectNodeResponse, error) {
	selected, err := p.core.inspectByNodeID(ctx, req.NodeID)
	if err != nil {
		return SelectNodeResponse{}, err
	}
	return SelectNodeResponse{Selected: TreeNodeDTO{NodeID: selected.NodeID, Name: selected.Name, ControlType: selected.ControlType, ClassName: selected.ClassName}}, nil
}

func (p *windowsProvider) GetFocusedElement(ctx context.Context, req GetFocusedElementRequest) (GetFocusedElementResponse, error) {
	el, err := p.core.focused(ctx)
	if err != nil {
		return GetFocusedElementResponse{}, err
	}
	return GetFocusedElementResponse{Element: el}, nil
}

func (p *windowsProvider) GetElementUnderCursor(ctx context.Context, req GetElementUnderCursorRequest) (GetElementUnderCursorResponse, error) {
	el, err := p.core.underCursor(ctx)
	if err != nil {
		return GetElementUnderCursorResponse{}, err
	}
	return GetElementUnderCursorResponse{Element: el}, nil
}

func (p *windowsProvider) HighlightNode(context.Context, HighlightNodeRequest) (HighlightNodeResponse, error) {
	return HighlightNodeResponse{}, ErrProviderActionUnsupported
}

func (p *windowsProvider) ClearHighlight(context.Context, ClearHighlightRequest) (ClearHighlightResponse, error) {
	return ClearHighlightResponse{}, ErrProviderActionUnsupported
}

func (p *windowsProvider) CopyBestSelector(ctx context.Context, req CopyBestSelectorRequest) (CopyBestSelectorResponse, error) {
	selected, err := p.core.inspectByNodeID(ctx, req.NodeID)
	if err != nil {
		return CopyBestSelectorResponse{}, err
	}
	if selected.BestSelector == nil {
		return CopyBestSelectorResponse{}, nil
	}
	selector := selected.BestSelector.AutomationID
	if selector == "" {
		selector = selected.BestSelector.Name
	}
	return CopyBestSelectorResponse{Selector: selector, ClipboardUpdated: false}, nil
}

func (p *windowsProvider) GetPatternActions(ctx context.Context, req GetPatternActionsRequest) (GetPatternActionsResponse, error) {
	selected, err := p.core.inspectByNodeID(ctx, req.NodeID)
	if err != nil {
		return GetPatternActionsResponse{}, err
	}
	actions := make([]PatternActionDTO, 0, len(selected.Patterns))
	for _, a := range selected.Patterns {
		actions = append(actions, PatternActionDTO{Name: a.Action, PayloadSchema: a.PayloadSchema})
	}
	return GetPatternActionsResponse{NodeID: req.NodeID, Actions: actions}, nil
}

func (p *windowsProvider) InvokePattern(context.Context, InvokePatternRequest) (InvokePatternResponse, error) {
	return InvokePatternResponse{}, ErrProviderActionUnsupported
}

func (p *windowsProvider) ToggleFollowCursor(context.Context, ToggleFollowCursorRequest) (ToggleFollowCursorResponse, error) {
	return ToggleFollowCursorResponse{}, ErrProviderActionUnsupported
}

func (p *windowsProvider) RefreshWindows(context.Context, RefreshWindowsRequest) (RefreshWindowsResponse, error) {
	return RefreshWindowsResponse{}, ErrProviderActionUnsupported
}

type unsupportedUIAAdapter struct{}

func newUnsupportedUIAAdapter() uiaAdapter { return unsupportedUIAAdapter{} }

func (unsupportedUIAAdapter) ResolveWindowRoot(context.Context, string) (*uiaElement, error) {
	return nil, ErrProviderActionUnsupported
}
func (unsupportedUIAAdapter) GetFocusedElement(context.Context) (*uiaElement, error) {
	return nil, ErrProviderActionUnsupported
}
func (unsupportedUIAAdapter) GetCursorPosition(context.Context) (int, int, error) {
	return 0, 0, ErrProviderActionUnsupported
}
func (unsupportedUIAAdapter) ElementFromPoint(context.Context, int, int) (*uiaElement, error) {
	return nil, ErrProviderActionUnsupported
}
func (unsupportedUIAAdapter) GetElementByRef(context.Context, string) (*uiaElement, error) {
	return nil, ErrProviderActionUnsupported
}
func (unsupportedUIAAdapter) GetParent(context.Context, string) (*uiaElement, error) {
	return nil, ErrProviderActionUnsupported
}
func (unsupportedUIAAdapter) GetChildren(context.Context, string) ([]*uiaElement, error) {
	return nil, ErrProviderActionUnsupported
}
