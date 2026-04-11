package inspect

import (
	"encoding/json"
	"testing"
)

func TestCanonicalDTO_JSONFieldNames(t *testing.T) {
	t.Parallel()

	childCount := 3
	payload := struct {
		Refresh RefreshWindowsRequest  `json:"refresh"`
		Window  WindowSummary          `json:"window"`
		Tree    TreeNodeDTO            `json:"tree"`
		Details GetNodeDetailsResponse `json:"details"`
	}{
		Refresh: RefreshWindowsRequest{Filter: "note", VisibleOnly: true, TitleOnly: true},
		Window:  WindowSummary{HWND: "0x1", Title: "Notepad", ProcessName: "notepad.exe", ClassName: "Notepad", ProcessID: 100},
		Tree: TreeNodeDTO{
			NodeID: "node-1", Name: "Root", ControlType: "Window", ClassName: "Notepad", HasChildren: true, ParentNodeID: "",
			Patterns: []string{"Invoke"}, ChildCount: &childCount,
		},
		Details: GetNodeDetailsResponse{
			WindowInfo: WindowInfoDTO{HWND: "0x1", Title: "Notepad", Text: "Notepad", Class: "Notepad", Process: "notepad.exe", PID: 100},
			Element:    ElementPropertiesDTO{Name: "Root", ControlType: "Window", IsEnabled: true},
			Properties: []PropertyDTO{{Name: "name", Value: "Root"}},
			Patterns:   []PatternActionDTO{{Name: "invoke", Pattern: "Invoke", DisplayName: "Invoke"}},
			StatusText: "ok", BestSelector: "#main",
			Path: []TreeNodeDTO{{NodeID: "node-root", Name: "Root", HasChildren: true}},
			SelectorPath: SelectorPathDTO{
				BestSelector:        &Selector{AutomationID: "main"},
				FullPath:            []TreeNodeDTO{{NodeID: "node-root", Name: "Root", HasChildren: true}},
				SelectorSuggestions: []SelectorCandidate{{Rank: 1, Selector: Selector{AutomationID: "main"}}},
			},
		},
	}

	b, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	refresh := decoded["refresh"].(map[string]any)
	if _, ok := refresh["visibleOnly"]; !ok {
		t.Fatalf("missing visibleOnly in refresh payload: %s", string(b))
	}
	window := decoded["window"].(map[string]any)
	if _, ok := window["processID"]; !ok {
		t.Fatalf("missing processID in window payload: %s", string(b))
	}
	tree := decoded["tree"].(map[string]any)
	if _, ok := tree["nodeID"]; !ok {
		t.Fatalf("missing nodeID in tree payload: %s", string(b))
	}
	details := decoded["details"].(map[string]any)
	if _, ok := details["windowInfo"]; !ok {
		t.Fatalf("missing windowInfo in details payload: %s", string(b))
	}
	if _, ok := details["element"]; !ok {
		t.Fatalf("missing element in details payload: %s", string(b))
	}
	if _, ok := details["selectorPath"]; !ok {
		t.Fatalf("missing selectorPath in details payload: %s", string(b))
	}
}

func TestCanonicalDTO_AliasCompatibility(t *testing.T) {
	t.Parallel()

	var _ WindowListRequestDTO = RefreshWindowsRequest{}
	var _ WindowListItemDTO = WindowSummary{}
	var _ TreeNodeCanonicalDTO = TreeNodeDTO{}
	var _ NodeDetailsRequestDTO = GetNodeDetailsRequest{}
	var _ NodeDetailsDTO = GetNodeDetailsResponse{}
}
