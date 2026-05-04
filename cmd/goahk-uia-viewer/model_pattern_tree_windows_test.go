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
	if nodes[0].Label != "InvokePattern" {
		t.Fatalf("unexpected deterministic group order: %q", nodes[0].Label)
	}
}

func TestMapPatternTree_SupportsCanonicalAndLegacyAliases(t *testing.T) {
	nodes := mapPatternTree([]inspect.PatternActionDTO{{Pattern: "InvokePattern", Name: "invoke"}, {Pattern: "InvokePattern", Name: "doDefaultAction"}, {Pattern: "ValuePattern", Name: "set_value"}, {Pattern: "ExpandCollapsePattern", Name: "expand"}, {Pattern: "ExpandCollapsePattern", Name: "collapse"}, {Pattern: "TogglePattern", Name: "toggle"}})
	m := newPatternTreeModel()
	m.SetRoots(nodes)
	if got, ok := m.ActionForNode("InvokePattern/doDefaultAction"); !ok || got != "doDefaultAction" {
		t.Fatalf("action mismatch got=%q ok=%v", got, ok)
	}
	if got := callableActionLabel("set_value"); got != "SetValue()" {
		t.Fatalf("label=%q", got)
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
