//go:build windows

package main

import "goahk/internal/inspect"

var propertyOrderAHK = []string{
	"ControlType", "LocalizedControlType", "Name", "Value", "AutomationId", "BoundingRectangle", "ClassName", "HelpText", "AccessKey", "AcceleratorKey", "HasKeyboardFocus", "IsKeyboardFocusable", "ItemType", "ProcessId", "IsEnabled", "IsPassword", "IsOffscreen", "FrameworkId", "IsRequiredForForm", "ItemStatus", "LabeledBy",
}

type propertyTableRow struct {
	Name   string
	Value  string
	Status string
}

func mapPropertyTableRows(props []inspect.PropertyDTO) []propertyTableRow {
	byName := make(map[string]inspect.PropertyDTO, len(props))
	for _, p := range props {
		byName[p.Name] = p
	}
	rows := make([]propertyTableRow, 0, len(propertyOrderAHK))
	for _, name := range propertyOrderAHK {
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
	return rows
}
