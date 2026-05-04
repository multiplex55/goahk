//go:build windows

package main

import (
	"sort"

	"github.com/lxn/walk"
	"goahk/internal/inspect"
)

var propertyOrderAHK = []string{
	"ControlType", "LocalizedControlType", "Name", "Value", "AutomationId", "BoundingRectangle", "ClassName", "HelpText", "AccessKey", "AcceleratorKey", "HasKeyboardFocus", "IsKeyboardFocusable", "ItemType", "ProcessId", "IsEnabled", "IsPassword", "IsOffscreen", "FrameworkId", "IsRequiredForForm", "ItemStatus", "LabeledBy",
}

type propertyTableRow struct {
	Name   string
	Value  string
	Status string
}

type propertyTableModel struct {
	walk.TableModelBase
	rows []propertyTableRow
}

func newPropertyTableModel() *propertyTableModel { return &propertyTableModel{} }

func (m *propertyTableModel) RowCount() int { return len(m.rows) }
func (m *propertyTableModel) Value(row, col int) any {
	if row < 0 || row >= len(m.rows) {
		return ""
	}
	r := m.rows[row]
	switch col {
	case 0:
		return r.Name
	case 1:
		return r.Value
	case 2:
		return r.Status
	default:
		return ""
	}
}

func (m *propertyTableModel) SetRows(rows []propertyTableRow) {
	m.rows = append([]propertyTableRow(nil), rows...)
	m.PublishRowsReset()
}

func (m *propertyTableModel) RowAt(row int) (propertyTableRow, bool) {
	if row < 0 || row >= len(m.rows) {
		return propertyTableRow{}, false
	}
	return m.rows[row], true
}

func mapPropertyTableRows(props []inspect.PropertyDTO) []propertyTableRow {
	byName := make(map[string]inspect.PropertyDTO, len(props))
	for _, p := range props {
		byName[p.Name] = p
	}
	rows := make([]propertyTableRow, 0, len(propertyOrderAHK))
	seen := make(map[string]struct{}, len(propertyOrderAHK))
	for _, name := range propertyOrderAHK {
		seen[name] = struct{}{}
		row := propertyTableRow{Name: name, Status: "unsupported"}
		if p, ok := byName[name]; ok {
			if p.Value != nil {
				row.Value = *p.Value
			}
			if p.Status != "" {
				row.Status = p.Status
			}
		}
		rows = append(rows, row)
	}
	unknown := make([]inspect.PropertyDTO, 0)
	for _, p := range props {
		if _, ok := seen[p.Name]; ok {
			continue
		}
		unknown = append(unknown, p)
	}
	sort.SliceStable(unknown, func(i, j int) bool { return unknown[i].Name < unknown[j].Name })
	for _, p := range unknown {
		row := propertyTableRow{Name: p.Name, Status: "unsupported"}
		if p.Value != nil {
			row.Value = *p.Value
		}
		if p.Status != "" {
			row.Status = p.Status
		}
		rows = append(rows, row)
	}
	return rows
}
