package goahk

import (
	"context"
	"fmt"
	stdruntime "runtime"

	"goahk/internal/actions"
	internalapp "goahk/internal/app"
	"goahk/internal/config"
	"goahk/internal/runtime"
)

func (a *App) Run(ctx context.Context) error {
	var registry *actions.Registry
	if a.validateActions {
		registry = actions.NewRegistry()
	}

	if _, err := internalapp.CompileRuntimeBindingsFromProgram(a.toProgram(), registry); err != nil {
		return fmt.Errorf("compile app program: %w", err)
	}

	if stdruntime.GOOS != "windows" {
		return nil
	}

	bootstrap := runtime.NewBootstrap()
	cfg := a.toConfig()
	bootstrap.LoadConfig = func(context.Context, string) (config.Config, error) {
		return cfg, nil
	}
	return bootstrap.Run(ctx, "")
}

func (a *App) toConfig() config.Config {
	cfg := config.Config{Hotkeys: make([]config.HotkeyBinding, 0, len(a.bindings))}
	for i, b := range a.bindings {
		steps := make([]config.Step, 0, len(b.actions))
		for _, action := range b.actions {
			params := make(map[string]string, len(action.params))
			for k, v := range action.params {
				params[k] = v
			}
			steps = append(steps, config.Step{Action: action.name, Params: params})
		}
		cfg.Hotkeys = append(cfg.Hotkeys, config.HotkeyBinding{ID: bindingID(i), Hotkey: b.hotkey, Steps: steps})
	}
	return cfg
}

func (a *App) MustRun(ctx context.Context) {
	if err := a.Run(ctx); err != nil {
		panic(err)
	}
}
