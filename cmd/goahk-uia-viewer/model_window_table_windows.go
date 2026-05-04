//go:build windows

package main

import (
	"sort"

	"github.com/lxn/walk"
	"goahk/internal/inspect"
)

type windowTableRow struct {
	Title   string
	Process string
	ID      string
}

type windowTableModel struct {
	walk.TableModelBase
	rows []windowTableRow
}

func newWindowTableModel() *windowTableModel { return &windowTableModel{} }

func (m *windowTableModel) RowCount() int { return len(m.rows) }

func (m *windowTableModel) Value(row, col int) any {
	if row < 0 || row >= len(m.rows) {
		return ""
	}
	r := m.rows[row]
	switch col {
	case 0:
		return r.Title
	case 1:
		return r.Process
	case 2:
		return r.ID
	default:
		return ""
	}
}

func (m *windowTableModel) Sort(col int, order walk.SortOrder) error {
	asc := order != walk.SortDescending
	sort.SliceStable(m.rows, func(i, j int) bool {
		a, b := m.rows[i], m.rows[j]
		cmp := 0
		switch col {
		case 0:
			cmp = compareStrings(a.Title, b.Title)
		case 1:
			cmp = compareStrings(a.Process, b.Process)
		case 2:
			cmp = compareStrings(a.ID, b.ID)
		default:
			cmp = compareStrings(a.Title, b.Title)
		}
		if cmp == 0 {
			cmp = compareStrings(a.ID, b.ID)
		}
		if asc {
			return cmp < 0
		}
		return cmp > 0
	})
	m.PublishRowsReset()
	return nil
}

func (m *windowTableModel) SetRows(rows []windowTableRow) {
	m.rows = append([]windowTableRow(nil), rows...)
	m.PublishRowsReset()
}

func (m *windowTableModel) WindowAt(row int) (windowTableRow, bool) {
	if row < 0 || row >= len(m.rows) {
		return windowTableRow{}, false
	}
	return m.rows[row], true
}

func compareStrings(a, b string) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

func mapWindowTableRows(windows []inspect.WindowSummary, stableSort bool) []windowTableRow {
	rows := make([]windowTableRow, 0, len(windows))
	for _, w := range windows {
		rows = append(rows, windowTableRow{Title: w.Title, Process: w.ProcessName, ID: w.HWND})
	}
	if stableSort {
		sort.SliceStable(rows, func(i, j int) bool {
			if rows[i].Title != rows[j].Title {
				return rows[i].Title < rows[j].Title
			}
			if rows[i].Process != rows[j].Process {
				return rows[i].Process < rows[j].Process
			}
			return rows[i].ID < rows[j].ID
		})
	}
	return rows
}
