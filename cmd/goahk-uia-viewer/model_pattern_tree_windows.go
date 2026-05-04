//go:build windows

package main

import (
	"sort"
	"strings"

	"goahk/internal/inspect"
)

type patternTreeNode struct {
	NodeID   string
	Label    string
	ActionID string
	Children []patternTreeNode
}

type patternTreeModel struct {
	nodes map[string]patternTreeNode
}

func newPatternTreeModel() *patternTreeModel {
	return &patternTreeModel{nodes: map[string]patternTreeNode{}}
}

func (m *patternTreeModel) SetRoots(roots []patternTreeNode) {
	m.nodes = map[string]patternTreeNode{}
	for _, root := range roots {
		m.indexNode(root)
	}
}

func (m *patternTreeModel) indexNode(n patternTreeNode) {
	m.nodes[n.NodeID] = n
	for _, child := range n.Children {
		m.indexNode(child)
	}
}

func (m *patternTreeModel) NodeByID(nodeID string) (patternTreeNode, bool) {
	n, ok := m.nodes[nodeID]
	return n, ok
}

func (m *patternTreeModel) ActionForNode(nodeID string) (string, bool) {
	n, ok := m.NodeByID(nodeID)
	if !ok || n.ActionID == "" {
		return "", false
	}
	return n.ActionID, true
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
		childID := pat + "/" + action
		groups[pat] = append(groups[pat], patternTreeNode{NodeID: childID, Label: callableActionLabel(action), ActionID: action})
	}
	sort.Strings(order)
	out := make([]patternTreeNode, 0, len(order))
	for _, pat := range order {
		children := groups[pat]
		sort.SliceStable(children, func(i, j int) bool {
			return children[i].ActionID < children[j].ActionID
		})
		out = append(out, patternTreeNode{NodeID: pat, Label: pat, Children: children})
	}
	return out
}

func normalizePatternActionName(name string) string {
	normalized := strings.TrimSpace(name)
	switch normalized {
	case "do_default_action", "doDefaultAction":
		return "doDefaultAction"
	case "set_value", "setValue":
		return "setValue"
	default:
		return normalized
	}
}

func callableActionLabel(name string) string {
	switch normalizePatternActionName(name) {
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
