//go:build windows

package main

import (
	"context"
	"testing"

	"goahk/internal/inspect"
)

func TestMapPatternTreeGroupsAndLabels(t *testing.T) {
	nodes := mapPatternTree([]inspect.PatternActionDTO{{Pattern: "InvokePattern", Name: "invoke"}, {Pattern: "LegacyIAccessiblePattern", Name: "do_default_action"}, {Pattern: "SelectionItemPattern", Name: "select"}, {Pattern: "ValuePattern", Name: "set_value"}})
	if len(nodes) != 4 {
		t.Fatalf("unexpected group count: %d", len(nodes))
	}
	if nodes[0].label != "InvokePattern" {
		t.Fatalf("unexpected deterministic group order: %q", nodes[0].label)
	}
}

func TestPatternTreeModel_ActionDispatchUsesActionID(t *testing.T) {
	nodes := mapPatternTree([]inspect.PatternActionDTO{{Pattern: "ValuePattern", Name: "set_value"}})
	m := newPatternTreeModel()
	m.SetRoots(nodes)
	n, ok := m.NodeByID("ValuePattern/setValue")
	if !ok || n.ActionID() != ActionID("setValue") {
		t.Fatalf("action mismatch got=%v ok=%v", n, ok)
	}
}

func TestPatternTreeUsesDisplayNameWhenAvailable(t *testing.T) {
	nodes := mapPatternTree([]inspect.PatternActionDTO{{Pattern: "InvokePattern", Name: "invoke", DisplayName: "Click"}, {Pattern: "InvokePattern", Name: "toggle"}})
	if got := nodes[0].children[0].label; got != "Click" {
		t.Fatalf("expected display name, got %q", got)
	}
	if got := nodes[0].children[1].label; got != "Toggle()" {
		t.Fatalf("expected callable fallback, got %q", got)
	}
}

func TestPatternParentNodesAreNonActionable(t *testing.T) {
	nodes := mapPatternTree([]inspect.PatternActionDTO{{Pattern: "InvokePattern", Name: "invoke"}})
	parent := nodes[0]
	if parent.IsActionableLeaf() {
		t.Fatalf("pattern parent node should be non-actionable")
	}
	child := parent.children[0]
	if !child.IsActionableLeaf() {
		t.Fatalf("action child node should be actionable")
	}
}

func TestController_InvokeSetValue_DialogFlow(t *testing.T) {
	svc := &fakeInspectService{nodeDetailsResp: inspect.GetNodeDetailsResponse{Element: inspect.ElementPropertiesDTO{NodeID: "node-1"}}}
	c := NewController(context.Background(), svc)
	c.selectedNodeID = "node-1"

	c.SetDialogs(&fakeDialogs{ok: false})
	if _, accepted, err := c.InvokeSetValue(); err != nil {
		t.Fatalf("cancel should be noop: %v", err)
	} else if accepted {
		t.Fatalf("cancel should not accept")
	}
	if len(svc.invokeReqs) != 0 {
		t.Fatalf("expected no invoke on cancel")
	}

	c.SetDialogs(&fakeDialogs{ok: true, value: "hello"})
	if _, accepted, err := c.InvokeSetValue(); err != nil {
		t.Fatalf("confirm failed: %v", err)
	} else if !accepted {
		t.Fatalf("confirm should accept")
	}
	if len(svc.invokeReqs) != 1 || svc.invokeReqs[0].Action != "setValue" || svc.invokeReqs[0].Payload["value"] != "hello" {
		t.Fatalf("unexpected invoke req: %+v", svc.invokeReqs)
	}
}

func TestPatternTreePrimaryPathShowsGroupedPatterns(t *testing.T) {
	m := newPatternTreeModel()
	m.SetRoots(mapPatternTree([]inspect.PatternActionDTO{{Pattern: "InvokePattern", Name: "invoke"}}))
	if m.RootCount() != 1 {
		t.Fatalf("expected one grouped root, got %d", m.RootCount())
	}
	root := m.RootAt(0)
	node, ok := root.(*patternTreeNode)
	if !ok || node.label != "InvokePattern" {
		t.Fatalf("expected InvokePattern root, got %#v", root)
	}
	if node.ChildCount() != 1 {
		t.Fatalf("expected one action child, got %d", node.ChildCount())
	}
}

func TestPatternTreeShowsNoSupportedPatternsPlaceholder(t *testing.T) {
	m := newPatternTreeModel()
	m.SetRoots(nil)
	if m.RootCount() != 1 {
		t.Fatalf("expected placeholder root count 1, got %d", m.RootCount())
	}
	root := m.RootAt(0)
	node, ok := root.(*patternTreeNode)
	if !ok {
		t.Fatalf("expected *patternTreeNode root, got %T", root)
	}
	if node.id != emptyPatternNodeID || node.label != "No supported patterns" {
		t.Fatalf("unexpected placeholder node: %#v", node)
	}
	if node.IsActionableLeaf() {
		t.Fatalf("placeholder node should not be actionable")
	}
}
