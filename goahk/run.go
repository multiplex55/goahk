package goahk

import (
	"context"
	"fmt"

	"goahk/internal/actions"
	internalapp "goahk/internal/app"
)

func (a *App) Run(ctx context.Context) error {
	_ = ctx
	var registry *actions.Registry
	if a.validateActions {
		registry = actions.NewRegistry()
	}
	if _, err := internalapp.CompileRuntimeBindingsFromProgram(a.toProgram(), registry); err != nil {
		return fmt.Errorf("compile app program: %w", err)
	}
	return nil
}

func (a *App) MustRun(ctx context.Context) {
	if err := a.Run(ctx); err != nil {
		panic(err)
	}
}
