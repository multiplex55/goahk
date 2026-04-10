package uia

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

func ElementJSON(el Element) ([]byte, error) {
	return json.MarshalIndent(el, "", "  ")
}

func TreeJSON(node *Node) ([]byte, error) {
	return json.MarshalIndent(node, "", "  ")
}

type InspectResult struct {
	Element             Element              `json:"element"`
	SupportedPatterns   []string             `json:"supportedPatterns"`
	SelectorSuggestions []SelectorSuggestion `json:"selectorSuggestions,omitempty"`
	BestSelector        *Selector            `json:"bestSelector,omitempty"`
}

type SelectorSuggestion struct {
	Rank      int      `json:"rank"`
	Selector  Selector `json:"selector"`
	Rationale string   `json:"rationale"`
}

func FormatElementText(el Element) string {
	result := BuildInspectResult(el)
	b := &strings.Builder{}
	fmt.Fprintf(b, "ID: %s\n", emptyOrUnknown(result.Element.ID))
	fmt.Fprintf(b, "Name: %s\n", ptrString(result.Element.Name))
	fmt.Fprintf(b, "ControlType: %s\n", ptrString(result.Element.ControlType))
	fmt.Fprintf(b, "ClassName: %s\n", ptrString(result.Element.ClassName))
	fmt.Fprintf(b, "AutomationID: %s\n", ptrString(result.Element.AutomationID))
	fmt.Fprintf(b, "FrameworkID: %s\n", ptrString(result.Element.FrameworkID))
	if len(result.SupportedPatterns) == 0 {
		fmt.Fprint(b, "SupportedPatterns: (none)")
	} else {
		fmt.Fprintf(b, "SupportedPatterns: %s", strings.Join(result.SupportedPatterns, ", "))
	}
	if len(result.SelectorSuggestions) > 0 {
		fmt.Fprint(b, "\nSelectorSuggestions:")
		for _, suggestion := range result.SelectorSuggestions {
			raw, _ := json.Marshal(suggestion.Selector)
			fmt.Fprintf(b, "\n%d. %s | rationale: %s", suggestion.Rank, string(raw), suggestion.Rationale)
		}
	}
	return b.String()
}

func BuildInspectResult(el Element) InspectResult {
	patterns := append([]string(nil), el.Patterns...)
	sort.Strings(patterns)
	suggestions := SuggestSelectors(el)
	result := InspectResult{
		Element:             el,
		SupportedPatterns:   patterns,
		SelectorSuggestions: suggestions,
	}
	if len(suggestions) > 0 {
		result.BestSelector = &suggestions[0].Selector
	}
	return result
}

func SuggestSelectors(el Element) []SelectorSuggestion {
	type candidate struct {
		selector  Selector
		rationale string
		score     int
	}
	var candidates []candidate
	automationID := strings.TrimSpace(ptrValue(el.AutomationID))
	name := strings.TrimSpace(ptrValue(el.Name))
	controlType := strings.TrimSpace(ptrValue(el.ControlType))

	if automationID != "" {
		candidates = append(candidates, candidate{
			selector:  Selector{AutomationID: automationID},
			rationale: "AutomationID is usually stable across UI changes.",
			score:     100,
		})
	}
	if automationID != "" && controlType != "" {
		candidates = append(candidates, candidate{
			selector:  Selector{AutomationID: automationID, ControlType: controlType},
			rationale: "AutomationID + ControlType narrows matches when IDs repeat.",
			score:     90,
		})
	}
	if name != "" && controlType != "" {
		candidates = append(candidates, candidate{
			selector:  Selector{Name: name, ControlType: controlType},
			rationale: "Name + ControlType works when AutomationID is absent.",
			score:     70,
		})
	}
	if name != "" {
		candidates = append(candidates, candidate{
			selector:  Selector{Name: name},
			rationale: "Name-only selectors are readable but can be ambiguous.",
			score:     50,
		})
	}
	if controlType != "" {
		candidates = append(candidates, candidate{
			selector:  Selector{ControlType: controlType},
			rationale: "ControlType-only selectors are broad and should be a fallback.",
			score:     30,
		})
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})

	out := make([]SelectorSuggestion, 0, len(candidates))
	for i, c := range candidates {
		out = append(out, SelectorSuggestion{
			Rank:      i + 1,
			Selector:  c.selector,
			Rationale: c.rationale,
		})
	}
	return out
}

func ptrValue(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func FormatTreeText(node *Node) string {
	if node == nil {
		return "(empty tree)"
	}
	var lines []string
	walk(node, "", &lines)
	return strings.Join(lines, "\n")
}

func walk(node *Node, indent string, lines *[]string) {
	suffix := ""
	if node.Cycle {
		suffix = " [cycle]"
	}
	*lines = append(*lines, fmt.Sprintf("%s- %s%s", indent, node.Element.Summary(), suffix))
	for _, c := range node.Children {
		walk(c, indent+"  ", lines)
	}
}
