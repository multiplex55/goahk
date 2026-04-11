//go:build windows
// +build windows

package inspect

import (
	"context"
	"errors"
	"sort"
	"strconv"
	"strings"
	"sync"

	"goahk/internal/window"
)

type windowAdapter interface {
	EnumerateWindows(context.Context) ([]window.Info, error)
	ActivateWindow(context.Context, window.HWND) error
}

type windowsProvider struct {
	core       *providerCore
	highlights *highlightController
	windows    windowAdapter

	followMu            sync.RWMutex
	followCursorEnabled bool
	focusedUnderCursor  TreeNodeDTO
}

func newWindowsProvider() WindowsProvider {
	return newWindowsProviderWithDeps(newUIAAdapter(newNativeUIADeps()), window.NewOSProvider())
}

func newWindowsProviderWithDeps(adapter uiaAdapter, windows windowAdapter) WindowsProvider {
	if adapter == nil {
		adapter = newUIAAdapter(nil)
	}
	if windows == nil {
		windows = window.NewOSProvider()
	}
	return &windowsProvider{
		core:       newProviderCore(adapter),
		highlights: newHighlightController(newNativeHighlightOverlay()),
		windows:    windows,
	}
}

func (p *windowsProvider) ListWindows(ctx context.Context, req ListWindowsRequest) (ListWindowsResponse, error) {
	infos, err := p.windows.EnumerateWindows(ctx)
	if err != nil {
		return ListWindowsResponse{}, err
	}
	filter := strings.TrimSpace(req.TitleContains)
	className := strings.TrimSpace(req.ClassName)
	windows := summarizeAndFilterWindows(infos, func(info window.Info) bool {
		if filter != "" && !containsFold(info.Title, filter) {
			return false
		}
		if className != "" && !containsFold(info.Class, className) {
			return false
		}
		return true
	})
	return ListWindowsResponse{Windows: windows}, nil
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

func (p *windowsProvider) GetNodeDetails(ctx context.Context, req GetNodeDetailsRequest) (GetNodeDetailsResponse, error) {
	selected, err := p.core.inspectByNodeID(ctx, req.NodeID)
	if err != nil {
		return GetNodeDetailsResponse{}, err
	}
	properties := []PropertyDTO{
		{Name: "name", Value: selected.Name},
		{Name: "controlType", Value: selected.ControlType},
		{Name: "className", Value: selected.ClassName},
	}
	if selected.AutomationID != "" {
		properties = append(properties, PropertyDTO{Name: "automationId", Value: selected.AutomationID})
	}
	patterns, err := p.core.getPatternActions(ctx, req.NodeID)
	if err != nil {
		return GetNodeDetailsResponse{}, err
	}
	bestSelector := ""
	if selected.BestSelector != nil {
		bestSelector = selected.BestSelector.AutomationID
		if bestSelector == "" {
			bestSelector = selected.BestSelector.Name
		}
	}
	return GetNodeDetailsResponse{
		Properties:   properties,
		Patterns:     patterns,
		StatusText:   selected.Name,
		BestSelector: bestSelector,
	}, nil
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
	p.followMu.Lock()
	if p.followCursorEnabled {
		p.focusedUnderCursor = el
	}
	p.followMu.Unlock()
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

func (p *windowsProvider) ToggleFollowCursor(_ context.Context, req ToggleFollowCursorRequest) (ToggleFollowCursorResponse, error) {
	p.followMu.Lock()
	p.followCursorEnabled = req.Enabled
	if !req.Enabled {
		p.focusedUnderCursor = TreeNodeDTO{}
	}
	p.followMu.Unlock()
	return ToggleFollowCursorResponse{Enabled: req.Enabled}, nil
}

func (p *windowsProvider) ActivateWindow(ctx context.Context, req ActivateWindowRequest) (ActivateWindowResponse, error) {
	hwnd, err := parseHWND(req.HWND)
	if err != nil {
		return ActivateWindowResponse{}, err
	}
	if err := p.windows.ActivateWindow(ctx, hwnd); err != nil {
		return ActivateWindowResponse{}, err
	}
	return ActivateWindowResponse{Activated: true}, nil
}

func (p *windowsProvider) RefreshWindows(ctx context.Context, req RefreshWindowsRequest) (RefreshWindowsResponse, error) {
	_ = p.highlights.Clear(ctx)
	p.core.invalidateWindowCache("")

	infos, err := p.windows.EnumerateWindows(ctx)
	if err != nil {
		return RefreshWindowsResponse{}, err
	}
	filter := strings.TrimSpace(req.Filter)
	windows := summarizeAndFilterWindows(infos, func(info window.Info) bool {
		if req.VisibleOnly && info.Visible != nil && !*info.Visible {
			return false
		}
		if filter == "" {
			return true
		}
		if req.TitleOnly {
			return containsFold(info.Title, filter)
		}
		if containsFold(info.Title, filter) || containsFold(info.Class, filter) || containsFold(info.Exe, filter) || containsFold(info.HWND.String(), filter) {
			return true
		}
		return false
	})
	return RefreshWindowsResponse{Windows: windows}, nil
}

func summarizeAndFilterWindows(infos []window.Info, keep func(window.Info) bool) []WindowSummary {
	out := make([]WindowSummary, 0, len(infos))
	for _, info := range infos {
		if keep != nil && !keep(info) {
			continue
		}
		out = append(out, WindowSummary{
			HWND:        info.HWND.String(),
			Title:       info.Title,
			ProcessName: info.Exe,
			ClassName:   info.Class,
			ProcessID:   int(info.PID),
		})
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Title == out[j].Title {
			return out[i].HWND < out[j].HWND
		}
		return out[i].Title < out[j].Title
	})
	return out
}

func containsFold(haystack, needle string) bool {
	if needle == "" {
		return true
	}
	return strings.Contains(strings.ToLower(strings.TrimSpace(haystack)), strings.ToLower(strings.TrimSpace(needle)))
}

func parseHWND(raw string) (window.HWND, error) {
	norm := strings.TrimSpace(raw)
	v, err := strconv.ParseUint(norm, 0, 64)
	if err != nil {
		return 0, errors.Join(ErrInvalidNodeID, err)
	}
	return window.HWND(v), nil
}

type unsupportedUIAAdapter struct{}

func newUnsupportedUIAAdapter() uiaAdapter { return unsupportedUIAAdapter{} }

type unsupportedUIADeps struct{}

func (unsupportedUIADeps) ResolveWindowRoot(context.Context, string) (*uiaElement, error) {
	return nil, ErrProviderActionUnsupported
}
func (unsupportedUIADeps) GetFocusedElement(context.Context) (*uiaElement, error) {
	return nil, ErrProviderActionUnsupported
}
func (unsupportedUIADeps) GetCursorPosition(context.Context) (int, int, error) {
	return 0, 0, ErrProviderActionUnsupported
}
func (unsupportedUIADeps) ElementFromPoint(context.Context, int, int) (*uiaElement, error) {
	return nil, ErrProviderActionUnsupported
}
func (unsupportedUIADeps) GetElementByRef(context.Context, string) (*uiaElement, error) {
	return nil, ErrProviderActionUnsupported
}
func (unsupportedUIADeps) GetParent(context.Context, string) (*uiaElement, error) {
	return nil, ErrProviderActionUnsupported
}
func (unsupportedUIADeps) GetChildren(context.Context, string) ([]*uiaElement, error) {
	return nil, ErrProviderActionUnsupported
}
func (unsupportedUIADeps) GetChildCount(context.Context, string) (int, bool, error) {
	return 0, false, nil
}
func (unsupportedUIADeps) Invoke(context.Context, string) error {
	return ErrProviderActionUnsupported
}
func (unsupportedUIADeps) Select(context.Context, string) error {
	return ErrProviderActionUnsupported
}
func (unsupportedUIADeps) SetValue(context.Context, string, string) error {
	return ErrProviderActionUnsupported
}
func (unsupportedUIADeps) DoDefaultAction(context.Context, string) error {
	return ErrProviderActionUnsupported
}
func (unsupportedUIADeps) Toggle(context.Context, string) error {
	return ErrProviderActionUnsupported
}
func (unsupportedUIADeps) Expand(context.Context, string) error {
	return ErrProviderActionUnsupported
}
func (unsupportedUIADeps) Collapse(context.Context, string) error {
	return ErrProviderActionUnsupported
}

func (unsupportedUIAAdapter) ResolveWindowRoot(ctx context.Context, hwnd string) (*uiaElement, error) {
	return unsupportedUIADeps{}.ResolveWindowRoot(ctx, hwnd)
}
func (unsupportedUIAAdapter) GetFocusedElement(ctx context.Context) (*uiaElement, error) {
	return unsupportedUIADeps{}.GetFocusedElement(ctx)
}
func (unsupportedUIAAdapter) GetCursorPosition(ctx context.Context) (int, int, error) {
	return unsupportedUIADeps{}.GetCursorPosition(ctx)
}
func (unsupportedUIAAdapter) ElementFromPoint(ctx context.Context, x, y int) (*uiaElement, error) {
	return unsupportedUIADeps{}.ElementFromPoint(ctx, x, y)
}
func (unsupportedUIAAdapter) GetElementByRef(ctx context.Context, ref string) (*uiaElement, error) {
	return unsupportedUIADeps{}.GetElementByRef(ctx, ref)
}
func (unsupportedUIAAdapter) GetParent(ctx context.Context, ref string) (*uiaElement, error) {
	return unsupportedUIADeps{}.GetParent(ctx, ref)
}
func (unsupportedUIAAdapter) GetChildren(ctx context.Context, ref string) ([]*uiaElement, error) {
	return unsupportedUIADeps{}.GetChildren(ctx, ref)
}
func (unsupportedUIAAdapter) GetChildCount(ctx context.Context, ref string) (int, bool, error) {
	return unsupportedUIADeps{}.GetChildCount(ctx, ref)
}
func (unsupportedUIAAdapter) Invoke(ctx context.Context, ref string) error {
	return unsupportedUIADeps{}.Invoke(ctx, ref)
}
func (unsupportedUIAAdapter) Select(ctx context.Context, ref string) error {
	return unsupportedUIADeps{}.Select(ctx, ref)
}
func (unsupportedUIAAdapter) SetValue(ctx context.Context, ref, value string) error {
	return unsupportedUIADeps{}.SetValue(ctx, ref, value)
}
func (unsupportedUIAAdapter) DoDefaultAction(ctx context.Context, ref string) error {
	return unsupportedUIADeps{}.DoDefaultAction(ctx, ref)
}
func (unsupportedUIAAdapter) Toggle(ctx context.Context, ref string) error {
	return unsupportedUIADeps{}.Toggle(ctx, ref)
}
func (unsupportedUIAAdapter) Expand(ctx context.Context, ref string) error {
	return unsupportedUIADeps{}.Expand(ctx, ref)
}
func (unsupportedUIAAdapter) Collapse(ctx context.Context, ref string) error {
	return unsupportedUIADeps{}.Collapse(ctx, ref)
}
