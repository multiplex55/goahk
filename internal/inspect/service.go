package inspect

import (
	"context"
	"errors"
	"fmt"
	"strings"
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
	PauseFollowCursor(context.Context, PauseFollowCursorRequest) (PauseFollowCursorResponse, error)
	ResumeFollowCursor(context.Context, ResumeFollowCursorRequest) (ResumeFollowCursorResponse, error)
	LockFollowCursor(context.Context, LockFollowCursorRequest) (LockFollowCursorResponse, error)
	UnlockFollowCursor(context.Context, UnlockFollowCursorRequest) (UnlockFollowCursorResponse, error)
	RefreshWindows(context.Context, RefreshWindowsRequest) (RefreshWindowsResponse, error)
	RefreshTreeRoot(context.Context, RefreshTreeRootRequest) (RefreshTreeRootResponse, error)
	RefreshNodeChildren(context.Context, RefreshNodeChildrenRequest) (RefreshNodeChildrenResponse, error)
	RefreshNodeDetails(context.Context, RefreshNodeDetailsRequest) (RefreshNodeDetailsResponse, error)
	GetDiagnostics(context.Context, GetDiagnosticsRequest) (GetDiagnosticsResponse, error)
	DumpTree(context.Context, DumpTreeRequest) (DumpTreeResponse, error)
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
	PauseFollowCursor(context.Context, PauseFollowCursorRequest) (PauseFollowCursorResponse, error)
	ResumeFollowCursor(context.Context, ResumeFollowCursorRequest) (ResumeFollowCursorResponse, error)
	LockFollowCursor(context.Context, LockFollowCursorRequest) (LockFollowCursorResponse, error)
	UnlockFollowCursor(context.Context, UnlockFollowCursorRequest) (UnlockFollowCursorResponse, error)
	RefreshWindows(context.Context, RefreshWindowsRequest) (RefreshWindowsResponse, error)
	RefreshTreeRoot(context.Context, RefreshTreeRootRequest) (RefreshTreeRootResponse, error)
	RefreshNodeChildren(context.Context, RefreshNodeChildrenRequest) (RefreshNodeChildrenResponse, error)
	RefreshNodeDetails(context.Context, RefreshNodeDetailsRequest) (RefreshNodeDetailsResponse, error)
	GetDiagnostics(context.Context, GetDiagnosticsRequest) (GetDiagnosticsResponse, error)
}

type DumpTreeRequest struct {
	HWND    string      `json:"hwnd,omitempty"`
	NodeID  string      `json:"nodeID,omitempty"`
	Mode    InspectMode `json:"mode,omitempty"`
	Depth   int         `json:"depth,omitempty"`
	Refresh bool        `json:"refresh,omitempty"`
}

type DumpTreeResponse struct {
	RootNodeID   string            `json:"rootNodeID"`
	Depth        int               `json:"depth"`
	Mode         InspectMode       `json:"mode"`
	Metadata     ProviderSourceDTO `json:"metadata"`
	NodeCount    int               `json:"nodeCount"`
	WarningCount int               `json:"warningCount"`
	Warnings     []string          `json:"warnings,omitempty"`
	Root         DumpNode          `json:"root"`
	Text         string            `json:"text"`
}

type DumpNode struct {
	NodeID               string     `json:"nodeID"`
	ControlType          string     `json:"controlType,omitempty"`
	LocalizedControlType string     `json:"localizedControlType,omitempty"`
	Name                 string     `json:"name,omitempty"`
	Children             []DumpNode `json:"children,omitempty"`
}

type service struct {
	provider WindowsProvider
}

type InspectMode string

const (
	InspectModeAuto       InspectMode = "AUTO"
	InspectModeUIATree    InspectMode = "UIA_TREE"
	InspectModeUIAOnly    InspectMode = "UIA_ONLY"
	InspectModeWindowTree InspectMode = "WINDOW_TREE"
	InspectModeHWNDTree   InspectMode = "HWND_TREE"
)

type InspectModeState struct {
	RequestedMode InspectMode `json:"requestedMode,omitempty"`
	ActiveMode    InspectMode `json:"activeMode"`
	SatisfiedMode InspectMode `json:"satisfiedMode,omitempty"`
	FallbackUsed  bool        `json:"fallbackUsed"`
	Provider      string      `json:"provider,omitempty"`
	Backend       string      `json:"backend,omitempty"`
	DegradeReason string      `json:"degradeReason,omitempty"`
	FailureStage  string      `json:"failureStage,omitempty"`
	GuidanceText  string      `json:"guidanceText,omitempty"`
}

type ProviderSourceDTO struct {
	Provider   string      `json:"provider,omitempty"`
	Source     string      `json:"source,omitempty"`
	Backend    string      `json:"backend,omitempty"`
	Mode       InspectMode `json:"mode,omitempty"`
	Traversal  string      `json:"traversal,omitempty"`
	Fallback   string      `json:"fallback,omitempty"`
	NodeCount  int         `json:"nodeCount,omitempty"`
	ChildCount int         `json:"childCount,omitempty"`
}

type InspectDiagnostics struct {
	Stage         string      `json:"stage,omitempty"`
	ErrorCode     string      `json:"errorCode,omitempty"`
	HResult       string      `json:"hresult,omitempty"`
	Message       string      `json:"message,omitempty"`
	FallbackMode  InspectMode `json:"fallbackMode,omitempty"`
	PrivilegeHint string      `json:"privilegeHint,omitempty"`
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
	HWND string      `json:"hwnd"`
	Mode InspectMode `json:"mode,omitempty"`
}

type InspectWindowResponse struct {
	Window      WindowSummary       `json:"window"`
	RootNodeID  string              `json:"rootNodeID,omitempty"`
	State       InspectModeState    `json:"state"`
	Diagnostics *InspectDiagnostics `json:"diagnostics,omitempty"`
}

type GetTreeRootRequest struct {
	HWND    string      `json:"hwnd"`
	Refresh bool        `json:"refresh,omitempty"`
	Mode    InspectMode `json:"mode,omitempty"`
}

type TreeNodeDTO struct {
	NodeID               string       `json:"nodeID"`
	NodeId               string       `json:"nodeId,omitempty"`
	RuntimeID            string       `json:"runtimeId,omitempty"`
	HWND                 string       `json:"hwnd,omitempty"`
	Name                 string       `json:"name,omitempty"`
	ControlType          string       `json:"controlType,omitempty"`
	LocalizedControlType string       `json:"localizedControlType,omitempty"`
	DisplayLabel         string       `json:"displayLabel,omitempty"`
	DebugMeta            DebugMetaDTO `json:"debugMeta,omitempty"`
	ClassName            string       `json:"className,omitempty"`
	HasChildren          bool         `json:"hasChildren"`
	ParentNodeID         string       `json:"parentNodeID,omitempty"`
	Patterns             []string     `json:"patterns,omitempty"`
	ChildCount           *int         `json:"childCount,omitempty"`
	Expanded             bool         `json:"expanded,omitempty"`
	Cycle                bool         `json:"cycle,omitempty"`
}

type DebugMetaDTO struct {
	ClassName    string `json:"className,omitempty"`
	HWND         string `json:"hwnd,omitempty"`
	AutomationID string `json:"automationId,omitempty"`
	RuntimeID    string `json:"runtimeId,omitempty"`
}

type GetTreeRootResponse struct {
	Root        TreeNodeDTO         `json:"root"`
	State       InspectModeState    `json:"state"`
	Diagnostics *InspectDiagnostics `json:"diagnostics,omitempty"`
	Source      ProviderSourceDTO   `json:"source,omitempty"`
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
	Selected TreeNodeDTO       `json:"selected"`
	State    InspectModeState  `json:"state,omitempty"`
	Source   ProviderSourceDTO `json:"source,omitempty"`
}

type GetNodeDetailsRequest struct {
	NodeID string `json:"nodeID"`
}

type PropertyDTO struct {
	Name   string  `json:"name"`
	Group  string  `json:"group"`
	Value  *string `json:"value"`
	Status string  `json:"status"`
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

type SelectorResolutionDTO struct {
	Best       *SelectorCandidate  `json:"best,omitempty"`
	Alternates []SelectorCandidate `json:"alternates,omitempty"`
}

type GetNodeDetailsResponse struct {
	WindowInfo      WindowInfoDTO         `json:"windowInfo"`
	Element         ElementPropertiesDTO  `json:"element"`
	Properties      []PropertyDTO         `json:"properties"`
	Patterns        []PatternActionDTO    `json:"patterns"`
	StatusText      string                `json:"statusText,omitempty"`
	BestSelector    string                `json:"bestSelector,omitempty"`
	Path            []TreeNodeDTO         `json:"path,omitempty"`
	SelectorPath    SelectorPathDTO       `json:"selectorPath"`
	SelectorOptions SelectorResolutionDTO `json:"selectorOptions,omitempty"`
	ACCPath         string                `json:"accPath,omitempty"`
	Source          ProviderSourceDTO     `json:"source,omitempty"`
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
	Name          string               `json:"name"`
	Pattern       string               `json:"pattern,omitempty"`
	DisplayName   string               `json:"displayName,omitempty"`
	PayloadSchema string               `json:"payloadSchema,omitempty"`
	RequiredArgs  []string             `json:"requiredArgs,omitempty"`
	Supported     bool                 `json:"supported"`
	Enabled       bool                 `json:"enabled"`
	Preconditions []PreconditionStatus `json:"preconditions,omitempty"`
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
	NodeID  string                 `json:"nodeID"`
	Action  string                 `json:"action"`
	Invoked bool                   `json:"invoked"`
	Result  string                 `json:"result,omitempty"`
	Error   *PatternActionErrorDTO `json:"error,omitempty"`
}

type PatternActionErrorClass string

const (
	patternErrorClassNotSupported   PatternActionErrorClass = "not_supported"
	patternErrorClassInvalidInput   PatternActionErrorClass = "invalid_input"
	patternErrorClassTransientState PatternActionErrorClass = "transient_state"
	patternErrorClassAccessDenied   PatternActionErrorClass = "access_denied"
)

type PatternActionErrorDTO struct {
	Class     PatternActionErrorClass `json:"class"`
	Code      string                  `json:"code"`
	Message   string                  `json:"message"`
	Retryable bool                    `json:"retryable"`
}

type patternActionError struct {
	Class  PatternActionErrorClass
	Code   string
	Msg    string
	Action string
	NodeID string
	Err    error
}

func (e *patternActionError) Error() string {
	return fmt.Sprintf("inspect pattern action error: class=%s code=%s action=%s node=%s: %s", e.Class, e.Code, e.Action, e.NodeID, e.Msg)
}

func (e *patternActionError) Unwrap() error { return e.Err }

func newPatternActionError(class PatternActionErrorClass, code, msg, action, nodeID string, err error) error {
	return &patternActionError{Class: class, Code: code, Msg: msg, Action: action, NodeID: nodeID, Err: err}
}

func asPatternActionError(err error) (*patternActionError, bool) {
	if err == nil {
		return nil, false
	}
	var target *patternActionError
	if errors.As(err, &target) {
		return target, true
	}
	return nil, false
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

type PauseFollowCursorRequest struct{}
type PauseFollowCursorResponse struct {
	Paused bool `json:"paused"`
}

type ResumeFollowCursorRequest struct{}
type ResumeFollowCursorResponse struct {
	Paused bool `json:"paused"`
}

type LockFollowCursorRequest struct {
	NodeID string `json:"nodeID,omitempty"`
}
type LockFollowCursorResponse struct {
	Locked bool   `json:"locked"`
	NodeID string `json:"nodeID,omitempty"`
}

type UnlockFollowCursorRequest struct{}
type UnlockFollowCursorResponse struct {
	Locked bool `json:"locked"`
}

type RefreshWindowsRequest struct {
	Filter      string `json:"filter,omitempty"`
	VisibleOnly bool   `json:"visibleOnly"`
	TitleOnly   bool   `json:"titleOnly"`
}
type RefreshWindowsResponse struct {
	Windows []WindowSummary `json:"windows"`
}

type RefreshTreeRootRequest struct {
	HWND string      `json:"hwnd"`
	Mode InspectMode `json:"mode,omitempty"`
}
type RefreshTreeRootResponse struct {
	Root        TreeNodeDTO         `json:"root"`
	State       InspectModeState    `json:"state"`
	Diagnostics *InspectDiagnostics `json:"diagnostics,omitempty"`
}

type RefreshNodeChildrenRequest struct {
	NodeID string `json:"nodeID"`
}
type RefreshNodeChildrenResponse struct {
	ParentNodeID string        `json:"parentNodeID"`
	Children     []TreeNodeDTO `json:"children"`
}

type RefreshNodeDetailsRequest struct {
	NodeID string `json:"nodeID"`
}
type RefreshNodeDetailsResponse struct {
	Details GetNodeDetailsResponse `json:"details"`
}

type GetDiagnosticsRequest struct{}
type GetDiagnosticsResponse struct {
	Diagnostics *InspectDiagnostics `json:"diagnostics,omitempty"`
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
	if err == nil {
		return resp, nil
	}
	mapped := mapProviderError(err)
	if actionErr, ok := asPatternActionError(mapped); ok {
		return InvokePatternResponse{
			NodeID:  req.NodeID,
			Action:  req.Action,
			Invoked: false,
			Error: &PatternActionErrorDTO{
				Class:     actionErr.Class,
				Code:      actionErr.Code,
				Message:   actionErr.Msg,
				Retryable: actionErr.Class == patternErrorClassTransientState,
			},
		}, nil
	}
	return InvokePatternResponse{}, mapped
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

func (s service) PauseFollowCursor(ctx context.Context, req PauseFollowCursorRequest) (PauseFollowCursorResponse, error) {
	resp, err := s.provider.PauseFollowCursor(ctx, req)
	return resp, mapProviderError(err)
}

func (s service) ResumeFollowCursor(ctx context.Context, req ResumeFollowCursorRequest) (ResumeFollowCursorResponse, error) {
	resp, err := s.provider.ResumeFollowCursor(ctx, req)
	return resp, mapProviderError(err)
}

func (s service) LockFollowCursor(ctx context.Context, req LockFollowCursorRequest) (LockFollowCursorResponse, error) {
	resp, err := s.provider.LockFollowCursor(ctx, req)
	return resp, mapProviderError(err)
}

func (s service) UnlockFollowCursor(ctx context.Context, req UnlockFollowCursorRequest) (UnlockFollowCursorResponse, error) {
	resp, err := s.provider.UnlockFollowCursor(ctx, req)
	return resp, mapProviderError(err)
}

func (s service) RefreshWindows(ctx context.Context, req RefreshWindowsRequest) (RefreshWindowsResponse, error) {
	resp, err := s.provider.RefreshWindows(ctx, req)
	return resp, mapProviderError(err)
}

func (s service) RefreshTreeRoot(ctx context.Context, req RefreshTreeRootRequest) (RefreshTreeRootResponse, error) {
	if req.HWND == "" {
		return RefreshTreeRootResponse{}, ErrInvalidNodeID
	}
	resp, err := s.provider.RefreshTreeRoot(ctx, req)
	return resp, mapProviderError(err)
}

func (s service) RefreshNodeChildren(ctx context.Context, req RefreshNodeChildrenRequest) (RefreshNodeChildrenResponse, error) {
	if req.NodeID == "" {
		return RefreshNodeChildrenResponse{}, ErrInvalidNodeID
	}
	resp, err := s.provider.RefreshNodeChildren(ctx, req)
	return resp, mapProviderError(err)
}

func (s service) RefreshNodeDetails(ctx context.Context, req RefreshNodeDetailsRequest) (RefreshNodeDetailsResponse, error) {
	if req.NodeID == "" {
		return RefreshNodeDetailsResponse{}, ErrInvalidNodeID
	}
	resp, err := s.provider.RefreshNodeDetails(ctx, req)
	return resp, mapProviderError(err)
}

func (s service) GetDiagnostics(ctx context.Context, req GetDiagnosticsRequest) (GetDiagnosticsResponse, error) {
	resp, err := s.provider.GetDiagnostics(ctx, req)
	return resp, mapProviderError(err)
}

func (s service) DumpTree(ctx context.Context, req DumpTreeRequest) (DumpTreeResponse, error) {
	depth := req.Depth
	if depth <= 0 {
		depth = 3
	}
	rootID := req.NodeID
	var metadata ProviderSourceDTO
	var warnings []string
	if rootID == "" {
		if req.HWND == "" {
			return DumpTreeResponse{}, ErrInvalidNodeID
		}
		root, err := s.GetTreeRoot(ctx, GetTreeRootRequest{HWND: req.HWND, Refresh: req.Refresh, Mode: req.Mode})
		if err != nil {
			return DumpTreeResponse{}, err
		}
		rootID = root.Root.NodeID
		metadata = root.Source
		if root.State.FallbackUsed {
			warnings = append(warnings, "fallback mode active")
		}
		if strings.TrimSpace(root.State.GuidanceText) != "" {
			warnings = append(warnings, root.State.GuidanceText)
		}
	}
	node, err := s.GetNodeDetails(ctx, GetNodeDetailsRequest{NodeID: rootID})
	if err != nil {
		return DumpTreeResponse{}, err
	}
	if metadata.Provider == "" {
		metadata = node.Source
	}
	root := DumpNode{
		NodeID:               rootID,
		ControlType:          node.Element.ControlType,
		LocalizedControlType: node.Element.LocalizedControlType,
		Name:                 node.Element.Name,
	}
	nodeCount := 1
	if depth > 1 {
		children, childCount, err := s.dumpNodeChildren(ctx, rootID, 1, depth)
		if err != nil {
			return DumpTreeResponse{}, err
		}
		root.Children = children
		nodeCount += childCount
	}
	lines := []string{formatDumpTreeLine(root, 0)}
	appendDumpTreeLines(root.Children, 1, &lines)
	return DumpTreeResponse{
		RootNodeID:   rootID,
		Depth:        depth,
		Mode:         req.Mode,
		Metadata:     metadata,
		NodeCount:    nodeCount,
		WarningCount: len(warnings),
		Warnings:     warnings,
		Root:         root,
		Text:         strings.Join(lines, "\n"),
	}, nil
}

func (s service) dumpNodeChildren(ctx context.Context, parentID string, level, maxDepth int) ([]DumpNode, int, error) {
	if level >= maxDepth {
		return nil, 0, nil
	}
	children, err := s.GetNodeChildren(ctx, GetNodeChildrenRequest{NodeID: parentID})
	if err != nil {
		return nil, 0, err
	}
	nodes := make([]DumpNode, 0, len(children.Children))
	nodeCount := 0
	for _, child := range children.Children {
		n := DumpNode{
			NodeID:               child.NodeID,
			ControlType:          child.ControlType,
			LocalizedControlType: child.LocalizedControlType,
			Name:                 child.Name,
		}
		nodeCount++
		grandChildren, grandCount, err := s.dumpNodeChildren(ctx, child.NodeID, level+1, maxDepth)
		if err != nil {
			return nil, 0, err
		}
		n.Children = grandChildren
		nodeCount += grandCount
		nodes = append(nodes, n)
	}
	return nodes, nodeCount, nil
}

func appendDumpTreeLines(nodes []DumpNode, level int, lines *[]string) {
	for _, child := range nodes {
		*lines = append(*lines, strings.Repeat("  ", level)+formatDumpTreeLine(child, level))
		appendDumpTreeLines(child.Children, level+1, lines)
	}
}

func formatDumpTreeLine(node DumpNode, level int) string {
	return formatDumpNodeLine(node.ControlType, node.LocalizedControlType, node.Name, level == 0)
}

func formatDumpNodeLine(controlType, localizedControlType, name string, root bool) string {
	typ := strings.TrimSpace(controlType)
	if strings.TrimSpace(localizedControlType) != "" {
		typ = strings.TrimSpace(localizedControlType)
	}
	if typ == "" {
		typ = "unknown"
	}
	if root {
		return fmt.Sprintf("%s \"%s\"", strings.ToLower(typ), strings.TrimSpace(name))
	}
	return fmt.Sprintf("%s \"%s\"", typ, strings.TrimSpace(name))
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
