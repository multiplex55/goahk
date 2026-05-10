//go:build windows

package main

import (
	"goahk/internal/inspect"
	"strings"

	"github.com/lxn/walk"
)

type NodeID string
type TreeExpansionSnapshot struct {
	ExpandedIDs  []string
	CollapsedIDs []string
	SelectedID   string
}

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
	root             *uiaTreeNode
	nodes            map[NodeID]*uiaTreeNode
	allRoot          *uiaTreeNode
	allNodes         map[NodeID]*uiaTreeNode
	allChildren      map[NodeID][]NodeID
	loadedChildren   map[NodeID]bool
	expanded         map[NodeID]bool
	userCollapsed    map[NodeID]bool
	filterText       string
	filterActive     bool
	matchedFilterIDs map[NodeID]bool
	visibleFilterIDs map[NodeID]bool
}

type TreeFilterResult struct {
	FilterText   string
	Active       bool
	MatchCount   int
	VisibleCount int
}

func newUIATreeModel() *uiaTreeModel {
	return &uiaTreeModel{
		nodes:            map[NodeID]*uiaTreeNode{},
		allNodes:         map[NodeID]*uiaTreeNode{},
		allChildren:      map[NodeID][]NodeID{},
		loadedChildren:   map[NodeID]bool{},
		expanded:         map[NodeID]bool{},
		userCollapsed:    map[NodeID]bool{},
		matchedFilterIDs: map[NodeID]bool{},
		visibleFilterIDs: map[NodeID]bool{},
	}
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
	if n, ok := m.allNodes[id]; ok {
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
func (m *uiaTreeModel) MarkExpanded(nodeID string) {
	id := NodeID(nodeID)
	m.expanded[id] = true
	delete(m.userCollapsed, id)
}
func (m *uiaTreeModel) MarkCollapsed(nodeID string) {
	id := NodeID(nodeID)
	m.expanded[id] = false
	m.userCollapsed[id] = true
}
func (m *uiaTreeModel) WasUserCollapsed(nodeID string) bool { return m.userCollapsed[NodeID(nodeID)] }
func (m *uiaTreeModel) ShouldAutoExpand(nodeID string) bool {
	if strings.TrimSpace(nodeID) == "" {
		return false
	}
	return !m.WasUserCollapsed(nodeID)
}

func (m *uiaTreeModel) SetRoot(root inspect.TreeNodeDTO) {
	n := &uiaTreeNode{TreeNodeDTO: root, id: NodeID(root.NodeID), maybeHasChildren: true}
	m.allRoot = n
	m.allNodes = map[NodeID]*uiaTreeNode{n.id: n}
	m.allChildren = map[NodeID][]NodeID{}
	m.loadedChildren = map[NodeID]bool{}
	m.expanded = map[NodeID]bool{}
	m.userCollapsed = map[NodeID]bool{}
	m.filterText = ""
	m.filterActive = false
	m.matchedFilterIDs = map[NodeID]bool{}
	m.visibleFilterIDs = map[NodeID]bool{}
	m.attachPlaceholder(n)
	m.rebuildVisibleProjection()
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
	m.allRoot = nil
	m.nodes = map[NodeID]*uiaTreeNode{}
	m.allNodes = map[NodeID]*uiaTreeNode{}
	m.allChildren = map[NodeID][]NodeID{}
	m.loadedChildren = map[NodeID]bool{}
	m.expanded = map[NodeID]bool{}
	m.userCollapsed = map[NodeID]bool{}
	m.filterText = ""
	m.filterActive = false
	m.matchedFilterIDs = map[NodeID]bool{}
	m.visibleFilterIDs = map[NodeID]bool{}
	m.PublishItemsReset(nil)
}

func (m *uiaTreeModel) SetChildren(nodeID string, children []inspect.TreeNodeDTO) {
	pid := NodeID(nodeID)
	parent, ok := m.allNodes[pid]
	if !ok {
		return
	}

	nextChildren := make([]*uiaTreeNode, 0, len(children))
	nextIDs := make([]NodeID, 0, len(children))
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
		if child.maybeHasChildren && !m.AreChildrenLoaded(ch.NodeID) {
			m.attachPlaceholder(child)
		} else {
			child.children = filterOutPlaceholders(child.children)
		}
		nextChildren = append(nextChildren, child)
		nextIDs = append(nextIDs, cid)
	}
	parent.children = nextChildren
	m.allChildren[pid] = nextIDs
	m.MarkChildrenLoaded(nodeID)
	m.rebuildVisibleProjection()
	m.PublishItemsReset(parent)
}

func (m *uiaTreeModel) AppendChildren(nodeID string, children []inspect.TreeNodeDTO) {
	pid := NodeID(nodeID)
	parent, ok := m.allNodes[pid]
	if !ok {
		return
	}
	existing := map[NodeID]bool{}
	for _, ch := range parent.children {
		if ch != nil && !ch.placeholder {
			existing[ch.id] = true
		}
	}
	parent.children = filterOutPlaceholders(parent.children)
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
		child.parent = parent
		child.placeholder = false
		child.maybeHasChildren = maybeHasChildren(dto)
		if child.maybeHasChildren && !m.AreChildrenLoaded(dto.NodeID) {
			m.attachPlaceholder(child)
		} else {
			child.children = filterOutPlaceholders(child.children)
		}
		parent.children = append(parent.children, child)
		m.allChildren[pid] = append(m.allChildren[pid], cid)
	}
	m.MarkChildrenLoaded(nodeID)
	m.rebuildVisibleProjection()
	m.PublishItemsReset(parent)
}

func filterOutPlaceholders(items []*uiaTreeNode) []*uiaTreeNode { /* unchanged */
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

func (m *uiaTreeModel) SnapshotExpansion() *TreeExpansionSnapshot {
	if m == nil {
		return nil
	}
	s := &TreeExpansionSnapshot{}
	for id, expanded := range m.expanded {
		if expanded {
			s.ExpandedIDs = append(s.ExpandedIDs, string(id))
			continue
		}
		s.CollapsedIDs = append(s.CollapsedIDs, string(id))
	}
	return s
}

func (m *uiaTreeModel) RestoreExpansion(snapshot *TreeExpansionSnapshot) {
	if m == nil || snapshot == nil {
		return
	}
	nextExpanded := map[NodeID]bool{}
	nextUserCollapsed := map[NodeID]bool{}
	for _, id := range snapshot.ExpandedIDs {
		nextExpanded[NodeID(id)] = true
	}
	for _, id := range snapshot.CollapsedIDs {
		nid := NodeID(id)
		nextExpanded[nid] = false
		nextUserCollapsed[nid] = true
	}
	m.expanded = nextExpanded
	m.userCollapsed = nextUserCollapsed
}

func (m *uiaTreeModel) ApplyFilter(filterText string) TreeFilterResult {
	m.filterText = filterText
	tokens := normalizeTreeFilter(filterText)
	m.filterActive = len(tokens) > 0
	if !m.filterActive {
		m.matchedFilterIDs = map[NodeID]bool{}
		m.visibleFilterIDs = map[NodeID]bool{}
		m.rebuildUnfilteredProjection()
		return TreeFilterResult{FilterText: m.filterText}
	}
	visible, matched := m.visibleIDsForFilter(tokens)
	m.visibleFilterIDs = visible
	m.matchedFilterIDs = matched
	m.rebuildVisibleProjection()
	return TreeFilterResult{
		FilterText:   m.filterText,
		Active:       true,
		MatchCount:   len(m.matchedFilterIDs),
		VisibleCount: len(m.visibleFilterIDs),
	}
}

func (m *uiaTreeModel) ClearFilter() TreeFilterResult {
	return m.ApplyFilter("")
}
func (m *uiaTreeModel) IsFiltered() bool   { return m.filterActive }
func (m *uiaTreeModel) FilterText() string { return m.filterText }
func (m *uiaTreeModel) FilteredMatchCount() int {
	return len(m.matchedFilterIDs)
}
func (m *uiaTreeModel) FilteredVisibleCount() int {
	return len(m.visibleFilterIDs)
}

func normalizeTreeFilter(filterText string) []string {
	fields := strings.Fields(strings.ToLower(strings.TrimSpace(filterText)))
	return fields
}

func nodeMatchesFilter(n *uiaTreeNode, tokens []string) bool {
	if n == nil || len(tokens) == 0 || n.placeholder {
		return false
	}
	haystack := strings.ToLower(strings.Join([]string{
		n.Text(),
		n.DisplayLabel,
		n.Name,
		n.LocalizedControlType,
		n.ControlType,
		n.AutomationID,
		n.ClassName,
		n.HWND,
	}, "\n"))
	for _, tok := range tokens {
		if !strings.Contains(haystack, tok) {
			return false
		}
	}
	return true
}

func (m *uiaTreeModel) visibleIDsForFilter(tokens []string) (visible map[NodeID]bool, matched map[NodeID]bool) {
	visible = map[NodeID]bool{}
	matched = map[NodeID]bool{}
	if m == nil || m.allRoot == nil || len(tokens) == 0 {
		return visible, matched
	}
	for id, n := range m.allNodes {
		if nodeMatchesFilter(n, tokens) {
			matched[id] = true
			visible[id] = true
			for p := n.parent; p != nil; p = p.parent {
				visible[p.id] = true
			}
		}
	}
	return visible, matched
}

func (m *uiaTreeModel) rebuildVisibleProjection() {
	if m == nil {
		return
	}
	if !m.filterActive {
		m.rebuildUnfilteredProjection()
		return
	}
	if m.allRoot == nil || !m.visibleFilterIDs[m.allRoot.id] {
		m.root = nil
		m.nodes = map[NodeID]*uiaTreeNode{}
		return
	}
	root := m.cloneFilteredNode(m.allRoot.id, nil, m.visibleFilterIDs)
	m.root = root
	m.nodes = map[NodeID]*uiaTreeNode{}
	var visit func(*uiaTreeNode)
	visit = func(n *uiaTreeNode) {
		if n == nil {
			return
		}
		m.nodes[n.id] = n
		for _, ch := range n.children {
			visit(ch)
		}
	}
	visit(root)
}

func (m *uiaTreeModel) rebuildUnfilteredProjection() {
	m.root = m.allRoot
	m.nodes = map[NodeID]*uiaTreeNode{}
	for id, n := range m.allNodes {
		m.nodes[id] = n
	}
}

func (m *uiaTreeModel) cloneFilteredNode(id NodeID, parent *uiaTreeNode, visible map[NodeID]bool) *uiaTreeNode {
	src, ok := m.allNodes[id]
	if !ok || src == nil || !visible[id] {
		return nil
	}
	dst := &uiaTreeNode{
		TreeNodeDTO:      src.TreeNodeDTO,
		id:               src.id,
		parent:           parent,
		loaded:           src.loaded,
		maybeHasChildren: src.maybeHasChildren,
		placeholder:      src.placeholder,
		ChildrenLoaded:   src.ChildrenLoaded,
		Loading:          src.Loading,
		LoadErr:          src.LoadErr,
	}
	for _, childID := range m.allChildren[id] {
		if !visible[childID] {
			continue
		}
		clonedChild := m.cloneFilteredNode(childID, dst, visible)
		if clonedChild != nil {
			dst.children = append(dst.children, clonedChild)
		}
	}
	return dst
}
