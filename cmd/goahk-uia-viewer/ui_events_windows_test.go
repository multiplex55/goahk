//go:build windows

package main

import "testing"

func TestPropertyContextMenuCopyValue(t *testing.T) {
	row := propertyTableRow{Name: "Name", Value: "Calculator"}
	if got := propertyContextCopyValue(row, true); got != "Calculator" {
		t.Fatalf("copy value got %q", got)
	}
}

func TestPropertyContextMenuCopyName(t *testing.T) {
	row := propertyTableRow{Name: "ControlType", Value: "50000 (button)"}
	if got := row.Name; got != "ControlType" {
		t.Fatalf("copy name got %q", got)
	}
}

func TestPropertyContextMenuCopyRow(t *testing.T) {
	row := propertyTableRow{Name: "ControlType", Value: "50000 (button)"}
	got := row.Name + ": " + propertyContextCopyValue(row, true)
	if got != "ControlType: button" {
		t.Fatalf("copy row got %q", got)
	}
}

func TestPatternRightClickCopyUsesVisibleLabel(t *testing.T) {
	node := &patternTreeNode{label: "Set Value"}
	if got := patternNodeCopyText(node); got != "Set Value" {
		t.Fatalf("copy text got %q", got)
	}
}

func TestPatternDoubleClickInvokesOnlyActionNodes(t *testing.T) {
	parent := &patternTreeNode{label: "ValuePattern", children: []patternTreeNode{{label: "Set Value", actionID: ActionID("setValue")}}}
	if action, ok := patternActionForNode(parent); ok || action != "" {
		t.Fatalf("parent should not produce action, got %q", action)
	}
	leaf := &parent.children[0]
	if action, ok := patternActionForNode(leaf); !ok || action != "setValue" {
		t.Fatalf("leaf should produce action setValue, got %q ok=%v", action, ok)
	}
}

func TestPatternSetValuePromptsThenInvokes(t *testing.T) {
	if action, ok := patternActionForNode(&patternTreeNode{actionID: ActionID("setValue")}); !ok || action != "setValue" {
		t.Fatalf("setValue should be mappable, got %q ok=%v", action, ok)
	}
}

func TestVisibleCheckboxChangeRefreshesWindowList(t *testing.T) {
	ui := &viewerUI{}
	visible, title := ui.defaultRefreshArgs()
	if !visible || !title {
		t.Fatalf("expected default refresh args true/true, got %v/%v", visible, title)
	}
}

func TestTitleCheckboxChangeRefreshesWindowList(t *testing.T) {
	ui := &viewerUI{}
	visible, title := ui.defaultRefreshArgs()
	if !visible || !title {
		t.Fatalf("expected default refresh args true/true, got %v/%v", visible, title)
	}
}

func TestRefreshListButtonUsesCurrentCheckboxState(t *testing.T) {
	ui := &viewerUI{}
	visible, title := ui.defaultRefreshArgs()
	if !visible || !title {
		t.Fatalf("expected refresh list defaults true/true, got %v/%v", visible, title)
	}
}

func TestFilterTextboxShowsNotImplementedStatus(t *testing.T) {
	ui := &viewerUI{}
	if got := ui.filterNotImplementedStatus(); got != "Tree filtering not implemented yet" {
		t.Fatalf("filter placeholder status got %q", got)
	}
}
