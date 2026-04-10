package runtime

import (
	"encoding/json"
	"fmt"
	"strings"

	"goahk/internal/actions"
	"goahk/internal/flow"
	"goahk/internal/hotkey"
	"goahk/internal/program"
	"goahk/internal/uia"
)

type RuntimeBinding struct {
	ID             string
	Chord          hotkey.Chord
	Plan           actions.Plan
	Flow           *flow.Definition
	ControlCommand string
	Policy         program.ConcurrencyPolicy
}

// CompileRuntimeBindings is the canonical compile entrypoint for runtime bindings.
func CompileRuntimeBindings(p program.Program, registry *actions.Registry) ([]RuntimeBinding, error) {
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
		if command, isControl := compileControlCommand(b, parsed[i].Chord, p.Options.EnableImplicitEscapeControls); isControl {
			rb.ControlCommand = command
			compiled = append(compiled, rb)
			continue
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
				sel, err := parseUIASelector(params, selectors)
				if err != nil {
					return nil, fmt.Errorf("binding %q action %q selector: %w", b.ID, step.Action, err)
				}
				raw, err := encodeSelectorJSON(sel)
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

func compileControlCommand(binding program.BindingSpec, chord hotkey.Chord, allowImplicitEscape bool) (string, bool) {
	if len(binding.Steps) == 1 {
		actionName := strings.ToLower(strings.TrimSpace(binding.Steps[0].Action))
		switch actionName {
		case "runtime.control_stop":
			return "stop", true
		case "runtime.control_hard_stop":
			return "hard_stop", true
		}
	}
	if !allowImplicitEscape {
		return "", false
	}
	switch {
	case strings.EqualFold(chord.Key, "escape") && chord.Modifiers == 0:
		return "stop", true
	case strings.EqualFold(chord.Key, "escape") && chord.Modifiers == hotkey.ModShift:
		return "hard_stop", true
	default:
		return "", false
	}
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

func convertUIASelectors(in map[string]program.UIASelectorSpec) map[string]uia.Selector {
	out := make(map[string]uia.Selector, len(in))
	for name, sel := range in {
		out[name] = convertUIASelector(sel)
	}
	return out
}

func convertUIASelector(in program.UIASelectorSpec) uia.Selector {
	out := uia.Selector{AutomationID: in.AutomationID, Name: in.Name, ControlType: in.ControlType}
	for _, anc := range in.Ancestors {
		out.Ancestors = append(out.Ancestors, convertUIASelector(anc))
	}
	return out
}

func parseUIASelector(params map[string]string, defs map[string]uia.Selector) (uia.Selector, error) {
	ref := strings.TrimSpace(params["selector"])
	var sel uia.Selector
	if ref != "" {
		def, ok := defs[ref]
		if !ok {
			return uia.Selector{}, fmt.Errorf("unknown uia selector %q", ref)
		}
		sel = def
	}
	if raw := strings.TrimSpace(params["selector_json"]); raw != "" {
		if err := json.Unmarshal([]byte(raw), &sel); err != nil {
			return uia.Selector{}, fmt.Errorf("decode selector_json: %w", err)
		}
	}
	if sel.AutomationID == "" {
		sel.AutomationID = strings.TrimSpace(params["automationId"])
	}
	if sel.Name == "" {
		sel.Name = strings.TrimSpace(params["name"])
	}
	if sel.ControlType == "" {
		sel.ControlType = strings.TrimSpace(params["controlType"])
	}
	if err := sel.Validate(); err != nil {
		return uia.Selector{}, err
	}
	return sel, nil
}

func encodeSelectorJSON(sel uia.Selector) (string, error) {
	raw, err := json.Marshal(sel)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}
