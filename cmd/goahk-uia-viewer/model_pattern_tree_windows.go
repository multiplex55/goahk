//go:build windows

package main

import (
	"sort"
	"strings"

	"github.com/lxn/walk"
	"goahk/internal/inspect"
)

type ActionID string

type patternTreeNode struct {
	id       string
	label    string
	actionID ActionID
	enabled  bool
	parent   *patternTreeNode
	children []patternTreeNode
}

const emptyPatternNodeID = "patterns.empty"

func (n *patternTreeNode) Text() string { return n.label }
func (n *patternTreeNode) Parent() walk.TreeItem {
	if n == nil || n.parent == nil {
		return nil
	}
	return n.parent
}
func (n *patternTreeNode) ChildCount() int { return len(n.children) }
func (n *patternTreeNode) ChildAt(index int) walk.TreeItem {
	if index < 0 || index >= len(n.children) {
		return nil
	}
	return &n.children[index]
}
func (n *patternTreeNode) ActionID() ActionID { return n.actionID }
func (n *patternTreeNode) IsActionableLeaf() bool {
	if n == nil {
		return false
	}
	return len(n.children) == 0 && n.enabled && isSupportedPatternAction(string(n.actionID))
}

type patternTreeModel struct {
	walk.TreeModelBase
	roots []*patternTreeNode
	nodes map[string]*patternTreeNode
}

func newPatternTreeModel() *patternTreeModel {
	return &patternTreeModel{nodes: map[string]*patternTreeNode{}}
}
func (m *patternTreeModel) LazyPopulation() bool { return false }
func (m *patternTreeModel) RootCount() int       { return len(m.roots) }
func (m *patternTreeModel) RootAt(index int) walk.TreeItem {
	if index < 0 || index >= len(m.roots) {
		return nil
	}
	return m.roots[index]
}

func (m *patternTreeModel) SetRoots(roots []patternTreeNode) {
	if len(roots) == 0 {
		roots = []patternTreeNode{{
			id:    emptyPatternNodeID,
			label: "No supported patterns",
		}}
	}
	m.nodes = map[string]*patternTreeNode{}
	m.roots = m.roots[:0]
	for i := range roots {
		r := m.cloneAndIndex(nil, &roots[i])
		m.roots = append(m.roots, r)
	}
	m.PublishItemsReset(nil)
}

func (m *patternTreeModel) cloneAndIndex(parent *patternTreeNode, src *patternTreeNode) *patternTreeNode {
	n := &patternTreeNode{id: src.id, label: src.label, actionID: src.actionID, enabled: src.enabled, parent: parent}
	m.nodes[n.id] = n
	for i := range src.children {
		child := m.cloneAndIndex(n, &src.children[i])
		n.children = append(n.children, *child)
	}
	return n
}

func (m *patternTreeModel) NodeByID(nodeID string) (*patternTreeNode, bool) {
	n, ok := m.nodes[nodeID]
	return n, ok
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
		label := strings.TrimSpace(a.DisplayName)
		if label == "" {
			label = callableActionLabel(action)
		}
		enabled := true
		if a.Supported || a.Enabled {
			enabled = a.Supported && a.Enabled
		}
		groups[pat] = append(groups[pat], patternTreeNode{id: childID, label: label, actionID: ActionID(action), enabled: enabled})
	}
	sort.Strings(order)
	out := make([]patternTreeNode, 0, len(order))
	for _, pat := range order {
		children := groups[pat]
		sort.SliceStable(children, func(i, j int) bool { return children[i].actionID < children[j].actionID })
		group := patternTreeNode{id: pat, label: pat}
		for i := range children {
			child := children[i]
			group.children = append(group.children, child)
		}
		out = append(out, group)
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

func isSupportedPatternAction(name string) bool {
	switch normalizePatternActionName(name) {
	case "invoke", "doDefaultAction", "select", "setValue", "toggle", "expand", "collapse":
		return true
	default:
		return false
	}
}
