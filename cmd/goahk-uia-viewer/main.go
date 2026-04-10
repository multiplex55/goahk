package main

import (
	"embed"
	"fmt"

	"goahk/internal/inspect"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var viewerFrontendAssets embed.FS

func newViewerApp() *ViewerApp {
	return NewViewerApp(inspect.NewService())
}

func buildWailsOptions(app *ViewerApp, assets embed.FS) *options.App {
	return &options.App{
		Title:       "GoAHK UIA Viewer",
		Width:       1280,
		Height:      800,
		MinWidth:    960,
		MinHeight:   640,
		AssetServer: &assetserver.Options{Assets: assets},
		OnStartup:   app.OnStartup,
		OnShutdown:  app.OnShutdown,
		Bind:        []any{app},
	}
}

func runViewer(app *ViewerApp, assets embed.FS) error {
	return wails.Run(buildWailsOptions(app, assets))
}

func main() {
	if err := runViewer(newViewerApp(), viewerFrontendAssets); err != nil {
		panic(fmt.Errorf("failed to run goahk-uia-viewer: %w", err))
	}
}
