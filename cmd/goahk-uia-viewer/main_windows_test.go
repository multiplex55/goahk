//go:build windows

package main

import (
	"errors"
	"os"
	"strings"
	"testing"
)

type stubViewerWindow struct {
	runErr error
}

func (w stubViewerWindow) Run() error {
	return w.runErr
}

func TestRunWindows_SymbolAvailable(t *testing.T) {
	t.Parallel()
	_ = runWindows
}

func TestRunWindows_NewViewerWindowErrorWrapped(t *testing.T) {
	t.Parallel()

	orig := newViewerWindow
	origBootstrap := bootstrapCheck
	t.Cleanup(func() { newViewerWindow = orig })
	t.Cleanup(func() { bootstrapCheck = origBootstrap })
	bootstrapCheck = func() error { return nil }

	want := errors.New("build failed")
	newViewerWindow = func(*Controller) (viewerWindow, error) {
		return nil, want
	}

	err := runWindows()
	if !errors.Is(err, want) {
		t.Fatalf("runWindows error = %v, want wrapped %v", err, want)
	}
	if !strings.Contains(err.Error(), "failed to create goahk-uia-viewer window") {
		t.Fatalf("runWindows error = %v, want startup context", err)
	}
}

func TestRunWindows_NilViewerWindow(t *testing.T) {
	t.Parallel()

	orig := newViewerWindow
	origBootstrap := bootstrapCheck
	t.Cleanup(func() { newViewerWindow = orig })
	t.Cleanup(func() { bootstrapCheck = origBootstrap })
	bootstrapCheck = func() error { return nil }

	newViewerWindow = func(*Controller) (viewerWindow, error) {
		return nil, nil
	}

	err := runWindows()
	if err == nil {
		t.Fatal("runWindows error = nil, want error")
	}
	if !strings.Contains(err.Error(), "NewViewerWindow returned nil") {
		t.Fatalf("runWindows error = %v, want nil-window message", err)
	}
}

func TestRunWindows_RunErrorWrapped(t *testing.T) {
	t.Parallel()

	orig := newViewerWindow
	origBootstrap := bootstrapCheck
	t.Cleanup(func() { newViewerWindow = orig })
	t.Cleanup(func() { bootstrapCheck = origBootstrap })
	bootstrapCheck = func() error { return nil }

	want := errors.New("event loop failed")
	newViewerWindow = func(*Controller) (viewerWindow, error) {
		return stubViewerWindow{runErr: want}, nil
	}

	err := runWindows()
	if !errors.Is(err, want) {
		t.Fatalf("runWindows error = %v, want wrapped %v", err, want)
	}
	if !strings.Contains(err.Error(), "goahk-uia-viewer event loop failed") {
		t.Fatalf("runWindows error = %v, want event-loop context", err)
	}
}

func TestRunWindows_BootstrapFailureActionable(t *testing.T) {
	t.Parallel()

	origBootstrap := bootstrapCheck
	t.Cleanup(func() { bootstrapCheck = origBootstrap })

	bootstrapCheck = func() error { return errors.New("walk bootstrap failed") }

	err := runWindows()
	if err == nil {
		t.Fatal("runWindows error = nil, want error")
	}
	if !strings.Contains(err.Error(), "build\\build-uia-viewer.bat") {
		t.Fatalf("runWindows error = %v, want rebuild instruction", err)
	}
}

func TestRunWindows_BootstrapAlwaysRuns(t *testing.T) {
	t.Parallel()

	origBootstrap := bootstrapCheck
	origEnv := os.Getenv("GOAHK_UIA_VIEWER_BOOTSTRAP_CHECK")
	t.Cleanup(func() {
		bootstrapCheck = origBootstrap
		_ = os.Setenv("GOAHK_UIA_VIEWER_BOOTSTRAP_CHECK", origEnv)
	})

	if err := os.Setenv("GOAHK_UIA_VIEWER_BOOTSTRAP_CHECK", "0"); err != nil {
		t.Fatalf("Setenv failed: %v", err)
	}

	called := false
	bootstrapCheck = func() error {
		called = true
		return errors.New("stop after bootstrap")
	}

	_ = runWindows()
	if !called {
		t.Fatal("bootstrapCheck was not called")
	}
}
