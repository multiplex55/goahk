//go:build windows

package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

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
				row.Value = formatPropertyValue(name, *p.Value, byName)
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
			row.Value = formatPropertyValue(p.Name, *p.Value, byName)
		}
		if p.Status != "" {
			row.Status = p.Status
		}
		rows = append(rows, row)
	}
	return rows
}

func mapPropertyRowsFromDetails(details inspect.GetNodeDetailsResponse) []propertyTableRow {
	if len(details.Properties) > 0 {
		return mapPropertyTableRows(details.Properties)
	}

	e := details.Element
	bounds := ""
	if e.Bounds != nil {
		bounds = fmt.Sprintf("%d %d %d %d", e.Bounds.Left, e.Bounds.Top, e.Bounds.Width, e.Bounds.Height)
	}
	processID := ""
	if details.WindowInfo.PID > 0 {
		processID = strconv.Itoa(details.WindowInfo.PID)
	}

	status := func(name string, fallback string) string {
		_ = name
		return fallback
	}

	props := []inspect.PropertyDTO{
		{Name: "ControlType", Value: stringPtrOrNil(e.ControlType), Status: statusFromValue(e.ControlType)},
		{Name: "LocalizedControlType", Value: stringPtrOrNil(e.LocalizedControlType), Status: status("LocalizedControlType", statusFromValue(e.LocalizedControlType))},
		{Name: "Name", Value: stringPtrOrNil(e.Name), Status: status("Name", statusFromValue(e.Name))},
		{Name: "Value", Value: stringPtrOrNil(e.Value), Status: status("Value", statusFromValue(e.Value))},
		{Name: "AutomationId", Value: stringPtrOrNil(e.AutomationID), Status: status("AutomationId", statusFromValue(e.AutomationID))},
		{Name: "BoundingRectangle", Value: stringPtrOrNil(bounds), Status: status("BoundingRectangle", statusFromValue(bounds))},
		{Name: "ClassName", Value: stringPtrOrNil(details.WindowInfo.Class), Status: status("ClassName", statusFromValue(details.WindowInfo.Class))},
		{Name: "HelpText", Value: stringPtrOrNil(e.HelpText), Status: status("HelpText", statusFromValue(e.HelpText))},
		{Name: "AccessKey", Value: stringPtrOrNil(e.AccessKey), Status: status("AccessKey", statusFromValue(e.AccessKey))},
		{Name: "AcceleratorKey", Value: stringPtrOrNil(e.AcceleratorKey), Status: status("AcceleratorKey", statusFromValue(e.AcceleratorKey))},
		{Name: "HasKeyboardFocus", Value: stringPtrOrNil(strconv.FormatBool(e.HasKeyboardFocus)), Status: status("HasKeyboardFocus", "ok")},
		{Name: "IsKeyboardFocusable", Value: stringPtrOrNil(strconv.FormatBool(e.IsKeyboardFocusable)), Status: status("IsKeyboardFocusable", "ok")},
		{Name: "ItemType", Value: stringPtrOrNil(e.ItemType), Status: status("ItemType", statusFromValue(e.ItemType))},
		{Name: "ProcessId", Value: stringPtrOrNil(processID), Status: status("ProcessId", statusFromValue(processID))},
		{Name: "IsEnabled", Value: stringPtrOrNil(strconv.FormatBool(e.IsEnabled)), Status: status("IsEnabled", "ok")},
		{Name: "IsPassword", Value: stringPtrOrNil(strconv.FormatBool(e.IsPassword)), Status: status("IsPassword", "ok")},
		{Name: "IsOffscreen", Value: stringPtrOrNil(strconv.FormatBool(e.IsOffscreen)), Status: status("IsOffscreen", "ok")},
		{Name: "FrameworkId", Value: stringPtrOrNil(e.FrameworkID), Status: status("FrameworkId", statusFromValue(e.FrameworkID))},
		{Name: "IsRequiredForForm", Value: stringPtrOrNil(strconv.FormatBool(e.IsRequiredForForm)), Status: status("IsRequiredForForm", "ok")},
		{Name: "ItemStatus", Value: stringPtrOrNil(e.ItemStatus), Status: status("ItemStatus", statusFromValue(e.ItemStatus))},
		{Name: "LabeledBy", Value: nil, Status: status("LabeledBy", "unsupported")},
	}

	return mapPropertyTableRows(props)
}

func stringPtrOrNil(v string) *string {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	return &v
}

func statusFromValue(v string) string {
	if strings.TrimSpace(v) == "" {
		return "unsupported"
	}
	return "ok"
}

func formatPropertyValue(name, value string, byName map[string]inspect.PropertyDTO) string {
	switch name {
	case "BoundingRectangle":
		if formatted, ok := formatBoundingRectangle(value); ok {
			return formatted
		}
	case "ControlType":
		if localized := propertyValueFromMap(byName, "LocalizedControlType"); localized != "" {
			if strings.Contains(value, "(") {
				return value
			}
			return fmt.Sprintf("%s (%s)", value, localized)
		}
	}
	return value
}

func formatBoundingRectangle(value string) (string, bool) {
	fields := strings.FieldsFunc(value, func(r rune) bool { return r == ',' || r == ';' || r == ' ' || r == '\t' })
	if len(fields) != 4 {
		return "", false
	}
	nums := make([]int, 0, 4)
	for _, f := range fields {
		n, err := strconv.Atoi(strings.TrimSpace(f))
		if err != nil {
			return "", false
		}
		nums = append(nums, n)
	}
	return fmt.Sprintf("x:%d y:%d w:%d h:%d | l:%d t:%d r:%d b:%d", nums[0], nums[1], nums[2], nums[3], nums[0], nums[1], nums[0]+nums[2], nums[1]+nums[3]), true
}
