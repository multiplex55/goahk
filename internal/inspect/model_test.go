package inspect

import (
	"encoding/json"
	"testing"
)

func TestInspectModels_JSONRoundTrip(t *testing.T) {
	t.Parallel()

	helpText := "Press to submit"
	value := "Submit"
	status := "Ready"
	itemStatus := "Hot"
	labeledBy := "node-label-1"

	model := InspectWindow{
		HWND:           "0x1020",
		Title:          "Demo App",
		ClassName:      "DemoWindow",
		Rect:           Rect{Left: 10, Top: 20, Width: 1280, Height: 720},
		ProcessID:      1234,
		ProcessName:    "demo.exe",
		ExecutablePath: `C:\\Demo\\demo.exe`,
		IsVisible:      true,
		IsActive:       true,
		IsMinimized:    false,
		RootNodeID:     "node-root",
		RootElement: &TreeNodeSummary{
			NodeID:               "node-root",
			Name:                 "Root",
			ControlType:          "Window",
			LocalizedControlType: "window",
			AutomationID:         "MainWindow",
			ClassName:            "DemoWindow",
			FrameworkID:          "Win32",
			BoundingRect:         &Rect{Left: 10, Top: 20, Width: 1280, Height: 720},
			IsOffscreen:          false,
			ChildCount:           1,
			HasChildren:          true,
			Patterns:             []PatternAction{{Pattern: "Window", Action: "setFocus", DisplayName: "Set Focus"}},
			BestSelector:         &Selector{AutomationID: "MainWindow", ControlType: "Window"},
			SelectorSuggestions: []SelectorCandidate{{
				Rank: 1, Selector: Selector{AutomationID: "MainWindow"}, Rationale: "Stable id", Score: 100, Source: "heuristic",
			}},
		},
	}

	element := InspectElement{
		NodeID:               "node-1",
		RuntimeID:            "42.1.9",
		ParentNodeID:         "node-root",
		Name:                 "Submit",
		LocalizedControlType: "button",
		ControlType:          "Button",
		AutomationID:         "SubmitButton",
		ClassName:            "Button",
		FrameworkID:          "WPF",
		ProcessID:            1234,
		HelpText:             &helpText,
		Status:               &status,
		Value:                &value,
		ItemStatus:           &itemStatus,
		LabeledBy:            &labeledBy,
		BoundingRect:         &Rect{Left: 100, Top: 200, Width: 90, Height: 30},
		IsEnabled:            true,
		IsKeyboardFocusable:  true,
		HasKeyboardFocus:     false,
		IsOffscreen:          false,
		IsContentElement:     true,
		IsControlElement:     true,
		IsPassword:           false,
		Patterns:             []PatternAction{{Pattern: "Invoke", Action: "invoke"}},
		BestSelector:         &Selector{AutomationID: "SubmitButton", ControlType: "Button"},
		SelectorSuggestions:  []SelectorCandidate{{Rank: 1, Selector: Selector{AutomationID: "SubmitButton"}}},
	}

	for _, tc := range []struct {
		name string
		in   any
		out  any
	}{
		{name: "inspect_window", in: model, out: &InspectWindow{}},
		{name: "inspect_element", in: element, out: &InspectElement{}},
		{name: "tree_node_summary", in: *model.RootElement, out: &TreeNodeSummary{}},
		{name: "pattern_action", in: PatternAction{Pattern: "Invoke", Action: "invoke"}, out: &PatternAction{}},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			b, err := json.Marshal(tc.in)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			if err := json.Unmarshal(b, tc.out); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			b2, err := json.Marshal(tc.out)
			if err != nil {
				t.Fatalf("marshal round-trip: %v", err)
			}
			if string(b) != string(b2) {
				t.Fatalf("round-trip mismatch\norig: %s\nout:  %s", string(b), string(b2))
			}
		})
	}
}

func TestInspectModels_ZeroValueAndOptionalFields(t *testing.T) {
	t.Parallel()

	b, err := json.Marshal(InspectElement{})
	if err != nil {
		t.Fatalf("marshal zero inspect element: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("unmarshal map: %v", err)
	}

	if _, exists := got["labeledBy"]; exists {
		t.Fatalf("expected labeledBy to be omitted when nil")
	}
	if _, exists := got["bestSelector"]; exists {
		t.Fatalf("expected bestSelector to be omitted when nil")
	}
	if _, exists := got["selectorSuggestions"]; exists {
		t.Fatalf("expected selectorSuggestions to be omitted when empty")
	}

	if got["isEnabled"] != false {
		t.Fatalf("expected isEnabled false by default, got %v", got["isEnabled"])
	}
}

func TestSelectorSuggestions_OrderAndShape(t *testing.T) {
	t.Parallel()

	node := TreeNodeSummary{
		NodeID: "node-1",
		SelectorSuggestions: []SelectorCandidate{
			{Rank: 1, Selector: Selector{AutomationID: "Primary"}, Rationale: "best"},
			{Rank: 2, Selector: Selector{Name: "Fallback", ControlType: "Button"}, Rationale: "fallback"},
		},
	}

	b, err := json.Marshal(node)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded TreeNodeSummary
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(decoded.SelectorSuggestions) != 2 {
		t.Fatalf("expected 2 suggestions, got %d", len(decoded.SelectorSuggestions))
	}
	if decoded.SelectorSuggestions[0].Rank != 1 || decoded.SelectorSuggestions[1].Rank != 2 {
		t.Fatalf("expected rank ordering to survive round-trip, got %+v", decoded.SelectorSuggestions)
	}
	if decoded.SelectorSuggestions[0].Selector.AutomationID == "" {
		t.Fatalf("expected first suggestion selector shape with automationId")
	}
	if decoded.SelectorSuggestions[1].Selector.Name == "" || decoded.SelectorSuggestions[1].Selector.ControlType == "" {
		t.Fatalf("expected second suggestion selector shape with name/controlType")
	}
}
