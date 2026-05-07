//go:build windows

package main

import (
	"testing"

	"goahk/internal/inspect"
)

func TestMapPropertyTableRowsFixedOrderingAndFallback(t *testing.T) {
	v := "Button"
	rows := mapPropertyTableRows([]inspect.PropertyDTO{{Name: "ControlType", Value: &v, Status: "ok"}})
	if len(rows) != len(propertyOrderAHK) {
		t.Fatalf("row count mismatch: %d", len(rows))
	}
	if rows[0].Name != "ControlType" || rows[0].Value != "Button" || rows[0].Status != "ok" {
		t.Fatalf("unexpected first row: %#v", rows[0])
	}
	if rows[1].Name != "LocalizedControlType" || rows[1].Value != "" || rows[1].Status != "unsupported" {
		t.Fatalf("unexpected fallback row: %#v", rows[1])
	}
}

func TestMapPropertyTableRows_UnknownPropertiesAppendedDeterministically(t *testing.T) {
	v1, v2 := "1", "2"
	rows := mapPropertyTableRows([]inspect.PropertyDTO{{Name: "ZZZ", Value: &v1, Status: "ok"}, {Name: "AAA", Value: &v2, Status: "ok"}})
	if got := rows[len(rows)-2].Name; got != "AAA" {
		t.Fatalf("expected AAA first unknown, got %q", got)
	}
	if got := rows[len(rows)-1].Name; got != "ZZZ" {
		t.Fatalf("expected ZZZ second unknown, got %q", got)
	}
}

func TestPropertyTableModelSetRowsNonEmpty(t *testing.T) {
	value := "ok"
	model := newPropertyTableModel()
	rows := mapPropertyTableRows([]inspect.PropertyDTO{{Name: "Name", Value: &value, Status: "ok"}})
	model.SetRows(rows)
	if model.RowCount() == 0 {
		t.Fatal("expected non-empty rows after refresh mapping")
	}
	row, found := model.RowAt(2)
	if !found || row.Name != "Name" || row.Value != "ok" {
		t.Fatalf("unexpected mapped row at Name index: found=%v row=%#v", found, row)
	}
}

func TestPropertyRowsPopulateAHKOrderWithValues(t *testing.T) {
	name := "Calculator"
	rows := mapPropertyTableRows([]inspect.PropertyDTO{{Name: "Name", Value: &name, Status: "ok"}})
	if len(rows) < 3 {
		t.Fatalf("unexpected row count: %d", len(rows))
	}
	if rows[2].Name != "Name" || rows[2].Value != "Calculator" {
		t.Fatalf("expected Name row populated at AHK order index, got %#v", rows[2])
	}
}

func TestPropertyBoundingRectangleFormatting(t *testing.T) {
	v := "10,20,30,40"
	rows := mapPropertyTableRows([]inspect.PropertyDTO{{Name: "BoundingRectangle", Value: &v, Status: "ok"}})
	var got string
	for _, row := range rows {
		if row.Name == "BoundingRectangle" {
			got = row.Value
			break
		}
	}
	want := "x:10 y:20 w:30 h:40 | l:10 t:20 r:40 b:60"
	if got != want {
		t.Fatalf("BoundingRectangle=%q want=%q", got, want)
	}
}

func TestPropertyControlTypeFormatting(t *testing.T) {
	ct := "50004"
	localized := "edit"
	rows := mapPropertyTableRows([]inspect.PropertyDTO{{Name: "ControlType", Value: &ct, Status: "ok"}, {Name: "LocalizedControlType", Value: &localized, Status: "ok"}})
	if rows[0].Value != "50004 (edit)" {
		t.Fatalf("ControlType=%q", rows[0].Value)
	}
}

func TestPropertyValueFromMap_ExistingKeyWithValue(t *testing.T) {
	v := "value"
	byName := map[string]inspect.PropertyDTO{"Name": {Name: "Name", Value: &v}}
	if got := propertyValueFromMap(byName, "Name"); got != "value" {
		t.Fatalf("propertyValueFromMap existing=%q want=%q", got, "value")
	}
}

func TestPropertyValueFromMap_MissingOrNilReturnsEmpty(t *testing.T) {
	v := "value"
	var nilPtr *string
	byName := map[string]inspect.PropertyDTO{
		"Name":  {Name: "Name", Value: &v},
		"Value": {Name: "Value", Value: nilPtr},
	}
	if got := propertyValueFromMap(byName, "Missing"); got != "" {
		t.Fatalf("propertyValueFromMap missing=%q want empty", got)
	}
	if got := propertyValueFromMap(byName, "Value"); got != "" {
		t.Fatalf("propertyValueFromMap nil=%q want empty", got)
	}
}

func TestPropertyValueFromMap_TrimsWhitespace(t *testing.T) {
	v := "  padded value\t"
	byName := map[string]inspect.PropertyDTO{"Name": {Name: "Name", Value: &v}}
	if got := propertyValueFromMap(byName, "Name"); got != "padded value" {
		t.Fatalf("propertyValueFromMap trimmed=%q want=%q", got, "padded value")
	}
}

func TestFormatPropertyValue_ControlTypeUsesLocalizedLookup(t *testing.T) {
	localized := "button"
	byName := map[string]inspect.PropertyDTO{
		"LocalizedControlType": {Name: "LocalizedControlType", Value: &localized},
	}
	got := formatPropertyValue("ControlType", "50000", byName)
	if got != "50000 (button)" {
		t.Fatalf("formatPropertyValue ControlType=%q want=%q", got, "50000 (button)")
	}
}

func TestPropertyRowsPrimaryPathUsesDetailsProperties(t *testing.T) {
	name := "From Properties"
	rows := mapPropertyRowsFromDetails(inspect.GetNodeDetailsResponse{Properties: []inspect.PropertyDTO{{Name: "Name", Value: &name, Status: "ok"}}})
	if rows[2].Name != "Name" || rows[2].Value != "From Properties" || rows[2].Status != "ok" {
		t.Fatalf("expected Name row from primary property path, got %#v", rows[2])
	}
}

func TestPropertyRowsFallbackToElementWhenPropertiesEmpty(t *testing.T) {
	details := inspect.GetNodeDetailsResponse{
		Element: inspect.ElementPropertiesDTO{
			Name:                 "Fallback Name",
			Value:                "Fallback Value",
			ControlType:          "50004",
			LocalizedControlType: "edit",
			AutomationID:         "auto-id",
			LabeledBy:            "label-source",
			Bounds:               &inspect.Rect{Left: 1, Top: 2, Width: 3, Height: 4},
			PropertyStates: map[string]string{
				"Value":             "unsupported",
				"BoundingRectangle": "unsupported",
				"LabeledBy":         "ok",
			},
		},
		WindowInfo: inspect.WindowInfoDTO{Class: "Edit", PID: 314},
	}

	rows := mapPropertyRowsFromDetails(details)
	if rows[0].Value != "50004 (edit)" {
		t.Fatalf("expected ControlType fallback formatting, got %q", rows[0].Value)
	}
	if rows[2].Value != "Fallback Name" || rows[2].Status != "ok" {
		t.Fatalf("expected Name fallback row populated, got %#v", rows[2])
	}
	if rows[3].Value != "Fallback Value" || rows[3].Status != "unsupported" {
		t.Fatalf("expected Value fallback row populated, got %#v", rows[3])
	}
	if rows[5].Value != "x:1 y:2 w:3 h:4 | l:1 t:2 r:4 b:6" || rows[5].Status != "unsupported" {
		t.Fatalf("expected BoundingRectangle fallback formatting, got %#v", rows[5])
	}
	if rows[6].Value != "Edit" {
		t.Fatalf("expected ClassName fallback from WindowInfo, got %#v", rows[6])
	}
	if rows[13].Value != "314" {
		t.Fatalf("expected ProcessId fallback from WindowInfo, got %#v", rows[13])
	}
	if rows[20].Value != "label-source" || rows[20].Status != "ok" {
		t.Fatalf("expected LabeledBy row populated, got %#v", rows[20])
	}
}
