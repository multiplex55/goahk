//go:build windows

package main

import (
	"testing"

	"goahk/internal/inspect"
)

func TestMapWindowTableRowsMapping(t *testing.T) {
	rows := mapWindowTableRows([]inspect.WindowSummary{{HWND: "0x1", Title: "A", ProcessName: "p"}}, false)
	if len(rows) != 1 || rows[0].Title != "A" || rows[0].Process != "p" || rows[0].ID != "0x1" {
		t.Fatalf("unexpected rows: %#v", rows)
	}
}

func TestMapWindowTableRowsStableSort(t *testing.T) {
	in := []inspect.WindowSummary{{HWND: "0x2", Title: "B", ProcessName: "p2"}, {HWND: "0x1", Title: "A", ProcessName: "p1"}}
	rows := mapWindowTableRows(in, true)
	if rows[0].ID != "0x1" || rows[1].ID != "0x2" {
		t.Fatalf("expected sorted rows, got %#v", rows)
	}
}
