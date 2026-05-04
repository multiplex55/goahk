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
	return nil, fmt.Errorf("failed to start goahk-uia-viewer: NewViewerWindow is not wired")
}

func runWindows() (err error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	svc := inspect.NewService()
	controller := NewController(ctx, svc)
	controller.SetClipboard(walkClipboard{})
	controller.SetDialogs(walkDialogs{})
	defer controller.Shutdown()

	ui, err := newViewerWindow(controller)
	if err != nil {
		return err
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
