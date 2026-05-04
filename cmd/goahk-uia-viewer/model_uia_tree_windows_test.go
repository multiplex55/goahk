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
	if got := m.Label(inspect.TreeNodeDTO{NodeID: "n3", ControlType: "Button", Name: "OK"}); got != "Button \"OK\"" {
		t.Fatalf("got %q", got)
	}
	if got := m.Label(inspect.TreeNodeDTO{NodeID: "n4"}); got != "n4" {
		t.Fatalf("got %q", got)
	}
}

func TestUIATreeLoadedAndExpandedState(t *testing.T) {
	m := newUIATreeModel()
	if m.AreChildrenLoaded("n1") || m.IsExpanded("n1") {
		t.Fatalf("unexpected initial state")
	}
	m.MarkChildrenLoaded("n1")
	m.SetExpanded("n1", true)
	if !m.AreChildrenLoaded("n1") || !m.IsExpanded("n1") {
		t.Fatalf("expected true state")
	}
	m.SetExpanded("n1", false)
	if !m.AreChildrenLoaded("n1") || m.IsExpanded("n1") {
		t.Fatalf("expected loaded true expanded false")
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
	m.SetChildren("root", []inspect.TreeNodeDTO{{NodeID: "c1"}, {NodeID: "c2"}})
	if m.ShouldShowLazyPlaceholder("root") {
		t.Fatal("should not show placeholder once children loaded")
	}
	ids := m.ChildrenOf("root")
	if len(ids) != 2 || ids[0] != "c1" || ids[1] != "c2" {
		t.Fatalf("unexpected children: %#v", ids)
	}
	m.Reset()
	if m.RootID() != "" {
		t.Fatal("expected cleared root")
	}
}
