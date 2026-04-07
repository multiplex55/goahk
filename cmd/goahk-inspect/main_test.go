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
		window: windowProviderFunc{
			active: func(context.Context) (window.Info, error) {
				called = true
				return window.Info{HWND: 0x1, Title: "Editor", Active: true}, nil
			},
			list: func(context.Context) ([]window.Info, error) { return nil, nil },
		},
		uia: uiaProviderFunc{
			focused: func(context.Context) (uia.Element, error) { return uia.Element{}, nil },
			under:   func(context.Context) (uia.Element, error) { return uia.Element{}, nil },
			tree:    func(context.Context, int) (*uia.Node, error) { return nil, nil },
		},
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
		window: windowProviderFunc{
			active: func(context.Context) (window.Info, error) {
				return window.Info{HWND: 0x2, Title: "Term"}, nil
			},
			list: func(context.Context) ([]window.Info, error) { return nil, nil },
		},
		uia: uiaProviderFunc{
			focused: func(context.Context) (uia.Element, error) { return uia.Element{}, nil },
			under:   func(context.Context) (uia.Element, error) { return uia.Element{}, nil },
			tree:    func(context.Context, int) (*uia.Node, error) { return nil, nil },
		},
	}
	var out bytes.Buffer
	if err := run(context.Background(), []string{"--format", "json", "window", "active"}, &out, &bytes.Buffer{}, d); err != nil {
		t.Fatalf("run() error = %v", err)
	}
	if !strings.Contains(out.String(), "\"Title\": \"Term\"") {
		t.Fatalf("expected json output, got %q", out.String())
	}
}

func TestParseGlobal_TableDriven(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantFmt string
		want    []string
		wantErr string
	}{
		{name: "default", args: []string{"window", "active"}, wantFmt: "text", want: []string{"window", "active"}},
		{name: "format flag", args: []string{"--format", "json", "window", "active"}, wantFmt: "json", want: []string{"window", "active"}},
		{name: "format equals", args: []string{"--format=text", "uia", "focused"}, wantFmt: "text", want: []string{"uia", "focused"}},
		{name: "missing value", args: []string{"--format"}, wantErr: "requires a value"},
		{name: "unsupported", args: []string{"--format=yaml", "window", "list"}, wantErr: "unsupported format"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			format := "text"
			got, err := parseGlobal(tt.args, &format)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected err containing %q, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseGlobal err: %v", err)
			}
			if format != tt.wantFmt {
				t.Fatalf("format=%q want %q", format, tt.wantFmt)
			}
			if strings.Join(got, ",") != strings.Join(tt.want, ",") {
				t.Fatalf("args=%v want %v", got, tt.want)
			}
		})
	}
}

type windowProviderFunc struct {
	active func(context.Context) (window.Info, error)
	list   func(context.Context) ([]window.Info, error)
}

func (f windowProviderFunc) Active(ctx context.Context) (window.Info, error) { return f.active(ctx) }
func (f windowProviderFunc) List(ctx context.Context) ([]window.Info, error) {
	if f.list == nil {
		return nil, nil
	}
	return f.list(ctx)
}

type uiaProviderFunc struct {
	focused func(context.Context) (uia.Element, error)
	under   func(context.Context) (uia.Element, error)
	tree    func(context.Context, int) (*uia.Node, error)
}

func (f uiaProviderFunc) Focused(ctx context.Context) (uia.Element, error)     { return f.focused(ctx) }
func (f uiaProviderFunc) UnderCursor(ctx context.Context) (uia.Element, error) { return f.under(ctx) }
func (f uiaProviderFunc) ActiveWindowTree(ctx context.Context, depth int) (*uia.Node, error) {
	return f.tree(ctx, depth)
}
