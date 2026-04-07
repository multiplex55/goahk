package main

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"goahk/internal/uia"
	"goahk/internal/window"
)

func TestRun_WindowActiveRoutes(t *testing.T) {
	called := false
	d := deps{
		windowActive: func(context.Context) (window.Info, error) {
			called = true
			return window.Info{HWND: 0x1, Title: "Editor", Active: true}, nil
		},
		windowList: func(context.Context) ([]window.Info, error) { return nil, nil },
		uiaFocused: func(context.Context) (uia.Element, error) { return uia.Element{}, nil },
		uiaUnder:   func(context.Context) (uia.Element, error) { return uia.Element{}, nil },
		uiaTree:    func(context.Context, int) (*uia.Node, error) { return nil, nil },
	}
	var out bytes.Buffer
	if err := run(context.Background(), []string{"window", "active"}, &out, &bytes.Buffer{}, d); err != nil {
		t.Fatalf("run() error = %v", err)
	}
	if !called {
		t.Fatal("expected window active route to be called")
	}
	if !strings.Contains(out.String(), "HWND: 0x1") {
		t.Fatalf("unexpected output: %q", out.String())
	}
}

func TestRun_UIATreeRequiresFlag(t *testing.T) {
	d := deps{}
	err := run(context.Background(), []string{"uia", "tree"}, &bytes.Buffer{}, &bytes.Buffer{}, d)
	if err == nil || !strings.Contains(err.Error(), "requires --active-window") {
		t.Fatalf("expected active-window validation error, got %v", err)
	}
}

func TestRun_ParsesJSONFormat(t *testing.T) {
	d := deps{
		windowActive: func(context.Context) (window.Info, error) {
			return window.Info{HWND: 0x2, Title: "Term"}, nil
		},
		windowList: func(context.Context) ([]window.Info, error) { return nil, nil },
		uiaFocused: func(context.Context) (uia.Element, error) { return uia.Element{}, nil },
		uiaUnder:   func(context.Context) (uia.Element, error) { return uia.Element{}, nil },
		uiaTree:    func(context.Context, int) (*uia.Node, error) { return nil, nil },
	}
	var out bytes.Buffer
	if err := run(context.Background(), []string{"--format", "json", "window", "active"}, &out, &bytes.Buffer{}, d); err != nil {
		t.Fatalf("run() error = %v", err)
	}
	if !strings.Contains(out.String(), "\"Title\": \"Term\"") {
		t.Fatalf("expected json output, got %q", out.String())
	}
}
