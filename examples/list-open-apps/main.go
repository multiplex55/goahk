package main

import (
	"context"
	"fmt"
	"log"
	"runtime"

	"goahk/internal/actions"
	"goahk/internal/window"
)

func main() {
	if runtime.GOOS != "windows" {
		log.Println("list-open-apps example is Windows-only")
		return
	}

	provider := window.NewOSProvider()
	registry := actions.NewRegistry()
	executor := actions.NewExecutor(registry)
	ctx := actions.ActionContext{
		Context:  context.Background(),
		Metadata: map[string]string{},
		Services: actions.Services{
			WindowList: func(ctx context.Context) ([]window.Info, error) {
				return window.Enumerate(ctx, provider)
			},
		},
	}

	plan := actions.Plan{{
		Name:   "window.list_open_applications",
		Params: map[string]string{"save_as": "open_apps", "include_background": "true", "dedupe_by": "window"},
	}}

	result := executor.Execute(ctx, plan)
	if !result.Success {
		log.Fatalf("action failed: %s", result.Steps[0].Error)
	}

	fmt.Println("Open applications JSON:")
	fmt.Println(ctx.Metadata["open_apps"])
}
