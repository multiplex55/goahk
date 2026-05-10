//go:build windows

package main

import (
	"strings"

	"github.com/lxn/walk"
)

func (ui *viewerUI) attachTreeNavigationKeys() {
	if ui == nil || ui.treeView == nil {
		return
	}
	ui.treeView.KeyDown().Attach(func(key walk.Key) {
		switch key {
		case walk.KeyNumpad8, walk.KeyUp:
			ui.navigateTreeVisible(-1)
		case walk.KeyNumpad2, walk.KeyDown:
			ui.navigateTreeVisible(1)
		}
	})
}

func (ui *viewerUI) navigateTreeVisible(delta int) {
	if ui == nil || ui.treeView == nil || delta == 0 {
		return
	}
	items := ui.visibleTreeItems()
	if len(items) == 0 {
		return
	}
	currentIndex := ui.currentVisibleTreeIndex(items)
	if currentIndex < 0 {
		currentIndex = 0
	}
	nextIndex := currentIndex + delta
	if nextIndex < 0 {
		nextIndex = 0
	}
	if nextIndex >= len(items) {
		nextIndex = len(items) - 1
	}
	nextNode := items[nextIndex]
	if nextNode == nil {
		return
	}
	ui.treeView.SetCurrentItem(nextNode)
	ui.treeView.SetFocus()
}

func (ui *viewerUI) visibleTreeItems() []*uiaTreeNode {
	if ui == nil || ui.treeModel == nil || ui.treeModel.root == nil {
		return nil
	}
	items := make([]*uiaTreeNode, 0)
	ui.collectVisibleTreeItems(ui.treeModel.root, &items)
	return items
}

func (ui *viewerUI) collectVisibleTreeItems(node *uiaTreeNode, out *[]*uiaTreeNode) {
	if ui == nil || node == nil || out == nil {
		return
	}
	if !node.placeholder && strings.TrimSpace(node.Text()) != "Loading..." {
		*out = append(*out, node)
	}
	if ui.treeView == nil || !ui.treeView.Expanded(node) {
		return
	}
	for _, child := range node.children {
		ui.collectVisibleTreeItems(child, out)
	}
}

func (ui *viewerUI) currentVisibleTreeIndex(items []*uiaTreeNode) int {
	if ui == nil || ui.treeView == nil || len(items) == 0 {
		return -1
	}
	current, ok := ui.treeView.CurrentItem().(*uiaTreeNode)
	if !ok || current == nil {
		return -1
	}
	for i, item := range items {
		if item == current {
			return i
		}
	}
	return -1
}
