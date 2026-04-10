package uia

import (
	"context"
	"reflect"
	"testing"
)

type mockNavigator struct {
	elements map[string]Element
	children map[string][]string
}

func (m mockNavigator) ElementByID(_ context.Context, id string) (Element, error) {
	return m.elements[id], nil
}

func (m mockNavigator) ChildrenIDs(_ context.Context, id string) ([]string, error) {
	return append([]string(nil), m.children[id]...), nil
}

func TestBuildTree_DepthLimit(t *testing.T) {
	nav := mockNavigator{
		elements: map[string]Element{"root": {ID: "root"}, "a": {ID: "a"}, "b": {ID: "b"}},
		children: map[string][]string{"root": {"a"}, "a": {"b"}},
	}
	node, err := BuildTree(context.Background(), nav, "root", TreeOptions{MaxDepth: 1})
	if err != nil {
		t.Fatalf("BuildTree() error = %v", err)
	}
	if len(node.Children) != 1 || len(node.Children[0].Children) != 0 {
		t.Fatalf("unexpected depth limiting result: %#v", node)
	}
}

func TestBuildTree_CycleProtection(t *testing.T) {
	nav := mockNavigator{
		elements: map[string]Element{"root": {ID: "root"}, "child": {ID: "child"}},
		children: map[string][]string{"root": {"child"}, "child": {"root"}},
	}
	node, err := BuildTree(context.Background(), nav, "root", TreeOptions{MaxDepth: 5})
	if err != nil {
		t.Fatalf("BuildTree() error = %v", err)
	}
	if len(node.Children) != 1 || len(node.Children[0].Children) != 1 || !node.Children[0].Children[0].Cycle {
		t.Fatalf("expected cycle marker, got %#v", node)
	}
}

func TestBuildTree_DeterministicTraversalOrder(t *testing.T) {
	nav := mockNavigator{
		elements: map[string]Element{"root": {ID: "root"}, "c": {ID: "c"}, "a": {ID: "a"}, "b": {ID: "b"}},
		children: map[string][]string{"root": {"c", "a", "b"}},
	}
	node, err := BuildTree(context.Background(), nav, "root", TreeOptions{MaxDepth: 1})
	if err != nil {
		t.Fatalf("BuildTree() error = %v", err)
	}
	ids := []string{node.Children[0].Element.ID, node.Children[1].Element.ID, node.Children[2].Element.ID}
	if !reflect.DeepEqual(ids, []string{"a", "b", "c"}) {
		t.Fatalf("child order = %v, want [a b c]", ids)
	}
}

func TestBuildTree_FindIntegrationScenario(t *testing.T) {
	window, pane, edit := "Window", "Pane", "Edit"
	nav := mockNavigator{
		elements: map[string]Element{
			"root":   {ID: "root", Name: strPtr("App"), ControlType: &window},
			"panel":  {ID: "panel", Name: strPtr("Login"), ControlType: &pane},
			"target": {ID: "target", AutomationID: strPtr("username"), ControlType: &edit},
		},
		children: map[string][]string{"root": {"panel"}, "panel": {"target"}},
	}
	built, err := BuildTree(context.Background(), nav, "root", TreeOptions{MaxDepth: 3})
	if err != nil {
		t.Fatalf("BuildTree() error = %v", err)
	}
	if built.Element.ID != "root" || len(built.Children) != 1 {
		t.Fatalf("unexpected tree: %#v", built)
	}
	found, _, err := Find(context.Background(), nav, "root", Selector{AutomationID: "username", Ancestors: []Selector{{Name: "Login"}}})
	if err != nil || found.ID != "target" {
		t.Fatalf("Find() got=%+v err=%v", found, err)
	}
}
