package app

import (
	"context"
	"fmt"
	"io"

	"goahk/internal/actions"
	"goahk/internal/config"
	"goahk/internal/hotkey"
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

type RuntimeBinding struct {
	ID    string
	Chord hotkey.Chord
	Plan  actions.Plan
}

func CompileRuntimeBindings(cfg config.Config, registry *actions.Registry) ([]RuntimeBinding, error) {
	parsed := make([]hotkey.Binding, 0, len(cfg.Hotkeys))
	for _, b := range cfg.Hotkeys {
		binding, err := hotkey.ParseBinding(b.ID, b.Hotkey)
		if err != nil {
			return nil, err
		}
		parsed = append(parsed, binding)
	}
	if err := hotkey.DetectConflicts(parsed); err != nil {
		return nil, err
	}

	compiled := make([]RuntimeBinding, 0, len(cfg.Hotkeys))
	for i, b := range cfg.Hotkeys {
		plan := make(actions.Plan, 0, len(b.Steps))
		for _, step := range b.Steps {
			if registry != nil {
				if _, ok := registry.Lookup(step.Action); !ok {
					return nil, fmt.Errorf("binding %q references unknown action %q", b.ID, step.Action)
				}
			}
			plan = append(plan, actions.Step{Name: step.Action, Params: step.Params})
		}
		compiled = append(compiled, RuntimeBinding{ID: b.ID, Chord: parsed[i].Chord, Plan: plan})
	}
	return compiled, nil
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
