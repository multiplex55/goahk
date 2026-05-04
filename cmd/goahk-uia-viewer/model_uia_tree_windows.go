//go:build windows

package main

import "goahk/internal/inspect"

type uiaTreeNode struct {
	inspect.TreeNodeDTO
	children []string
	loaded   bool
}

type uiaTreeModel struct {
	nodes          map[string]*uiaTreeNode
	rootID         string
	loadedChildren map[string]bool
	expanded       map[string]bool
}

func newUIATreeModel() *uiaTreeModel {
	return &uiaTreeModel{nodes: map[string]*uiaTreeNode{}, loadedChildren: map[string]bool{}, expanded: map[string]bool{}}
}

func (m *uiaTreeModel) Label(n inspect.TreeNodeDTO) string {
	return uiaNodeLabel(n.NodeID, n.DisplayLabel, n.LocalizedControlType, n.ControlType, n.Name)
}

func (m *uiaTreeModel) MarkChildrenLoaded(nodeID string) {
	m.loadedChildren[nodeID] = true
	if n, ok := m.nodes[nodeID]; ok {
		n.loaded = true
	}
}

func (m *uiaTreeModel) AreChildrenLoaded(nodeID string) bool {
	return m.loadedChildren[nodeID]
}

func (m *uiaTreeModel) SetExpanded(nodeID string, expanded bool) {
	m.expanded[nodeID] = expanded
}

func (m *uiaTreeModel) IsExpanded(nodeID string) bool {
	return m.expanded[nodeID]
}

func (m *uiaTreeModel) SetRoot(root inspect.TreeNodeDTO) {
	m.rootID = root.NodeID
	m.nodes = map[string]*uiaTreeNode{root.NodeID: {TreeNodeDTO: root}}
	m.loadedChildren = map[string]bool{}
	m.expanded = map[string]bool{}
}

func (m *uiaTreeModel) RootID() string { return m.rootID }

func (m *uiaTreeModel) Reset() {
	m.rootID = ""
	m.nodes = map[string]*uiaTreeNode{}
	m.loadedChildren = map[string]bool{}
	m.expanded = map[string]bool{}
}

func (m *uiaTreeModel) SetChildren(nodeID string, children []inspect.TreeNodeDTO) {
	n, ok := m.nodes[nodeID]
	if !ok {
		n = &uiaTreeNode{TreeNodeDTO: inspect.TreeNodeDTO{NodeID: nodeID}}
		m.nodes[nodeID] = n
	}
	n.children = n.children[:0]
	for _, ch := range children {
		cid := ch.NodeID
		n.children = append(n.children, cid)
		if existing, ok := m.nodes[cid]; ok {
			existing.TreeNodeDTO = ch
		} else {
			m.nodes[cid] = &uiaTreeNode{TreeNodeDTO: ch}
		}
	}
	m.MarkChildrenLoaded(nodeID)
}

func (m *uiaTreeModel) ChildrenOf(nodeID string) []string {
	n, ok := m.nodes[nodeID]
	if !ok {
		return nil
	}
	return append([]string(nil), n.children...)
}

func (m *uiaTreeModel) ShouldShowLazyPlaceholder(nodeID string) bool {
	if !m.AreChildrenLoaded(nodeID) {
		return true
	}
	return len(m.ChildrenOf(nodeID)) == 0
}
