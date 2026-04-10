package main

import (
	"bytes"
	"context"
	"testing"

	"goahk/internal/testutil"
	"goahk/internal/uia"
	"goahk/internal/window"
)

func TestRun_OutputFormatsGolden(t *testing.T) {
	button := "Button"
	name := "Search"
	d := deps{
		window: windowProviderFunc{
			active: func(context.Context) (window.Info, error) {
				return window.Info{HWND: 0x2, Title: "Term", Class: "ConsoleWindowClass", PID: 42, Exe: "WindowsTerminal.exe", Active: true}, nil
			},
			list: func(context.Context) ([]window.Info, error) {
				return []window.Info{
					{HWND: 0x2, Title: "Term", Class: "ConsoleWindowClass", PID: 42, Exe: "WindowsTerminal.exe", Active: true},
					{HWND: 0x3, Title: "Editor", Class: "Notepad", PID: 73, Exe: "notepad.exe", Active: false},
				}, nil
			},
		},
		uia: uiaProviderFunc{
			focused: func(context.Context) (uia.Element, error) {
				return uia.Element{ID: "root", Name: &name}, nil
			},
			under: func(context.Context) (uia.Element, error) {
				return uia.Element{ID: "btn-1", Name: &name, ControlType: &button, Patterns: []string{"Invoke"}}, nil
			},
			tree: func(context.Context, int) (*uia.Node, error) {
				return &uia.Node{Element: uia.Element{ID: "root", Name: &name}, Children: []*uia.Node{{Element: uia.Element{ID: "btn-1", Name: &name, ControlType: &button}}}}, nil
			},
		},
	}

	cases := []struct {
		name string
		args []string
		path string
	}{
		{name: "window text", args: []string{"window", "active"}, path: "testdata/golden/inspect/window_active_text.txt"},
		{name: "window json", args: []string{"--format", "json", "window", "active"}, path: "testdata/golden/inspect/window_active_json.txt"},
		{name: "window list text", args: []string{"window", "list"}, path: "testdata/golden/inspect/window_list_text.txt"},
		{name: "uia focused text", args: []string{"uia", "focused"}, path: "testdata/golden/inspect/uia_focused_text.txt"},
		{name: "uia under cursor text", args: []string{"uia", "under-cursor"}, path: "testdata/golden/inspect/uia_under_cursor_text.txt"},
		{name: "uia tree text", args: []string{"uia", "tree", "--active-window", "--depth", "3"}, path: "testdata/golden/inspect/uia_tree_text.txt"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var out bytes.Buffer
			if err := run(context.Background(), tc.args, &out, &bytes.Buffer{}, d); err != nil {
				t.Fatalf("run err: %v", err)
			}
			testutil.AssertGolden(t, tc.path, out.String())
		})
	}
}
