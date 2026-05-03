//go:build windows

package main

import "goahk/internal/inspect"

type uiaTreeModel struct {
	loadedChildren map[string]bool
	expanded       map[string]bool
}

func newUIATreeModel() *uiaTreeModel {
	return &uiaTreeModel{loadedChildren: map[string]bool{}, expanded: map[string]bool{}}
}

func (m *uiaTreeModel) Label(n inspect.TreeNodeDTO) string {
	return uiaNodeLabel(n.NodeID, n.DisplayLabel, n.LocalizedControlType, n.ControlType, n.Name)
}

func (m *uiaTreeModel) MarkChildrenLoaded(nodeID string) {
	m.loadedChildren[nodeID] = true
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
