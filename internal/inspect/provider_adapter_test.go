package inspect

import (
	"context"
	"errors"
	"testing"
)

type fakeAdapter struct {
	root     *uiaElement
	focused  *uiaElement
	under    *uiaElement
	byRef    map[string]*uiaElement
	kids     map[string][]*uiaElement
	cursorX  int
	cursorY  int
	focusErr error
	pointErr error
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
	if children, ok := f.kids[ref]; ok {
		return children, nil
	}
	return nil, nil
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

func TestProviderAdapter_TraversalParentChildrenAndStaleNode(t *testing.T) {
	adapter := &fakeAdapter{root: &uiaElement{Ref: "root", Name: "Root"}, kids: map[string][]*uiaElement{"root": {{Ref: "c1", ParentRef: "root", Name: "C1"}}}}
	core := newProviderCore(adapter)
	root, err := core.treeRoot(context.Background(), "0x1")
	if err != nil {
		t.Fatalf("treeRoot: %v", err)
	}
	children, err := core.nodeChildren(context.Background(), root.NodeID)
	if err != nil {
		t.Fatalf("nodeChildren: %v", err)
	}
	if len(children) != 1 || children[0].NodeID == "" {
		t.Fatalf("children mapping failed: %+v", children)
	}
	if _, err := core.nodeChildren(context.Background(), "node:missing"); !errors.Is(err, ErrStaleCache) {
		t.Fatalf("expected stale error for unknown node, got %v", err)
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
