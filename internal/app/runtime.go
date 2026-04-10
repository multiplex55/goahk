package app

import (
	"context"
	"io"

	"goahk/internal/actions"
	"goahk/internal/config"
	"goahk/internal/hotkey"
	"goahk/internal/program"
	"goahk/internal/runtime"
)

type RuntimeDeps struct {
	Bootstrap       Bootstrap
	InitLogging     func(context.Context, config.LoggingConfig) error
	InitServices    func(context.Context, config.Config) (io.Closer, error)
	RegisterHotkeys func(context.Context, []config.HotkeyBinding) (io.Closer, error)
	RunMessageLoop  func(context.Context) error
}

type Runtime struct {
	deps RuntimeDeps
}

func NewRuntime(deps RuntimeDeps) *Runtime {
	if deps.Bootstrap.Load == nil {
		deps.Bootstrap = NewBootstrap()
	}
	return &Runtime{deps: deps}
}

type RuntimeBinding = runtime.RuntimeBinding

func CompileRuntimeBindings(cfg config.Config, registry *actions.Registry) ([]RuntimeBinding, error) {
	// Deprecated: compatibility adapter from config.Config to the canonical program compiler.
	// Prefer calling internal/runtime.CompileRuntimeBindings from program.Program directly.
	p, err := config.ToProgram(cfg)
	if err != nil {
		return nil, err
	}
	return CompileRuntimeBindingsFromProgram(p, registry)
}

func CompileRuntimeBindingsFromProgram(p program.Program, registry *actions.Registry) ([]RuntimeBinding, error) {
	// Deprecated: thin adapter retained for compatibility.
	// Prefer calling internal/runtime.CompileRuntimeBindings directly.
	return runtime.CompileRuntimeBindings(p, registry)
}

func DispatchHotkeyEvents(ctx context.Context, events <-chan hotkey.TriggerEvent, plans map[string]actions.Plan, executor *actions.Executor, base actions.ActionContext) <-chan actions.ExecutionResult {
	results := make(chan actions.ExecutionResult, 16)
	go func() {
		defer close(results)
		for {
			select {
			case <-ctx.Done():
				return
			case ev, ok := <-events:
				if !ok {
					return
				}
				plan, exists := plans[ev.BindingID]
				if !exists {
					continue
				}
				actionCtx := base
				actionCtx.Context = ctx
				actionCtx.BindingID = ev.BindingID
				actionCtx.TriggerText = ev.Chord.String()
				res := executor.Execute(actionCtx, plan)
				select {
				case results <- res:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return results
}
