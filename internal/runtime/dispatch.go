package runtime

import (
	"context"
	"time"

	"goahk/internal/actions"
	"goahk/internal/hotkey"
)

type DispatchResult struct {
	BindingID string
	Actions   []string
	Outcomes  []actions.StepResult
	Duration  time.Duration
	Error     string
	Timestamp time.Time
	Execution actions.ExecutionResult
}

type DispatchLogEntry struct {
	Event        string
	BindingID    string
	KnownCount   int
	Actions      []string
	Duration     time.Duration
	Error        string
	Timestamp    time.Time
	FailedAction string
}

type DispatchLogSink func(context.Context, DispatchLogEntry)

func DispatchHotkeyEvents(
	ctx context.Context,
	shutdown <-chan struct{},
	events <-chan hotkey.TriggerEvent,
	plans map[string]actions.Plan,
	executor *actions.Executor,
	base actions.ActionContext,
	logSink DispatchLogSink,
) <-chan DispatchResult {
	results := make(chan DispatchResult, 16)
	if logSink == nil {
		logSink = func(context.Context, DispatchLogEntry) {}
	}

	go func() {
		defer close(results)
		logSink(ctx, DispatchLogEntry{Event: "dispatch_startup", KnownCount: len(plans), Timestamp: time.Now().UTC()})
		for {
			select {
			case <-ctx.Done():
				return
			case <-shutdown:
				return
			case ev, ok := <-events:
				if !ok {
					return
				}
				plan, exists := plans[ev.BindingID]
				if !exists {
					logSink(ctx, DispatchLogEntry{Event: "dispatch_unknown_binding", BindingID: ev.BindingID, Error: "binding plan not found", Timestamp: time.Now().UTC()})
					continue
				}
				actionCtx := base
				actionCtx.Context = ctx
				actionCtx.BindingID = ev.BindingID
				actionCtx.TriggerText = ev.Chord.String()
				execResult := executor.Execute(actionCtx, plan)
				envelope := buildDispatchResult(ev.BindingID, plan, execResult)
				logSink(ctx, DispatchLogEntry{Event: "dispatch_trigger_result", BindingID: envelope.BindingID, Actions: envelope.Actions, Duration: envelope.Duration, Timestamp: envelope.Timestamp, Error: envelope.Error})
				if envelope.Error != "" {
					logSink(ctx, DispatchLogEntry{Event: "dispatch_failure_detail", BindingID: envelope.BindingID, Error: envelope.Error, FailedAction: firstFailedAction(execResult), Timestamp: envelope.Timestamp})
				}
				select {
				case results <- envelope:
				default:
					select {
					case results <- envelope:
					case <-ctx.Done():
						return
					case <-shutdown:
						return
					}
				}
			}
		}
	}()
	return results
}

func buildDispatchResult(bindingID string, plan actions.Plan, res actions.ExecutionResult) DispatchResult {
	actionsList := make([]string, 0, len(plan))
	for _, step := range plan {
		actionsList = append(actionsList, step.Name)
	}
	ts := res.EndedAt
	if ts.IsZero() {
		ts = time.Now().UTC()
	}
	return DispatchResult{
		BindingID: bindingID,
		Actions:   actionsList,
		Outcomes:  append([]actions.StepResult(nil), res.Steps...),
		Duration:  res.Duration,
		Error:     extractExecutionError(res),
		Timestamp: ts,
		Execution: res,
	}
}

func extractExecutionError(res actions.ExecutionResult) string {
	for _, step := range res.Steps {
		if step.Status == actions.StepStatusFailed {
			if step.Error != "" {
				return step.Error
			}
			return "action execution failed"
		}
	}
	return ""
}

func firstFailedAction(res actions.ExecutionResult) string {
	for _, step := range res.Steps {
		if step.Status == actions.StepStatusFailed {
			return step.Action
		}
	}
	return ""
}
