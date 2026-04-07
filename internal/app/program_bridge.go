package app

import (
	"goahk/internal/config"
	"goahk/internal/program"
)

func ConfigToProgram(cfg config.Config) program.Program {
	p := program.Program{
		Bindings: make([]program.BindingSpec, 0, len(cfg.Hotkeys)),
		Options: program.Options{
			Flows:        make([]program.FlowSpec, 0, len(cfg.Flows)),
			UIASelectors: make(map[string]program.UIASelectorSpec, len(cfg.UIASelectors)),
		},
	}
	for _, b := range cfg.Hotkeys {
		p.Bindings = append(p.Bindings, program.BindingSpec{
			ID:     b.ID,
			Hotkey: b.Hotkey,
			Flow:   b.Flow,
			Steps:  configStepsToProgram(b.Steps),
		})
	}
	for _, f := range cfg.Flows {
		p.Options.Flows = append(p.Options.Flows, configFlowToProgram(f))
	}
	for name, sel := range cfg.UIASelectors {
		p.Options.UIASelectors[name] = configSelectorToProgram(sel)
	}
	return p
}

func configStepsToProgram(steps []config.Step) []program.StepSpec {
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

func configFlowToProgram(f config.Flow) program.FlowSpec {
	return program.FlowSpec{ID: f.ID, Timeout: f.Timeout.ToStd(), Steps: configFlowStepsToProgram(f.Steps)}
}

func configFlowStepsToProgram(steps []config.FlowStep) []program.FlowStepSpec {
	out := make([]program.FlowStepSpec, 0, len(steps))
	for _, s := range steps {
		params := make(map[string]any, len(s.Params))
		for k, v := range s.Params {
			params[k] = v
		}
		step := program.FlowStepSpec{Name: s.Name, Action: s.Action, Params: params, Timeout: s.Timeout.ToStd()}
		if s.If != nil {
			step.If = &program.FlowIfSpec{
				WindowMatches: s.If.WindowMatches,
				ElementExists: s.If.ElementExists,
				Then:          configFlowStepsToProgram(s.If.Then),
				Else:          configFlowStepsToProgram(s.If.Else),
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
			step.Repeat = &program.FlowRepeatSpec{Times: s.Repeat.Times, Steps: configFlowStepsToProgram(s.Repeat.Steps)}
		}
		out = append(out, step)
	}
	return out
}

func configSelectorToProgram(in config.UIASelector) program.UIASelectorSpec {
	out := program.UIASelectorSpec{AutomationID: in.AutomationID, Name: in.Name, ControlType: in.ControlType}
	for _, anc := range in.Ancestors {
		out.Ancestors = append(out.Ancestors, configSelectorToProgram(anc))
	}
	return out
}
