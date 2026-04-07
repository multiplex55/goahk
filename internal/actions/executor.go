package actions

import (
	"context"
	"errors"
	"time"
)

type Executor struct {
	registry *Registry
}

func NewExecutor(registry *Registry) *Executor {
	return &Executor{registry: registry}
}

func (e *Executor) Execute(ctx ActionContext, plan Plan) ExecutionResult {
	started := time.Now().UTC()
	result := ExecutionResult{StartedAt: started, Success: true, Steps: make([]StepResult, 0, len(plan))}
	for _, step := range plan {
		stepResult := StepResult{Action: step.Name, StartedAt: time.Now().UTC()}
		handler, ok := e.registry.Lookup(step.Name)
		if !ok {
			stepResult.Status = StepStatusFailed
			stepResult.Error = "action not registered"
			stepResult.EndedAt = time.Now().UTC()
			stepResult.Duration = stepResult.EndedAt.Sub(stepResult.StartedAt)
			result.Success = false
			result.Steps = append(result.Steps, stepResult)
			break
		}

		stepCtx := ctx
		baseCtx := ctx.Context
		if baseCtx == nil {
			baseCtx = context.Background()
		}
		var cancel context.CancelFunc = func() {}
		timeout := step.Timeout
		if timeout <= 0 {
			timeout = ctx.Timeout
		}
		if timeout > 0 {
			baseCtx, cancel = context.WithTimeout(baseCtx, timeout)
		}
		stepCtx = stepCtx.withContext(baseCtx)
		err := handler(stepCtx, step)
		cancel()

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
	}
	result.EndedAt = time.Now().UTC()
	result.Duration = result.EndedAt.Sub(result.StartedAt)
	return result
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
