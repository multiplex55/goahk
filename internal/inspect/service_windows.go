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
	uiaCore    *providerCore
	windowCore *providerCore
	highlights *highlightController
	windows    windowAdapter
	modeMu     sync.RWMutex
	activeMode InspectMode

	followMu            sync.RWMutex
	followCursorEnabled bool
	focusedUnderCursor  TreeNodeDTO
}

func newWindowsProvider() WindowsProvider {
	return newWindowsProviderWithModeAdapters(newUIAAdapter(newNativeUIADeps()), newUIAAdapter(newNativeUIADeps()), window.NewOSProvider())
}

func newWindowsProviderWithDeps(adapter uiaAdapter, windows windowAdapter) WindowsProvider {
	return newWindowsProviderWithModeAdapters(adapter, adapter, windows)
}

func newWindowsProviderWithModeAdapters(uiaAdapter uiaAdapter, windowTreeAdapter uiaAdapter, windows windowAdapter) WindowsProvider {
	if uiaAdapter == nil {
		uiaAdapter = newUIAAdapter(nil)
	}
	if windowTreeAdapter == nil {
		windowTreeAdapter = newUIAAdapter(newNativeUIADeps())
	}
	if windows == nil {
		windows = window.NewOSProvider()
	}
	return &windowsProvider{
		uiaCore:    newProviderCore(uiaAdapter),
		windowCore: newProviderCore(windowTreeAdapter),
		highlights: newHighlightController(newNativeHighlightOverlay()),
		windows:    windows,
		activeMode: InspectModeUIATree,
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
	mode := normalizeInspectMode(req.Mode)
	root, resolvedMode, err := p.resolveTreeRoot(ctx, req.HWND, mode, false)
	if err != nil {
		return InspectWindowResponse{}, err
	}
	p.setActiveMode(resolvedMode)
	return InspectWindowResponse{Window: WindowSummary{HWND: req.HWND}, RootNodeID: root.NodeID}, nil
}

func (p *windowsProvider) GetTreeRoot(ctx context.Context, req GetTreeRootRequest) (GetTreeRootResponse, error) {
	if req.Refresh {
		_ = p.highlights.Clear(ctx)
	}
	mode := normalizeInspectMode(req.Mode)
	root, resolvedMode, err := p.resolveTreeRoot(ctx, req.HWND, mode, req.Refresh)
	if err != nil {
		return GetTreeRootResponse{}, err
	}
	p.setActiveMode(resolvedMode)
	_ = p.highlights.ClearOnWindowSwitch(ctx, req.HWND)
	state := InspectModeState{
		ActiveMode:   resolvedMode,
		FallbackUsed: mode != resolvedMode,
	}
	if state.FallbackUsed {
		state.FailureStage = "ResolveWindowRoot"
		state.GuidanceText = "UIA tree is unavailable. Switch to Window Tree mode to continue inspecting this window."
	}
	return GetTreeRootResponse{Root: root, State: state}, nil
}

func (p *windowsProvider) GetNodeChildren(ctx context.Context, req GetNodeChildrenRequest) (GetNodeChildrenResponse, error) {
	children, err := p.activeCore().nodeChildren(ctx, req.NodeID)
	if err != nil {
		return GetNodeChildrenResponse{}, err
	}
	return GetNodeChildrenResponse{ParentNodeID: req.NodeID, Children: children}, nil
}

func (p *windowsProvider) SelectNode(ctx context.Context, req SelectNodeRequest) (SelectNodeResponse, error) {
	selected, err := p.activeCore().inspectByNodeID(ctx, req.NodeID)
	if err != nil {
		_ = p.highlights.Clear(ctx)
		return SelectNodeResponse{}, err
	}
	_ = p.highlights.ClearOnDeselection(ctx, selected)
	return SelectNodeResponse{Selected: TreeNodeDTO{
		NodeID:               selected.NodeID,
		NodeId:               selected.NodeID,
		HWND:                 selected.HWND,
		Name:                 selected.Name,
		ControlType:          selected.ControlType,
		LocalizedControlType: selected.LocalizedControlType,
		ClassName:            selected.ClassName,
	}}, nil
}

func (p *windowsProvider) GetNodeDetails(ctx context.Context, req GetNodeDetailsRequest) (GetNodeDetailsResponse, error) {
	core := p.activeCore()
	selected, err := core.inspectByNodeID(ctx, req.NodeID)
	if err != nil {
		return GetNodeDetailsResponse{}, err
	}
	properties := buildPropertyList(selected)
	patterns, err := core.getPatternActions(ctx, req.NodeID)
	if err != nil {
		return GetNodeDetailsResponse{}, err
	}
	path := p.nodePath(ctx, req.NodeID)
	if len(path) == 0 {
		path = []TreeNodeDTO{{
			NodeID:               selected.NodeID,
			NodeId:               selected.NodeID,
			HWND:                 selected.HWND,
			Name:                 selected.Name,
			ControlType:          selected.ControlType,
			LocalizedControlType: selected.LocalizedControlType,
			ClassName:            selected.ClassName,
			ParentNodeID:         selected.ParentNodeID,
		}}
	}
	windowInfo := p.windowInfoForSelection(ctx, selected.ProcessID)
	bestSelector := selectorToCanonicalString(selected.BestSelector)
	statusText := "Loaded node details"
	if strings.TrimSpace(selected.Name) != "" {
		statusText = "Loaded node details: " + selected.Name
	}
	return GetNodeDetailsResponse{
		WindowInfo: windowInfo,
		Element: ElementPropertiesDTO{
			NodeID:               selected.NodeID,
			NodeId:               selected.NodeID,
			HWND:                 selected.HWND,
			ControlType:          selected.ControlType,
			LocalizedControlType: selected.LocalizedControlType,
			Name:                 selected.Name,
			Value:                ptrString(selected.Value),
			AutomationID:         selected.AutomationID,
			Bounds:               selected.BoundingRect,
			HelpText:             ptrString(selected.HelpText),
			AccessKey:            ptrString(selected.AccessKey),
			AcceleratorKey:       ptrString(selected.AcceleratorKey),
			IsKeyboardFocusable:  selected.IsKeyboardFocusable,
			HasKeyboardFocus:     selected.HasKeyboardFocus,
			ItemType:             ptrString(selected.ItemType),
			ItemStatus:           ptrString(selected.ItemStatus),
			IsEnabled:            selected.IsEnabled,
			IsPassword:           selected.IsPassword,
			IsOffscreen:          selected.IsOffscreen,
			FrameworkID:          selected.FrameworkID,
			IsRequiredForForm:    selected.IsRequiredForForm,
			Status:               ptrString(selected.Status),
		},
		Properties:   properties,
		Patterns:     patterns,
		StatusText:   statusText,
		BestSelector: bestSelector,
		Path:         path,
		SelectorPath: SelectorPathDTO{
			BestSelector:        selected.BestSelector,
			FullPath:            path,
			SelectorSuggestions: selected.SelectorSuggestions,
		},
	}, nil
}

func buildPropertyList(selected InspectElement) []PropertyDTO {
	properties := []PropertyDTO{
		{Name: "name", Value: selected.Name},
		{Name: "controlType", Value: selected.ControlType},
		{Name: "localizedControlType", Value: selected.LocalizedControlType},
		{Name: "automationId", Value: selected.AutomationID},
		{Name: "className", Value: selected.ClassName},
		{Name: "frameworkId", Value: selected.FrameworkID},
		{Name: "value", Value: ptrString(selected.Value)},
		{Name: "helpText", Value: ptrString(selected.HelpText)},
		{Name: "accessKey", Value: ptrString(selected.AccessKey)},
		{Name: "acceleratorKey", Value: ptrString(selected.AcceleratorKey)},
		{Name: "status", Value: ptrString(selected.Status)},
		{Name: "itemType", Value: ptrString(selected.ItemType)},
		{Name: "itemStatus", Value: ptrString(selected.ItemStatus)},
		{Name: "isRequiredForForm", Value: strconv.FormatBool(selected.IsRequiredForForm)},
	}
	return properties
}

func ptrString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func selectorToCanonicalString(sel *Selector) string {
	if sel == nil {
		return ""
	}
	parts := make([]string, 0, 5)
	if sel.AutomationID != "" {
		parts = append(parts, "automationId="+sel.AutomationID)
	}
	if sel.Name != "" {
		parts = append(parts, "name="+sel.Name)
	}
	if sel.ControlType != "" {
		parts = append(parts, "controlType="+sel.ControlType)
	}
	if sel.ClassName != "" {
		parts = append(parts, "className="+sel.ClassName)
	}
	if sel.FrameworkID != "" {
		parts = append(parts, "frameworkId="+sel.FrameworkID)
	}
	return strings.Join(parts, ";")
}

func (p *windowsProvider) nodePath(ctx context.Context, nodeID string) []TreeNodeDTO {
	var reversed []TreeNodeDTO
	current := nodeID
	for i := 0; i < 256 && current != ""; i++ {
		details, err := p.activeCore().inspectByNodeID(ctx, current)
		if err != nil {
			break
		}
		reversed = append(reversed, TreeNodeDTO{
			NodeID:               details.NodeID,
			NodeId:               details.NodeID,
			HWND:                 details.HWND,
			Name:                 details.Name,
			ControlType:          details.ControlType,
			LocalizedControlType: details.LocalizedControlType,
			ClassName:            details.ClassName,
			ParentNodeID:         details.ParentNodeID,
		})
		current = details.ParentNodeID
	}
	for i, j := 0, len(reversed)-1; i < j; i, j = i+1, j-1 {
		reversed[i], reversed[j] = reversed[j], reversed[i]
	}
	return reversed
}

func (p *windowsProvider) windowInfoForSelection(ctx context.Context, processID int) WindowInfoDTO {
	hwnd := p.activeCore().childrenCache.window()
	info := WindowInfoDTO{HWND: hwnd}
	if hwnd == "" {
		return info
	}
	infos, err := p.windows.EnumerateWindows(ctx)
	if err != nil {
		return info
	}
	for _, candidate := range infos {
		if candidate.HWND.String() != hwnd {
			continue
		}
		info.Title = candidate.Title
		info.Text = candidate.Title
		info.Class = candidate.Class
		info.Process = candidate.Exe
		info.PID = int(candidate.PID)
		if processID > 0 {
			info.PID = processID
		}
		if candidate.Rect != nil {
			info.Rect = &Rect{
				Left:   candidate.Rect.Left,
				Top:    candidate.Rect.Top,
				Width:  candidate.Rect.Width(),
				Height: candidate.Rect.Height(),
			}
		}
		return info
	}
	if processID > 0 {
		info.PID = processID
	}
	return info
}

func (p *windowsProvider) GetFocusedElement(ctx context.Context, req GetFocusedElementRequest) (GetFocusedElementResponse, error) {
	el, err := p.activeCore().focused(ctx)
	if err != nil {
		return GetFocusedElementResponse{}, err
	}
	return GetFocusedElementResponse{Element: el}, nil
}

func (p *windowsProvider) GetElementUnderCursor(ctx context.Context, req GetElementUnderCursorRequest) (GetElementUnderCursorResponse, error) {
	el, err := p.activeCore().underCursor(ctx)
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
	core := p.activeCore()
	selected, err := core.inspectByNodeID(ctx, req.NodeID)
	if err != nil {
		_ = p.highlights.Clear(ctx)
		return HighlightNodeResponse{}, err
	}
	highlighted, err := p.highlights.ShowNode(ctx, req.NodeID, selected, core.childrenCache.window())
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
	selected, err := p.activeCore().inspectByNodeID(ctx, req.NodeID)
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
	actions, err := p.activeCore().getPatternActions(ctx, req.NodeID)
	if err != nil {
		return GetPatternActionsResponse{}, err
	}
	return GetPatternActionsResponse{NodeID: req.NodeID, Actions: actions}, nil
}

func (p *windowsProvider) InvokePattern(ctx context.Context, req InvokePatternRequest) (InvokePatternResponse, error) {
	return p.activeCore().invokePattern(ctx, req)
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
	p.uiaCore.invalidateWindowCache("")
	p.windowCore.invalidateWindowCache("")

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

func (p *windowsProvider) resolveTreeRoot(ctx context.Context, hwnd string, mode InspectMode, refresh bool) (TreeNodeDTO, InspectMode, error) {
	core := p.coreForMode(mode)
	root, err := core.treeRoot(ctx, hwnd, refresh)
	if err == nil {
		return root, mode, nil
	}
	if mode != InspectModeUIATree || !shouldFallbackToWindowTree(err) {
		return TreeNodeDTO{}, mode, err
	}
	root, fallbackErr := p.windowCore.treeRoot(ctx, hwnd, refresh)
	if fallbackErr != nil {
		return TreeNodeDTO{}, mode, err
	}
	return root, InspectModeWindowTree, nil
}

func shouldFallbackToWindowTree(err error) bool {
	if errors.Is(err, ErrProviderActionUnsupported) {
		return true
	}
	var pErr *ProviderCallError
	if errors.As(err, &pErr) {
		msg := strings.ToLower(pErr.Err.Error())
		return strings.Contains(msg, "access is denied") || strings.Contains(msg, "e_accessdenied")
	}
	return false
}

func normalizeInspectMode(mode InspectMode) InspectMode {
	if mode == InspectModeWindowTree {
		return InspectModeWindowTree
	}
	return InspectModeUIATree
}

func (p *windowsProvider) coreForMode(mode InspectMode) *providerCore {
	if mode == InspectModeWindowTree {
		return p.windowCore
	}
	return p.uiaCore
}

func (p *windowsProvider) activeCore() *providerCore {
	p.modeMu.RLock()
	mode := p.activeMode
	p.modeMu.RUnlock()
	return p.coreForMode(normalizeInspectMode(mode))
}

func (p *windowsProvider) setActiveMode(mode InspectMode) {
	p.modeMu.Lock()
	p.activeMode = normalizeInspectMode(mode)
	p.modeMu.Unlock()
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
