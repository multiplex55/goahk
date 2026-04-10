package goahk

import (
	"context"
	"errors"
	"fmt"
	stdruntime "runtime"
	"strings"

	"goahk/internal/actions"
	"goahk/internal/config"
	"goahk/internal/program"
	"goahk/internal/runtime"
)

const callbackActionPlaceholder = "goahk.callback"

var ErrUnsupportedPlatform = errors.New("goahk: runtime execution is unsupported on this platform")

type unsupportedPlatformError struct {
	goos string
}

func (e unsupportedPlatformError) Error() string {
	return fmt.Sprintf("%s (goos=%s, require=windows)", ErrUnsupportedPlatform.Error(), e.goos)
}

func (e unsupportedPlatformError) Unwrap() error {
	return ErrUnsupportedPlatform
}

type callbackRegistration struct {
	ref string
	fn  ActionFunc
}

// Run validates configured bindings, then starts the runtime loop.
func (a *App) Run(ctx context.Context) error {
	logger := a.actionLogger()
	if len(a.buildErrors) > 0 {
		logger.Error("goahk.invalid_binding_wiring", map[string]any{"error_count": len(a.buildErrors), "errors": strings.Join(a.buildErrors, "; ")})
		return fmt.Errorf("invalid binding wiring: %s", strings.Join(a.buildErrors, "; "))
	}

	p, _, callbacks := a.runtimeArtifacts()
	logger.Info("goahk.runtime_startup", map[string]any{"binding_count": len(p.Bindings), "callback_count": len(callbacks), "validate_actions": a.validateActions})

	var registry *actions.Registry
	if a.validateActions || len(callbacks) > 0 {
		registry = buildRegistryWithCallbacks(a.state, callbacks, logger)
	}

	if _, err := runtime.CompileRuntimeBindings(p, registry); err != nil {
		logger.Error("goahk.compile_runtime_bindings_failed", map[string]any{"error": err.Error()})
		return fmt.Errorf("compile app program: %w", err)
	}

	if stdruntime.GOOS != "windows" {
		logger.Info("goahk.runtime_not_started", map[string]any{"reason": "non_windows", "goos": stdruntime.GOOS})
		return unsupportedPlatformError{goos: stdruntime.GOOS}
	}

	bootstrap := runtime.NewBootstrap()
	bootstrap.BaseActionCtx.Logger = logger
	bootstrap.LoadProgram = func(context.Context, string) (program.Program, error) {
		return p, nil
	}
	bootstrap.BuildRegistry = func(context.Context, program.Program) (*actions.Registry, error) {
		return buildRegistryWithCallbacks(a.state, callbacks, logger), nil
	}
	bootstrap.LogDispatch = a.dispatchLogSink(logger)
	logger.Info("goahk.binding_registration_summary", map[string]any{"binding_count": len(p.Bindings)})
	return bootstrap.Run(ctx, "")
}

func (a *App) dispatchLogSink(logger actions.Logger) runtime.DispatchLogSink {
	return func(ctx context.Context, entry runtime.DispatchLogEntry) {
		fields := map[string]any{
			"event":         entry.Event,
			"binding_id":    entry.BindingID,
			"known_count":   entry.KnownCount,
			"actions":       entry.Actions,
			"duration":      entry.Duration.String(),
			"error":         entry.Error,
			"timestamp":     entry.Timestamp,
			"failed_action": entry.FailedAction,
		}
		if entry.Error != "" {
			logger.Error("goahk.dispatch", fields)
			return
		}
		logger.Info("goahk.dispatch", fields)
	}
}

func (a *App) toConfig() config.Config {
	_, cfg, _ := a.runtimeArtifacts()
	return cfg
}

func buildRegistryWithCallbacks(state StateStore, callbacks []callbackRegistration, logger actions.Logger) *actions.Registry {
	r := actions.NewRegistry()
	for _, cb := range callbacks {
		cb := cb
		r.MustRegisterCallback(cb.ref, func(callbackCtx actions.CallbackContext) error {
			callbackLogger := callbackCtx.Log()
			if callbackLogger == nil {
				callbackLogger = logger
			}
			ctx := newContext(&actions.ActionContext{
				Context:     callbackCtx.Context(),
				Services:    callbackCtx.Window(),
				Metadata:    callbackCtx.StateBag(),
				BindingID:   callbackCtx.BindingID(),
				TriggerText: callbackCtx.TriggerText(),
				Stop:        func(reason string) { callbackCtx.StopRuntime(reason) },
				Logger:      callbackLogger,
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
		p.Bindings = append(p.Bindings, program.BindingSpec{
			ID:                id,
			Hotkey:            b.hotkey,
			Steps:             programSteps,
			ConcurrencyPolicy: b.concurrencyPolicy,
		})
		cfg.Hotkeys = append(cfg.Hotkeys, config.HotkeyBinding{
			ID:                id,
			Hotkey:            b.hotkey,
			Steps:             configSteps,
			ConcurrencyPolicy: string(b.concurrencyPolicy),
		})
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
