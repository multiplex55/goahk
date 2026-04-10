package main

import (
	"embed"
	"testing"

	"goahk/internal/inspect"
)

func TestBuildWailsOptions_IncludesExpectedRuntimeConfiguration(t *testing.T) {
	app := NewViewerApp(&fakeInspectService{})
	var assets embed.FS

	opts := buildWailsOptions(app, assets)
	if opts == nil {
		t.Fatalf("expected options to be assembled")
	}
	if opts.Title != "GoAHK UIA Viewer" {
		t.Fatalf("unexpected title %q", opts.Title)
	}
	if opts.Width != 1280 || opts.Height != 800 {
		t.Fatalf("unexpected window size %dx%d", opts.Width, opts.Height)
	}
	if opts.MinWidth != 960 || opts.MinHeight != 640 {
		t.Fatalf("unexpected min bounds %dx%d", opts.MinWidth, opts.MinHeight)
	}
	if opts.AssetServer == nil {
		t.Fatalf("expected AssetServer to be configured")
	}
	if opts.OnStartup == nil || opts.OnShutdown == nil {
		t.Fatalf("expected lifecycle hooks to be configured")
	}
	if len(opts.Bind) != 1 {
		t.Fatalf("expected exactly one binding, got %d", len(opts.Bind))
	}
	if got, ok := opts.Bind[0].(*ViewerApp); !ok || got != app {
		t.Fatalf("expected bound object to be the viewer app")
	}
}

func TestNewViewerApp_ConstructsWithInspectService(t *testing.T) {
	app := newViewerApp()
	if app == nil {
		t.Fatalf("expected app to be constructed")
	}
	if app.service == nil {
		t.Fatalf("expected inspect service dependency")
	}
	if _, ok := app.service.(inspect.Service); !ok {
		t.Fatalf("expected app service to implement inspect.Service")
	}
}
