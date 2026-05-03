//go:build windows

package main

import (
	"sort"

	"goahk/internal/inspect"
)

type windowTableRow struct {
	Title   string
	Process string
	ID      string
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
