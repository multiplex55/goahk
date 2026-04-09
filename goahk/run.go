package goahk

import (
	"context"
	"fmt"
	stdruntime "runtime"

	"goahk/internal/actions"
	internalapp "goahk/internal/app"
	"goahk/internal/config"
	"goahk/internal/program"
	"goahk/internal/runtime"
)

const callbackActionPlaceholder = "goahk.callback"

type callbackRegistration struct {
	actionName string
	fn         ActionFunc
}

func (a *App) Run(ctx context.Context) error {
	p, _, callbacks := a.runtimeArtifacts()

	var registry *actions.Registry
	if a.validateActions || len(callbacks) > 0 {
		registry = buildRegistryWithCallbacks(a.state, callbacks)
	}

	if _, err := internalapp.CompileRuntimeBindingsFromProgram(p, registry); err != nil {
		return fmt.Errorf("compile app program: %w", err)
	}

	if stdruntime.GOOS != "windows" {
		return nil
	}

	bootstrap := runtime.NewBootstrap()
	bootstrap.LoadProgram = func(context.Context, string) (program.Program, error) {
		return p, nil
	}
	bootstrap.BuildRegistry = func(context.Context, program.Program) (*actions.Registry, error) {
		return buildRegistryWithCallbacks(a.state, callbacks), nil
	}
	return bootstrap.Run(ctx, "")
}

func (a *App) toConfig() config.Config {
	_, cfg, _ := a.runtimeArtifacts()
	return cfg
}

func buildRegistryWithCallbacks(state StateStore, callbacks []callbackRegistration) *actions.Registry {
	r := actions.NewRegistry()
	for _, cb := range callbacks {
		cb := cb
		r.MustRegister(cb.actionName, func(actionCtx actions.ActionContext, _ actions.Step) error {
			ctx := newContext(&actionCtx, state)
			err := cb.fn(ctx)
			syncVarsToActionContext(ctx)
			return err
		})
	}
	return r
}

func (a *App) runtimeArtifacts() (program.Program, config.Config, []callbackRegistration) {
	p := program.Program{Bindings: make([]program.BindingSpec, 0, len(a.bindings))}
	cfg := config.Config{Hotkeys: make([]config.HotkeyBinding, 0, len(a.bindings))}
	callbacks := make([]callbackRegistration, 0)

	for i, b := range a.bindings {
		programSteps := make([]program.StepSpec, 0, len(b.steps))
		configSteps := make([]config.Step, 0, len(b.steps))
		for stepIndex, step := range b.steps {
			spec := step.stepSpec()
			actionName := spec.Action
			if cbStep, ok := step.(callbackStep); ok {
				actionName = callbackActionName(i, stepIndex)
				callbacks = append(callbacks, callbackRegistration{actionName: actionName, fn: cbStep.fn})
			}
			spec.Action = actionName
			programSteps = append(programSteps, spec)
			configSteps = append(configSteps, config.Step{Action: actionName, Params: stringifyStepParams(spec.Params)})
		}
		id := bindingID(i)
		p.Bindings = append(p.Bindings, program.BindingSpec{ID: id, Hotkey: b.hotkey, Steps: programSteps})
		cfg.Hotkeys = append(cfg.Hotkeys, config.HotkeyBinding{ID: id, Hotkey: b.hotkey, Steps: configSteps})
	}
	return p, cfg, callbacks
}

func callbackActionName(bindingIndex, stepIndex int) string {
	return fmt.Sprintf("goahk.callback.%s.step_%d", bindingID(bindingIndex), stepIndex+1)
}

func stringifyStepParams(params map[string]any) map[string]string {
	if len(params) == 0 {
		return map[string]string{}
	}
	out := make(map[string]string, len(params))
	for key, raw := range params {
		if value, ok := raw.(string); ok {
			out[key] = value
		}
	}
	return out
}

func (a *App) MustRun(ctx context.Context) {
	if err := a.Run(ctx); err != nil {
		panic(err)
	}
}
