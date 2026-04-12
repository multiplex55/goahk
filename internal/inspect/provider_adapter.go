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
	SupportedPatterns    []string
}

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
	adapter       uiaAdapter
	mu            sync.RWMutex
	nodeToRef     map[string]string
	parentByID    map[string]string
	childrenCache *nodeChildrenCache
}

func newProviderCore(adapter uiaAdapter) *providerCore {
	return &providerCore{
		adapter:       adapter,
		nodeToRef:     map[string]string{},
		parentByID:    map[string]string{},
		childrenCache: newNodeChildrenCache(),
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
	return toInspectElement(nodeID, p.parentOf(nodeID), el), nil
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
			Supported:     a.Supported,
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
	patterns := patternActionsFromSupported(el.SupportedPatterns)
	names := make([]string, 0, len(patterns))
	for _, p := range patterns {
		names = append(names, p.Action)
	}
	node := TreeNodeDTO{
		NodeID: nodeID, NodeId: nodeID, HWND: strings.TrimSpace(el.HWND),
		Name: el.Name, ControlType: normalizeControlType(el.ControlType, el.LocalizedControlType),
		LocalizedControlType: normalizeLocalizedControlType(el.LocalizedControlType, el.ControlType),
		ClassName:            el.ClassName, Patterns: names,
	}
	if count, ok, err := p.adapter.GetChildCount(context.Background(), el.Ref); err == nil && ok {
		node.ChildCount = &count
		node.HasChildren = count > 0
	}
	return node
}

func runtimeNodeID(runtimeID, ref string) string {
	if rid := strings.TrimSpace(runtimeID); rid != "" {
		return "node:rid:" + rid
	}
	return "node:ref:" + strings.TrimSpace(ref)
}

func (p *providerCore) cacheNodeRef(el *uiaElement) string {
	nodeID := runtimeNodeID(el.RuntimeID, el.Ref)
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
	best, alts := selectorCandidatesForElement(el)
	return InspectElement{
		NodeID: nodeID, RuntimeID: strings.TrimSpace(el.RuntimeID), ParentNodeID: parentNodeID,
		HWND: strings.TrimSpace(el.HWND),
		Name: el.Name, LocalizedControlType: normalizeLocalizedControlType(el.LocalizedControlType, el.ControlType), ControlType: normalizeControlType(el.ControlType, el.LocalizedControlType),
		AutomationID: el.AutomationID, ClassName: el.ClassName, FrameworkID: el.FrameworkID, ProcessID: el.ProcessID,
		HelpText: el.HelpText, AccessKey: el.AccessKey, AcceleratorKey: el.AcceleratorKey, Status: el.Status, Value: el.Value,
		ItemType: el.ItemType, ItemStatus: el.ItemStatus, IsRequiredForForm: el.IsRequiredForForm, LabeledBy: el.LabeledBy,
		BoundingRect: toRect(el.BoundingRect), IsEnabled: el.IsEnabled, IsKeyboardFocusable: el.IsKeyboardFocusable, HasKeyboardFocus: el.HasKeyboardFocus,
		IsOffscreen: el.IsOffscreen, IsContentElement: el.IsContentElement, IsControlElement: el.IsControlElement, IsPassword: el.IsPassword,
		UnsupportedProps: cloneUnsupportedProps(el.UnsupportedProps),
		Patterns:         patternActionsFromSupported(el.SupportedPatterns), BestSelector: best, SelectorSuggestions: alts,
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
	m := map[string][]PatternAction{
		"Invoke":         {{Pattern: "Invoke", Action: "invoke", DisplayName: "Invoke", Supported: true, Preconditions: []PreconditionStatus{{Name: "enabled", Satisfied: true}}}},
		"ExpandCollapse": {{Pattern: "ExpandCollapse", Action: "expand", DisplayName: "Expand", Supported: true, Preconditions: []PreconditionStatus{{Name: "enabled", Satisfied: true}}}, {Pattern: "ExpandCollapse", Action: "collapse", DisplayName: "Collapse", Supported: true, Preconditions: []PreconditionStatus{{Name: "enabled", Satisfied: true}}}},
		"SelectionItem":  {{Pattern: "SelectionItem", Action: "select", DisplayName: "Select", Supported: true, Preconditions: []PreconditionStatus{{Name: "enabled", Satisfied: true}}}},
		"Value":          {{Pattern: "Value", Action: "setValue", DisplayName: "Set Value", PayloadSchema: `{"type":"object","required":["value"]}`, Supported: true, Preconditions: []PreconditionStatus{{Name: "enabled", Satisfied: true}, {Name: "input:value", Satisfied: true, Reason: "requires payload.value"}}}},
		"Toggle":         {{Pattern: "Toggle", Action: "toggle", DisplayName: "Toggle", Supported: true, Preconditions: []PreconditionStatus{{Name: "enabled", Satisfied: true}}}},
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

func selectorCandidatesForElement(el *uiaElement) (*Selector, []SelectorCandidate) {
	if el == nil {
		return nil, nil
	}
	type cand struct {
		sel   Selector
		why   string
		score int
		src   string
	}
	var cands []cand
	if v := strings.TrimSpace(el.AutomationID); v != "" {
		cands = append(cands, cand{sel: Selector{AutomationID: v}, why: "automation id is stable", score: 100, src: "automationId"})
	}
	ct := normalizeControlType(el.ControlType, el.LocalizedControlType)
	if aid := strings.TrimSpace(el.AutomationID); aid != "" && ct != "" {
		cands = append(cands, cand{sel: Selector{AutomationID: aid, ControlType: ct}, why: "automation id with control type narrows duplicates", score: 95, src: "automationId+controlType"})
	}
	if n := strings.TrimSpace(el.Name); n != "" && ct != "" {
		cands = append(cands, cand{sel: Selector{Name: n, ControlType: ct, ClassName: strings.TrimSpace(el.ClassName)}, why: "name+type fallback", score: 70, src: "name+controlType"})
	}
	if strings.TrimSpace(el.Name) != "" {
		cands = append(cands, cand{sel: Selector{Name: strings.TrimSpace(el.Name)}, why: "name-only broad fallback", score: 40, src: "name"})
	}
	if strings.TrimSpace(el.ClassName) != "" && strings.TrimSpace(el.FrameworkID) != "" {
		cands = append(cands, cand{sel: Selector{ClassName: strings.TrimSpace(el.ClassName), FrameworkID: strings.TrimSpace(el.FrameworkID), ControlType: ct}, why: "framework class fallback", score: 35, src: "class+framework"})
	}
	sort.SliceStable(cands, func(i, j int) bool { return cands[i].score > cands[j].score })
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
