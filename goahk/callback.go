package goahk

import (
	"goahk/internal/program"
)

// ActionFunc is an imperative callback action that receives trigger context.
type ActionFunc func(*Context) error

type callbackStep struct {
	fn ActionFunc
}

// Func wraps a callback so it can be used alongside declarative Action steps.
func Func(fn ActionFunc) callbackStep {
	return callbackStep{fn: fn}
}

func (c callbackStep) stepSpec() program.StepSpec {
	return program.StepSpec{Action: callbackActionPlaceholder, Params: map[string]any{}}
}
