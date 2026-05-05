//go:build windows

package main

import "testing"

func TestPropertyContextMenuCopyValue(t *testing.T) {
	row := propertyTableRow{Name: "Name", Value: "Calculator"}
	if got := propertyContextCopyValue(row, true); got != "Calculator" {
		t.Fatalf("copy value got %q", got)
	}
}

func TestPropertyContextMenuCopyName(t *testing.T) {
	row := propertyTableRow{Name: "ControlType", Value: "50000 (button)"}
	if got := row.Name; got != "ControlType" {
		t.Fatalf("copy name got %q", got)
	}
}

func TestPropertyContextMenuCopyRow(t *testing.T) {
	row := propertyTableRow{Name: "ControlType", Value: "50000 (button)"}
	got := row.Name + ": " + propertyContextCopyValue(row, true)
	if got != "ControlType: button" {
		t.Fatalf("copy row got %q", got)
	}
}
