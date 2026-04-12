package inspect

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"syscall"
	"testing"
)

type fakeAdapter struct {
	root              *uiaElement
	focused           *uiaElement
	under             *uiaElement
	byRef             map[string]*uiaElement
	kids              map[string][]*uiaElement
	childCount        map[string]int
	childrenCallCount map[string]int
	cursorX           int
	cursorY           int
	focusErr          error
	pointErr          error
	getChildrenErr    map[string]error
	invokeErr         error
	selectErr         error
	setValueErr       error
	doDefaultErr      error
	toggleErr         error
	expandErr         error
	collapseErr       error
	invokeCount       int
	selectCount       int
	setValueCount     int
	doDefaultCount    int
	toggleCount       int
	expandCount       int
	collapseCount     int
	lastSetValue      string
	resolveRootCalls  int
}

func (f *fakeAdapter) ResolveWindowRoot(context.Context, string) (*uiaElement, error) {
	f.resolveRootCalls++
	return f.root, nil
}
func (f *fakeAdapter) GetFocusedElement(context.Context) (*uiaElement, error) {
	return f.focused, f.focusErr
}
func (f *fakeAdapter) GetCursorPosition(context.Context) (int, int, error) {
	return f.cursorX, f.cursorY, nil
}
func (f *fakeAdapter) ElementFromPoint(context.Context, int, int) (*uiaElement, error) {
	return f.under, f.pointErr
}
func (f *fakeAdapter) GetElementByRef(_ context.Context, ref string) (*uiaElement, error) {
	if f.byRef != nil {
		if el, ok := f.byRef[ref]; ok {
			return el, nil
		}
	}
	return nil, errors.New("not implemented")
}
func (f *fakeAdapter) GetParent(context.Context, string) (*uiaElement, error) {
	return nil, errors.New("not implemented")
}
func (f *fakeAdapter) GetChildren(_ context.Context, ref string) ([]*uiaElement, error) {
	if f.childrenCallCount == nil {
		f.childrenCallCount = map[string]int{}
	}
	f.childrenCallCount[ref]++
	if f.getChildrenErr != nil {
		if err, ok := f.getChildrenErr[ref]; ok {
			return nil, err
		}
	}
	if children, ok := f.kids[ref]; ok {
		return children, nil
	}
	return nil, nil
}
func (f *fakeAdapter) GetChildCount(_ context.Context, ref string) (int, bool, error) {
	if f.childCount == nil {
		return 0, false, nil
	}
	count, ok := f.childCount[ref]
	return count, ok, nil
}
func (f *fakeAdapter) Invoke(context.Context, string) error {
	f.invokeCount++
	return f.invokeErr
}
func (f *fakeAdapter) Select(context.Context, string) error {
	f.selectCount++
	return f.selectErr
}
func (f *fakeAdapter) SetValue(_ context.Context, _ string, value string) error {
	f.setValueCount++
	f.lastSetValue = value
	return f.setValueErr
}
func (f *fakeAdapter) DoDefaultAction(context.Context, string) error {
	f.doDefaultCount++
	if f.doDefaultErr != nil {
		return f.doDefaultErr
	}
	return nil
}
func (f *fakeAdapter) Toggle(context.Context, string) error {
	f.toggleCount++
	if f.toggleErr != nil {
		return f.toggleErr
	}
	return nil
}
func (f *fakeAdapter) Expand(context.Context, string) error {
	f.expandCount++
	if f.expandErr != nil {
		return f.expandErr
	}
	return nil
}
func (f *fakeAdapter) Collapse(context.Context, string) error {
	f.collapseCount++
	if f.collapseErr != nil {
		return f.collapseErr
	}
	return nil
}

func TestProviderAdapter_PropertyMappingAndNormalization(t *testing.T) {
	help := "h"
	el := &uiaElement{
		RuntimeID: " 1.2 ", Name: "Save", ControlType: "button", LocalizedControlType: "BUTTON", AutomationID: "SaveBtn",
		BoundingRect: &uiaRect{Left: 1, Top: 2, Width: 3, Height: 4}, HelpText: &help, SupportedPatterns: []string{"Invoke", "Value"},
		UnsupportedProps: map[string]bool{"HelpText": true},
	}
	mapped := toInspectElement("node:x", "node:p", el)
	if mapped.ControlType != "Button" || mapped.LocalizedControlType != "button" {
		t.Fatalf("unexpected control type normalization: %#v", mapped)
	}
	if mapped.BoundingRect == nil || mapped.BoundingRect.Width != 3 {
		t.Fatalf("bounding rect mapping failed: %#v", mapped.BoundingRect)
	}
	if mapped.HelpText != nil {
		t.Fatalf("helpText should be cleared when property is unsupported")
	}
	if len(mapped.Patterns) != 2 {
		t.Fatalf("pattern extraction failed: %+v", mapped.Patterns)
	}
	if !mapped.UnsupportedProps["HelpText"] {
		t.Fatalf("unsupported property metadata not mapped")
	}
}

func TestProviderAdapter_NormalizationSemanticStates(t *testing.T) {
	t.Run("string normalization states", func(t *testing.T) {
		cases := []struct {
			name        string
			raw         string
			unsupported bool
			state       string
			wantValue   string
			wantStatus  string
		}{
			{name: "unsupported overrides", raw: "value", unsupported: true, state: propertyStatusOK, wantStatus: propertyStatusUnsupported, wantValue: "value"},
			{name: "empty inferred from blank", raw: "   ", wantStatus: propertyStatusEmpty},
			{name: "explicit unavailable retained", raw: "", state: propertyStatusUnavailable, wantStatus: propertyStatusUnavailable},
			{name: "stale retained", raw: "", state: propertyStatusStale, wantStatus: propertyStatusStale},
			{name: "ok with value", raw: " value ", wantStatus: propertyStatusOK, wantValue: "value"},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				gotValue, gotStatus := normalizeStringField(tc.raw, tc.unsupported, tc.state)
				if gotValue != tc.wantValue || gotStatus != tc.wantStatus {
					t.Fatalf("normalizeStringField(%q)=(%q,%q) want (%q,%q)", tc.raw, gotValue, gotStatus, tc.wantValue, tc.wantStatus)
				}
			})
		}
	})

	t.Run("bool normalization keeps explicit false", func(t *testing.T) {
		gotValue, gotStatus := normalizeBoolField(false, false, propertyStatusOK)
		if gotStatus != propertyStatusOK || gotValue {
			t.Fatalf("normalizeBoolField false should remain explicit false, got (%v,%q)", gotValue, gotStatus)
		}
	})

	t.Run("scalar normalization", func(t *testing.T) {
		_, gotEmpty := normalizeScalarField(0, false, propertyStatusOK)
		if gotEmpty != propertyStatusEmpty {
			t.Fatalf("expected empty scalar status for zero, got %q", gotEmpty)
		}
		gotValue, gotOK := normalizeScalarField(42, false, propertyStatusOK)
		if gotValue != 42 || gotOK != propertyStatusOK {
			t.Fatalf("expected scalar value to survive normalization, got (%d,%q)", gotValue, gotOK)
		}
	})

	t.Run("rect normalization", func(t *testing.T) {
		rect, status := normalizeRectField(&Rect{Left: 0, Top: 0, Width: 0, Height: 10}, false, propertyStatusOK)
		if rect != nil || status != propertyStatusEmpty {
			t.Fatalf("expected invalid rect to normalize as empty, got rect=%+v status=%q", rect, status)
		}
	})
}

func TestProviderAdapter_CacheBehavior(t *testing.T) {
	t.Run("first load misses and second load hits cache", func(t *testing.T) {
		adapter := &fakeAdapter{root: &uiaElement{Ref: "root", Name: "Root"}, kids: map[string][]*uiaElement{"root": {{Ref: "c1", ParentRef: "root", Name: "C1"}}}}
		core := newProviderCore(adapter)
		root, err := core.treeRoot(context.Background(), "0x1", false)
		if err != nil {
			t.Fatalf("treeRoot: %v", err)
		}
		if _, err := core.nodeChildren(context.Background(), root.NodeID); err != nil {
			t.Fatalf("nodeChildren miss: %v", err)
		}
		if _, err := core.nodeChildren(context.Background(), root.NodeID); err != nil {
			t.Fatalf("nodeChildren hit: %v", err)
		}
		if got := adapter.childrenCallCount["root"]; got != 1 {
			t.Fatalf("expected one adapter call after cache hit, got %d", got)
		}
	})

	t.Run("window switch invalidates cache", func(t *testing.T) {
		adapter := &fakeAdapter{root: &uiaElement{Ref: "root", Name: "Root"}, kids: map[string][]*uiaElement{"root": {{Ref: "c1", ParentRef: "root", Name: "C1"}}}}
		core := newProviderCore(adapter)
		root, _ := core.treeRoot(context.Background(), "0x1", false)
		_, _ = core.nodeChildren(context.Background(), root.NodeID)
		_, _ = core.treeRoot(context.Background(), "0x2", false)
		root2, _ := core.treeRoot(context.Background(), "0x1", false)
		_, _ = core.nodeChildren(context.Background(), root2.NodeID)
		if got := adapter.childrenCallCount["root"]; got < 2 {
			t.Fatalf("expected cache invalidation across window switch, calls=%d", got)
		}
	})

	t.Run("refresh invalidates cache", func(t *testing.T) {
		adapter := &fakeAdapter{root: &uiaElement{Ref: "root", Name: "Root"}, kids: map[string][]*uiaElement{"root": {{Ref: "c1", ParentRef: "root", Name: "C1"}}}}
		core := newProviderCore(adapter)
		root, _ := core.treeRoot(context.Background(), "0x1", false)
		_, _ = core.nodeChildren(context.Background(), root.NodeID)
		_, _ = core.treeRoot(context.Background(), "0x1", true)
		rootRefreshed, _ := core.treeRoot(context.Background(), "0x1", false)
		_, _ = core.nodeChildren(context.Background(), rootRefreshed.NodeID)
		if got := adapter.childrenCallCount["root"]; got < 2 {
			t.Fatalf("expected reload after refresh, calls=%d", got)
		}
	})

	t.Run("stale node fallback reloads", func(t *testing.T) {
		adapter := &fakeAdapter{
			root:           &uiaElement{Ref: "root", Name: "Root"},
			kids:           map[string][]*uiaElement{"root": {{Ref: "c1", ParentRef: "root", Name: "C1"}}},
			getChildrenErr: map[string]error{"root": errStaleElementReference},
		}
		core := newProviderCore(adapter)
		root, _ := core.treeRoot(context.Background(), "0x1", false)
		if _, err := core.nodeChildren(context.Background(), root.NodeID); !errors.Is(err, ErrStaleCache) {
			t.Fatalf("expected stale cache error when provider remains stale, got %v", err)
		}
		adapter.getChildrenErr = nil
		children, err := core.nodeChildren(context.Background(), root.NodeID)
		if err != nil {
			t.Fatalf("expected fallback to recover after stale clears: %v", err)
		}
		if len(children) != 1 {
			t.Fatalf("expected one child after reload, got %d", len(children))
		}
	})
}

func TestProviderAdapter_TreeAPIContract(t *testing.T) {
	adapter := &fakeAdapter{
		root: &uiaElement{Ref: "root", RuntimeID: "1", Name: "Root"},
		kids: map[string][]*uiaElement{
			"root": {{Ref: "c1", RuntimeID: "2", ParentRef: "root", Name: "C1"}},
			"c1":   {{Ref: "root", RuntimeID: "1", ParentRef: "c1", Name: "RootCycle"}},
		},
		childCount: map[string]int{"root": 1},
	}
	core := newProviderCore(adapter)

	root, err := core.treeRoot(context.Background(), "0x1", false)
	if err != nil {
		t.Fatalf("treeRoot: %v", err)
	}
	if root.Expanded {
		t.Fatalf("root should be collapsed in root-only payload")
	}
	if root.ChildCount == nil || *root.ChildCount != 1 {
		t.Fatalf("expected cheap child count on root, got %+v", root.ChildCount)
	}

	children, err := core.nodeChildren(context.Background(), root.NodeID)
	if err != nil {
		t.Fatalf("nodeChildren: %v", err)
	}
	if len(children) != 1 {
		t.Fatalf("expected direct children only, got %d", len(children))
	}
	if children[0].NodeID != "node:rid:2" {
		t.Fatalf("expected direct child c1, got %s", children[0].NodeID)
	}

	grandchildren, err := core.nodeChildren(context.Background(), children[0].NodeID)
	if err != nil {
		t.Fatalf("grandchildren: %v", err)
	}
	if len(grandchildren) != 1 || !grandchildren[0].Cycle {
		t.Fatalf("expected cycle marker on back-edge node, got %+v", grandchildren)
	}
}

func TestProviderAdapter_RuntimeIDStabilityAcrossRefresh(t *testing.T) {
	adapter := &fakeAdapter{
		root: &uiaElement{Ref: "root#v1", RuntimeID: "42", Name: "Root"},
	}
	core := newProviderCore(adapter)
	first, err := core.treeRoot(context.Background(), "0x1", false)
	if err != nil {
		t.Fatalf("treeRoot first: %v", err)
	}
	adapter.root = &uiaElement{Ref: "root#v2", RuntimeID: "42", Name: "Root Updated"}
	second, err := core.treeRoot(context.Background(), "0x1", true)
	if err != nil {
		t.Fatalf("treeRoot second: %v", err)
	}
	if first.NodeID != second.NodeID {
		t.Fatalf("expected stable node identity from runtime ID, got %q vs %q", first.NodeID, second.NodeID)
	}
}

func TestProviderAdapter_FixtureDrivenNotepadLikeTree(t *testing.T) {
	fixture := map[string][]*uiaElement{
		"root": {
			{Ref: "menu", RuntimeID: "101", ParentRef: "root", Name: "Application", ControlType: "MenuBar"},
			{Ref: "editor", RuntimeID: "102", ParentRef: "root", Name: "Text Editor", ControlType: "Edit"},
			{Ref: "status", RuntimeID: "103", ParentRef: "root", Name: "Ready", ControlType: "StatusBar"},
		},
	}
	adapter := &fakeAdapter{
		root: &uiaElement{Ref: "root", RuntimeID: "100", Name: "Untitled - Notepad", ControlType: "Window"},
		kids: fixture,
	}
	core := newProviderCore(adapter)
	root, err := core.treeRoot(context.Background(), "0x1", false)
	if err != nil {
		t.Fatalf("treeRoot: %v", err)
	}
	children, err := core.nodeChildren(context.Background(), root.NodeID)
	if err != nil {
		t.Fatalf("nodeChildren: %v", err)
	}
	if len(children) != 3 || children[1].ControlType != "Edit" {
		t.Fatalf("unexpected notepad fixture descendants: %+v", children)
	}
}

func TestProviderAdapter_PatternMappingAndSelectorFallbacks(t *testing.T) {
	el := &uiaElement{Name: "OK", ControlType: "Button", ClassName: "ButtonClass", FrameworkID: "Win32", SupportedPatterns: []string{"ExpandCollapse", "Toggle"}}
	best, suggestions := selectorCandidatesForElement(el)
	if best == nil || best.Name != "OK" {
		t.Fatalf("expected name-based selector fallback, got %+v", best)
	}
	if len(suggestions) == 0 {
		t.Fatalf("expected selector suggestions")
	}
	patterns := patternActionsFromSupported(el.SupportedPatterns)
	if len(patterns) != 3 {
		t.Fatalf("expected expand/collapse/toggle actions, got %+v", patterns)
	}
}

func TestSelectorCandidatesForElement_DeterministicScoring(t *testing.T) {
	fixtures := []struct {
		name       string
		element    *uiaElement
		wantSource []string
		wantScores []int
	}{
		{
			name: "automation id dominates",
			element: &uiaElement{
				Name:         "Search",
				ControlType:  "Edit",
				AutomationID: "SearchBox",
				ClassName:    "Edit",
				FrameworkID:  "Win32",
			},
			wantSource: []string{"automationId", "automationId+controlType", "name+controlType", "name", "class+framework"},
			wantScores: []int{100, 95, 70, 40, 35},
		},
		{
			name: "name fallback only",
			element: &uiaElement{
				Name:        "Untitled - Notepad",
				ControlType: "Window",
			},
			wantSource: []string{"name+controlType", "name"},
			wantScores: []int{70, 40},
		},
	}

	for _, tc := range fixtures {
		t.Run(tc.name, func(t *testing.T) {
			_, got := selectorCandidatesForElement(tc.element)
			if len(got) != len(tc.wantSource) {
				t.Fatalf("selector count=%d want=%d (%+v)", len(got), len(tc.wantSource), got)
			}
			for i := range got {
				if got[i].Source != tc.wantSource[i] {
					t.Fatalf("source[%d]=%q want %q", i, got[i].Source, tc.wantSource[i])
				}
				if got[i].Score != tc.wantScores[i] {
					t.Fatalf("score[%d]=%d want %d", i, got[i].Score, tc.wantScores[i])
				}
			}
		})
	}
}

func TestProviderAdapter_EdgeCases(t *testing.T) {
	if got := toInspectElement("node:nil", "", nil); got.NodeID != "node:nil" {
		t.Fatalf("nil element should still map node metadata")
	}
	if rect := toRect(nil); rect != nil {
		t.Fatalf("nil rect should stay nil")
	}
	core := newProviderCore(&fakeAdapter{focusErr: errNilElementReference})
	if _, err := core.focused(context.Background()); !errors.Is(err, ErrStaleCache) {
		t.Fatalf("nil element reference should map to stale cache, got %v", err)
	}
	core = newProviderCore(&fakeAdapter{pointErr: errors.New("com failed")})
	if _, err := core.underCursor(context.Background()); err == nil {
		t.Fatalf("expected structured error")
	} else {
		var callErr *ProviderCallError
		if !errors.As(err, &callErr) {
			t.Fatalf("expected ProviderCallError, got %T", err)
		}
	}
}

func TestProviderAdapter_DetailsAndFocusCursorResolution(t *testing.T) {
	adapter := &fakeAdapter{
		root:    &uiaElement{Ref: "root", Name: "Root"},
		focused: &uiaElement{Ref: "focus", Name: "Focused", SupportedPatterns: []string{"Invoke"}},
		under:   &uiaElement{Ref: "under", Name: "Under", SupportedPatterns: []string{"Value"}},
		cursorX: 7,
		cursorY: 8,
		byRef: map[string]*uiaElement{
			"root":  {Ref: "root", Name: "Root", ControlType: "Button", AutomationID: "rootBtn", SupportedPatterns: []string{"Invoke", "Value"}},
			"focus": {Ref: "focus", Name: "Focused", SupportedPatterns: []string{"Invoke"}},
		},
	}
	core := newProviderCore(adapter)

	root, err := core.treeRoot(context.Background(), "0x1", false)
	if err != nil {
		t.Fatalf("treeRoot: %v", err)
	}
	details, err := core.inspectByNodeID(context.Background(), root.NodeID)
	if err != nil {
		t.Fatalf("inspectByNodeID: %v", err)
	}
	if details.AutomationID != "rootBtn" {
		t.Fatalf("expected automation id in node details, got %+v", details)
	}
	actions, err := core.getPatternActions(context.Background(), root.NodeID)
	if err != nil {
		t.Fatalf("getPatternActions: %v", err)
	}
	if len(actions) < 2 {
		t.Fatalf("expected invoke/value actions, got %+v", actions)
	}

	focused, err := core.focused(context.Background())
	if err != nil {
		t.Fatalf("focused: %v", err)
	}
	if focused.NodeID == "" {
		t.Fatalf("focused element should resolve node id")
	}
	under, err := core.underCursor(context.Background())
	if err != nil {
		t.Fatalf("underCursor: %v", err)
	}
	if under.NodeID == "" {
		t.Fatalf("under-cursor element should resolve node id")
	}
}

func TestProviderAdapter_InvokePatternDispatcher(t *testing.T) {
	t.Parallel()

	setup := func(patterns []string) (*providerCore, *fakeAdapter, string) {
		adapter := &fakeAdapter{
			root:  &uiaElement{Ref: "root", Name: "Root"},
			kids:  map[string][]*uiaElement{},
			byRef: map[string]*uiaElement{"root": {Ref: "root", SupportedPatterns: patterns}},
		}
		core := newProviderCore(adapter)
		root, _ := core.treeRoot(context.Background(), "0x1", false)
		return core, adapter, root.NodeID
	}

	t.Run("valid routing per action", func(t *testing.T) {
		core, adapter, nodeID := setup([]string{"Invoke", "SelectionItem", "Value", "ExpandCollapse"})
		_, err := core.invokePattern(context.Background(), InvokePatternRequest{NodeID: nodeID, Action: "invoke"})
		if err != nil {
			t.Fatalf("invoke route failed: %v", err)
		}
		_, err = core.invokePattern(context.Background(), InvokePatternRequest{NodeID: nodeID, Action: "select"})
		if err != nil {
			t.Fatalf("select route failed: %v", err)
		}
		_, err = core.invokePattern(context.Background(), InvokePatternRequest{
			NodeID:  nodeID,
			Action:  "setValue",
			Payload: map[string]any{"value": "hello"},
		})
		if err != nil {
			t.Fatalf("setValue route failed: %v", err)
		}
		if adapter.invokeCount != 1 || adapter.selectCount != 1 || adapter.setValueCount != 1 {
			t.Fatalf("unexpected call counts invoke=%d select=%d setValue=%d", adapter.invokeCount, adapter.selectCount, adapter.setValueCount)
		}
		if adapter.lastSetValue != "hello" {
			t.Fatalf("expected propagated value, got %q", adapter.lastSetValue)
		}
		_, err = core.invokePattern(context.Background(), InvokePatternRequest{NodeID: nodeID, Action: "expand"})
		if err != nil {
			t.Fatalf("expand route failed: %v", err)
		}
		_, err = core.invokePattern(context.Background(), InvokePatternRequest{NodeID: nodeID, Action: "collapse"})
		if err != nil {
			t.Fatalf("collapse route failed: %v", err)
		}
		if adapter.expandCount != 1 || adapter.collapseCount != 1 {
			t.Fatalf("unexpected expand/collapse calls expand=%d collapse=%d", adapter.expandCount, adapter.collapseCount)
		}
	})

	t.Run("unknown action rejected", func(t *testing.T) {
		core, _, nodeID := setup([]string{"Invoke"})
		_, err := core.invokePattern(context.Background(), InvokePatternRequest{NodeID: nodeID, Action: "unknown"})
		if !errors.Is(err, ErrUnsupportedPatternAction) {
			t.Fatalf("expected unsupported action, got %v", err)
		}
	})

	t.Run("input required for set value", func(t *testing.T) {
		core, _, nodeID := setup([]string{"Value"})
		_, err := core.invokePattern(context.Background(), InvokePatternRequest{NodeID: nodeID, Action: "setValue"})
		if !errors.Is(err, ErrMissingPatternInput) {
			t.Fatalf("expected missing input, got %v", err)
		}
	})
}

func TestProviderAdapter_InvokePatternCapabilityGating(t *testing.T) {
	t.Parallel()

	t.Run("actions disabled when unsupported", func(t *testing.T) {
		adapter := &fakeAdapter{
			root:  &uiaElement{Ref: "root", Name: "Root"},
			byRef: map[string]*uiaElement{"root": {Ref: "root", SupportedPatterns: []string{}}},
		}
		core := newProviderCore(adapter)
		root, _ := core.treeRoot(context.Background(), "0x1", false)
		_, err := core.invokePattern(context.Background(), InvokePatternRequest{NodeID: root.NodeID, Action: "invoke"})
		if !errors.Is(err, ErrUnsupportedPatternAction) {
			t.Fatalf("expected unsupported action, got %v", err)
		}
	})

	t.Run("supported action invokes correct provider method", func(t *testing.T) {
		adapter := &fakeAdapter{
			root:  &uiaElement{Ref: "root", Name: "Root"},
			byRef: map[string]*uiaElement{"root": {Ref: "root", SupportedPatterns: []string{"SelectionItem"}}},
		}
		core := newProviderCore(adapter)
		root, _ := core.treeRoot(context.Background(), "0x1", false)
		_, err := core.invokePattern(context.Background(), InvokePatternRequest{NodeID: root.NodeID, Action: "select"})
		if err != nil {
			t.Fatalf("select failed: %v", err)
		}
		if adapter.selectCount != 1 || adapter.invokeCount != 0 || adapter.setValueCount != 0 {
			t.Fatalf("unexpected adapter calls invoke=%d select=%d setValue=%d", adapter.invokeCount, adapter.selectCount, adapter.setValueCount)
		}
	})
}

func TestProviderAdapter_InvokePatternFailureModes(t *testing.T) {
	t.Parallel()

	adapter := &fakeAdapter{
		root:      &uiaElement{Ref: "root", Name: "Root"},
		byRef:     map[string]*uiaElement{"root": {Ref: "root", SupportedPatterns: []string{"Invoke"}}},
		invokeErr: fmt.Errorf("com invoke failed"),
	}
	core := newProviderCore(adapter)
	root, _ := core.treeRoot(context.Background(), "0x1", false)

	_, err := core.invokePattern(context.Background(), InvokePatternRequest{NodeID: root.NodeID, Action: "invoke"})
	if !errors.Is(err, ErrPatternExecutionFailure) {
		t.Fatalf("expected execution failure wrapper, got %v", err)
	}
	if err == nil || !strings.Contains(err.Error(), "action=invoke") {
		t.Fatalf("expected action context in error, got %v", err)
	}
}

func TestProviderAdapter_InvokePatternErrorClasses(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		invokeErr error
		class     PatternActionErrorClass
	}{
		{name: "unsupported", invokeErr: ErrProviderActionUnsupported, class: patternErrorClassNotSupported},
		{name: "transient", invokeErr: ErrStaleCache, class: patternErrorClassTransientState},
		{name: "access denied", invokeErr: syscall.EACCES, class: patternErrorClassAccessDenied},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			adapter := &fakeAdapter{
				root:      &uiaElement{Ref: "root", Name: "Root"},
				byRef:     map[string]*uiaElement{"root": {Ref: "root", SupportedPatterns: []string{"Invoke"}}},
				invokeErr: tc.invokeErr,
			}
			core := newProviderCore(adapter)
			root, _ := core.treeRoot(context.Background(), "0x1", false)
			_, err := core.invokePattern(context.Background(), InvokePatternRequest{NodeID: root.NodeID, Action: "invoke"})
			var actionErr *patternActionError
			if !errors.As(err, &actionErr) {
				t.Fatalf("expected patternActionError, got %v", err)
			}
			if actionErr.Class != tc.class {
				t.Fatalf("expected class %s, got %s", tc.class, actionErr.Class)
			}
		})
	}
}
