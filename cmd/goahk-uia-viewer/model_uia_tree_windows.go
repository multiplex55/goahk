//go:build windows

package main

import (
	"goahk/internal/inspect"
	"strings"

	"github.com/lxn/walk"
)

type NodeID string

type uiaTreeNode struct {
	inspect.TreeNodeDTO
	id               NodeID
	parent           *uiaTreeNode
	children         []*uiaTreeNode
	loaded           bool
	maybeHasChildren bool
	placeholder      bool
	ChildrenLoaded   bool
	Loading          bool
	LoadErr          error
}

func (n *uiaTreeNode) Text() string {
	if n != nil && n.placeholder {
		return "Loading..."
	}
	return uiaNodeLabel(n.NodeID, n.DisplayLabel, n.LocalizedControlType, n.ControlType, n.Name)
}
func (n *uiaTreeNode) Parent() walk.TreeItem {
	if n == nil || n.parent == nil {
		return nil
	}
	return n.parent
}
func (n *uiaTreeNode) ChildCount() int { return len(n.children) }
func (n *uiaTreeNode) ChildAt(index int) walk.TreeItem {
	if index < 0 || index >= len(n.children) {
		return nil
	}
	return n.children[index]
}

type uiaTreeModel struct {
	walk.TreeModelBase
	root           *uiaTreeNode
	nodes          map[NodeID]*uiaTreeNode
	allRoot        *uiaTreeNode
	allNodes       map[NodeID]*uiaTreeNode
	allChildren    map[NodeID][]NodeID
	filterText     string
	matchedNodes   map[NodeID]bool
	loadedChildren map[NodeID]bool
	expanded       map[NodeID]bool
}

func newUIATreeModel() *uiaTreeModel {
	return &uiaTreeModel{nodes: map[NodeID]*uiaTreeNode{}, allNodes: map[NodeID]*uiaTreeNode{}, allChildren: map[NodeID][]NodeID{}, matchedNodes: map[NodeID]bool{}, loadedChildren: map[NodeID]bool{}, expanded: map[NodeID]bool{}}
}

func (m *uiaTreeModel) LazyPopulation() bool { return true }
func (m *uiaTreeModel) RootCount() int {
	if m.root == nil {
		return 0
	}
	return 1
}
func (m *uiaTreeModel) RootAt(index int) walk.TreeItem {
	if index != 0 || m.root == nil {
		return nil
	}
	return m.root
}

func (m *uiaTreeModel) MarkChildrenLoaded(nodeID string) {
	id := NodeID(nodeID)
	m.loadedChildren[id] = true
	if n, ok := m.nodes[id]; ok {
		n.loaded = true
		n.ChildrenLoaded = true
		n.Loading = false
		n.LoadErr = nil
	}
}
func (m *uiaTreeModel) AreChildrenLoaded(nodeID string) bool { return m.loadedChildren[NodeID(nodeID)] }
func (m *uiaTreeModel) SetExpanded(nodeID string, expanded bool) {
	m.expanded[NodeID(nodeID)] = expanded
}
func (m *uiaTreeModel) IsExpanded(nodeID string) bool { return m.expanded[NodeID(nodeID)] }

func (m *uiaTreeModel) SetRoot(root inspect.TreeNodeDTO) {
	n := &uiaTreeNode{TreeNodeDTO: root, id: NodeID(root.NodeID), maybeHasChildren: true}
	m.allRoot = n
	m.allNodes = map[NodeID]*uiaTreeNode{n.id: n}
	m.allChildren = map[NodeID][]NodeID{}
	m.loadedChildren = map[NodeID]bool{}
	m.expanded = map[NodeID]bool{}
	m.filterText = ""
	m.matchedNodes = map[NodeID]bool{}
	m.rebuildVisibleGraph()
}
func (m *uiaTreeModel) RootID() string {
	if m.root == nil {
		return ""
	}
	return m.root.NodeID
}
func (m *uiaTreeModel) NodeCount() int { return len(m.nodes) }
func (m *uiaTreeModel) Reset() {
	m.root = nil
	m.nodes = map[NodeID]*uiaTreeNode{}
	m.allRoot = nil
	m.allNodes = map[NodeID]*uiaTreeNode{}
	m.allChildren = map[NodeID][]NodeID{}
	m.filterText = ""
	m.matchedNodes = map[NodeID]bool{}
	m.loadedChildren = map[NodeID]bool{}
	m.expanded = map[NodeID]bool{}
	m.PublishItemsReset(nil)
}

func (m *uiaTreeModel) SetChildren(nodeID string, children []inspect.TreeNodeDTO) {
	pid := NodeID(nodeID)
	parent, ok := m.nodes[pid]
	if !ok {
		parent, ok = m.allNodes[pid]
	}
	if !ok {
		return
	}
	m.allChildren[pid] = m.allChildren[pid][:0]
	for _, ch := range children {
		cid := NodeID(ch.NodeID)
		child, ok := m.allNodes[cid]
		if !ok {
			child = &uiaTreeNode{id: cid}
			m.allNodes[cid] = child
		}
		child.TreeNodeDTO = ch
		child.parent = parent
		child.placeholder = false
		child.maybeHasChildren = maybeHasChildren(ch)
		m.allChildren[pid] = append(m.allChildren[pid], cid)
	}
	m.MarkChildrenLoaded(nodeID)
	m.rebuildVisibleGraph()
}

func (m *uiaTreeModel) AppendChildren(nodeID string, children []inspect.TreeNodeDTO) {
	pid := NodeID(nodeID)
	parent, ok := m.allNodes[pid]
	if !ok {
		return
	}
	existing := map[NodeID]bool{}
	for _, cid := range m.allChildren[pid] {
		existing[cid] = true
	}
	_ = parent
	for _, dto := range children {
		cid := NodeID(dto.NodeID)
		if existing[cid] {
			continue
		}
		child, ok := m.allNodes[cid]
		if !ok {
			child = &uiaTreeNode{id: cid}
			m.allNodes[cid] = child
		}
		child.TreeNodeDTO = dto
		child.parent = m.allNodes[pid]
		child.placeholder = false
		child.maybeHasChildren = maybeHasChildren(dto)
		m.allChildren[pid] = append(m.allChildren[pid], cid)
	}
	m.MarkChildrenLoaded(nodeID)
	m.rebuildVisibleGraph()
}

func (m *uiaTreeModel) SetFilter(text string) {
	m.filterText = strings.ToLower(strings.TrimSpace(text))
	m.rebuildVisibleGraph()
}

func (m *uiaTreeModel) VisibleMatchCount() int { return len(m.matchedNodes) }

func (m *uiaTreeModel) ExpandedIDsForFilter() []string {
	if m.filterText == "" {
		return nil
	}
	ids := map[NodeID]bool{}
	for id := range m.matchedNodes {
		n := m.nodes[id]
		for p := n.parent; p != nil; p = p.parent {
			ids[p.id] = true
		}
	}
	out := make([]string, 0, len(ids))
	for id := range ids {
		out = append(out, string(id))
	}
	return out
}

func (m *uiaTreeModel) rebuildVisibleGraph() {
	m.nodes = map[NodeID]*uiaTreeNode{}
	m.matchedNodes = map[NodeID]bool{}
	if m.allRoot == nil {
		m.root = nil
		m.PublishItemsReset(nil)
		return
	}
	var clone func(n *uiaTreeNode, parent *uiaTreeNode) *uiaTreeNode
	clone = func(n *uiaTreeNode, parent *uiaTreeNode) *uiaTreeNode {
		if n == nil {
			return nil
		}
		current := &uiaTreeNode{TreeNodeDTO: n.TreeNodeDTO, id: n.id, parent: parent, loaded: n.loaded, maybeHasChildren: n.maybeHasChildren, ChildrenLoaded: n.ChildrenLoaded, Loading: n.Loading, LoadErr: n.LoadErr}
		matchSelf := m.filterText == "" || treeNodeMatchesFilter(current, m.filterText)
		includedKids := []*uiaTreeNode{}
		for _, cid := range m.allChildren[n.id] {
			if child := clone(m.allNodes[cid], current); child != nil {
				includedKids = append(includedKids, child)
			}
		}
		if !matchSelf && len(includedKids) == 0 {
			return nil
		}
		if matchSelf && m.filterText != "" {
			m.matchedNodes[current.id] = true
		}
		current.children = includedKids
		if len(current.children) == 0 && current.maybeHasChildren && !m.AreChildrenLoaded(current.NodeID) {
			m.attachPlaceholder(current)
		}
		m.nodes[current.id] = current
		return current
	}
	m.root = clone(m.allRoot, nil)
	m.PublishItemsReset(nil)
}

func treeNodeMatchesFilter(node *uiaTreeNode, filter string) bool {
	if node == nil || filter == "" {
		return false
	}
	fields := []string{node.DisplayLabel, node.Name, node.ControlType, node.LocalizedControlType, node.ClassName, node.DebugMeta.AutomationID, node.RuntimeID, node.NodeID}
	for _, field := range fields {
		if strings.Contains(strings.ToLower(strings.TrimSpace(field)), filter) {
			return true
		}
	}
	return false
}

func filterOutPlaceholders(items []*uiaTreeNode) []*uiaTreeNode {
	out := make([]*uiaTreeNode, 0, len(items))
	for _, it := range items {
		if it != nil && !it.placeholder {
			out = append(out, it)
		}
	}
	return out
}

func (m *uiaTreeModel) attachPlaceholder(parent *uiaTreeNode) {
	if parent == nil || !parent.maybeHasChildren || m.AreChildrenLoaded(parent.NodeID) {
		return
	}
	for _, existing := range parent.children {
		if existing != nil && existing.placeholder {
			return
		}
	}
	placeholder := &uiaTreeNode{id: NodeID(parent.NodeID + "#placeholder"), parent: parent, placeholder: true}
	parent.children = append(parent.children, placeholder)
}

func maybeHasChildren(node inspect.TreeNodeDTO) bool {
	if node.ChildCount != nil {
		return *node.ChildCount > 0
	}
	// Unknown child count keeps lazy-loading possible for discovered nodes.
	return true
}

func (m *uiaTreeModel) NodeByID(nodeID string) (inspect.TreeNodeDTO, bool) {
	n, ok := m.nodes[NodeID(nodeID)]
	if !ok {
		return inspect.TreeNodeDTO{}, false
	}
	return n.TreeNodeDTO, true
}
func (m *uiaTreeModel) ItemByID(nodeID string) (*uiaTreeNode, bool) {
	n, ok := m.nodes[NodeID(nodeID)]
	return n, ok
}
func (m *uiaTreeModel) ShouldShowLazyPlaceholder(nodeID string) bool {
	if !m.AreChildrenLoaded(nodeID) {
		return true
	}
	n, ok := m.nodes[NodeID(nodeID)]
	if !ok {
		return false
	}
	return len(n.children) == 0
}
