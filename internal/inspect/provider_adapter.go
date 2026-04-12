package inspect

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

var errNilElementReference = errors.New("inspect: nil element reference")
var errStaleElementReference = errors.New("inspect: stale element reference")

type ProviderCallError struct {
	Op     string
	NodeID string
	HWND   string
	Err    error
}

func (e *ProviderCallError) Error() string {
	parts := []string{"inspect provider call failed", e.Op}
	if e.NodeID != "" {
		parts = append(parts, "node="+e.NodeID)
	}
	if e.HWND != "" {
		parts = append(parts, "hwnd="+e.HWND)
	}
	return strings.Join(parts, " ") + ": " + e.Err.Error()
}

func (e *ProviderCallError) Unwrap() error { return e.Err }

type uiaRect struct{ Left, Top, Width, Height int }

type uiaElement struct {
	Ref                  string
	RuntimeID            string
	HWND                 string
	ParentRef            string
	Name                 string
	LocalizedControlType string
	ControlType          string
	AutomationID         string
	ClassName            string
	FrameworkID          string
	ProcessID            int
	HelpText             *string
	AccessKey            *string
	AcceleratorKey       *string
	Status               *string
	Value                *string
	ItemType             *string
	ItemStatus           *string
	IsRequiredForForm    bool
	LabeledBy            *string
	BoundingRect         *uiaRect
	IsEnabled            bool
	IsKeyboardFocusable  bool
	HasKeyboardFocus     bool
	IsOffscreen          bool
	IsContentElement     bool
	IsControlElement     bool
	IsPassword           bool
	UnsupportedProps     map[string]bool
	PropertyStates       map[string]string
	SupportedPatterns    []string
}

const (
	propertyStatusOK          = "ok"
	propertyStatusUnsupported = "unsupported"
	propertyStatusEmpty       = "empty"
	propertyStatusUnavailable = "unavailable"
	propertyStatusStale       = "stale"
)

type uiaAdapter interface {
	ResolveWindowRoot(context.Context, string) (*uiaElement, error)
	GetFocusedElement(context.Context) (*uiaElement, error)
	GetCursorPosition(context.Context) (int, int, error)
	ElementFromPoint(context.Context, int, int) (*uiaElement, error)
	GetElementByRef(context.Context, string) (*uiaElement, error)
	GetParent(context.Context, string) (*uiaElement, error)
	GetChildren(context.Context, string) ([]*uiaElement, error)
	GetChildCount(context.Context, string) (int, bool, error)
	Invoke(context.Context, string) error
	Select(context.Context, string) error
	SetValue(context.Context, string, string) error
	DoDefaultAction(context.Context, string) error
	Toggle(context.Context, string) error
	Expand(context.Context, string) error
	Collapse(context.Context, string) error
}

type providerCore struct {
	adapter         uiaAdapter
	probeChildCount bool
	nodeNamespace   string
	mu              sync.RWMutex
	nodeToRef       map[string]string
	parentByID      map[string]string
	childrenCache   *nodeChildrenCache
}

func newProviderCore(adapter uiaAdapter) *providerCore {
	return newProviderCoreWithChildCountProbe(adapter, true)
}

func newProviderCoreWithChildCountProbe(adapter uiaAdapter, probeChildCount bool) *providerCore {
	return newProviderCoreWithNamespace(adapter, probeChildCount, "")
}

func newProviderCoreWithNamespace(adapter uiaAdapter, probeChildCount bool, nodeNamespace string) *providerCore {
	return &providerCore{
		adapter:         adapter,
		probeChildCount: probeChildCount,
		nodeNamespace:   strings.TrimSpace(nodeNamespace),
		nodeToRef:       map[string]string{},
		parentByID:      map[string]string{},
		childrenCache:   newNodeChildrenCache(),
	}
}

func (p *providerCore) treeRoot(ctx context.Context, hwnd string, refresh bool) (TreeNodeDTO, error) {
	if refresh {
		p.invalidateWindowCache(hwnd)
	}
	p.childrenCache.setSelectedWindow(hwnd)
	root, err := p.adapter.ResolveWindowRoot(ctx, hwnd)
	if err != nil {
		return TreeNodeDTO{}, p.wrapErr("ResolveWindowRoot", "", hwnd, err)
	}
	node := p.cacheNode(root)
	node.Expanded = false
	return node, nil
}

func (p *providerCore) focused(ctx context.Context) (TreeNodeDTO, error) {
	el, err := p.adapter.GetFocusedElement(ctx)
	if err != nil {
		return TreeNodeDTO{}, p.wrapErr("GetFocusedElement", "", "", err)
	}
	return p.cacheNode(el), nil
}

func (p *providerCore) underCursor(ctx context.Context) (TreeNodeDTO, error) {
	x, y, err := p.adapter.GetCursorPosition(ctx)
	if err != nil {
		return TreeNodeDTO{}, p.wrapErr("GetCursorPosition", "", "", err)
	}
	el, err := p.adapter.ElementFromPoint(ctx, x, y)
	if err != nil {
		return TreeNodeDTO{}, p.wrapErr("ElementFromPoint", "", "", err)
	}
	return p.cacheNode(el), nil
}

func (p *providerCore) nodeChildren(ctx context.Context, nodeID string) ([]TreeNodeDTO, error) {
	windowID := p.childrenCache.window()
	if windowID == "" {
		return nil, ErrStaleCache
	}
	if cached, ok := p.childrenCache.get(windowID, nodeID); ok {
		return cached, nil
	}

	children, err := p.loadNodeChildren(ctx, nodeID)
	if err != nil {
		if !errors.Is(err, ErrStaleCache) {
			return nil, err
		}
		// Stale node fallback: refresh window root and retry once.
		p.invalidateWindowCache(windowID)
		if _, rootErr := p.treeRoot(ctx, windowID, false); rootErr != nil {
			return nil, rootErr
		}
		children, err = p.loadNodeChildren(ctx, nodeID)
		if err != nil {
			return nil, err
		}
	}
	p.childrenCache.put(windowID, nodeID, children)
	return children, nil
}

func (p *providerCore) loadNodeChildren(ctx context.Context, nodeID string) ([]TreeNodeDTO, error) {
	ref, ok := p.lookupRef(nodeID)
	if !ok {
		return nil, ErrStaleCache
	}
	rawChildren, err := p.adapter.GetChildren(ctx, ref)
	if err != nil {
		return nil, p.wrapErr("GetChildren", nodeID, "", err)
	}
	out := make([]TreeNodeDTO, 0, len(rawChildren))
	for _, child := range rawChildren {
		n := p.cacheNode(child)
		n.Expanded = false
		n.Cycle = p.hasAncestor(nodeID, n.NodeID)
		out = append(out, n)
	}
	return out, nil
}

func (p *providerCore) nodeParent(ctx context.Context, nodeID string) (TreeNodeDTO, error) {
	ref, ok := p.lookupRef(nodeID)
	if !ok {
		return TreeNodeDTO{}, ErrStaleCache
	}
	parent, err := p.adapter.GetParent(ctx, ref)
	if err != nil {
		return TreeNodeDTO{}, p.wrapErr("GetParent", nodeID, "", err)
	}
	return p.cacheNode(parent), nil
}

func (p *providerCore) inspectByNodeID(ctx context.Context, nodeID string) (InspectElement, error) {
	ref, ok := p.lookupRef(nodeID)
	if !ok {
		return InspectElement{}, ErrStaleCache
	}
	el, err := p.adapter.GetElementByRef(ctx, ref)
	if err != nil {
		return InspectElement{}, p.wrapErr("GetElementByRef", nodeID, "", err)
	}
	if el.ParentRef != "" {
		parentID := p.parentOf(nodeID)
		if parentID == "" {
			parentID = "node:ref:" + el.ParentRef
		}
		p.mu.Lock()
		p.parentByID[nodeID] = parentID
		p.mu.Unlock()
	}
	selected := toInspectElement(nodeID, p.parentOf(nodeID), el)
	best, suggestions := p.selectorCandidatesForNode(ctx, nodeID, el)
	selected.BestSelector = best
	selected.SelectorSuggestions = suggestions
	return selected, nil
}

func (p *providerCore) selectorCandidatesForNode(ctx context.Context, nodeID string, el *uiaElement) (*Selector, []SelectorCandidate) {
	ref, ok := p.lookupRef(nodeID)
	if !ok {
		return selectorCandidatesForElement(el)
	}
	selectorCtx := selectorContext{}
	if parentRef := strings.TrimSpace(el.ParentRef); parentRef != "" {
		if siblings, err := p.adapter.GetChildren(ctx, parentRef); err == nil {
			selectorCtx.siblings = siblings
		}
		current := parentRef
		for i := 0; i < 4 && current != ""; i++ {
			parent, err := p.adapter.GetElementByRef(ctx, current)
			if err != nil || parent == nil {
				break
			}
			selectorCtx.ancestry = append(selectorCtx.ancestry, parent)
			current = strings.TrimSpace(parent.ParentRef)
		}
	}
	if len(selectorCtx.siblings) == 0 {
		if siblings, err := p.adapter.GetChildren(ctx, ref); err == nil {
			selectorCtx.siblings = siblings
		}
	}
	return selectorCandidatesForElementWithContext(el, selectorCtx)
}

func (p *providerCore) invokePattern(ctx context.Context, req InvokePatternRequest) (InvokePatternResponse, error) {
	available, err := p.getPatternActions(ctx, req.NodeID)
	if err != nil {
		return InvokePatternResponse{}, err
	}
	action := strings.TrimSpace(req.Action)
	if action == "" {
		return InvokePatternResponse{}, newPatternActionError(patternErrorClassNotSupported, "action_required", "action is required", req.Action, req.NodeID, ErrUnsupportedPatternAction)
	}
	var allowed bool
	for _, a := range available {
		if a.Name == action {
			allowed = true
			break
		}
	}
	if !allowed {
		return InvokePatternResponse{}, newPatternActionError(patternErrorClassNotSupported, "action_not_supported", "action is not supported for this node", action, req.NodeID, ErrUnsupportedPatternAction)
	}

	switch action {
	case "invoke":
		err = p.Invoke(ctx, req.NodeID)
	case "select":
		err = p.Select(ctx, req.NodeID)
	case "setValue":
		value, ok := valueFromPayload(req.Payload)
		if !ok {
			return InvokePatternResponse{}, newPatternActionError(patternErrorClassInvalidInput, "missing_input", "setValue requires a non-empty payload.value", action, req.NodeID, ErrMissingPatternInput)
		}
		err = p.SetValue(ctx, req.NodeID, value)
	case "doDefaultAction":
		err = p.DoDefaultAction(ctx, req.NodeID)
	case "toggle":
		err = p.Toggle(ctx, req.NodeID)
	case "expand":
		err = p.Expand(ctx, req.NodeID)
	case "collapse":
		err = p.Collapse(ctx, req.NodeID)
	default:
		return InvokePatternResponse{}, newPatternActionError(patternErrorClassNotSupported, "action_not_supported", "action is not supported for this node", action, req.NodeID, ErrUnsupportedPatternAction)
	}
	if err != nil {
		return InvokePatternResponse{}, p.wrapActionErr(action, req.NodeID, err)
	}
	return InvokePatternResponse{NodeID: req.NodeID, Action: action, Invoked: true}, nil
}

func (p *providerCore) getPatternActions(ctx context.Context, nodeID string) ([]PatternActionDTO, error) {
	selected, err := p.inspectByNodeID(ctx, nodeID)
	if err != nil {
		return nil, err
	}
	actions := make([]PatternActionDTO, 0, len(selected.Patterns))
	for _, a := range selected.Patterns {
		actions = append(actions, PatternActionDTO{
			Name:          a.Action,
			Pattern:       a.Pattern,
			DisplayName:   a.DisplayName,
			PayloadSchema: a.PayloadSchema,
			RequiredArgs:  append([]string(nil), a.RequiredArgs...),
			Supported:     a.Supported,
			Enabled:       a.Enabled,
			Preconditions: a.Preconditions,
		})
	}
	return actions, nil
}

func (p *providerCore) Invoke(ctx context.Context, nodeID string) error {
	ref, ok := p.lookupRef(nodeID)
	if !ok {
		return ErrStaleCache
	}
	if err := p.adapter.Invoke(ctx, ref); err != nil {
		return &ProviderCallError{Op: "Invoke", NodeID: nodeID, Err: err}
	}
	return nil
}

func (p *providerCore) Select(ctx context.Context, nodeID string) error {
	ref, ok := p.lookupRef(nodeID)
	if !ok {
		return ErrStaleCache
	}
	if err := p.adapter.Select(ctx, ref); err != nil {
		return &ProviderCallError{Op: "Select", NodeID: nodeID, Err: err}
	}
	return nil
}

func (p *providerCore) SetValue(ctx context.Context, nodeID, value string) error {
	ref, ok := p.lookupRef(nodeID)
	if !ok {
		return ErrStaleCache
	}
	if err := p.adapter.SetValue(ctx, ref, value); err != nil {
		return &ProviderCallError{Op: "SetValue", NodeID: nodeID, Err: err}
	}
	return nil
}

func (p *providerCore) DoDefaultAction(ctx context.Context, nodeID string) error {
	ref, ok := p.lookupRef(nodeID)
	if !ok {
		return ErrStaleCache
	}
	if err := p.adapter.DoDefaultAction(ctx, ref); err != nil {
		return &ProviderCallError{Op: "DoDefaultAction", NodeID: nodeID, Err: err}
	}
	return nil
}

func (p *providerCore) Toggle(ctx context.Context, nodeID string) error {
	ref, ok := p.lookupRef(nodeID)
	if !ok {
		return ErrStaleCache
	}
	if err := p.adapter.Toggle(ctx, ref); err != nil {
		return &ProviderCallError{Op: "Toggle", NodeID: nodeID, Err: err}
	}
	return nil
}

func (p *providerCore) Expand(ctx context.Context, nodeID string) error {
	ref, ok := p.lookupRef(nodeID)
	if !ok {
		return ErrStaleCache
	}
	if err := p.adapter.Expand(ctx, ref); err != nil {
		return &ProviderCallError{Op: "Expand", NodeID: nodeID, Err: err}
	}
	return nil
}

func (p *providerCore) Collapse(ctx context.Context, nodeID string) error {
	ref, ok := p.lookupRef(nodeID)
	if !ok {
		return ErrStaleCache
	}
	if err := p.adapter.Collapse(ctx, ref); err != nil {
		return &ProviderCallError{Op: "Collapse", NodeID: nodeID, Err: err}
	}
	return nil
}

func valueFromPayload(payload map[string]any) (string, bool) {
	if payload == nil {
		return "", false
	}
	value, ok := payload["value"]
	if !ok {
		return "", false
	}
	str, ok := value.(string)
	return str, ok && strings.TrimSpace(str) != ""
}

func (p *providerCore) wrapActionErr(action, nodeID string, err error) error {
	if err == nil {
		return nil
	}
	class := patternErrorClassTransientState
	code := "action_failed"
	message := "action failed due to transient state"
	switch {
	case errors.Is(err, ErrProviderActionUnsupported), errors.Is(err, ErrUnsupportedPatternAction):
		class = patternErrorClassNotSupported
		code = "action_not_supported"
		message = "action is not supported by the provider"
	case errors.Is(err, ErrMissingPatternInput), errors.Is(err, ErrInvalidNodeID):
		class = patternErrorClassInvalidInput
		code = "invalid_input"
		message = "request input is invalid"
	case errors.Is(err, syscall.EACCES):
		class = patternErrorClassAccessDenied
		code = "access_denied"
		message = "access denied while invoking action"
	case errors.Is(err, ErrStaleCache), errors.Is(err, errStaleElementReference), errors.Is(err, ErrProviderTransientFailure), errors.Is(err, ErrTransientFailure):
		class = patternErrorClassTransientState
		code = "transient_state"
		message = "action could not run due to transient state"
	}
	return newPatternActionError(class, code, message, action, nodeID, fmt.Errorf("%w: %v", ErrPatternExecutionFailure, err))
}

func (p *providerCore) cacheNode(el *uiaElement) TreeNodeDTO {
	if el == nil || el.Ref == "" {
		return TreeNodeDTO{}
	}
	nodeID := p.cacheNodeRef(el)
	if el.ParentRef != "" {
		parentID := p.parentOf(nodeID)
		if parentID == "" {
			parentID = "node:ref:" + el.ParentRef
		}
		p.mu.Lock()
		p.parentByID[nodeID] = parentID
		p.mu.Unlock()
	}
	patterns := patternActionsForElement(el)
	names := make([]string, 0, len(patterns))
	for _, p := range patterns {
		names = append(names, p.Action)
	}
	node := TreeNodeDTO{
		NodeID: nodeID, NodeId: nodeID, RuntimeID: strings.TrimSpace(el.RuntimeID), HWND: strings.TrimSpace(el.HWND),
		Name: el.Name, ControlType: normalizeControlType(el.ControlType, el.LocalizedControlType),
		LocalizedControlType: normalizeLocalizedControlType(el.LocalizedControlType, el.ControlType),
		DisplayLabel:         formatDisplayLabel(el.Name, normalizeLocalizedControlType(el.LocalizedControlType, el.ControlType), normalizeControlType(el.ControlType, el.LocalizedControlType)),
		DebugMeta:            buildDebugMeta(el),
		ClassName:            el.ClassName, Patterns: names,
	}
	if p.probeChildCount {
		if count, ok, err := p.adapter.GetChildCount(context.Background(), el.Ref); err == nil && ok {
			node.ChildCount = &count
			node.HasChildren = count > 0
		}
	}
	return node
}

func runtimeNodeID(runtimeID, ref string) string {
	return runtimeNodeIDWithNamespace("", runtimeID, ref)
}

func runtimeNodeIDWithNamespace(namespace, runtimeID, ref string) string {
	ns := strings.TrimSpace(namespace)
	if rid := strings.TrimSpace(runtimeID); rid != "" {
		if ns != "" {
			return "node:" + ns + ":rid:" + rid
		}
		return "node:rid:" + rid
	}
	if ns != "" {
		return "node:" + ns + ":ref:" + strings.TrimSpace(ref)
	}
	return "node:ref:" + strings.TrimSpace(ref)
}

func (p *providerCore) cacheNodeRef(el *uiaElement) string {
	nodeID := runtimeNodeIDWithNamespace(p.nodeNamespace, el.RuntimeID, el.Ref)
	p.mu.Lock()
	p.nodeToRef[nodeID] = el.Ref
	if el.ParentRef != "" {
		for existingID, existingRef := range p.nodeToRef {
			if existingRef == el.ParentRef {
				p.parentByID[nodeID] = existingID
				break
			}
		}
	}
	p.mu.Unlock()
	return nodeID
}

func (p *providerCore) lookupRef(nodeID string) (string, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	ref, ok := p.nodeToRef[nodeID]
	return ref, ok
}
func (p *providerCore) parentOf(nodeID string) string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.parentByID[nodeID]
}

func (p *providerCore) invalidateWindowCache(_ string) {
	p.mu.Lock()
	p.nodeToRef = map[string]string{}
	p.parentByID = map[string]string{}
	p.mu.Unlock()
	p.childrenCache.invalidateAll()
}

func (p *providerCore) hasAncestor(nodeID, ancestorID string) bool {
	if nodeID == "" || ancestorID == "" {
		return false
	}
	current := nodeID
	for i := 0; i < 128; i++ {
		parent := p.parentOf(current)
		if parent == "" {
			return false
		}
		if parent == ancestorID {
			return true
		}
		current = parent
	}
	return true
}

func (p *providerCore) wrapErr(op, nodeID, hwnd string, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, errNilElementReference) || errors.Is(err, errStaleElementReference) {
		return ErrStaleCache
	}
	return &ProviderCallError{Op: op, NodeID: nodeID, HWND: hwnd, Err: err}
}

func toInspectElement(nodeID, parentNodeID string, el *uiaElement) InspectElement {
	if el == nil {
		return InspectElement{NodeID: nodeID, ParentNodeID: parentNodeID}
	}
	runtimeID, runtimeState := normalizeRuntimeIDField(el.RuntimeID, el.UnsupportedProps["RuntimeID"], el.PropertyStates["RuntimeID"])
	name, nameState := normalizeStringField(el.Name, el.UnsupportedProps["Name"], el.PropertyStates["Name"])
	localizedType, localizedTypeState := normalizeStringField(normalizeLocalizedControlType(el.LocalizedControlType, el.ControlType), el.UnsupportedProps["LocalizedControlType"], el.PropertyStates["LocalizedControlType"])
	controlType, controlTypeState := normalizeStringField(normalizeControlType(el.ControlType, el.LocalizedControlType), el.UnsupportedProps["ControlType"], el.PropertyStates["ControlType"])
	automationID, automationIDState := normalizeStringField(el.AutomationID, el.UnsupportedProps["AutomationId"], el.PropertyStates["AutomationId"])
	className, classNameState := normalizeStringField(el.ClassName, el.UnsupportedProps["ClassName"], el.PropertyStates["ClassName"])
	frameworkID, frameworkIDState := normalizeStringField(el.FrameworkID, el.UnsupportedProps["FrameworkId"], el.PropertyStates["FrameworkId"])
	processID, processIDState := normalizeScalarField(el.ProcessID, el.UnsupportedProps["ProcessId"], el.PropertyStates["ProcessId"])
	helpText, helpTextState := normalizeOptionalStringField(el.HelpText, el.UnsupportedProps["HelpText"], el.PropertyStates["HelpText"])
	accessKey, accessKeyState := normalizeOptionalStringField(el.AccessKey, el.UnsupportedProps["AccessKey"], el.PropertyStates["AccessKey"])
	accelKey, accelKeyState := normalizeOptionalStringField(el.AcceleratorKey, el.UnsupportedProps["AcceleratorKey"], el.PropertyStates["AcceleratorKey"])
	status, statusState := normalizeOptionalStringField(el.Status, el.UnsupportedProps["Status"], el.PropertyStates["Status"])
	value, valueState := normalizeOptionalStringField(el.Value, el.UnsupportedProps["Value"], el.PropertyStates["Value"])
	itemType, itemTypeState := normalizeOptionalStringField(el.ItemType, el.UnsupportedProps["ItemType"], el.PropertyStates["ItemType"])
	itemStatus, itemStatusState := normalizeOptionalStringField(el.ItemStatus, el.UnsupportedProps["ItemStatus"], el.PropertyStates["ItemStatus"])
	labeledBy, labeledByState := normalizeOptionalStringField(el.LabeledBy, el.UnsupportedProps["LabeledBy"], el.PropertyStates["LabeledBy"])
	rect, rectState := normalizeRectField(toRect(el.BoundingRect), el.UnsupportedProps["BoundingRectangle"], el.PropertyStates["BoundingRectangle"])
	isEnabled, isEnabledState := normalizeBoolField(el.IsEnabled, el.UnsupportedProps["IsEnabled"], el.PropertyStates["IsEnabled"])
	isKeyboardFocusable, focusableState := normalizeBoolField(el.IsKeyboardFocusable, el.UnsupportedProps["IsKeyboardFocusable"], el.PropertyStates["IsKeyboardFocusable"])
	hasKeyboardFocus, hasFocusState := normalizeBoolField(el.HasKeyboardFocus, el.UnsupportedProps["HasKeyboardFocus"], el.PropertyStates["HasKeyboardFocus"])
	isOffscreen, offscreenState := normalizeBoolField(el.IsOffscreen, el.UnsupportedProps["IsOffscreen"], el.PropertyStates["IsOffscreen"])
	isContentElement, contentState := normalizeBoolField(el.IsContentElement, el.UnsupportedProps["IsContentElement"], el.PropertyStates["IsContentElement"])
	isControlElement, controlElementState := normalizeBoolField(el.IsControlElement, el.UnsupportedProps["IsControlElement"], el.PropertyStates["IsControlElement"])
	isPassword, passwordState := normalizeBoolField(el.IsPassword, el.UnsupportedProps["IsPassword"], el.PropertyStates["IsPassword"])
	isRequiredForForm, requiredState := normalizeBoolField(el.IsRequiredForForm, el.UnsupportedProps["IsRequiredForForm"], el.PropertyStates["IsRequiredForForm"])

	propertyStates := clonePropertyStates(el.PropertyStates)
	if propertyStates == nil {
		propertyStates = map[string]string{}
	}
	mergeNormalizedPropertyState(propertyStates, "RuntimeID", runtimeState)
	mergeNormalizedPropertyState(propertyStates, "Name", nameState)
	mergeNormalizedPropertyState(propertyStates, "LocalizedControlType", localizedTypeState)
	mergeNormalizedPropertyState(propertyStates, "ControlType", controlTypeState)
	mergeNormalizedPropertyState(propertyStates, "AutomationId", automationIDState)
	mergeNormalizedPropertyState(propertyStates, "ClassName", classNameState)
	mergeNormalizedPropertyState(propertyStates, "FrameworkId", frameworkIDState)
	mergeNormalizedPropertyState(propertyStates, "ProcessId", processIDState)
	mergeNormalizedPropertyState(propertyStates, "HelpText", helpTextState)
	mergeNormalizedPropertyState(propertyStates, "AccessKey", accessKeyState)
	mergeNormalizedPropertyState(propertyStates, "AcceleratorKey", accelKeyState)
	mergeNormalizedPropertyState(propertyStates, "Status", statusState)
	mergeNormalizedPropertyState(propertyStates, "Value", valueState)
	mergeNormalizedPropertyState(propertyStates, "ItemType", itemTypeState)
	mergeNormalizedPropertyState(propertyStates, "ItemStatus", itemStatusState)
	mergeNormalizedPropertyState(propertyStates, "LabeledBy", labeledByState)
	mergeNormalizedPropertyState(propertyStates, "BoundingRectangle", rectState)
	mergeNormalizedPropertyState(propertyStates, "IsEnabled", isEnabledState)
	mergeNormalizedPropertyState(propertyStates, "IsKeyboardFocusable", focusableState)
	mergeNormalizedPropertyState(propertyStates, "HasKeyboardFocus", hasFocusState)
	mergeNormalizedPropertyState(propertyStates, "IsOffscreen", offscreenState)
	mergeNormalizedPropertyState(propertyStates, "IsContentElement", contentState)
	mergeNormalizedPropertyState(propertyStates, "IsControlElement", controlElementState)
	mergeNormalizedPropertyState(propertyStates, "IsPassword", passwordState)
	mergeNormalizedPropertyState(propertyStates, "IsRequiredForForm", requiredState)

	best, alts := selectorCandidatesForElement(el)
	return InspectElement{
		NodeID: nodeID, RuntimeID: runtimeID, ParentNodeID: parentNodeID,
		HWND: strings.TrimSpace(el.HWND),
		Name: name, LocalizedControlType: localizedType, ControlType: controlType,
		AutomationID: automationID, ClassName: className, FrameworkID: frameworkID, ProcessID: processID,
		HelpText: helpText, AccessKey: accessKey, AcceleratorKey: accelKey, Status: status, Value: value,
		ItemType: itemType, ItemStatus: itemStatus, IsRequiredForForm: isRequiredForForm, LabeledBy: labeledBy,
		BoundingRect: rect, IsEnabled: isEnabled, IsKeyboardFocusable: isKeyboardFocusable, HasKeyboardFocus: hasKeyboardFocus,
		IsOffscreen: isOffscreen, IsContentElement: isContentElement, IsControlElement: isControlElement, IsPassword: isPassword,
		UnsupportedProps: cloneUnsupportedProps(el.UnsupportedProps),
		PropertyStates:   propertyStates,
		Patterns:         patternActionsForElement(el), BestSelector: best, SelectorSuggestions: alts,
	}
}

func cloneUnsupportedProps(src map[string]bool) map[string]bool {
	if len(src) == 0 {
		return nil
	}
	dst := make(map[string]bool, len(src))
	for key, unsupported := range src {
		dst[key] = unsupported
	}
	return dst
}

func clonePropertyStates(src map[string]string) map[string]string {
	if len(src) == 0 {
		return nil
	}
	dst := make(map[string]string, len(src))
	for key, status := range src {
		dst[key] = normalizePropertyStatus(status)
	}
	return dst
}

func mergeNormalizedPropertyState(states map[string]string, property, status string) {
	normalized := normalizePropertyStatus(status)
	if normalized == propertyStatusOK {
		return
	}
	states[property] = normalized
}

func normalizeRuntimeIDField(raw string, unsupported bool, status string) (string, string) {
	return normalizeStringField(raw, unsupported, status)
}

func normalizeStringField(raw string, unsupported bool, status string) (string, string) {
	normalized := normalizePropertyStatus(status)
	if unsupported {
		normalized = propertyStatusUnsupported
	}
	trimmed := strings.TrimSpace(raw)
	if normalized == propertyStatusOK && trimmed == "" {
		normalized = propertyStatusEmpty
	}
	return trimmed, normalized
}

func normalizeOptionalStringField(raw *string, unsupported bool, status string) (*string, string) {
	normalized := normalizePropertyStatus(status)
	if unsupported {
		return nil, propertyStatusUnsupported
	}
	if raw == nil {
		if normalized == propertyStatusOK {
			normalized = propertyStatusEmpty
		}
		return nil, normalized
	}
	trimmed := strings.TrimSpace(*raw)
	if trimmed == "" {
		if normalized == propertyStatusOK {
			normalized = propertyStatusEmpty
		}
		return nil, normalized
	}
	return &trimmed, normalized
}

func normalizeScalarField(raw int, unsupported bool, status string) (int, string) {
	normalized := normalizePropertyStatus(status)
	if unsupported {
		return 0, propertyStatusUnsupported
	}
	if raw <= 0 && normalized == propertyStatusOK {
		normalized = propertyStatusEmpty
	}
	return raw, normalized
}

func normalizeBoolField(raw bool, unsupported bool, status string) (bool, string) {
	normalized := normalizePropertyStatus(status)
	if unsupported {
		return false, propertyStatusUnsupported
	}
	return raw, normalized
}

func normalizeRectField(raw *Rect, unsupported bool, status string) (*Rect, string) {
	normalized := normalizePropertyStatus(status)
	if unsupported {
		return nil, propertyStatusUnsupported
	}
	if raw == nil || raw.Width <= 0 || raw.Height <= 0 {
		if normalized == propertyStatusOK {
			normalized = propertyStatusEmpty
		}
		return nil, normalized
	}
	return raw, normalized
}

func normalizePropertyStatus(status string) string {
	switch strings.TrimSpace(strings.ToLower(status)) {
	case propertyStatusUnsupported:
		return propertyStatusUnsupported
	case propertyStatusEmpty:
		return propertyStatusEmpty
	case propertyStatusUnavailable:
		return propertyStatusUnavailable
	case propertyStatusStale:
		return propertyStatusStale
	default:
		return propertyStatusOK
	}
}

func toRect(r *uiaRect) *Rect {
	if r == nil {
		return nil
	}
	return &Rect{Left: r.Left, Top: r.Top, Width: r.Width, Height: r.Height}
}

func normalizeControlType(controlType, localized string) string {
	c := strings.TrimSpace(controlType)
	if c != "" {
		return strings.ToUpper(c[:1]) + c[1:]
	}
	l := strings.TrimSpace(localized)
	if l == "" {
		return ""
	}
	return strings.ToUpper(l[:1]) + l[1:]
}

func normalizeLocalizedControlType(localized, controlType string) string {
	l := strings.TrimSpace(localized)
	if l != "" {
		return strings.ToLower(l)
	}
	c := strings.TrimSpace(controlType)
	if c == "" {
		return ""
	}
	return strings.ToLower(c)
}

func patternActionsFromSupported(patterns []string) []PatternAction {
	return patternActionsForElement(&uiaElement{SupportedPatterns: patterns, IsEnabled: true})
}

func patternActionsForElement(el *uiaElement) []PatternAction {
	if el == nil {
		return nil
	}
	patterns := el.SupportedPatterns
	enabled := el.IsEnabled
	toggleState := firstNonEmpty(el.PropertyStates["ToggleState"], el.PropertyStates["Toggle.ToggleState"])
	expandCollapseState := firstNonEmpty(el.PropertyStates["ExpandCollapseState"], el.PropertyStates["ExpandCollapse.ExpandCollapseState"])
	valueState := firstNonEmpty(el.PropertyStates["Value"], el.PropertyStates["Value.Value"], el.PropertyStates["Value.IsReadOnly"])
	selectionHint := firstNonEmpty(el.PropertyStates["SelectionItem.IsSelected"], el.PropertyStates["Selection.IsSelectionRequired"], el.PropertyStates["Selection.CanSelectMultiple"])

	stateReason := func(prefix, value string) string {
		if strings.TrimSpace(value) == "" {
			return ""
		}
		return prefix + ":" + strings.TrimSpace(value)
	}

	base := []PreconditionStatus{{Name: "enabled", Satisfied: enabled}}
	m := map[string][]PatternAction{
		"Invoke": {
			{Pattern: "Invoke", Action: "invoke", DisplayName: "Invoke", Supported: true, Enabled: enabled, Preconditions: append([]PreconditionStatus(nil), base...)},
		},
		"LegacyIAccessible": {
			{Pattern: "LegacyIAccessible", Action: "doDefaultAction", DisplayName: "Default Action", Supported: true, Enabled: enabled, Preconditions: append([]PreconditionStatus(nil), base...)},
		},
		"ExpandCollapse": {
			{Pattern: "ExpandCollapse", Action: "expand", DisplayName: "Expand", Supported: true, Enabled: enabled, Preconditions: append(append([]PreconditionStatus(nil), base...), PreconditionStatus{Name: "expandCollapseState", Satisfied: true, Reason: stateReason("state", expandCollapseState)})},
			{Pattern: "ExpandCollapse", Action: "collapse", DisplayName: "Collapse", Supported: true, Enabled: enabled, Preconditions: append(append([]PreconditionStatus(nil), base...), PreconditionStatus{Name: "expandCollapseState", Satisfied: true, Reason: stateReason("state", expandCollapseState)})},
		},
		"SelectionItem": {
			{Pattern: "SelectionItem", Action: "select", DisplayName: "Select", Supported: true, Enabled: enabled, Preconditions: append(append([]PreconditionStatus(nil), base...), PreconditionStatus{Name: "selectionHint", Satisfied: true, Reason: stateReason("selection", selectionHint)})},
		},
		"Value": {
			{Pattern: "Value", Action: "setValue", DisplayName: "Set Value", PayloadSchema: `{"type":"object","required":["value"]}`, RequiredArgs: []string{"value"}, Supported: true, Enabled: enabled, Preconditions: append(append([]PreconditionStatus(nil), base...), PreconditionStatus{Name: "valueState", Satisfied: true, Reason: stateReason("value", valueState)}, PreconditionStatus{Name: "input:value", Satisfied: true, Reason: "requires payload.value"})},
		},
		"Toggle": {
			{Pattern: "Toggle", Action: "toggle", DisplayName: "Toggle", Supported: true, Enabled: enabled, Preconditions: append(append([]PreconditionStatus(nil), base...), PreconditionStatus{Name: "toggleState", Satisfied: true, Reason: stateReason("state", toggleState)})},
		},
	}
	seen := map[string]bool{}
	var out []PatternAction
	for _, p := range patterns {
		for _, a := range m[strings.TrimSpace(p)] {
			key := a.Pattern + ":" + a.Action
			if !seen[key] {
				seen[key] = true
				out = append(out, a)
			}
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Pattern == out[j].Pattern {
			return out[i].Action < out[j].Action
		}
		return out[i].Pattern < out[j].Pattern
	})
	return out
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

type selectorContext struct {
	siblings []*uiaElement
	ancestry []*uiaElement
}

func selectorCandidatesForElement(el *uiaElement) (*Selector, []SelectorCandidate) {
	return selectorCandidatesForElementWithContext(el, selectorContext{})
}

func selectorCandidatesForElementWithContext(el *uiaElement, ctx selectorContext) (*Selector, []SelectorCandidate) {
	if el == nil {
		return nil, nil
	}
	type cand struct {
		sel   Selector
		why   string
		score int
		src   string
		tie   int
	}
	var cands []cand
	ct := normalizeControlType(el.ControlType, el.LocalizedControlType)
	if aid := strings.TrimSpace(el.AutomationID); aid != "" && ct != "" {
		score, rationale, tie := candidateUniquenessScore(ctx.siblings, el, func(other *uiaElement) bool {
			return strings.EqualFold(strings.TrimSpace(other.AutomationID), aid) && normalizeControlType(other.ControlType, other.LocalizedControlType) == ct
		}, "automation id + control type is unique among siblings")
		cands = append(cands, cand{sel: Selector{AutomationID: aid, ControlType: ct}, why: rationale, score: 100 + score, src: "automationId+controlType", tie: tie})
	}
	if n := strings.TrimSpace(el.Name); n != "" && ct != "" {
		score, rationale, tie := candidateUniquenessScore(ctx.siblings, el, func(other *uiaElement) bool {
			return strings.EqualFold(strings.TrimSpace(other.Name), n) && normalizeControlType(other.ControlType, other.LocalizedControlType) == ct
		}, "name + control type is unique among siblings")
		cands = append(cands, cand{sel: Selector{Name: n, ControlType: ct}, why: rationale, score: 80 + score, src: "name+controlType", tie: tie})
	}
	if n := strings.TrimSpace(el.Name); n != "" {
		ancestor := firstNonEmpty(ancestryControlTypes(ctx.ancestry)...)
		score, rationale, tie := candidateUniquenessScore(ctx.siblings, el, func(other *uiaElement) bool {
			return strings.EqualFold(strings.TrimSpace(other.Name), n)
		}, "name with ancestry disambiguates siblings")
		cands = append(cands, cand{sel: Selector{Name: n, FrameworkID: ancestor}, why: rationale, score: 60 + score, src: "name+ancestry", tie: tie})
	}
	if class := strings.TrimSpace(el.ClassName); class != "" {
		ancestor := firstNonEmpty(ancestryControlTypes(ctx.ancestry)...)
		score, rationale, tie := candidateUniquenessScore(ctx.siblings, el, func(other *uiaElement) bool {
			return strings.EqualFold(strings.TrimSpace(other.ClassName), class)
		}, "class name with ancestry fallback")
		cands = append(cands, cand{sel: Selector{ClassName: class, FrameworkID: ancestor}, why: rationale, score: 40 + score, src: "className+ancestry", tie: tie})
	}
	if aid := strings.TrimSpace(el.AutomationID); aid != "" {
		cands = append(cands, cand{sel: Selector{AutomationID: aid}, why: "fallback automation id", score: 20, src: "fallback", tie: len(cands)})
	} else if n := strings.TrimSpace(el.Name); n != "" {
		cands = append(cands, cand{sel: Selector{Name: n}, why: "fallback name", score: 10, src: "fallback", tie: len(cands)})
	}
	sort.SliceStable(cands, func(i, j int) bool {
		if cands[i].score == cands[j].score {
			return cands[i].tie < cands[j].tie
		}
		return cands[i].score > cands[j].score
	})
	out := make([]SelectorCandidate, 0, len(cands))
	for i, c := range cands {
		out = append(out, SelectorCandidate{Rank: i + 1, Selector: c.sel, Rationale: c.why, Score: c.score, Source: c.src, Meta: map[string]any{"index": strconv.Itoa(i)}})
	}
	if len(out) == 0 {
		return nil, nil
	}
	best := out[0].Selector
	return &best, out
}

func candidateUniquenessScore(siblings []*uiaElement, self *uiaElement, matches func(*uiaElement) bool, uniqueReason string) (int, string, int) {
	if len(siblings) == 0 {
		return 0, uniqueReason, 1
	}
	matchesCount := 0
	for _, sibling := range siblings {
		if sibling == nil || sibling.Ref == self.Ref {
			continue
		}
		if matches(sibling) {
			matchesCount++
		}
	}
	if matchesCount == 0 {
		return 10, uniqueReason, 0
	}
	return 0, "candidate may match sibling peers", matchesCount
}

func ancestryControlTypes(ancestry []*uiaElement) []string {
	out := make([]string, 0, len(ancestry))
	for _, ancestor := range ancestry {
		if ancestor == nil {
			continue
		}
		out = append(out, normalizeControlType(ancestor.ControlType, ancestor.LocalizedControlType))
	}
	return out
}
