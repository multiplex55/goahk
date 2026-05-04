//go:build windows

package main

import (
	"testing"

	"github.com/lxn/walk"
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

func TestWindowTableModel_WindowAtAndSort(t *testing.T) {
	m := newWindowTableModel()
	m.SetRows([]windowTableRow{{Title: "B", Process: "p2", ID: "2"}, {Title: "A", Process: "p1", ID: "1"}})
	if _, ok := m.WindowAt(-1); ok {
		t.Fatal("expected invalid index false")
	}
	if err := m.Sort(0, walk.SortAscending); err != nil {
		t.Fatalf("sort failed: %v", err)
	}
	r, ok := m.WindowAt(0)
	if !ok || r.ID != "1" {
		t.Fatalf("unexpected first row after sort: %#v", r)
	}
}
