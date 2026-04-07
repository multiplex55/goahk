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
	d := deps{
		window: windowProviderFunc{
			active: func(context.Context) (window.Info, error) {
				return window.Info{HWND: 0x2, Title: "Term", Class: "ConsoleWindowClass", PID: 42, Exe: "WindowsTerminal.exe", Active: true}, nil
			},
		},
		uia: uiaProviderFunc{
			focused: func(context.Context) (uia.Element, error) {
				name := "Search"
				return uia.Element{ID: "root", Name: &name}, nil
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
		{name: "uia focused text", args: []string{"uia", "focused"}, path: "testdata/golden/inspect/uia_focused_text.txt"},
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
