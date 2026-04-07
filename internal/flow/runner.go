package flow

import (
	"context"
	"fmt"
	"time"
)

type ActionHandler func(context.Context, Step) error

type ActionResolver interface {
	ResolveAction(string) (ActionHandler, bool)
}

type Trace struct {
	Name     string
	Kind     string
	Status   string
	Error    string
	Started  time.Time
	Ended    time.Time
	Duration time.Duration
	Nested   []Trace
}

type Result struct {
	Success  bool
	Started  time.Time
	Ended    time.Time
	Duration time.Duration
	Traces   []Trace
}

type Runner struct {
	Actions    ActionResolver
	Conditions ConditionEvaluator
}

func (r Runner) Run(ctx context.Context, def Definition) Result {
	start := time.Now().UTC()
	flowCtx, cancel := withTimeout(ctx, def.Timeout)
	defer cancel()
	res := Result{Success: true, Started: start, Traces: make([]Trace, 0, len(def.Steps))}
	for _, step := range def.Steps {
		tr, cont := r.runStep(flowCtx, def.Timeout, step)
		res.Traces = append(res.Traces, tr)
		if !cont {
			res.Success = false
			break
		}
	}
	res.Ended = time.Now().UTC()
	res.Duration = res.Ended.Sub(res.Started)
	return res
}

func (r Runner) runStep(ctx context.Context, flowTimeout time.Duration, step Step) (Trace, bool) {
	tr := Trace{Name: stepName(step), Kind: step.kind(), Started: time.Now().UTC()}
	defer func() {
		tr.Ended = time.Now().UTC()
		tr.Duration = tr.Ended.Sub(tr.Started)
	}()

	stepCtx, cancel := withTimeout(ctx, effectiveTimeout(flowTimeout, step.Timeout))
	defer cancel()

	switch {
	case step.Action != "":
		h, ok := r.Actions.ResolveAction(step.Action)
		if !ok {
			tr.Status = "failed"
			tr.Error = "action not registered"
			return tr, false
		}
		if err := h(stepCtx, step); err != nil {
			tr.Status = "failed"
			tr.Error = err.Error()
			return tr, false
		}
		tr.Status = "success"
		return tr, true
	case step.If != nil:
		matched, err := r.Conditions.Evaluate(stepCtx, step.If.Condition)
		if err != nil {
			tr.Status = "failed"
			tr.Error = err.Error()
			return tr, false
		}
		branch := step.If.Else
		if matched {
			branch = step.If.Then
		}
		for _, nested := range branch {
			nestedTrace, cont := r.runStep(stepCtx, flowTimeout, nested)
			tr.Nested = append(tr.Nested, nestedTrace)
			if !cont {
				tr.Status = "failed"
				tr.Error = nestedTrace.Error
				return tr, false
			}
		}
		tr.Status = "success"
		return tr, true
	case step.WaitUntil != nil:
		deadline := step.WaitUntil.Timeout
		if deadline <= 0 {
			deadline = effectiveTimeout(flowTimeout, step.Timeout)
		}
		waitCtx, waitCancel := withTimeout(stepCtx, deadline)
		defer waitCancel()
		interval := step.WaitUntil.Interval
		if interval <= 0 {
			interval = 10 * time.Millisecond
		}
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			ok, err := r.Conditions.Evaluate(waitCtx, step.WaitUntil.Condition)
			if err != nil {
				tr.Status = "failed"
				tr.Error = err.Error()
				return tr, false
			}
			if ok {
				tr.Status = "success"
				return tr, true
			}
			select {
			case <-waitCtx.Done():
				tr.Status = "failed"
				tr.Error = fmt.Sprintf("wait_until timeout: %v", waitCtx.Err())
				return tr, false
			case <-ticker.C:
			}
		}
	case step.Repeat != nil:
		for i := 0; i < step.Repeat.Times; i++ {
			for _, nested := range step.Repeat.Steps {
				nestedTrace, cont := r.runStep(stepCtx, flowTimeout, nested)
				tr.Nested = append(tr.Nested, nestedTrace)
				if !cont {
					tr.Status = "failed"
					tr.Error = nestedTrace.Error
					return tr, false
				}
			}
		}
		tr.Status = "success"
		return tr, true
	default:
		tr.Status = "failed"
		tr.Error = "unsupported step"
		return tr, false
	}
}

func stepName(step Step) string {
	if step.Name != "" {
		return step.Name
	}
	if step.Action != "" {
		return step.Action
	}
	return step.kind()
}
