//go:build windows

package main

import (
	"testing"

	"goahk/internal/inspect"
)

func TestMapPatternTreeGroupsAndLabels(t *testing.T) {
	nodes := mapPatternTree([]inspect.PatternActionDTO{{Pattern: "InvokePattern", Name: "invoke"}, {Pattern: "LegacyIAccessiblePattern", Name: "do_default_action"}, {Pattern: "SelectionItemPattern", Name: "select"}, {Pattern: "ValuePattern", Name: "set_value"}})
	if len(nodes) != 4 {
		t.Fatalf("unexpected group count: %d", len(nodes))
	}
	if nodes[0].Children[0].Label != "Invoke()" || nodes[0].Children[0].ActionID != "invoke" {
		t.Fatalf("unexpected invoke child: %#v", nodes[0].Children[0])
	}
}
