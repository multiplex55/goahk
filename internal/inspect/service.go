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
	GetNodeDetails(context.Context, GetNodeDetailsRequest) (GetNodeDetailsResponse, error)
	GetFocusedElement(context.Context, GetFocusedElementRequest) (GetFocusedElementResponse, error)
	GetElementUnderCursor(context.Context, GetElementUnderCursorRequest) (GetElementUnderCursorResponse, error)
	HighlightNode(context.Context, HighlightNodeRequest) (HighlightNodeResponse, error)
	ClearHighlight(context.Context, ClearHighlightRequest) (ClearHighlightResponse, error)
	CopyBestSelector(context.Context, CopyBestSelectorRequest) (CopyBestSelectorResponse, error)
	GetPatternActions(context.Context, GetPatternActionsRequest) (GetPatternActionsResponse, error)
	InvokePattern(context.Context, InvokePatternRequest) (InvokePatternResponse, error)
	ActivateWindow(context.Context, ActivateWindowRequest) (ActivateWindowResponse, error)
	ToggleFollowCursor(context.Context, ToggleFollowCursorRequest) (ToggleFollowCursorResponse, error)
	RefreshWindows(context.Context, RefreshWindowsRequest) (RefreshWindowsResponse, error)
}

type WindowsProvider interface {
	ListWindows(context.Context, ListWindowsRequest) (ListWindowsResponse, error)
	InspectWindow(context.Context, InspectWindowRequest) (InspectWindowResponse, error)
	GetTreeRoot(context.Context, GetTreeRootRequest) (GetTreeRootResponse, error)
	GetNodeChildren(context.Context, GetNodeChildrenRequest) (GetNodeChildrenResponse, error)
	SelectNode(context.Context, SelectNodeRequest) (SelectNodeResponse, error)
	GetNodeDetails(context.Context, GetNodeDetailsRequest) (GetNodeDetailsResponse, error)
	GetFocusedElement(context.Context, GetFocusedElementRequest) (GetFocusedElementResponse, error)
	GetElementUnderCursor(context.Context, GetElementUnderCursorRequest) (GetElementUnderCursorResponse, error)
	HighlightNode(context.Context, HighlightNodeRequest) (HighlightNodeResponse, error)
	ClearHighlight(context.Context, ClearHighlightRequest) (ClearHighlightResponse, error)
	CopyBestSelector(context.Context, CopyBestSelectorRequest) (CopyBestSelectorResponse, error)
	GetPatternActions(context.Context, GetPatternActionsRequest) (GetPatternActionsResponse, error)
	InvokePattern(context.Context, InvokePatternRequest) (InvokePatternResponse, error)
	ActivateWindow(context.Context, ActivateWindowRequest) (ActivateWindowResponse, error)
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
	HWND        string `json:"hwnd"`
	Title       string `json:"title"`
	ProcessName string `json:"processName,omitempty"`
	ClassName   string `json:"className,omitempty"`
	ProcessID   int    `json:"processID,omitempty"`
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
	NodeID               string   `json:"nodeID"`
	NodeId               string   `json:"nodeId,omitempty"`
	HWND                 string   `json:"hwnd,omitempty"`
	Name                 string   `json:"name,omitempty"`
	ControlType          string   `json:"controlType,omitempty"`
	LocalizedControlType string   `json:"localizedControlType,omitempty"`
	ClassName            string   `json:"className,omitempty"`
	HasChildren          bool     `json:"hasChildren"`
	ParentNodeID         string   `json:"parentNodeID,omitempty"`
	Patterns             []string `json:"patterns,omitempty"`
	ChildCount           *int     `json:"childCount,omitempty"`
	Expanded             bool     `json:"expanded,omitempty"`
	Cycle                bool     `json:"cycle,omitempty"`
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

type GetNodeDetailsRequest struct {
	NodeID string `json:"nodeID"`
}

type PropertyDTO struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type WindowInfoDTO struct {
	Title   string `json:"title,omitempty"`
	HWND    string `json:"hwnd,omitempty"`
	Text    string `json:"text,omitempty"`
	Rect    *Rect  `json:"rect,omitempty"`
	Class   string `json:"class,omitempty"`
	Process string `json:"process,omitempty"`
	PID     int    `json:"pid,omitempty"`
}

type ElementPropertiesDTO struct {
	NodeID               string `json:"nodeID,omitempty"`
	NodeId               string `json:"nodeId,omitempty"`
	HWND                 string `json:"hwnd,omitempty"`
	ControlType          string `json:"controlType,omitempty"`
	LocalizedControlType string `json:"localizedControlType,omitempty"`
	Name                 string `json:"name,omitempty"`
	Value                string `json:"value,omitempty"`
	AutomationID         string `json:"automationId,omitempty"`
	Bounds               *Rect  `json:"bounds,omitempty"`
	HelpText             string `json:"helpText,omitempty"`
	AccessKey            string `json:"accessKey,omitempty"`
	AcceleratorKey       string `json:"acceleratorKey,omitempty"`
	IsKeyboardFocusable  bool   `json:"isKeyboardFocusable"`
	HasKeyboardFocus     bool   `json:"hasKeyboardFocus"`
	ItemType             string `json:"itemType,omitempty"`
	ItemStatus           string `json:"itemStatus,omitempty"`
	IsEnabled            bool   `json:"isEnabled"`
	IsPassword           bool   `json:"isPassword"`
	IsOffscreen          bool   `json:"isOffscreen"`
	FrameworkID          string `json:"frameworkId,omitempty"`
	IsRequiredForForm    bool   `json:"isRequiredForForm"`
	Status               string `json:"status,omitempty"`
}

type SelectorPathDTO struct {
	BestSelector        *Selector           `json:"bestSelector,omitempty"`
	FullPath            []TreeNodeDTO       `json:"fullPath,omitempty"`
	SelectorSuggestions []SelectorCandidate `json:"selectorSuggestions,omitempty"`
}

type GetNodeDetailsResponse struct {
	WindowInfo   WindowInfoDTO        `json:"windowInfo"`
	Element      ElementPropertiesDTO `json:"element"`
	Properties   []PropertyDTO        `json:"properties"`
	Patterns     []PatternActionDTO   `json:"patterns"`
	StatusText   string               `json:"statusText,omitempty"`
	BestSelector string               `json:"bestSelector,omitempty"`
	Path         []TreeNodeDTO        `json:"path,omitempty"`
	SelectorPath SelectorPathDTO      `json:"selectorPath"`
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
	Pattern       string `json:"pattern,omitempty"`
	DisplayName   string `json:"displayName,omitempty"`
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

type ActivateWindowRequest struct {
	HWND string `json:"hwnd"`
}

type ActivateWindowResponse struct {
	Activated bool `json:"activated"`
}

type ToggleFollowCursorRequest struct {
	Enabled bool `json:"enabled"`
}
type ToggleFollowCursorResponse struct {
	Enabled bool `json:"enabled"`
}

type RefreshWindowsRequest struct {
	Filter      string `json:"filter,omitempty"`
	VisibleOnly bool   `json:"visibleOnly"`
	TitleOnly   bool   `json:"titleOnly"`
}
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

func (s service) GetNodeDetails(ctx context.Context, req GetNodeDetailsRequest) (GetNodeDetailsResponse, error) {
	if req.NodeID == "" {
		return GetNodeDetailsResponse{}, ErrInvalidNodeID
	}
	resp, err := s.provider.GetNodeDetails(ctx, req)
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

func (s service) ActivateWindow(ctx context.Context, req ActivateWindowRequest) (ActivateWindowResponse, error) {
	if req.HWND == "" {
		return ActivateWindowResponse{}, ErrInvalidNodeID
	}
	resp, err := s.provider.ActivateWindow(ctx, req)
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
