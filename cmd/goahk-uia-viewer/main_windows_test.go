//go:build windows

package main

import (
	"errors"
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
	t.Cleanup(func() { newViewerWindow = orig })

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
	t.Cleanup(func() { newViewerWindow = orig })

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
	t.Cleanup(func() { newViewerWindow = orig })

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
