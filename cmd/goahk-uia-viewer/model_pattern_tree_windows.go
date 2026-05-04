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
		action := normalizePatternActionName(a.Name)
		pat := a.Pattern
		if pat == "" {
			pat = "UnknownPattern"
		}
		if _, ok := groups[pat]; !ok {
			order = append(order, pat)
		}
		groups[pat] = append(groups[pat], patternTreeNode{Label: callableActionLabel(action), ActionID: action})
	}
	out := make([]patternTreeNode, 0, len(order))
	for _, pat := range order {
		out = append(out, patternTreeNode{Label: pat, Children: groups[pat]})
	}
	return out
}

func normalizePatternActionName(name string) string {
	switch name {
	case "do_default_action":
		return "doDefaultAction"
	case "set_value":
		return "setValue"
	default:
		return name
	}
}

func callableActionLabel(name string) string {
	switch name {
	case "invoke":
		return "Invoke()"
	case "doDefaultAction":
		return "DoDefaultAction()"
	case "select":
		return "Select()"
	case "setValue":
		return "SetValue()"
	case "toggle":
		return "Toggle()"
	case "expand":
		return "Expand()"
	case "collapse":
		return "Collapse()"
	default:
		return name + "()"
	}
}
