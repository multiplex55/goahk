//go:build windows

package main

import (
	"testing"

	"goahk/internal/inspect"
)

func TestInfoRowsPrimaryPathUsesWindowInfo(t *testing.T) {
	details := &inspect.GetNodeDetailsResponse{
		Element: inspect.ElementPropertiesDTO{
			Name:   "Element Name",
			Value:  "Element Value",
			HWND:   "0xbeef",
			Bounds: &inspect.Rect{Left: 1, Top: 2, Width: 3, Height: 4},
		},
		WindowInfo: inspect.WindowInfoDTO{
			Title:   "Window Title",
			Text:    "Window Text",
			HWND:    "0x123",
			Rect:    &inspect.Rect{Left: 10, Top: 20, Width: 30, Height: 40},
			Class:   "WindowClass",
			Process: "notepad.exe",
			PID:     1337,
		},
		Source: inspect.ProviderSourceDTO{Provider: "uia", Mode: inspect.InspectModeUIATree, Traversal: "raw-true-condition", Fallback: "none", NodeCount: 4, ChildCount: 2},
	}

	rows := mapWindowInfoRows(details)
	assertInfoRowValue(t, rows, "Title", "Window Title")
	assertInfoRowValue(t, rows, "Text", "Window Text")
	assertInfoRowValue(t, rows, "Hwnd", "ahk_id 0x123")
	assertInfoRowValue(t, rows, "Location", "x:10 y:20")
	assertInfoRowValue(t, rows, "Size", "w:30 h:40")
	assertInfoRowValue(t, rows, "Class(NN)", "WindowClass")
	assertInfoRowValue(t, rows, "Process", "notepad.exe")
	assertInfoRowValue(t, rows, "PID", "1337")
	assertInfoRowValue(t, rows, "Provider", "uia")
	assertInfoRowValue(t, rows, "Mode", "UIA_TREE")
	assertInfoRowValue(t, rows, "Traversal", "raw-true-condition")
	assertInfoRowValue(t, rows, "Fallback", "none")
	assertInfoRowValue(t, rows, "Node Count", "4")
	assertInfoRowValue(t, rows, "Children Count", "2")
}

func TestInfoRowsFallbackToElementWhenWindowInfoEmpty(t *testing.T) {
	className := "Button"
	processName := "calc.exe"
	processID := "99"
	details := &inspect.GetNodeDetailsResponse{
		Element: inspect.ElementPropertiesDTO{
			Name:   "Element Name",
			Value:  "Element Value",
			HWND:   "beef",
			Bounds: &inspect.Rect{Left: 3, Top: 4, Width: 5, Height: 6},
		},
		Properties: []inspect.PropertyDTO{
			{Name: "ClassName", Value: &className, Status: "ok"},
			{Name: "ProcessName", Value: &processName, Status: "ok"},
			{Name: "ProcessId", Value: &processID, Status: "ok"},
		},
		Source: inspect.ProviderSourceDTO{Provider: "acc", Mode: inspect.InspectModeWindowTree, Traversal: "control-view", Fallback: "active", NodeCount: 3, ChildCount: 1},
	}

	rows := mapWindowInfoRows(details)
	assertInfoRowValue(t, rows, "Title", "Element Name")
	assertInfoRowValue(t, rows, "Text", "Element Value")
	assertInfoRowValue(t, rows, "Hwnd", "ahk_id 0xbeef")
	assertInfoRowValue(t, rows, "Location", "x:3 y:4")
	assertInfoRowValue(t, rows, "Size", "w:5 h:6")
	assertInfoRowValue(t, rows, "Class(NN)", "Button")
	assertInfoRowValue(t, rows, "Process", "calc.exe")
	assertInfoRowValue(t, rows, "PID", "99")
	assertInfoRowValue(t, rows, "Provider", "acc")
	assertInfoRowValue(t, rows, "Mode", "WINDOW_TREE")
	assertInfoRowValue(t, rows, "Traversal", "control-view")
	assertInfoRowValue(t, rows, "Fallback", "active")
	assertInfoRowValue(t, rows, "Node Count", "3")
	assertInfoRowValue(t, rows, "Children Count", "1")
}

func TestPropertyValueFromList_ExactNameHit(t *testing.T) {
	v := "found"
	got := propertyValueFromList([]inspect.PropertyDTO{{Name: "ClassName", Value: &v}}, "ClassName")
	if got != "found" {
		t.Fatalf("propertyValueFromList exact hit=%q want=%q", got, "found")
	}
}

func TestPropertyValueFromList_TrimsWhitespace(t *testing.T) {
	v := "  padded  "
	got := propertyValueFromList([]inspect.PropertyDTO{{Name: "  ProcessName  ", Value: &v}}, "ProcessName")
	if got != "padded" {
		t.Fatalf("propertyValueFromList trim=%q want=%q", got, "padded")
	}
}

func TestPropertyValueFromList_MissingOrNilReturnsEmpty(t *testing.T) {
	v := "present"
	var nilPtr *string
	gotMissing := propertyValueFromList([]inspect.PropertyDTO{{Name: "Other", Value: &v}}, "ClassName")
	if gotMissing != "" {
		t.Fatalf("propertyValueFromList missing=%q want empty", gotMissing)
	}
	gotNil := propertyValueFromList([]inspect.PropertyDTO{{Name: "ClassName", Value: nilPtr}}, "ClassName")
	if gotNil != "" {
		t.Fatalf("propertyValueFromList nil=%q want empty", gotNil)
	}
}

func TestPropertyIntFromList_ExactNameHitAndTrim(t *testing.T) {
	v := "  42  "
	got := propertyIntFromList([]inspect.PropertyDTO{{Name: "  ProcessId", Value: &v}}, "ProcessId")
	if got != 42 {
		t.Fatalf("propertyIntFromList trimmed hit=%d want=%d", got, 42)
	}
}

func TestPropertyIntFromList_MissingNilInvalidOrNonPositiveReturnZero(t *testing.T) {
	v := "10"
	var nilPtr *string
	if got := propertyIntFromList([]inspect.PropertyDTO{{Name: "Other", Value: &v}}, "ProcessId"); got != 0 {
		t.Fatalf("propertyIntFromList missing=%d want=0", got)
	}
	if got := propertyIntFromList([]inspect.PropertyDTO{{Name: "ProcessId", Value: nilPtr}}, "ProcessId"); got != 0 {
		t.Fatalf("propertyIntFromList nil=%d want=0", got)
	}
	invalid := "abc"
	if got := propertyIntFromList([]inspect.PropertyDTO{{Name: "ProcessId", Value: &invalid}}, "ProcessId"); got != 0 {
		t.Fatalf("propertyIntFromList invalid=%d want=0", got)
	}
	zero := "0"
	if got := propertyIntFromList([]inspect.PropertyDTO{{Name: "ProcessId", Value: &zero}}, "ProcessId"); got != 0 {
		t.Fatalf("propertyIntFromList zero=%d want=0", got)
	}
	neg := "-7"
	if got := propertyIntFromList([]inspect.PropertyDTO{{Name: "ProcessId", Value: &neg}}, "ProcessId"); got != 0 {
		t.Fatalf("propertyIntFromList negative=%d want=0", got)
	}
}

func assertInfoRowValue(t *testing.T, rows []infoTableRow, property, want string) {
	t.Helper()
	for _, row := range rows {
		if row.Property != property {
			continue
		}
		if row.Value != want {
			t.Fatalf("%s value=%q want=%q", property, row.Value, want)
		}
		return
	}
	t.Fatalf("property %q not found in rows=%#v", property, rows)
}
