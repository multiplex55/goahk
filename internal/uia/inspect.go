package uia

import (
	"encoding/json"
	"fmt"
	"strings"
)

func ElementJSON(el Element) ([]byte, error) {
	return json.MarshalIndent(el, "", "  ")
}

func TreeJSON(node *Node) ([]byte, error) {
	return json.MarshalIndent(node, "", "  ")
}

func FormatElementText(el Element) string {
	b := &strings.Builder{}
	fmt.Fprintf(b, "ID: %s\n", emptyOrUnknown(el.ID))
	fmt.Fprintf(b, "Name: %s\n", ptrString(el.Name))
	fmt.Fprintf(b, "ControlType: %s\n", ptrString(el.ControlType))
	fmt.Fprintf(b, "ClassName: %s\n", ptrString(el.ClassName))
	fmt.Fprintf(b, "AutomationID: %s\n", ptrString(el.AutomationID))
	fmt.Fprintf(b, "FrameworkID: %s\n", ptrString(el.FrameworkID))
	if len(el.Patterns) == 0 {
		fmt.Fprint(b, "Patterns: (none)")
	} else {
		fmt.Fprintf(b, "Patterns: %s", strings.Join(el.Patterns, ", "))
	}
	return b.String()
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
