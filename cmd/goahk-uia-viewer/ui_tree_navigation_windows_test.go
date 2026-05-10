//go:build windows

package main

import (
	"testing"

	"goahk/internal/inspect"
)

func TestVisibleTreeItemsRespectsExpansion(t *testing.T) {
	m := newUIATreeModel()
	m.SetRoot(inspect.TreeNodeDTO{NodeID: "root"})
	m.SetChildren("root", []inspect.TreeNodeDTO{{NodeID: "child"}})
	ui := &viewerUI{treeModel: m}

	items := ui.visibleTreeItems()
	if len(items) != 1 || items[0].NodeID != "root" {
		t.Fatalf("expected collapsed traversal to return root only, got %#v", items)
	}
}

func TestNavigateTreeVisibleSkipsPlaceholders(t *testing.T) {
	ui := &viewerUI{}
	out := make([]*uiaTreeNode, 0)
	ui.collectVisibleTreeItems(&uiaTreeNode{TreeNodeDTO: inspect.TreeNodeDTO{NodeID: "loading"}, placeholder: true}, &out)
	if len(out) != 0 {
		t.Fatalf("expected placeholder node to be skipped, got %d", len(out))
	}
}

func TestNavigateTreeVisibleRespectsFilteredProjection(t *testing.T) {
	m := newUIATreeModel()
	m.SetRoot(inspect.TreeNodeDTO{NodeID: "root"})
	m.SetChildren("root", []inspect.TreeNodeDTO{{NodeID: "button", LocalizedControlType: "button"}, {NodeID: "text", LocalizedControlType: "text"}})
	m.ApplyFilter("button")
	ui := &viewerUI{treeModel: m}

	items := ui.visibleTreeItems()
	if len(items) != 1 || items[0].NodeID != "root" {
		t.Fatalf("expected filtered collapsed traversal to return projected root only, got %#v", items)
	}
	if m.IsVisibleNodeID("text") {
		t.Fatal("expected non-matching node hidden by filtered projection")
	}
}
