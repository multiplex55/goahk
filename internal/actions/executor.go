package actions

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"strings"
	"time"

	"goahk/internal/flow"
)

type Executor struct {
	registry *Registry
}

func NewExecutor(registry *Registry) *Executor {
	return &Executor{registry: registry}
}

func (e *Executor) Execute(ctx ActionContext, plan Plan) ExecutionResult {
	ctx = ctx.withContext(baseContext(ctx.Context))
	started := time.Now().UTC()
	result := ExecutionResult{StartedAt: started, Success: true, Steps: make([]StepResult, 0, len(plan))}
	for i, step := range plan {
		if ctx.isStopRequested() {
			result.Steps = append(result.Steps, skippedStepResults(plan[i:])...)
			break
		}
		stepResult := StepResult{Action: step.Name, Kind: "action", StartedAt: time.Now().UTC()}
		err := e.executeActionStep(ctx, step)
		stepResult.EndedAt = time.Now().UTC()
		stepResult.Duration = stepResult.EndedAt.Sub(stepResult.StartedAt)
		if err != nil {
			stepResult.Status = StepStatusFailed
			stepResult.Error = err.Error()
			stepResult.ErrorChain = unwrapErrors(err)
			result.Success = false
			result.Steps = append(result.Steps, stepResult)
			break
		}
		stepResult.Status = StepStatusSuccess
		result.Steps = append(result.Steps, stepResult)
		if ctx.isStopRequested() && i+1 < len(plan) {
			result.Steps = append(result.Steps, skippedStepResults(plan[i+1:])...)
			break
		}
	}
	result.EndedAt = time.Now().UTC()
	result.Duration = result.EndedAt.Sub(result.StartedAt)
	return result
}

func skippedStepResults(plan Plan) []StepResult {
	out := make([]StepResult, 0, len(plan))
	for _, step := range plan {
		now := time.Now().UTC()
		out = append(out, StepResult{
			Action:    step.Name,
			Kind:      "action",
			Status:    StepStatusSkipped,
			StartedAt: now,
			EndedAt:   now,
		})
	}
	return out
}

func (e *Executor) ExecuteFlow(ctx ActionContext, def flow.Definition, conditions flow.ConditionEvaluator) ExecutionResult {
	runner := flow.Runner{Actions: flowActionResolver{executor: e, base: ctx}, Conditions: conditions}
	res := runner.Run(baseContext(ctx.Context), def)
	out := ExecutionResult{StartedAt: res.Started, EndedAt: res.Ended, Duration: res.Duration, Success: res.Success, Steps: make([]StepResult, 0, len(res.Traces))}
	for _, tr := range res.Traces {
		out.Steps = append(out.Steps, convertTrace(tr))
	}
	return out
}

func (e *Executor) executeActionStep(ctx ActionContext, step Step) error {
	handler, ok := e.registry.resolve(step, ctx)
	if !ok {
		return fmt.Errorf("action not registered")
	}
	stepCtx := ctx
	baseCtx := baseContext(ctx.Context)
	var cancel context.CancelFunc = func() {}
	timeout := step.Timeout
	if timeout <= 0 {
		timeout = ctx.Timeout
	}
	if timeout > 0 {
		baseCtx, cancel = context.WithTimeout(baseCtx, timeout)
	}
	stepCtx = stepCtx.withContext(baseCtx)
	err := executeSafely(step.Name, func() error {
		return handler(stepCtx, step)
	})
	cancel()
	return err
}

func executeSafely(actionName string, run func() error) (err error) {
	defer func() {
		recovered := recover()
		if recovered == nil {
			return
		}
		err = fmt.Errorf("callback/action panic: %w", fmt.Errorf("%s: %v", actionName, recovered))
		err = fmt.Errorf("%w (stack: %s)", err, strings.TrimSpace(string(debug.Stack())))
	}()
	return run()
}

type flowActionResolver struct {
	executor *Executor
	base     ActionContext
}

func (r flowActionResolver) ResolveAction(name string) (flow.ActionHandler, bool) {
	if _, ok := r.executor.registry.Lookup(name); !ok {
		return nil, false
	}
	return func(ctx context.Context, step flow.Step) error {
		call := Step{Name: name, Params: cloneMap(step.Params), Timeout: step.Timeout}
		stepCtx := r.base
		stepCtx.Context = ctx
		stepCtx = stepCtx.withContext(ctx)
		return r.executor.executeActionStep(stepCtx, call)
	}, true
}

func convertTrace(in flow.Trace) StepResult {
	status := StepStatusSuccess
	if strings.EqualFold(in.Status, "failed") {
		status = StepStatusFailed
	}
	out := StepResult{Action: in.Name, Kind: in.Kind, Status: status, StartedAt: in.Started, EndedAt: in.Ended, Duration: in.Duration, Error: in.Error}
	for _, n := range in.Nested {
		out.Nested = append(out.Nested, convertTrace(n))
	}
	if out.Error != "" {
		out.ErrorChain = []string{out.Error}
	}
	return out
}

func baseContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}

func cloneMap(in map[string]string) map[string]string {
	if len(in) == 0 {
		return map[string]string{}
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func unwrapErrors(err error) []string {
	chain := []string{err.Error()}
	for {
		err = errors.Unwrap(err)
		if err == nil {
			return chain
		}
		chain = append(chain, err.Error())
	}
}
