//go:build windows

package main

import (
	"goahk/internal/inspect"

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
}

func (n *uiaTreeNode) Text() string {
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
	loadedChildren map[NodeID]bool
	expanded       map[NodeID]bool
}

func newUIATreeModel() *uiaTreeModel {
	return &uiaTreeModel{nodes: map[NodeID]*uiaTreeNode{}, loadedChildren: map[NodeID]bool{}, expanded: map[NodeID]bool{}}
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
	}
}
func (m *uiaTreeModel) AreChildrenLoaded(nodeID string) bool { return m.loadedChildren[NodeID(nodeID)] }
func (m *uiaTreeModel) SetExpanded(nodeID string, expanded bool) {
	m.expanded[NodeID(nodeID)] = expanded
}
func (m *uiaTreeModel) IsExpanded(nodeID string) bool { return m.expanded[NodeID(nodeID)] }

func (m *uiaTreeModel) SetRoot(root inspect.TreeNodeDTO) {
	n := &uiaTreeNode{TreeNodeDTO: root, id: NodeID(root.NodeID), maybeHasChildren: true}
	m.root = n
	m.nodes = map[NodeID]*uiaTreeNode{n.id: n}
	m.loadedChildren = map[NodeID]bool{}
	m.expanded = map[NodeID]bool{}
	m.attachPlaceholder(n)
	m.PublishItemsReset(nil)
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
	m.loadedChildren = map[NodeID]bool{}
	m.expanded = map[NodeID]bool{}
	m.PublishItemsReset(nil)
}

func (m *uiaTreeModel) SetChildren(nodeID string, children []inspect.TreeNodeDTO) {
	pid := NodeID(nodeID)
	parent, ok := m.nodes[pid]
	if !ok {
		return
	}
	parent.children = parent.children[:0]
	for _, ch := range children {
		cid := NodeID(ch.NodeID)
		child, ok := m.nodes[cid]
		if !ok {
			child = &uiaTreeNode{id: cid}
			m.nodes[cid] = child
		}
		child.TreeNodeDTO = ch
		child.parent = parent
		child.placeholder = false
		child.maybeHasChildren = true
		if !m.AreChildrenLoaded(ch.NodeID) {
			m.attachPlaceholder(child)
		}
		parent.children = append(parent.children, child)
	}
	m.MarkChildrenLoaded(nodeID)
	// A parent-local reset ensures Walk re-queries the updated child list
	// immediately after lazy expansion without resetting the entire tree.
	m.PublishItemsReset(parent)
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
