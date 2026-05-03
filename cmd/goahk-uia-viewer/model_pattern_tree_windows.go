//go:build windows

package main

import "goahk/internal/inspect"

type patternTreeNode struct {
	Label    string
	ActionID string
	Children []patternTreeNode
}

func mapPatternTree(actions []inspect.PatternActionDTO) []patternTreeNode {
	groups := map[string][]patternTreeNode{}
	order := []string{}
	for _, a := range actions {
		pat := a.Pattern
		if pat == "" {
			pat = "UnknownPattern"
		}
		if _, ok := groups[pat]; !ok {
			order = append(order, pat)
		}
		groups[pat] = append(groups[pat], patternTreeNode{Label: callableActionLabel(a.Name), ActionID: a.Name})
	}
	out := make([]patternTreeNode, 0, len(order))
	for _, pat := range order {
		out = append(out, patternTreeNode{Label: pat, Children: groups[pat]})
	}
	return out
}

func callableActionLabel(name string) string {
	switch name {
	case "invoke":
		return "Invoke()"
	case "do_default_action":
		return "DoDefaultAction()"
	case "select":
		return "Select()"
	case "set_value":
		return "SetValue()"
	default:
		return name + "()"
	}
}
