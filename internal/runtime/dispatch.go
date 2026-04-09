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
	control map[string]RuntimeControlCommand,
	executor *actions.Executor,
	base actions.ActionContext,
	logSink DispatchLogSink,
	onControl func(runtimeControlEvent),
) <-chan DispatchResult {
	if logSink == nil {
		logSink = func(context.Context, DispatchLogEntry) {}
	}
	supervisor := NewSupervisor(ctx, executor, base, logSink, onControl)
	supervisor.Start(4)
	results := supervisor.Results()

	go func() {
		logSink(ctx, DispatchLogEntry{Event: "dispatch_startup", KnownCount: len(plans), Timestamp: time.Now().UTC()})
		for {
			select {
			case <-ctx.Done():
				return
			case <-shutdown:
				supervisor.CloseWhenIdle(250 * time.Millisecond)
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
				if cmd, isControl := control[ev.BindingID]; isControl {
					supervisor.SubmitControl(runtimeControlEvent{BindingID: ev.BindingID, Command: cmd, Triggered: ev, Received: time.Now().UTC()})
					continue
				}
				supervisor.SubmitWork(supervisorJob{bindingID: ev.BindingID, trigger: ev, plan: plan})
			}
		}
	}()

	output := make(chan DispatchResult, 16)
	go func() {
		defer close(output)
		for {
			select {
			case envelope, ok := <-results:
				if !ok {
					return
				}
				logSink(ctx, DispatchLogEntry{Event: "dispatch_trigger_result", BindingID: envelope.BindingID, Actions: envelope.Actions, Duration: envelope.Duration, Timestamp: envelope.Timestamp, Error: envelope.Error})
				if envelope.Error != "" {
					logSink(ctx, DispatchLogEntry{Event: "dispatch_failure_detail", BindingID: envelope.BindingID, Error: envelope.Error, FailedAction: firstFailedAction(envelope.Execution), Timestamp: envelope.Timestamp})
				}
				select {
				case output <- envelope:
				case <-ctx.Done():
					return
				case <-shutdown:
					return
				}
			}
		}
	}()
	return output
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
