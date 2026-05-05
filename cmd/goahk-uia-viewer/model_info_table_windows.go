//go:build windows

package main

import (
	"fmt"
	"strings"

	"github.com/lxn/walk"
	"goahk/internal/inspect"
)

type infoTableRow struct {
	Property string
	Value    string
}

type infoTableModel struct {
	walk.TableModelBase
	rows []infoTableRow
}

func newInfoTableModel() *infoTableModel { return &infoTableModel{} }

func (m *infoTableModel) RowCount() int { return len(m.rows) }

func (m *infoTableModel) Value(row, col int) any {
	if row < 0 || row >= len(m.rows) {
		return ""
	}
	r := m.rows[row]
	switch col {
	case 0:
		return r.Property
	case 1:
		return r.Value
	default:
		return ""
	}
}

func (m *infoTableModel) SetRows(rows []infoTableRow) {
	m.rows = append([]infoTableRow(nil), rows...)
	m.PublishRowsReset()
}

func mapWindowInfoRows(details *inspect.GetNodeDetailsResponse) []infoTableRow {
	if details == nil {
		return nil
	}
	window := details.WindowInfo
	element := details.Element
	rect := window.Rect
	if rect == nil {
		rect = element.Bounds
	}
	classNN := strings.TrimSpace(window.Class)
	if classNN == "" {
		classNN = "N/A"
	}
	return []infoTableRow{
		{Property: "Title", Value: fallback(window.Title)},
		{Property: "Text", Value: fallback(window.Text)},
		{Property: "Hwnd", Value: formatAHKID(window.HWND, element.HWND)},
		{Property: "Location", Value: formatLocation(rect)},
		{Property: "Size", Value: formatSize(rect)},
		{Property: "Class(NN)", Value: classNN},
		{Property: "Process", Value: fallback(window.Process)},
		{Property: "PID", Value: intOrNA(window.PID)},
	}
}

func formatAHKID(windowHWND, elementHWND string) string {
	hwnd := strings.TrimSpace(windowHWND)
	if hwnd == "" {
		hwnd = strings.TrimSpace(elementHWND)
	}
	if hwnd == "" {
		return "N/A"
	}
	if strings.HasPrefix(strings.ToLower(hwnd), "0x") {
		return "ahk_id " + hwnd
	}
	return "ahk_id 0x" + hwnd
}

func formatLocation(rect *inspect.Rect) string {
	if rect == nil {
		return "N/A"
	}
	return fmt.Sprintf("x:%d y:%d", rect.Left, rect.Top)
}

func formatSize(rect *inspect.Rect) string {
	if rect == nil {
		return "N/A"
	}
	return fmt.Sprintf("w:%d h:%d", rect.Width, rect.Height)
}

func fallback(v string) string {
	if strings.TrimSpace(v) == "" {
		return "N/A"
	}
	return v
}

func intOrNA(v int) string {
	if v <= 0 {
		return "N/A"
	}
	return fmt.Sprintf("%d", v)
}
