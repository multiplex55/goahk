package goahk

import "goahk/internal/program"

type App struct {
	bindings        []bindingSpec
	logger          Logger
	validateActions bool
}

type bindingSpec struct {
	hotkey string
	steps  []stepSpecProvider
}

func NewApp(opts ...Option) *App {
	a := &App{validateActions: true}
	for _, opt := range opts {
		if opt != nil {
			opt(a)
		}
	}
	return a
}

func (a *App) toProgram() program.Program {
	out := program.Program{Bindings: make([]program.BindingSpec, 0, len(a.bindings))}
	for i, b := range a.bindings {
		steps := make([]program.StepSpec, 0, len(b.steps))
		for _, step := range b.steps {
			steps = append(steps, step.stepSpec())
		}
		out.Bindings = append(out.Bindings, program.BindingSpec{ID: bindingID(i), Hotkey: b.hotkey, Steps: steps})
	}
	return out
}
