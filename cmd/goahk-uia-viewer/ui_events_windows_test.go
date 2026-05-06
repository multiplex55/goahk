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
