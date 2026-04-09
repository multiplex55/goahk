package app

import (
	"context"
	"fmt"
	"io"
	"strings"

	"goahk/internal/actions"
	"goahk/internal/config"
	"goahk/internal/flow"
	"goahk/internal/hotkey"
	"goahk/internal/program"
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
	ID             string
	Chord          hotkey.Chord
	Plan           actions.Plan
	Flow           *flow.Definition
	ControlCommand string
	Policy         program.ConcurrencyPolicy
}

func CompileRuntimeBindings(cfg config.Config, registry *actions.Registry) ([]RuntimeBinding, error) {
	// Deprecated: prefer compiling from program.Program directly via
	// CompileRuntimeBindingsFromProgram or internal/runtime.CompileRuntimeBindings.
	p, err := config.ToProgram(cfg)
	if err != nil {
		return nil, err
	}
	return CompileRuntimeBindingsFromProgram(p, registry)
}

func CompileRuntimeBindingsFromProgram(p program.Program, registry *actions.Registry) ([]RuntimeBinding, error) {
	p = program.Normalize(p)
	if err := program.Validate(p); err != nil {
		return nil, err
	}

	parsed := make([]hotkey.Binding, 0, len(p.Bindings))
	for _, b := range p.Bindings {
		binding, err := hotkey.ParseBinding(b.ID, b.Hotkey)
		if err != nil {
			return nil, err
		}
		parsed = append(parsed, binding)
	}

	selectors := convertUIASelectors(p.Options.UIASelectors)
	flowsByID := map[string]flow.Definition{}
	for _, f := range p.Options.Flows {
		compiled, err := compileFlowDefinition(f, registry)
		if err != nil {
			return nil, err
		}
		flowsByID[strings.ToLower(strings.TrimSpace(f.ID))] = compiled
	}

	compiled := make([]RuntimeBinding, 0, len(p.Bindings))
	for i, b := range p.Bindings {
		rb := RuntimeBinding{ID: b.ID, Chord: parsed[i].Chord, Policy: b.ConcurrencyPolicy}
		switch {
		case strings.EqualFold(parsed[i].Chord.Key, "escape") && parsed[i].Chord.Modifiers == 0:
			rb.ControlCommand = "stop"
		case strings.EqualFold(parsed[i].Chord.Key, "escape") && parsed[i].Chord.Modifiers == hotkey.ModShift:
			rb.ControlCommand = "hard_stop"
		}
		if ref := strings.ToLower(strings.TrimSpace(b.Flow)); ref != "" {
			f, ok := flowsByID[ref]
			if !ok {
				return nil, fmt.Errorf("binding %q references unknown flow %q", b.ID, b.Flow)
			}
			rb.Flow = &f
			compiled = append(compiled, rb)
			continue
		}
		if err := validateBindingActions(b, registry); err != nil {
			return nil, err
		}
		plan := make(actions.Plan, 0, len(b.Steps))
		for stepIdx, step := range b.Steps {
			params, err := stringifyParams(b.ID, stepIdx, step.Params)
			if err != nil {
				return nil, err
			}
			if strings.HasPrefix(step.Action, "uia.") {
				sel, err := config.ParseUIASelector(params, selectors)
				if err != nil {
					return nil, fmt.Errorf("binding %q action %q selector: %w", b.ID, step.Action, err)
				}
				raw, err := config.EncodeSelectorJSON(sel)
				if err != nil {
					return nil, fmt.Errorf("binding %q action %q selector encode: %w", b.ID, step.Action, err)
				}
				params["selector_json"] = raw
			}
			plan = append(plan, actions.Step{Name: step.Action, Params: params})
		}
		rb.Plan = plan
		compiled = append(compiled, rb)
	}
	return compiled, nil
}

func validateBindingActions(binding program.BindingSpec, registry *actions.Registry) error {
	if registry == nil {
		return nil
	}
	for idx, step := range binding.Steps {
		if strings.EqualFold(step.Action, actions.CallbackActionName) {
			if _, ok := step.Params["callback_ref"]; ok {
				continue
			}
		}
		if _, ok := registry.Lookup(step.Action); ok {
			continue
		}
		return fmt.Errorf("binding %q binding/actions[%d]/name: unknown action %q", binding.ID, idx, step.Action)
	}
	return nil
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

func compileFlowDefinition(in program.FlowSpec, registry *actions.Registry) (flow.Definition, error) {
	out := flow.Definition{ID: in.ID, Timeout: in.Timeout, Steps: make([]flow.Step, 0, len(in.Steps))}
	for _, s := range in.Steps {
		step, err := compileFlowStep(s, registry)
		if err != nil {
			return flow.Definition{}, err
		}
		out.Steps = append(out.Steps, step)
	}
	return out, nil
}

func compileFlowStep(in program.FlowStepSpec, registry *actions.Registry) (flow.Step, error) {
	params, err := stringifyArbitraryParams(in.Params)
	if err != nil {
		return flow.Step{}, err
	}
	step := flow.Step{Name: in.Name, Action: in.Action, Params: params, Timeout: in.Timeout}
	if step.Action != "" && registry != nil {
		if _, ok := registry.Lookup(step.Action); !ok {
			return flow.Step{}, fmt.Errorf("flow step references unknown action %q", step.Action)
		}
	}
	if in.If != nil {
		cond := flow.Condition{WindowMatches: nil, ElementExists: nil}
		if in.If.WindowMatches != nil {
			cond.WindowMatches = &flow.WindowCondition{Matcher: *in.If.WindowMatches}
		}
		if in.If.ElementExists != nil {
			cond.ElementExists = &flow.ElementCondition{Selector: *in.If.ElementExists}
		}
		ifBlock := &flow.IfBlock{Condition: cond}
		for _, nested := range in.If.Then {
			n, err := compileFlowStep(nested, registry)
			if err != nil {
				return flow.Step{}, err
			}
			ifBlock.Then = append(ifBlock.Then, n)
		}
		for _, nested := range in.If.Else {
			n, err := compileFlowStep(nested, registry)
			if err != nil {
				return flow.Step{}, err
			}
			ifBlock.Else = append(ifBlock.Else, n)
		}
		step.If = ifBlock
	}
	if in.WaitUntil != nil {
		cond := flow.Condition{WindowMatches: nil, ElementExists: nil}
		if in.WaitUntil.WindowMatches != nil {
			cond.WindowMatches = &flow.WindowCondition{Matcher: *in.WaitUntil.WindowMatches}
		}
		if in.WaitUntil.ElementExists != nil {
			cond.ElementExists = &flow.ElementCondition{Selector: *in.WaitUntil.ElementExists}
		}
		step.WaitUntil = &flow.WaitUntilBlock{Condition: cond, Timeout: in.WaitUntil.Timeout, Interval: in.WaitUntil.Interval}
	}
	if in.Repeat != nil {
		r := &flow.RepeatBlock{Times: in.Repeat.Times}
		for _, nested := range in.Repeat.Steps {
			n, err := compileFlowStep(nested, registry)
			if err != nil {
				return flow.Step{}, err
			}
			r.Steps = append(r.Steps, n)
		}
		step.Repeat = r
	}
	return step, nil
}

func stringifyParams(bindingID string, stepIdx int, in map[string]any) (map[string]string, error) {
	params, err := stringifyArbitraryParams(in)
	if err != nil {
		return nil, fmt.Errorf("binding %q binding/actions[%d]/params: %w", bindingID, stepIdx, err)
	}
	return params, nil
}

func stringifyArbitraryParams(in map[string]any) (map[string]string, error) {
	if len(in) == 0 {
		return map[string]string{}, nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		s, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("param %q must be a string", k)
		}
		out[k] = s
	}
	return out, nil
}

func convertUIASelectors(in map[string]program.UIASelectorSpec) map[string]config.UIASelector {
	out := make(map[string]config.UIASelector, len(in))
	for name, sel := range in {
		out[name] = convertUIASelector(sel)
	}
	return out
}

func convertUIASelector(in program.UIASelectorSpec) config.UIASelector {
	out := config.UIASelector{AutomationID: in.AutomationID, Name: in.Name, ControlType: in.ControlType}
	for _, anc := range in.Ancestors {
		out.Ancestors = append(out.Ancestors, convertUIASelector(anc))
	}
	return out
}
