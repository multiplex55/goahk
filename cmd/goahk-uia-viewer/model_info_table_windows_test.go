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
