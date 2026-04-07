package uia

import (
	"context"
	"errors"
	"strings"
)

var ErrElementNotFound = errors.New("uia element not found")

func Find(ctx context.Context, nav Navigator, rootID string, sel Selector) (Element, int, error) {
	if err := sel.Validate(); err != nil {
		return Element{}, 0, err
	}
	root, err := buildTreeNode(ctx, nav, rootID, 0, -1, map[string]bool{})
	if err != nil {
		return Element{}, 0, err
	}
	var visited int
	match := findNode(root, nil, sel, &visited)
	if match == nil {
		return Element{}, visited, ErrElementNotFound
	}
	return match.Element, visited, nil
}

func findNode(node *Node, parents []Element, sel Selector, visited *int) *Node {
	*visited = *visited + 1
	if matchesSelector(node.Element, sel) && matchesAncestors(parents, sel.Ancestors) {
		return node
	}
	nextParents := append(parents, node.Element)
	for _, child := range node.Children {
		if got := findNode(child, nextParents, sel, visited); got != nil {
			return got
		}
	}
	return nil
}

func matchesSelector(el Element, sel Selector) bool {
	if sel.AutomationID != "" && !eqStringPtr(el.AutomationID, sel.AutomationID) {
		return false
	}
	if sel.Name != "" && !eqStringPtr(el.Name, sel.Name) {
		return false
	}
	if sel.ControlType != "" && !eqStringPtr(el.ControlType, sel.ControlType) {
		return false
	}
	return true
}

func matchesAncestors(parents []Element, ancestors []Selector) bool {
	if len(ancestors) == 0 {
		return true
	}
	if len(parents) == 0 {
		return false
	}
	idx := 0
	for _, p := range parents {
		if matchesSelector(p, ancestors[idx]) {
			idx++
			if idx == len(ancestors) {
				return true
			}
		}
	}
	return false
}

func eqStringPtr(ptr *string, expect string) bool {
	return ptr != nil && strings.EqualFold(strings.TrimSpace(*ptr), strings.TrimSpace(expect))
}
