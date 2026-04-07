package actions

import "time"

type Step struct {
	Name    string
	Params  map[string]string
	Timeout time.Duration
}

type Plan []Step

type Handler func(ActionContext, Step) error
