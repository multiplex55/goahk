package goahk

import (
	"goahk/internal/program"
)

type ActionFunc func(*Context) error

type callbackStep struct {
	fn ActionFunc
}

func Func(fn ActionFunc) callbackStep {
	return callbackStep{fn: fn}
}

func (c callbackStep) stepSpec() program.StepSpec {
	return program.StepSpec{Action: callbackActionPlaceholder, Params: map[string]any{}}
}
