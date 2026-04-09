package goahk

import (
	"context"
	"fmt"
	stdruntime "runtime"
	"strings"

	"goahk/internal/actions"
	internalapp "goahk/internal/app"
	"goahk/internal/config"
	"goahk/internal/program"
	"goahk/internal/runtime"
)

const callbackActionPlaceholder = "goahk.callback"

type callbackRegistration struct {
	ref string
	fn  ActionFunc
}

// Run validates configured bindings, then starts the runtime loop.
func (a *App) Run(ctx context.Context) error {
	if len(a.buildErrors) > 0 {
		return fmt.Errorf("invalid binding wiring: %s", strings.Join(a.buildErrors, "; "))
	}

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
		r.MustRegisterCallback(cb.ref, func(callbackCtx actions.CallbackContext) error {
			ctx := newContext(&actions.ActionContext{
				Context:     callbackCtx.Context(),
				Services:    callbackCtx.Window(),
				Metadata:    callbackCtx.StateBag(),
				BindingID:   callbackCtx.BindingID(),
				TriggerText: callbackCtx.TriggerText(),
				Stop:        func(reason string) { callbackCtx.StopRuntime(reason) },
				Logger:      callbackCtx.Log(),
			}, state)
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
			params := stringifyStepParams(spec.Params)
			if cbStep, ok := step.(callbackStep); ok {
				ref := callbackActionRef(i, stepIndex)
				callbacks = append(callbacks, callbackRegistration{ref: ref, fn: cbStep.fn})
				spec.Action = callbackActionPlaceholder
				if params == nil {
					params = map[string]string{}
				}
				params["callback_ref"] = ref
			}
			programSteps = append(programSteps, program.StepSpec{Action: spec.Action, Params: mapFromString(params)})
			configSteps = append(configSteps, config.Step{Action: spec.Action, Params: params})
		}
		id := bindingID(i)
		p.Bindings = append(p.Bindings, program.BindingSpec{ID: id, Hotkey: b.hotkey, Steps: programSteps})
		cfg.Hotkeys = append(cfg.Hotkeys, config.HotkeyBinding{ID: id, Hotkey: b.hotkey, Steps: configSteps})
	}
	return p, cfg, callbacks
}

func mapFromString(params map[string]string) map[string]any {
	if len(params) == 0 {
		return map[string]any{}
	}
	out := make(map[string]any, len(params))
	for k, v := range params {
		out[k] = v
	}
	return out
}

func callbackActionRef(bindingIndex, stepIndex int) string {
	return fmt.Sprintf("%s.step_%d", bindingID(bindingIndex), stepIndex+1)
}

func callbackActionName(bindingIndex, stepIndex int) string {
	return fmt.Sprintf("%s.%s", actions.CallbackActionName, callbackActionRef(bindingIndex, stepIndex))
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

// MustRun executes Run and panics if Run returns an error.
func (a *App) MustRun(ctx context.Context) {
	if err := a.Run(ctx); err != nil {
		panic(err)
	}
}
