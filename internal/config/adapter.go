package config

import (
	"fmt"
	"strings"

	"goahk/internal/program"
)

// ToProgram maps a loaded Config into the internal Program model.
func ToProgram(cfg Config) (program.Program, error) {
	p := program.Program{
		Bindings: make([]program.BindingSpec, 0, len(cfg.Hotkeys)),
		Options: program.Options{
			Flows:        make([]program.FlowSpec, 0, len(cfg.Flows)),
			UIASelectors: make(map[string]program.UIASelectorSpec, len(cfg.UIASelectors)),
		},
	}

	flowIDs := make(map[string]struct{}, len(cfg.Flows))
	for i, f := range cfg.Flows {
		if strings.TrimSpace(f.ID) == "" {
			return program.Program{}, fmt.Errorf("flows[%d].id is required", i)
		}
		flowKey := strings.ToLower(strings.TrimSpace(f.ID))
		flowIDs[flowKey] = struct{}{}

		spec, err := flowToProgram(i, f)
		if err != nil {
			return program.Program{}, err
		}
		p.Options.Flows = append(p.Options.Flows, spec)
	}

	for i, b := range cfg.Hotkeys {
		spec, err := hotkeyToProgram(i, b, flowIDs)
		if err != nil {
			return program.Program{}, err
		}
		p.Bindings = append(p.Bindings, spec)
	}

	for name, sel := range cfg.UIASelectors {
		p.Options.UIASelectors[name] = selectorToProgram(sel)
	}

	return p, nil
}

func hotkeyToProgram(idx int, in HotkeyBinding, flowIDs map[string]struct{}) (program.BindingSpec, error) {
	if strings.TrimSpace(in.ID) == "" {
		return program.BindingSpec{}, fmt.Errorf("hotkeys[%d].id is required", idx)
	}
	if strings.TrimSpace(in.Hotkey) == "" {
		return program.BindingSpec{}, fmt.Errorf("hotkeys[%d].hotkey is required", idx)
	}

	hasFlow := strings.TrimSpace(in.Flow) != ""
	hasSteps := len(in.Steps) > 0
	if !hasFlow && !hasSteps {
		return program.BindingSpec{}, fmt.Errorf("hotkeys[%d] requires steps or flow", idx)
	}
	if hasFlow && hasSteps {
		return program.BindingSpec{}, fmt.Errorf("hotkeys[%d] cannot set both steps and flow", idx)
	}
	if hasFlow {
		ref := strings.ToLower(strings.TrimSpace(in.Flow))
		if _, ok := flowIDs[ref]; !ok {
			return program.BindingSpec{}, fmt.Errorf("hotkeys[%d].flow references unknown flow %q", idx, in.Flow)
		}
	}

	return program.BindingSpec{
		ID:     in.ID,
		Hotkey: in.Hotkey,
		Flow:   in.Flow,
		Steps:  stepsToProgram(in.Steps),
	}, nil
}

func stepsToProgram(steps []Step) []program.StepSpec {
	out := make([]program.StepSpec, 0, len(steps))
	for _, s := range steps {
		params := make(map[string]any, len(s.Params))
		for k, v := range s.Params {
			params[k] = v
		}
		out = append(out, program.StepSpec{Action: s.Action, Params: params})
	}
	return out
}

func flowToProgram(flowIdx int, in Flow) (program.FlowSpec, error) {
	if len(in.Steps) == 0 {
		return program.FlowSpec{}, fmt.Errorf("flows[%d].steps is required", flowIdx)
	}
	steps, err := flowStepsToProgram(in.Steps, fmt.Sprintf("flows[%d].steps", flowIdx))
	if err != nil {
		return program.FlowSpec{}, err
	}
	return program.FlowSpec{ID: in.ID, Timeout: in.Timeout.ToStd(), Steps: steps}, nil
}

func flowStepsToProgram(steps []FlowStep, path string) ([]program.FlowStepSpec, error) {
	out := make([]program.FlowStepSpec, 0, len(steps))
	for i, s := range steps {
		currentPath := fmt.Sprintf("%s[%d]", path, i)
		if err := validateFlowStepShape(s, currentPath); err != nil {
			return nil, err
		}

		params := make(map[string]any, len(s.Params))
		for k, v := range s.Params {
			params[k] = v
		}

		step := program.FlowStepSpec{
			Name:    s.Name,
			Action:  s.Action,
			Params:  params,
			Timeout: s.Timeout.ToStd(),
		}

		if s.If != nil {
			thenSteps, err := flowStepsToProgram(s.If.Then, currentPath+".if.then")
			if err != nil {
				return nil, err
			}
			elseSteps, err := flowStepsToProgram(s.If.Else, currentPath+".if.else")
			if err != nil {
				return nil, err
			}
			step.If = &program.FlowIfSpec{
				WindowMatches: s.If.WindowMatches,
				ElementExists: s.If.ElementExists,
				Then:          thenSteps,
				Else:          elseSteps,
			}
		}

		if s.WaitUntil != nil {
			step.WaitUntil = &program.FlowWaitUntilSpec{
				WindowMatches: s.WaitUntil.WindowMatches,
				ElementExists: s.WaitUntil.ElementExists,
				Timeout:       s.WaitUntil.Timeout.ToStd(),
				Interval:      s.WaitUntil.Interval.ToStd(),
			}
		}

		if s.Repeat != nil {
			repeatSteps, err := flowStepsToProgram(s.Repeat.Steps, currentPath+".repeat.steps")
			if err != nil {
				return nil, err
			}
			step.Repeat = &program.FlowRepeatSpec{Times: s.Repeat.Times, Steps: repeatSteps}
		}

		out = append(out, step)
	}
	return out, nil
}

func validateFlowStepShape(step FlowStep, path string) error {
	blockCount := 0
	if strings.TrimSpace(step.Action) != "" {
		blockCount++
	}
	if step.If != nil {
		blockCount++
	}
	if step.WaitUntil != nil {
		blockCount++
	}
	if step.Repeat != nil {
		blockCount++
	}
	if blockCount == 0 {
		return fmt.Errorf("%s requires one of action, if, waitUntil, or repeat", path)
	}
	if blockCount > 1 {
		return fmt.Errorf("%s must set only one of action, if, waitUntil, or repeat", path)
	}
	for key := range step.Params {
		if strings.TrimSpace(key) == "" {
			return fmt.Errorf("%s.params contains an empty key", path)
		}
	}
	return nil
}

func selectorToProgram(in UIASelector) program.UIASelectorSpec {
	out := program.UIASelectorSpec{AutomationID: in.AutomationID, Name: in.Name, ControlType: in.ControlType}
	for _, anc := range in.Ancestors {
		out.Ancestors = append(out.Ancestors, selectorToProgram(anc))
	}
	return out
}
