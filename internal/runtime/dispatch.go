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

type DispatchHandle struct {
	Results           <-chan DispatchResult
	ForceTerminateAll func()
}

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
	return DispatchHotkeyEventsWithHandle(ctx, shutdown, events, plans, control, executor, base, logSink, onControl).Results
}

func DispatchHotkeyEventsWithHandle(
	ctx context.Context,
	shutdown <-chan struct{},
	events <-chan hotkey.TriggerEvent,
	plans map[string]actions.Plan,
	control map[string]RuntimeControlCommand,
	executor *actions.Executor,
	base actions.ActionContext,
	logSink DispatchLogSink,
	onControl func(runtimeControlEvent),
) DispatchHandle {
	bindings := make(map[string]actions.ExecutableBinding, len(plans))
	for id, plan := range plans {
		bindings[id] = actions.ExecutableBinding{ID: id, Kind: actions.BindingKindPlan, Plan: plan}
	}
	return DispatchHotkeyEventsWithBindingsHandle(ctx, shutdown, events, bindings, control, executor, base, logSink, onControl)
}

func DispatchHotkeyEventsWithBindingsHandle(
	ctx context.Context,
	shutdown <-chan struct{},
	events <-chan hotkey.TriggerEvent,
	bindings map[string]actions.ExecutableBinding,
	control map[string]RuntimeControlCommand,
	executor *actions.Executor,
	base actions.ActionContext,
	logSink DispatchLogSink,
	onControl func(runtimeControlEvent),
) DispatchHandle {
	if logSink == nil {
		logSink = func(context.Context, DispatchLogEntry) {}
	}
	supervisor := NewSupervisor(ctx, bindings, executor, base, logSink, onControl)
	supervisor.Start(4)
	results := supervisor.Results()

	go func() {
		logSink(ctx, DispatchLogEntry{Event: "dispatch_startup", KnownCount: len(bindings), Timestamp: time.Now().UTC()})
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
				if cmd, isControl := control[ev.BindingID]; isControl {
					supervisor.SubmitControl(runtimeControlEvent{BindingID: ev.BindingID, Command: cmd, Triggered: ev, Received: time.Now().UTC()})
					continue
				}
				supervisor.SubmitWork(supervisorJob{bindingID: ev.BindingID, trigger: ev})
			}
		}
	}()

	output := make(chan DispatchResult, 16)
	go func() {
		defer close(output)
		for envelope := range results {
			logSink(ctx, DispatchLogEntry{Event: "dispatch_trigger_result", BindingID: envelope.BindingID, Actions: envelope.Actions, Duration: envelope.Duration, Timestamp: envelope.Timestamp, Error: envelope.Error})
			if envelope.Error != "" {
				logSink(ctx, DispatchLogEntry{Event: "dispatch_failure_detail", BindingID: envelope.BindingID, Error: envelope.Error, FailedAction: firstFailedAction(envelope.Execution), Timestamp: envelope.Timestamp})
			}
			select {
			case output <- envelope:
			case <-time.After(25 * time.Millisecond):
			}
		}
	}()
	return DispatchHandle{
		Results:           output,
		ForceTerminateAll: supervisor.ForceTerminateAll,
	}
}

func buildDispatchResult(bindingID string, binding actions.ExecutableBinding, res actions.ExecutionResult) DispatchResult {
	ts := res.EndedAt
	if ts.IsZero() {
		ts = time.Now().UTC()
	}
	return DispatchResult{
		BindingID: bindingID,
		Actions:   dispatchActions(binding),
		Outcomes:  append([]actions.StepResult(nil), res.Steps...),
		Duration:  res.Duration,
		Error:     extractExecutionError(res),
		Timestamp: ts,
		Execution: res,
	}
}

func dispatchActions(binding actions.ExecutableBinding) []string {
	switch binding.Kind {
	case actions.BindingKindPlan:
		actionsList := make([]string, 0, len(binding.Plan))
		for _, step := range binding.Plan {
			actionsList = append(actionsList, step.Name)
		}
		return actionsList
	case actions.BindingKindFlow:
		if binding.Flow != nil && binding.Flow.ID != "" {
			return []string{"flow:" + binding.Flow.ID}
		}
		return []string{"flow"}
	case actions.BindingKindCallback:
		if binding.Policy.CallbackRef != "" {
			return []string{actions.CallbackActionName + ":" + binding.Policy.CallbackRef}
		}
		return []string{actions.CallbackActionName}
	default:
		return []string{string(binding.Kind)}
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
