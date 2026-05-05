//go:build windows

package main

import (
	"testing"

	"goahk/internal/inspect"
)

func TestWindowInfoRowsPopulateFromDetails(t *testing.T) {
	details := &inspect.GetNodeDetailsResponse{
		WindowInfo: inspect.WindowInfoDTO{Title: "Calculator", Text: "calc", Class: "ApplicationFrameWindow", Process: "CalculatorApp.exe", PID: 7788},
		Element:    inspect.ElementPropertiesDTO{HWND: "0042", Bounds: &inspect.Rect{Left: 10, Top: 22, Width: 600, Height: 400}},
	}
	rows := mapWindowInfoRows(details)
	if len(rows) != 8 {
		t.Fatalf("expected 8 rows, got %d", len(rows))
	}
	if rows[0].Property != "Title" || rows[0].Value != "Calculator" {
		t.Fatalf("unexpected title row: %#v", rows[0])
	}
	if rows[6].Property != "Process" || rows[6].Value != "CalculatorApp.exe" {
		t.Fatalf("unexpected process row: %#v", rows[6])
	}
}

func TestWindowInfoFormatsHwndLocationSize(t *testing.T) {
	details := &inspect.GetNodeDetailsResponse{
		WindowInfo: inspect.WindowInfoDTO{HWND: "0xABC", Rect: &inspect.Rect{Left: 120, Top: 250, Width: 900, Height: 700}},
	}
	rows := mapWindowInfoRows(details)
	assertRowValue(t, rows, "Hwnd", "ahk_id 0xABC")
	assertRowValue(t, rows, "Location", "x:120 y:250")
	assertRowValue(t, rows, "Size", "w:900 h:700")
}

func assertRowValue(t *testing.T, rows []infoTableRow, property, want string) {
	t.Helper()
	for _, row := range rows {
		if row.Property == property {
			if row.Value != want {
				t.Fatalf("property %s value=%q want=%q", property, row.Value, want)
			}
			return
		}
	}
	t.Fatalf("property %s not found", property)
}
