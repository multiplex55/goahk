package inspect

// API Model Compatibility Notes:
//   - These structs are part of the inspect JSON contract consumed by the frontend.
//   - Treat existing JSON field names and value semantics as stable API surface.
//   - Prefer additive changes only; avoid renaming/removing fields or changing types.
//   - For nullable semantics, use pointer fields and omit them when unavailable.

// Rect describes a screen-space rectangle in physical pixels.
//
// Compatibility: keep field names/types stable; add new geometry fields additively.
type Rect struct {
	Left   int `json:"left"`
	Top    int `json:"top"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// SelectorCandidate is a selector proposal ranked by match stability/uniqueness.
//
// Compatibility: ordering by Rank is part of the contract when provided.
type SelectorCandidate struct {
	Rank      int            `json:"rank"`
	Selector  Selector       `json:"selector"`
	Rationale string         `json:"rationale,omitempty"`
	Score     int            `json:"score,omitempty"`
	Source    string         `json:"source,omitempty"`
	Meta      map[string]any `json:"meta,omitempty"`
}

// Selector is a frontend-facing selector payload used for bestSelector and suggestions.
//
// Empty string means the attribute is not part of this selector.
type Selector struct {
	AutomationID string `json:"automationId,omitempty"`
	Name         string `json:"name,omitempty"`
	ControlType  string `json:"controlType,omitempty"`
	ClassName    string `json:"className,omitempty"`
	FrameworkID  string `json:"frameworkId,omitempty"`
}

// PatternAction describes an available UIA pattern operation for an element.
//
// Compatibility: action names are contract values used by clients.
type PatternAction struct {
	Pattern       string `json:"pattern"`
	Action        string `json:"action"`
	DisplayName   string `json:"displayName,omitempty"`
	PayloadSchema string `json:"payloadSchema,omitempty"`
}

// TreeNodeSummary is the lightweight tree row model shown in the inspect explorer.
//
// Nullability notes:
//   - boundingRect omitted when geometry is unavailable.
//   - bestSelector omitted when no deterministic selector can be generated.
type TreeNodeSummary struct {
	NodeID               string              `json:"nodeId"`
	ParentNodeID         string              `json:"parentNodeId,omitempty"`
	Name                 string              `json:"name,omitempty"`
	ControlType          string              `json:"controlType,omitempty"`
	LocalizedControlType string              `json:"localizedControlType,omitempty"`
	AutomationID         string              `json:"automationId,omitempty"`
	ClassName            string              `json:"className,omitempty"`
	FrameworkID          string              `json:"frameworkId,omitempty"`
	BoundingRect         *Rect               `json:"boundingRect,omitempty"`
	IsOffscreen          bool                `json:"isOffscreen"`
	ChildCount           int                 `json:"childCount,omitempty"`
	HasChildren          bool                `json:"hasChildren"`
	Patterns             []PatternAction     `json:"patterns,omitempty"`
	BestSelector         *Selector           `json:"bestSelector,omitempty"`
	SelectorSuggestions  []SelectorCandidate `json:"selectorSuggestions,omitempty"`
}

// InspectElement is the full metadata model for a selected UI Automation element.
//
// Nullability notes:
//   - labeledBy omitted when no labeling element is exposed.
//   - value/helpText/itemStatus/status omitted when unavailable.
//   - boundingRect omitted when provider cannot resolve geometry.
//   - bestSelector omitted when no selector candidate is suitable.
type InspectElement struct {
	NodeID               string              `json:"nodeId"`
	RuntimeID            string              `json:"runtimeId,omitempty"`
	ParentNodeID         string              `json:"parentNodeId,omitempty"`
	Name                 string              `json:"name,omitempty"`
	LocalizedControlType string              `json:"localizedControlType,omitempty"`
	ControlType          string              `json:"controlType,omitempty"`
	AutomationID         string              `json:"automationId,omitempty"`
	ClassName            string              `json:"className,omitempty"`
	FrameworkID          string              `json:"frameworkId,omitempty"`
	ProcessID            int                 `json:"processId,omitempty"`
	HelpText             *string             `json:"helpText,omitempty"`
	Status               *string             `json:"status,omitempty"`
	Value                *string             `json:"value,omitempty"`
	ItemStatus           *string             `json:"itemStatus,omitempty"`
	LabeledBy            *string             `json:"labeledBy,omitempty"`
	BoundingRect         *Rect               `json:"boundingRect,omitempty"`
	IsEnabled            bool                `json:"isEnabled"`
	IsKeyboardFocusable  bool                `json:"isKeyboardFocusable"`
	HasKeyboardFocus     bool                `json:"hasKeyboardFocus"`
	IsOffscreen          bool                `json:"isOffscreen"`
	IsContentElement     bool                `json:"isContentElement"`
	IsControlElement     bool                `json:"isControlElement"`
	IsPassword           bool                `json:"isPassword"`
	Patterns             []PatternAction     `json:"patterns,omitempty"`
	BestSelector         *Selector           `json:"bestSelector,omitempty"`
	SelectorSuggestions  []SelectorCandidate `json:"selectorSuggestions,omitempty"`
}

// InspectWindow is the full metadata model for a top-level inspected window.
//
// Nullability notes:
//   - executablePath omitted when not available.
//   - rootElement omitted when tree snapshot is unavailable.
//
// Compatibility: identity and activity fields are expected by the frontend; keep stable.
type InspectWindow struct {
	HWND           string           `json:"hwnd"`
	Title          string           `json:"title,omitempty"`
	ClassName      string           `json:"className,omitempty"`
	Rect           Rect             `json:"rect"`
	ProcessID      int              `json:"processId,omitempty"`
	ProcessName    string           `json:"processName,omitempty"`
	ExecutablePath string           `json:"executablePath,omitempty"`
	IsVisible      bool             `json:"isVisible"`
	IsActive       bool             `json:"isActive"`
	IsMinimized    bool             `json:"isMinimized"`
	RootNodeID     string           `json:"rootNodeId,omitempty"`
	RootElement    *TreeNodeSummary `json:"rootElement,omitempty"`
}
