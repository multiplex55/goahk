package uia

import "testing"

func TestBuildInspectResult_SortsPatternsAndSetsBestSelector(t *testing.T) {
	name := "Checkout"
	controlType := "Button"
	automationID := "submitBtn"
	res := BuildInspectResult(Element{
		Name:         &name,
		ControlType:  &controlType,
		AutomationID: &automationID,
		Patterns:     []string{"Value", "Invoke"},
	})

	if len(res.SupportedPatterns) != 2 || res.SupportedPatterns[0] != "Invoke" || res.SupportedPatterns[1] != "Value" {
		t.Fatalf("SupportedPatterns = %v", res.SupportedPatterns)
	}
	if res.BestSelector == nil || res.BestSelector.AutomationID != "submitBtn" {
		t.Fatalf("BestSelector = %#v", res.BestSelector)
	}
	if len(res.SelectorSuggestions) < 3 {
		t.Fatalf("SelectorSuggestions too short: %v", res.SelectorSuggestions)
	}
	if res.SelectorSuggestions[0].Rank != 1 || res.SelectorSuggestions[0].Selector.AutomationID == "" {
		t.Fatalf("ranked suggestion[0] = %#v", res.SelectorSuggestions[0])
	}
}

func TestSuggestSelectors_NoDataReturnsEmpty(t *testing.T) {
	if got := SuggestSelectors(Element{}); len(got) != 0 {
		t.Fatalf("SuggestSelectors() = %v, want empty", got)
	}
}
