//go:build windows
// +build windows

package inspect

import "context"

type windowsProvider struct {
	core       *providerCore
	highlights *highlightController
}

func newWindowsProvider() WindowsProvider {
	return &windowsProvider{
		core:       newProviderCore(newUnsupportedUIAAdapter()),
		highlights: newHighlightController(newNativeHighlightOverlay()),
	}
}

func (p *windowsProvider) ListWindows(context.Context, ListWindowsRequest) (ListWindowsResponse, error) {
	return ListWindowsResponse{}, ErrProviderActionUnsupported
}

func (p *windowsProvider) InspectWindow(ctx context.Context, req InspectWindowRequest) (InspectWindowResponse, error) {
	root, err := p.core.treeRoot(ctx, req.HWND, false)
	if err != nil {
		return InspectWindowResponse{}, err
	}
	return InspectWindowResponse{Window: WindowSummary{HWND: req.HWND}, RootNodeID: root.NodeID}, nil
}

func (p *windowsProvider) GetTreeRoot(ctx context.Context, req GetTreeRootRequest) (GetTreeRootResponse, error) {
	if req.Refresh {
		_ = p.highlights.Clear(ctx)
	}
	root, err := p.core.treeRoot(ctx, req.HWND, req.Refresh)
	if err != nil {
		return GetTreeRootResponse{}, err
	}
	_ = p.highlights.ClearOnWindowSwitch(ctx, req.HWND)
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
		_ = p.highlights.Clear(ctx)
		return SelectNodeResponse{}, err
	}
	_ = p.highlights.ClearOnDeselection(ctx, selected)
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

func (p *windowsProvider) HighlightNode(ctx context.Context, req HighlightNodeRequest) (HighlightNodeResponse, error) {
	selected, err := p.core.inspectByNodeID(ctx, req.NodeID)
	if err != nil {
		_ = p.highlights.Clear(ctx)
		return HighlightNodeResponse{}, err
	}
	highlighted, err := p.highlights.ShowNode(ctx, req.NodeID, selected, p.core.childrenCache.window())
	if err != nil {
		return HighlightNodeResponse{}, err
	}
	return HighlightNodeResponse{Highlighted: highlighted}, nil
}

func (p *windowsProvider) ClearHighlight(ctx context.Context, _ ClearHighlightRequest) (ClearHighlightResponse, error) {
	if err := p.highlights.Clear(ctx); err != nil {
		return ClearHighlightResponse{}, err
	}
	return ClearHighlightResponse{Cleared: true}, nil
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
	actions, err := p.core.getPatternActions(ctx, req.NodeID)
	if err != nil {
		return GetPatternActionsResponse{}, err
	}
	return GetPatternActionsResponse{NodeID: req.NodeID, Actions: actions}, nil
}

func (p *windowsProvider) InvokePattern(ctx context.Context, req InvokePatternRequest) (InvokePatternResponse, error) {
	return p.core.invokePattern(ctx, req)
}

func (p *windowsProvider) ToggleFollowCursor(context.Context, ToggleFollowCursorRequest) (ToggleFollowCursorResponse, error) {
	return ToggleFollowCursorResponse{}, ErrProviderActionUnsupported
}

func (p *windowsProvider) RefreshWindows(ctx context.Context, _ RefreshWindowsRequest) (RefreshWindowsResponse, error) {
	_ = p.highlights.Clear(ctx)
	p.core.invalidateWindowCache("")
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
func (unsupportedUIAAdapter) GetChildCount(context.Context, string) (int, bool, error) {
	return 0, false, nil
}
func (unsupportedUIAAdapter) Invoke(context.Context, string) error {
	return ErrProviderActionUnsupported
}
func (unsupportedUIAAdapter) Select(context.Context, string) error {
	return ErrProviderActionUnsupported
}
func (unsupportedUIAAdapter) SetValue(context.Context, string, string) error {
	return ErrProviderActionUnsupported
}
func (unsupportedUIAAdapter) DoDefaultAction(context.Context, string) error {
	return ErrProviderActionUnsupported
}
func (unsupportedUIAAdapter) Toggle(context.Context, string) error {
	return ErrProviderActionUnsupported
}
func (unsupportedUIAAdapter) Expand(context.Context, string) error {
	return ErrProviderActionUnsupported
}
func (unsupportedUIAAdapter) Collapse(context.Context, string) error {
	return ErrProviderActionUnsupported
}
