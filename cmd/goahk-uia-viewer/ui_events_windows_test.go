//go:build windows

package main

import (
	"strings"
	"testing"

	"goahk/internal/inspect"
)

func TestFormatSelectedInfo_MissingFieldsGraceful(t *testing.T) {
	got := formatSelectedInfo(&inspect.GetNodeDetailsResponse{})
	for _, want := range []string{"Title: N/A", "Process: N/A", "PID: N/A", "HWND: N/A", "Class: N/A", "NodeID: N/A"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in output: %s", want, got)
		}
	}
}

func TestFormatSelectedInfo_WithValues(t *testing.T) {
	got := formatSelectedInfo(&inspect.GetNodeDetailsResponse{
		WindowInfo: inspect.WindowInfoDTO{Title: "Notepad", Process: "notepad.exe", PID: 10, HWND: "0x1", Class: "Notepad"},
		Element:    inspect.ElementPropertiesDTO{NodeID: "node-1", Name: "OK", ControlType: "Button", LocalizedControlType: "button"},
		ACCPath:    "Window[1]/Button[1]",
	})
	for _, want := range []string{"Notepad", "notepad.exe", "PID: 10", "NodeID: node-1", "ACC Path: Window[1]/Button[1]"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in output: %s", want, got)
		}
	}
}
