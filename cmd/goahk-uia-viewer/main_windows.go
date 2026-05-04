//go:build windows

package main

import (
	"context"
	"fmt"
	"os"

	"goahk/internal/inspect"
)

type viewerWindow interface {
	Run() error
}

var newViewerWindow = func(controller *Controller) (viewerWindow, error) {
	return NewViewerWindow(controller)
}

func runWindows() (err error) {
	if os.Getenv("GOAHK_UIA_VIEWER_BOOTSTRAP_CHECK") == "1" {
		if err := walkBootstrapCheck(); err != nil {
			return fmt.Errorf("%w (try rebuilding with build/build-uia-viewer.bat to embed manifest resources)", err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	svc := inspect.NewService()
	controller := NewController(ctx, svc)
	controller.SetClipboard(walkClipboard{})
	controller.SetDialogs(walkDialogs{})
	defer controller.Shutdown()

	ui, err := newViewerWindow(controller)
	if err != nil {
		return fmt.Errorf("failed to create goahk-uia-viewer window: %w", err)
	}
	if ui == nil {
		return fmt.Errorf("failed to start goahk-uia-viewer: NewViewerWindow returned nil")
	}

	if err := ui.Run(); err != nil {
		return fmt.Errorf("goahk-uia-viewer event loop failed: %w", err)
	}

	return nil
}

func main() {
	if err := runWindows(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
