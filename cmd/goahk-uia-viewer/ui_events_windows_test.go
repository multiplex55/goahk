//go:build windows

package main

import (
	"testing"

	"goahk/internal/inspect"
)

func TestInspectModeFromComboIndex(t *testing.T) {
	cases := []struct {
		idx  int
		want inspect.InspectMode
	}{
		{idx: 0, want: inspect.InspectModeUIATree},
		{idx: 1, want: inspect.InspectModeWindowTree},
		{idx: 2, want: inspect.InspectModeHWNDTree},
		{idx: 99, want: inspect.InspectModeUIATree},
	}
	for _, tc := range cases {
		if got := inspectModeFromComboIndex(tc.idx); got != tc.want {
			t.Fatalf("idx=%d mode=%s want=%s", tc.idx, got, tc.want)
		}
	}
}

func TestTreeFilterStatus(t *testing.T) {
	ui := &viewerUI{treeModel: newUIATreeModel()}
	ui.treeModel.matchedNodes = map[NodeID]bool{"a": true, "b": true}
	if got := ui.treeFilterStatus(" button "); got != `tree filter "button": 2 matches` {
		t.Fatalf("unexpected status: %q", got)
	}
	if got := ui.treeFilterStatus(" "); got != "tree filter cleared" {
		t.Fatalf("unexpected clear status: %q", got)
	}
}
