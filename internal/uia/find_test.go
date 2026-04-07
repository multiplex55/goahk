package uia

import (
	"context"
	"errors"
	"testing"
)

func TestFind_MatchesByFields(t *testing.T) {
	button := "Button"
	nav := mockNavigator{
		elements: map[string]Element{
			"root": {ID: "root"},
			"a":    {ID: "a", AutomationID: strPtr("cancel"), ControlType: &button},
			"b":    {ID: "b", AutomationID: strPtr("submit"), ControlType: &button},
		},
		children: map[string][]string{"root": {"a", "b"}},
	}
	got, visited, err := Find(context.Background(), nav, "root", Selector{AutomationID: "submit", ControlType: "button"})
	if err != nil {
		t.Fatalf("Find() error = %v", err)
	}
	if got.ID != "b" || visited < 2 {
		t.Fatalf("got=%+v visited=%d", got, visited)
	}
}

func TestFind_AncestorConstraints(t *testing.T) {
	window, pane, edit := "Window", "Pane", "Edit"
	nav := mockNavigator{
		elements: map[string]Element{
			"root":   {ID: "root", ControlType: &window, Name: strPtr("Main")},
			"panel":  {ID: "panel", ControlType: &pane, Name: strPtr("Form")},
			"target": {ID: "target", ControlType: &edit, AutomationID: strPtr("username")},
		},
		children: map[string][]string{"root": {"panel"}, "panel": {"target"}},
	}
	_, _, err := Find(context.Background(), nav, "root", Selector{
		AutomationID: "username",
		Ancestors: []Selector{
			{Name: "Main", ControlType: "Window"},
			{Name: "Form"},
		},
	})
	if err != nil {
		t.Fatalf("Find() error = %v", err)
	}
}

func TestFind_NotFound(t *testing.T) {
	nav := mockNavigator{
		elements: map[string]Element{"root": {ID: "root"}},
	}
	_, _, err := Find(context.Background(), nav, "root", Selector{Name: "x"})
	if !errors.Is(err, ErrElementNotFound) {
		t.Fatalf("error=%v", err)
	}
}

func strPtr(v string) *string { return &v }
