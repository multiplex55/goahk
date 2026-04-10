package inspect

import (
	"context"
	"errors"
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
}

func (f *fakeAdapter) ResolveWindowRoot(context.Context, string) (*uiaElement, error) {
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
func (f *fakeAdapter) GetElementByRef(context.Context, string) (*uiaElement, error) {
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

func TestProviderAdapter_PropertyMappingAndNormalization(t *testing.T) {
	help := "h"
	el := &uiaElement{RuntimeID: " 1.2 ", Name: "Save", ControlType: "button", LocalizedControlType: "BUTTON", AutomationID: "SaveBtn", BoundingRect: &uiaRect{Left: 1, Top: 2, Width: 3, Height: 4}, HelpText: &help, SupportedPatterns: []string{"Invoke", "Value"}}
	mapped := toInspectElement("node:x", "node:p", el)
	if mapped.ControlType != "Button" || mapped.LocalizedControlType != "button" {
		t.Fatalf("unexpected control type normalization: %#v", mapped)
	}
	if mapped.BoundingRect == nil || mapped.BoundingRect.Width != 3 {
		t.Fatalf("bounding rect mapping failed: %#v", mapped.BoundingRect)
	}
	if mapped.HelpText == nil || *mapped.HelpText != "h" {
		t.Fatalf("helpText not mapped")
	}
	if len(mapped.Patterns) != 2 {
		t.Fatalf("pattern extraction failed: %+v", mapped.Patterns)
	}
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
		root: &uiaElement{Ref: "root", Name: "Root"},
		kids: map[string][]*uiaElement{
			"root": {{Ref: "c1", ParentRef: "root", Name: "C1"}},
			"c1":   {{Ref: "root", ParentRef: "c1", Name: "RootCycle"}},
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
	if children[0].NodeID != "node:c1" {
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
