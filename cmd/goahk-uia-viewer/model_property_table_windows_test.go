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
