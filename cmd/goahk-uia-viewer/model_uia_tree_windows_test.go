//go:build windows

package main

import (
	"testing"

	"goahk/internal/inspect"
)

func TestUIATreeLabelPrecedence(t *testing.T) {
	m := newUIATreeModel()
	if got := m.Label(inspect.TreeNodeDTO{NodeID: "n1", DisplayLabel: "Display"}); got != "Display" {
		t.Fatalf("got %q", got)
	}
	if got := m.Label(inspect.TreeNodeDTO{NodeID: "n2", LocalizedControlType: "button", Name: "OK"}); got != "button \"OK\"" {
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
	m.Reset()
	if m.RootID() != "" || m.NodeCount() != 0 {
		t.Fatal("expected reset state")
	}
}
