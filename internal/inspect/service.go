package inspect

import (
	"context"
	"errors"
	"fmt"
)

var (
	ErrInvalidNodeID             = errors.New("inspect: invalid node id")
	ErrStaleCache                = errors.New("inspect: stale cache")
	ErrUnsupportedPatternAction  = errors.New("inspect: unsupported pattern action")
	ErrMissingPatternInput       = errors.New("inspect: missing required pattern input")
	ErrPatternExecutionFailure   = errors.New("inspect: pattern execution failure")
	ErrProviderActionUnsupported = errors.New("inspect: operation not supported")
	ErrProviderTransientFailure  = errors.New("inspect: transient failure")
	ErrTransientFailure          = errors.New("inspect: transient failure")
)

type Service interface {
	ListWindows(context.Context, ListWindowsRequest) (ListWindowsResponse, error)
	InspectWindow(context.Context, InspectWindowRequest) (InspectWindowResponse, error)
	GetTreeRoot(context.Context, GetTreeRootRequest) (GetTreeRootResponse, error)
	GetNodeChildren(context.Context, GetNodeChildrenRequest) (GetNodeChildrenResponse, error)
	SelectNode(context.Context, SelectNodeRequest) (SelectNodeResponse, error)
	GetFocusedElement(context.Context, GetFocusedElementRequest) (GetFocusedElementResponse, error)
	GetElementUnderCursor(context.Context, GetElementUnderCursorRequest) (GetElementUnderCursorResponse, error)
	HighlightNode(context.Context, HighlightNodeRequest) (HighlightNodeResponse, error)
	ClearHighlight(context.Context, ClearHighlightRequest) (ClearHighlightResponse, error)
	CopyBestSelector(context.Context, CopyBestSelectorRequest) (CopyBestSelectorResponse, error)
	GetPatternActions(context.Context, GetPatternActionsRequest) (GetPatternActionsResponse, error)
	InvokePattern(context.Context, InvokePatternRequest) (InvokePatternResponse, error)
	ToggleFollowCursor(context.Context, ToggleFollowCursorRequest) (ToggleFollowCursorResponse, error)
	RefreshWindows(context.Context, RefreshWindowsRequest) (RefreshWindowsResponse, error)
}

type WindowsProvider interface {
	ListWindows(context.Context, ListWindowsRequest) (ListWindowsResponse, error)
	InspectWindow(context.Context, InspectWindowRequest) (InspectWindowResponse, error)
	GetTreeRoot(context.Context, GetTreeRootRequest) (GetTreeRootResponse, error)
	GetNodeChildren(context.Context, GetNodeChildrenRequest) (GetNodeChildrenResponse, error)
	SelectNode(context.Context, SelectNodeRequest) (SelectNodeResponse, error)
	GetFocusedElement(context.Context, GetFocusedElementRequest) (GetFocusedElementResponse, error)
	GetElementUnderCursor(context.Context, GetElementUnderCursorRequest) (GetElementUnderCursorResponse, error)
	HighlightNode(context.Context, HighlightNodeRequest) (HighlightNodeResponse, error)
	ClearHighlight(context.Context, ClearHighlightRequest) (ClearHighlightResponse, error)
	CopyBestSelector(context.Context, CopyBestSelectorRequest) (CopyBestSelectorResponse, error)
	GetPatternActions(context.Context, GetPatternActionsRequest) (GetPatternActionsResponse, error)
	InvokePattern(context.Context, InvokePatternRequest) (InvokePatternResponse, error)
	ToggleFollowCursor(context.Context, ToggleFollowCursorRequest) (ToggleFollowCursorResponse, error)
	RefreshWindows(context.Context, RefreshWindowsRequest) (RefreshWindowsResponse, error)
}

type service struct {
	provider WindowsProvider
}

func NewService() Service {
	return service{provider: newWindowsProvider()}
}

func newServiceWithProvider(provider WindowsProvider) Service {
	return service{provider: provider}
}

type ListWindowsRequest struct {
	TitleContains string `json:"titleContains,omitempty"`
	ClassName     string `json:"className,omitempty"`
}

type WindowSummary struct {
	HWND      string `json:"hwnd"`
	Title     string `json:"title"`
	ClassName string `json:"className,omitempty"`
}

type ListWindowsResponse struct {
	Windows []WindowSummary `json:"windows"`
}

type InspectWindowRequest struct {
	// InspectWindow is metadata-only and does not refresh the UIA tree cache.
	// Call GetTreeRoot with Refresh=true when a cache refresh is required.
	HWND string `json:"hwnd"`
}

type InspectWindowResponse struct {
	Window     WindowSummary `json:"window"`
	RootNodeID string        `json:"rootNodeID,omitempty"`
}

type GetTreeRootRequest struct {
	HWND    string `json:"hwnd"`
	Refresh bool   `json:"refresh,omitempty"`
}

type TreeNodeDTO struct {
	NodeID      string   `json:"nodeID"`
	Name        string   `json:"name,omitempty"`
	ControlType string   `json:"controlType,omitempty"`
	ClassName   string   `json:"className,omitempty"`
	Patterns    []string `json:"patterns,omitempty"`
	ChildCount  *int     `json:"childCount,omitempty"`
	HasChildren bool     `json:"hasChildren,omitempty"`
	Expanded    bool     `json:"expanded,omitempty"`
	Cycle       bool     `json:"cycle,omitempty"`
}

type GetTreeRootResponse struct {
	Root TreeNodeDTO `json:"root"`
}

type GetNodeChildrenRequest struct {
	NodeID string `json:"nodeID"`
}

type GetNodeChildrenResponse struct {
	ParentNodeID string        `json:"parentNodeID"`
	Children     []TreeNodeDTO `json:"children"`
}

type SelectNodeRequest struct {
	NodeID string `json:"nodeID"`
}
type SelectNodeResponse struct {
	Selected TreeNodeDTO `json:"selected"`
}

type GetFocusedElementRequest struct{}
type GetFocusedElementResponse struct {
	Element TreeNodeDTO `json:"element"`
}

type GetElementUnderCursorRequest struct{}
type GetElementUnderCursorResponse struct {
	Element TreeNodeDTO `json:"element"`
}

type HighlightNodeRequest struct {
	NodeID string `json:"nodeID"`
}
type HighlightNodeResponse struct {
	Highlighted bool `json:"highlighted"`
}

type ClearHighlightRequest struct{}
type ClearHighlightResponse struct {
	Cleared bool `json:"cleared"`
}

type CopyBestSelectorRequest struct {
	NodeID string `json:"nodeID"`
}

type CopyBestSelectorResponse struct {
	Selector         string `json:"selector"`
	ClipboardUpdated bool   `json:"clipboardUpdated"`
}

type GetPatternActionsRequest struct {
	NodeID string `json:"nodeID"`
}
type PatternActionDTO struct {
	Name          string `json:"name"`
	PayloadSchema string `json:"payloadSchema,omitempty"`
}
type GetPatternActionsResponse struct {
	NodeID  string             `json:"nodeID"`
	Actions []PatternActionDTO `json:"actions"`
}

type InvokePatternRequest struct {
	NodeID  string         `json:"nodeID"`
	Action  string         `json:"action"`
	Payload map[string]any `json:"payload,omitempty"`
}

type InvokePatternResponse struct {
	NodeID  string `json:"nodeID"`
	Action  string `json:"action"`
	Invoked bool   `json:"invoked"`
	Result  string `json:"result,omitempty"`
}

type ToggleFollowCursorRequest struct {
	Enabled bool `json:"enabled"`
}
type ToggleFollowCursorResponse struct {
	Enabled bool `json:"enabled"`
}

type RefreshWindowsRequest struct{}
type RefreshWindowsResponse struct {
	Windows []WindowSummary `json:"windows"`
}

func (s service) ListWindows(ctx context.Context, req ListWindowsRequest) (ListWindowsResponse, error) {
	resp, err := s.provider.ListWindows(ctx, req)
	return resp, mapProviderError(err)
}

func (s service) InspectWindow(ctx context.Context, req InspectWindowRequest) (InspectWindowResponse, error) {
	if req.HWND == "" {
		return InspectWindowResponse{}, ErrInvalidNodeID
	}
	resp, err := s.provider.InspectWindow(ctx, req)
	return resp, mapProviderError(err)
}

func (s service) GetTreeRoot(ctx context.Context, req GetTreeRootRequest) (GetTreeRootResponse, error) {
	if req.HWND == "" {
		return GetTreeRootResponse{}, ErrInvalidNodeID
	}
	resp, err := s.provider.GetTreeRoot(ctx, req)
	return resp, mapProviderError(err)
}

func (s service) GetNodeChildren(ctx context.Context, req GetNodeChildrenRequest) (GetNodeChildrenResponse, error) {
	if req.NodeID == "" {
		return GetNodeChildrenResponse{}, ErrInvalidNodeID
	}
	resp, err := s.provider.GetNodeChildren(ctx, req)
	return resp, mapProviderError(err)
}

func (s service) SelectNode(ctx context.Context, req SelectNodeRequest) (SelectNodeResponse, error) {
	if req.NodeID == "" {
		return SelectNodeResponse{}, ErrInvalidNodeID
	}
	resp, err := s.provider.SelectNode(ctx, req)
	return resp, mapProviderError(err)
}

func (s service) GetFocusedElement(ctx context.Context, req GetFocusedElementRequest) (GetFocusedElementResponse, error) {
	resp, err := s.provider.GetFocusedElement(ctx, req)
	return resp, mapProviderError(err)
}

func (s service) GetElementUnderCursor(ctx context.Context, req GetElementUnderCursorRequest) (GetElementUnderCursorResponse, error) {
	resp, err := s.provider.GetElementUnderCursor(ctx, req)
	return resp, mapProviderError(err)
}

func (s service) HighlightNode(ctx context.Context, req HighlightNodeRequest) (HighlightNodeResponse, error) {
	if req.NodeID == "" {
		return HighlightNodeResponse{}, ErrInvalidNodeID
	}
	resp, err := s.provider.HighlightNode(ctx, req)
	return resp, mapProviderError(err)
}

func (s service) ClearHighlight(ctx context.Context, req ClearHighlightRequest) (ClearHighlightResponse, error) {
	resp, err := s.provider.ClearHighlight(ctx, req)
	return resp, mapProviderError(err)
}

func (s service) CopyBestSelector(ctx context.Context, req CopyBestSelectorRequest) (CopyBestSelectorResponse, error) {
	if req.NodeID == "" {
		return CopyBestSelectorResponse{}, ErrInvalidNodeID
	}
	resp, err := s.provider.CopyBestSelector(ctx, req)
	return resp, mapProviderError(err)
}

func (s service) GetPatternActions(ctx context.Context, req GetPatternActionsRequest) (GetPatternActionsResponse, error) {
	if req.NodeID == "" {
		return GetPatternActionsResponse{}, ErrInvalidNodeID
	}
	resp, err := s.provider.GetPatternActions(ctx, req)
	return resp, mapProviderError(err)
}

func (s service) InvokePattern(ctx context.Context, req InvokePatternRequest) (InvokePatternResponse, error) {
	if req.NodeID == "" || req.Action == "" {
		return InvokePatternResponse{}, ErrInvalidNodeID
	}
	resp, err := s.provider.InvokePattern(ctx, req)
	return resp, mapProviderError(err)
}

func (s service) ToggleFollowCursor(ctx context.Context, req ToggleFollowCursorRequest) (ToggleFollowCursorResponse, error) {
	resp, err := s.provider.ToggleFollowCursor(ctx, req)
	return resp, mapProviderError(err)
}

func (s service) RefreshWindows(ctx context.Context, req RefreshWindowsRequest) (RefreshWindowsResponse, error) {
	resp, err := s.provider.RefreshWindows(ctx, req)
	return resp, mapProviderError(err)
}

func mapProviderError(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, ErrProviderActionUnsupported):
		return ErrUnsupportedPatternAction
	case errors.Is(err, ErrProviderTransientFailure):
		return fmt.Errorf("%w: %v", ErrTransientFailure, err)
	case errors.Is(err, ErrStaleCache):
		return ErrStaleCache
	default:
		return err
	}
}
