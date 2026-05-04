//go:build windows

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"goahk/internal/inspect"
)

type viewerWindow interface {
	Run() error
}

var newViewerWindow = func(controller *Controller) (viewerWindow, error) {
	return NewViewerWindow(controller)
}

var bootstrapCheck = walkBootstrapCheck

var (
	startupLogOnce sync.Once
	startupLogger  *log.Logger
)

func logStartup(msg string) {
	startupLogOnce.Do(func() {
		startupLogger = log.New(openStartupLogFile(), "", log.LstdFlags|log.Lmicroseconds)
	})
	startupLogger.Println(msg)
}

func openStartupLogFile() *os.File {
	preferred := filepath.Join("dist", "goahk-uia-viewer", "goahk-uia-viewer.log")
	if f, err := createLogFile(preferred); err == nil {
		return f
	}
	exe, err := os.Executable()
	if err != nil {
		return os.Stderr
	}
	fallback := filepath.Join(filepath.Dir(exe), "goahk-uia-viewer.log")
	f, err := createLogFile(fallback)
	if err != nil {
		return os.Stderr
	}
	return f
}

func createLogFile(path string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	return os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
}

func runWindows() (err error) {
	logStartup("startup begin")
	logStartup("bootstrap check begin")
	if err := bootstrapCheck(); err != nil {
		return fmt.Errorf("bootstrap initialization failed: %w. Rebuild using build\\build-uia-viewer.bat", err)
	}
	logStartup("bootstrap check end")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	svc := inspect.NewService()
	logStartup("controller init")
	controller := NewController(ctx, svc)
	controller.SetClipboard(walkClipboard{})
	controller.SetDialogs(walkDialogs{})
	defer func() {
		logStartup("shutdown")
		controller.Shutdown()
	}()

	logStartup("window init begin")
	ui, err := newViewerWindow(controller)
	if err != nil {
		return fmt.Errorf("failed to create goahk-uia-viewer window: %w", err)
	}
	if ui == nil {
		return fmt.Errorf("failed to start goahk-uia-viewer: NewViewerWindow returned nil")
	}
	logStartup("window init end")

	logStartup("run begin")
	if err := ui.Run(); err != nil {
		return fmt.Errorf("goahk-uia-viewer event loop failed: %w", err)
	}
	logStartup("run returned")

	return nil
}

func main() {
	runtime.LockOSThread()
	if err := runWindows(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
