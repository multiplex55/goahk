//go:build windows

package main

import (
	"testing"

	"goahk/internal/inspect"
)

func TestUIATreeLabelPrecedence(t *testing.T) {
	m := newUIATreeModel()
	if got := (&uiaTreeNode{TreeNodeDTO: inspect.TreeNodeDTO{NodeID: "n1", DisplayLabel: "Display"}}).Text(); got != "Display" {
		t.Fatalf("got %q", got)
	}
	if got := (&uiaTreeNode{TreeNodeDTO: inspect.TreeNodeDTO{NodeID: "n2", LocalizedControlType: "button", Name: "OK"}}).Text(); got != "button \"OK\"" {
		t.Fatalf("got %q", got)
	}
	_ = m
}

func TestUIATreeLabelUsesLocalizedControlTypeAndQuotedName(t *testing.T) {
	got := (&uiaTreeNode{TreeNodeDTO: inspect.TreeNodeDTO{NodeID: "n", LocalizedControlType: "pane", Name: "Main"}}).Text()
	if got != `pane "Main"` {
		t.Fatalf("got %q", got)
	}
}

func TestUIATreeLabelPreservesEmptyName(t *testing.T) {
	got := (&uiaTreeNode{TreeNodeDTO: inspect.TreeNodeDTO{NodeID: "n", LocalizedControlType: "pane", Name: ""}}).Text()
	if got != `pane ""` {
		t.Fatalf("got %q", got)
	}
}

func TestUIATreeLoadedAndExpandedState(t *testing.T) {
	m := newUIATreeModel()
	m.MarkChildrenLoaded("n1")
	m.SetExpanded("n1", true)
	if !m.AreChildrenLoaded("n1") || !m.IsExpanded("n1") {
		t.Fatalf("expected true state")
	}
}

func TestUIATreeRootAndChildrenLifecycle(t *testing.T) {
	m := newUIATreeModel()
	m.SetRoot(inspect.TreeNodeDTO{NodeID: "root"})
	if got := m.RootID(); got != "root" {
		t.Fatalf("root mismatch: %q", got)
	}
	if !m.ShouldShowLazyPlaceholder("root") {
		t.Fatal("expected lazy placeholder before load")
	}
	m.SetChildren("root", []inspect.TreeNodeDTO{{NodeID: "c1", Name: "first"}, {NodeID: "c2", Name: "second"}})
	if m.ShouldShowLazyPlaceholder("root") {
		t.Fatal("should not show placeholder once children loaded")
	}
	if m.NodeCount() != 3 {
		t.Fatalf("expected node count 3, got %d", m.NodeCount())
	}
	n, ok := m.NodeByID("c1")
	if !ok || n.Name != "first" {
		t.Fatalf("node lookup failed: %+v %v", n, ok)
	}
	if child, ok := m.ItemByID("c1"); !ok || child.Parent() == nil {
		t.Fatalf("expected child parent relation")
	}
	m.Reset()
	if m.RootID() != "" || m.NodeCount() != 0 {
		t.Fatal("expected reset state")
	}
}

func TestUIATreeDuplicateLabelSafety_UsesNodeID(t *testing.T) {
	m := newUIATreeModel()
	m.SetRoot(inspect.TreeNodeDTO{NodeID: "root"})
	m.SetChildren("root", []inspect.TreeNodeDTO{{NodeID: "a", DisplayLabel: "Button"}, {NodeID: "b", DisplayLabel: "Button"}})
	a, _ := m.ItemByID("a")
	b, _ := m.ItemByID("b")
	if a.NodeID == b.NodeID || a.Text() != b.Text() {
		t.Fatalf("expected duplicate labels with distinct node IDs")
	}
}

func TestUIATreeRootIsVisibleAfterSetRoot(t *testing.T) {
	m := newUIATreeModel()
	m.SetRoot(inspect.TreeNodeDTO{NodeID: "root"})
	if m.RootCount() != 1 || m.RootAt(0) == nil {
		t.Fatal("expected root to be visible")
	}
}

func TestUIATreeRootCanHaveLazyPlaceholder(t *testing.T) {
	m := newUIATreeModel()
	m.SetRoot(inspect.TreeNodeDTO{NodeID: "root"})
	root := m.RootAt(0).(*uiaTreeNode)
	if root.ChildCount() == 0 {
		t.Fatal("expected lazy placeholder child")
	}
}

func TestUIATreeSetChildrenRemovesPlaceholder(t *testing.T) {
	m := newUIATreeModel()
	m.SetRoot(inspect.TreeNodeDTO{NodeID: "root"})
	m.SetChildren("root", []inspect.TreeNodeDTO{{NodeID: "a"}})
	root := m.RootAt(0).(*uiaTreeNode)
	if root.ChildCount() != 1 {
		t.Fatalf("expected 1 real child, got %d", root.ChildCount())
	}
	if child := root.ChildAt(0).(*uiaTreeNode); child.placeholder {
		t.Fatal("placeholder should be removed")
	}
}

func TestUIATreeSetChildrenMakesNestedChildrenVisibleImmediately(t *testing.T) {
	m := newUIATreeModel()
	m.SetRoot(inspect.TreeNodeDTO{NodeID: "root"})
	m.SetChildren("root", []inspect.TreeNodeDTO{{NodeID: "parent"}})
	m.SetChildren("parent", []inspect.TreeNodeDTO{{NodeID: "leaf"}})

	parent, ok := m.ItemByID("parent")
	if !ok {
		t.Fatal("expected parent node")
	}
	if parent.ChildCount() != 1 {
		t.Fatalf("expected 1 visible child after update, got %d", parent.ChildCount())
	}
	leaf := parent.ChildAt(0).(*uiaTreeNode)
	if leaf.NodeID != "leaf" || leaf.placeholder {
		t.Fatalf("expected leaf child to be visible immediately, got %#v", leaf)
	}
}

func TestUIATreeIdentityUsesNodeIDNotLabel(t *testing.T) {
	m := newUIATreeModel()
	m.SetRoot(inspect.TreeNodeDTO{NodeID: "root"})
	m.SetChildren("root", []inspect.TreeNodeDTO{{NodeID: "a", DisplayLabel: "Button"}, {NodeID: "b", DisplayLabel: "Button"}})
	a, _ := m.ItemByID("a")
	b, _ := m.ItemByID("b")
	if a == b {
		t.Fatal("distinct IDs must map to distinct nodes")
	}
}

func TestUIATreeAppendChildrenAvoidsDuplicates(t *testing.T) {
	m := newUIATreeModel()
	m.SetRoot(inspect.TreeNodeDTO{NodeID: "root"})
	m.SetChildren("root", []inspect.TreeNodeDTO{{NodeID: "a"}})
	m.AppendChildren("root", []inspect.TreeNodeDTO{{NodeID: "a"}, {NodeID: "b"}})
	root := m.RootAt(0).(*uiaTreeNode)
	if root.ChildCount() != 2 {
		t.Fatalf("expected two unique children, got %d", root.ChildCount())
	}
}

func TestUIATreeUnknownChildCount_ShowsPlaceholderBeforeExpand(t *testing.T) {
	m := newUIATreeModel()
	m.SetRoot(inspect.TreeNodeDTO{NodeID: "root"})
	m.SetChildren("root", []inspect.TreeNodeDTO{{NodeID: "unknown", ChildCount: nil}})

	unknown, ok := m.ItemByID("unknown")
	if !ok {
		t.Fatal("expected unknown node")
	}
	if unknown.ChildCount() != 1 {
		t.Fatalf("expected placeholder for unknown child count, got %d children", unknown.ChildCount())
	}
	if !unknown.ChildAt(0).(*uiaTreeNode).placeholder {
		t.Fatal("expected placeholder child for unknown child count")
	}
}

func TestUIATreeZeroChildExpansion_RemovesPlaceholder(t *testing.T) {
	m := newUIATreeModel()
	m.SetRoot(inspect.TreeNodeDTO{NodeID: "root"})
	m.SetChildren("root", []inspect.TreeNodeDTO{{NodeID: "candidate"}})

	candidate, ok := m.ItemByID("candidate")
	if !ok || candidate.ChildCount() != 1 || !candidate.ChildAt(0).(*uiaTreeNode).placeholder {
		t.Fatalf("expected pre-expand placeholder, got ok=%v count=%d", ok, candidate.ChildCount())
	}

	m.SetChildren("candidate", nil)
	candidate, _ = m.ItemByID("candidate")
	if candidate.ChildCount() != 0 {
		t.Fatalf("expected placeholder removal after zero-child expansion, got %d children", candidate.ChildCount())
	}
}

func TestUIATreeRecursivePopulation_UnknownCountsUseSamePlaceholderRules(t *testing.T) {
	m := newUIATreeModel()
	m.SetRoot(inspect.TreeNodeDTO{NodeID: "root"})
	m.SetChildren("root", []inspect.TreeNodeDTO{{NodeID: "parent"}})
	m.SetChildren("parent", []inspect.TreeNodeDTO{{NodeID: "child", ChildCount: nil}})

	child, ok := m.ItemByID("child")
	if !ok {
		t.Fatal("expected recursive child")
	}
	if child.ChildCount() != 1 || !child.ChildAt(0).(*uiaTreeNode).placeholder {
		t.Fatalf("expected recursive unknown-count placeholder, got %d children", child.ChildCount())
	}

	m.SetChildren("child", []inspect.TreeNodeDTO{})
	child, _ = m.ItemByID("child")
	if child.ChildCount() != 0 {
		t.Fatalf("expected recursive zero-child expansion to clear placeholder, got %d", child.ChildCount())
	}
}

func TestUIATreeFakeNotepadTree_AHKStyleLabelsAndEmptyNameRetention(t *testing.T) {
	m := newUIATreeModel()
	fixture := buildFakeNotepadTreeFixture()
	m.SetRoot(fixture.Root)
	m.SetChildren("root", fixture.ChildrenBy["root"])
	m.SetChildren("pane", fixture.ChildrenBy["pane"])

	pane, ok := m.ItemByID("pane")
	if !ok {
		t.Fatal("expected pane node")
	}
	if got := pane.Text(); got != `pane ""` {
		t.Fatalf("expected AHK-style empty-name pane label, got %q", got)
	}
	for id, want := range map[string]string{"root": `window ""`, "menu": `menu bar ""`, "text": `text ""`, "dup-a": `button "Save"`, "dup-b": `button "Save"`} {
		n, ok := m.ItemByID(id)
		if !ok || n.Text() != want {
			t.Fatalf("node %s label = %q ok=%v want=%q", id, n.Text(), ok, want)
		}
	}
}

func TestUIATreeFakeNotepadTree_ConfiguredAutoExpandMarksExpandedNodes(t *testing.T) {
	m := newUIATreeModel()
	fixture := buildFakeNotepadTreeFixture()
	m.SetRoot(fixture.Root)
	m.SetChildren("root", fixture.ChildrenBy["root"])
	m.SetChildren("pane", fixture.ChildrenBy["pane"])

	for _, id := range []string{"root", "pane"} {
		m.SetExpanded(id, true)
	}
	if !m.IsExpanded("root") || !m.IsExpanded("pane") {
		t.Fatalf("expected configured auto-expand state for fake tree")
	}
	if m.IsExpanded("text") {
		t.Fatalf("leaf nodes should not be auto-expanded by default")
	}
}

func TestUIATreeSetChildren_NoPlaceholderForKnownLeaf(t *testing.T) {
	m := newUIATreeModel()
	m.SetRoot(inspect.TreeNodeDTO{NodeID: "root"})
	zero := 0
	m.SetChildren("root", []inspect.TreeNodeDTO{{NodeID: "leaf", HasChildren: true, ChildCount: &zero}})

	leaf, ok := m.ItemByID("leaf")
	if !ok {
		t.Fatal("expected leaf node")
	}
	if leaf.ChildCount() != 0 {
		t.Fatalf("expected known leaf to have no placeholder child, got %d", leaf.ChildCount())
	}
}

func TestUIATreeSetChildren_PlaceholderOnlyForPotentialChildren(t *testing.T) {
	m := newUIATreeModel()
	m.SetRoot(inspect.TreeNodeDTO{NodeID: "root"})
	one := 1
	m.SetChildren("root", []inspect.TreeNodeDTO{
		{NodeID: "unknown"},                        // no child count info
		{NodeID: "known-parent", ChildCount: &one}, // explicit children available
		{NodeID: "known-leaf", HasChildren: false}, // explicit leaf by flag
	})

	unknown, _ := m.ItemByID("unknown")
	if unknown.ChildCount() != 1 || !unknown.ChildAt(0).(*uiaTreeNode).placeholder {
		t.Fatalf("unknown node should keep lazy placeholder, got count=%d", unknown.ChildCount())
	}

	parent, _ := m.ItemByID("known-parent")
	if parent.ChildCount() != 1 || !parent.ChildAt(0).(*uiaTreeNode).placeholder {
		t.Fatalf("known parent should have placeholder child, got count=%d", parent.ChildCount())
	}

	leaf, _ := m.ItemByID("known-leaf")
	if leaf.ChildCount() != 0 {
		t.Fatalf("known leaf should not have placeholder, got %d", leaf.ChildCount())
	}
}

func TestUIATreeSetChildren_GrandchildrenRemainVisibleAfterParentExpansion(t *testing.T) {
	m := newUIATreeModel()
	m.SetRoot(inspect.TreeNodeDTO{NodeID: "root"})
	m.SetChildren("root", []inspect.TreeNodeDTO{{NodeID: "p"}})
	m.SetExpanded("p", true)
	m.SetChildren("p", []inspect.TreeNodeDTO{{NodeID: "c"}})
	m.SetExpanded("c", true)
	m.SetChildren("c", []inspect.TreeNodeDTO{{NodeID: "g"}})
	c, _ := m.ItemByID("c")
	if c.ChildCount() != 1 || c.ChildAt(0).(*uiaTreeNode).NodeID != "g" {
		t.Fatalf("expected expanded node to show deeper child, got count=%d", c.ChildCount())
	}
}

func TestUIATreeAutoExpandLevelOrder_RootThenChildLevels(t *testing.T) {
	m := newUIATreeModel()
	m.SetRoot(inspect.TreeNodeDTO{NodeID: "root"})
	m.SetChildren("root", []inspect.TreeNodeDTO{{NodeID: "child-1"}})
	m.SetExpanded("root", true)
	m.SetChildren("child-1", []inspect.TreeNodeDTO{{NodeID: "grandchild-1"}})
	m.SetExpanded("child-1", true)

	root, _ := m.ItemByID("root")
	if root.ChildCount() != 1 {
		t.Fatalf("expected root child level to render first, got %d", root.ChildCount())
	}
	child, _ := m.ItemByID("child-1")
	if child.ChildCount() != 1 || child.ChildAt(0).(*uiaTreeNode).NodeID != "grandchild-1" {
		t.Fatalf("expected child expansion to render next level, got count=%d", child.ChildCount())
	}
}

func TestUIATreeBulkExpansionKeepsModelAndVisualBookkeepingConsistent(t *testing.T) {
	m := newUIATreeModel()
	m.SetRoot(inspect.TreeNodeDTO{NodeID: "root"})
	m.SetChildren("root", []inspect.TreeNodeDTO{{NodeID: "a"}, {NodeID: "b"}})
	m.SetChildren("a", []inspect.TreeNodeDTO{{NodeID: "a1"}})
	m.SetChildren("b", []inspect.TreeNodeDTO{{NodeID: "b1"}})

	for _, id := range []string{"root", "a", "b"} {
		m.SetExpanded(id, true)
	}

	for _, id := range []string{"root", "a", "b"} {
		if !m.IsExpanded(id) {
			t.Fatalf("expected expanded model state for %s", id)
		}
		if !m.AreChildrenLoaded(id) {
			t.Fatalf("expected loaded child bookkeeping for %s", id)
		}
	}

	a, _ := m.ItemByID("a")
	b, _ := m.ItemByID("b")
	if a.ChildCount() != 1 || b.ChildCount() != 1 {
		t.Fatalf("expected visual expansion bookkeeping to keep children visible: a=%d b=%d", a.ChildCount(), b.ChildCount())
	}
}

func TestUIATreeSetChildren_ReusesPointersAcrossSequentialUpdates(t *testing.T) {
	m := newUIATreeModel()
	m.SetRoot(inspect.TreeNodeDTO{NodeID: "root"})
	m.SetChildren("root", []inspect.TreeNodeDTO{{NodeID: "a", Name: "first"}})

	first, _ := m.ItemByID("a")
	m.SetChildren("root", []inspect.TreeNodeDTO{{NodeID: "a", Name: "second"}})
	second, _ := m.ItemByID("a")

	if first != second {
		t.Fatal("expected SetChildren to reuse existing node pointer")
	}
	if second.Name != "second" {
		t.Fatalf("expected in-place field refresh, got %q", second.Name)
	}
}

func TestUIATreeNonFilteredMode_SharesAllNodePointers(t *testing.T) {
	m := newUIATreeModel()
	m.SetRoot(inspect.TreeNodeDTO{NodeID: "root"})
	if m.root != m.allRoot {
		t.Fatal("expected root and allRoot to alias in non-filtered mode")
	}

	m.SetChildren("root", []inspect.TreeNodeDTO{{NodeID: "a"}})
	if m.nodes[NodeID("a")] != m.allNodes[NodeID("a")] {
		t.Fatal("expected visible and all node maps to share child pointers")
	}
}
