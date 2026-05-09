package main

import "goahk/internal/inspect"

type fakeNotepadTreeFixture struct {
	Root       inspect.TreeNodeDTO
	ChildrenBy map[string][]inspect.TreeNodeDTO
	NonLeafIDs []string
	AllNodeIDs []string
}

func buildFakeNotepadTreeFixture() fakeNotepadTreeFixture {
	root := inspect.TreeNodeDTO{NodeID: "root", LocalizedControlType: "window", Name: ""}
	menuBar := inspect.TreeNodeDTO{NodeID: "menu", LocalizedControlType: "menu bar", Name: ""}
	pane := inspect.TreeNodeDTO{NodeID: "pane", LocalizedControlType: "pane", Name: ""}
	text := inspect.TreeNodeDTO{NodeID: "text", LocalizedControlType: "text", Name: ""}
	dupA := inspect.TreeNodeDTO{NodeID: "dup-a", LocalizedControlType: "button", Name: "Save"}
	dupB := inspect.TreeNodeDTO{NodeID: "dup-b", LocalizedControlType: "button", Name: "Save"}
	status := inspect.TreeNodeDTO{NodeID: "status", LocalizedControlType: "status bar", Name: "Ready"}

	childrenBy := map[string][]inspect.TreeNodeDTO{
		"root":   {menuBar, pane, status},
		"menu":   {},
		"pane":   {text, dupA, dupB},
		"text":   {},
		"dup-a":  {},
		"dup-b":  {},
		"status": {},
	}
	return fakeNotepadTreeFixture{
		Root:       root,
		ChildrenBy: childrenBy,
		NonLeafIDs: []string{"root", "pane"},
		AllNodeIDs: []string{"root", "menu", "pane", "text", "dup-a", "dup-b", "status"},
	}
}
