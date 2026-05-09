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

func TestAutoExpandModeOptionsAndDefault(t *testing.T) {
	want := []string{"Manual / lazy", "Expand 1 level", "Expand 2 levels", "Expand 3 levels", "AHK snapshot"}
	if len(autoExpandModeOptions) != len(want) {
		t.Fatalf("options len=%d want=%d", len(autoExpandModeOptions), len(want))
	}
	for i := range want {
		if autoExpandModeOptions[i] != want[i] {
			t.Fatalf("option[%d]=%q want=%q", i, autoExpandModeOptions[i], want[i])
		}
	}
	if defaultAutoExpandModeIndex != 4 {
		t.Fatalf("default index=%d want=4", defaultAutoExpandModeIndex)
	}
}

func TestAutoExpandModeMapping(t *testing.T) {
	cases := []struct {
		idx       int
		wantDepth int
		wantRec   bool
	}{
		{idx: 0, wantDepth: 0, wantRec: false},
		{idx: 1, wantDepth: 1, wantRec: false},
		{idx: 2, wantDepth: 2, wantRec: false},
		{idx: 3, wantDepth: 3, wantRec: false},
		{idx: 4, wantDepth: 2, wantRec: true},
	}
	for _, tc := range cases {
		depth, rec := autoExpandMode(tc.idx)
		if depth != tc.wantDepth || rec != tc.wantRec {
			t.Fatalf("idx=%d got=(%d,%t) want=(%d,%t)", tc.idx, depth, rec, tc.wantDepth, tc.wantRec)
		}
	}
}
