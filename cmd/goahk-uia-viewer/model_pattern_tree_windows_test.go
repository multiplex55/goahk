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
	if nodes[0].Children[0].Label != "Invoke()" || nodes[0].Children[0].ActionID != "invoke" {
		t.Fatalf("unexpected invoke child: %#v", nodes[0].Children[0])
	}
	if nodes[1].Children[0].ActionID != "doDefaultAction" || nodes[1].Children[0].Label != "DoDefaultAction()" {
		t.Fatalf("unexpected doDefaultAction child: %#v", nodes[1].Children[0])
	}
	if nodes[3].Children[0].ActionID != "setValue" || nodes[3].Children[0].Label != "SetValue()" {
		t.Fatalf("unexpected setValue child: %#v", nodes[3].Children[0])
	}
}

func TestMapPatternTree_SupportsCanonicalAndLegacyAliases(t *testing.T) {
	nodes := mapPatternTree([]inspect.PatternActionDTO{{Pattern: "InvokePattern", Name: "invoke"}, {Pattern: "InvokePattern", Name: "doDefaultAction"}, {Pattern: "ValuePattern", Name: "set_value"}, {Pattern: "ExpandCollapsePattern", Name: "expand"}, {Pattern: "ExpandCollapsePattern", Name: "collapse"}, {Pattern: "TogglePattern", Name: "toggle"}})
	if got := nodes[0].Children[1].Label; got != "DoDefaultAction()" {
		t.Fatalf("label=%q", got)
	}
	if got := nodes[1].Children[0].ActionID; got != "setValue" {
		t.Fatalf("action id=%q", got)
	}
	if got := nodes[2].Children[0].Label; got != "Expand()" {
		t.Fatalf("expand label=%q", got)
	}
	if got := nodes[2].Children[1].Label; got != "Collapse()" {
		t.Fatalf("collapse label=%q", got)
	}
	if got := nodes[3].Children[0].Label; got != "Toggle()" {
		t.Fatalf("toggle label=%q", got)
	}
}

func TestController_InvokeSetValue_DialogFlow(t *testing.T) {
	svc := &fakeInspectService{nodeDetailsResp: inspect.GetNodeDetailsResponse{Element: inspect.ElementPropertiesDTO{NodeID: "node-1"}}}
	c := NewController(context.Background(), svc)
	c.selectedNodeID = "node-1"

	c.SetDialogs(&fakeDialogs{ok: false})
	if _, err := c.InvokeSetValue(); err != nil {
		t.Fatalf("cancel should be noop: %v", err)
	}
	if len(svc.invokeReqs) != 0 {
		t.Fatalf("expected no invoke on cancel")
	}

	c.SetDialogs(&fakeDialogs{ok: true, value: "hello"})
	if _, err := c.InvokeSetValue(); err != nil {
		t.Fatalf("confirm failed: %v", err)
	}
	if len(svc.invokeReqs) != 1 || svc.invokeReqs[0].Action != "setValue" || svc.invokeReqs[0].Payload["value"] != "hello" {
		t.Fatalf("unexpected invoke req: %+v", svc.invokeReqs)
	}
}
