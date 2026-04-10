package goahk

import (
	"context"
	"testing"

	"goahk/internal/actions"
	internalapp "goahk/internal/app"
	"goahk/internal/runtime"
)

type logRecord struct {
	msg    string
	fields map[string]any
}

type fakeLogger struct {
	records []logRecord
}

func (f *fakeLogger) Info(msg string, fields map[string]any) {
	f.records = append(f.records, logRecord{msg: msg, fields: copyLogFields(fields)})
}

func TestRunCallback_RegistersAndExecutesThroughExecutor(t *testing.T) {
	t.Parallel()

	called := false
	a := NewApp()
	a.Bind("Ctrl+H", Func(func(*Context) error {
		called = true
		return nil
	}))

	bindings, executor := compileTestBindings(t, a)
	res := executor.Execute(actions.ActionContext{Context: context.Background(), BindingID: bindings[0].ID}, bindings[0].Plan)
	if !res.Success {
		t.Fatalf("execution failed: %#v", res)
	}
	if !called {
		t.Fatal("callback was not executed")
	}
}

func compileTestBindings(t *testing.T, a *App) ([]internalapp.RuntimeBinding, *actions.Executor) {
	t.Helper()

	program, _, callbacks := a.runtimeArtifacts()
	registry := buildRegistryWithCallbacks(a.state, callbacks, a.actionLogger())
	bindings, err := internalapp.CompileRuntimeBindingsFromProgram(program, registry)
	if err != nil {
		t.Fatalf("CompileRuntimeBindingsFromProgram() error = %v", err)
	}
	return bindings, actions.NewExecutor(registry)
}

func TestRunCallback_WithLoggerReceivesDispatchEvents(t *testing.T) {
	t.Parallel()

	collector := &fakeLogger{}
	a := NewApp(WithLogger(collector))
	sink := a.dispatchLogSink(a.actionLogger())
	sink(context.Background(), runtime.DispatchLogEntry{Event: "dispatch_startup", KnownCount: 2})

	if len(collector.records) != 1 {
		t.Fatalf("record count = %d, want 1", len(collector.records))
	}
	if collector.records[0].msg != "goahk.dispatch" {
		t.Fatalf("msg = %q, want goahk.dispatch", collector.records[0].msg)
	}
	if collector.records[0].fields["event"] != "dispatch_startup" {
		t.Fatalf("event field = %v, want dispatch_startup", collector.records[0].fields["event"])
	}
}

func TestRunCallback_ConfiguredLoggerUsedInsideCallbackContext(t *testing.T) {
	t.Parallel()

	collector := &fakeLogger{}
	a := NewApp(WithLogger(collector))
	a.Bind("Ctrl+H", Func(func(ctx *Context) error {
		ctx.actionCtx.Logger.Info("callback-log", map[string]any{"origin": "callback"})
		return nil
	}))

	bindings, executor := compileTestBindings(t, a)
	res := executor.Execute(actions.ActionContext{Context: context.Background(), BindingID: bindings[0].ID, Logger: a.actionLogger()}, bindings[0].Plan)
	if !res.Success {
		t.Fatalf("execution failed: %#v", res)
	}
	if len(collector.records) == 0 {
		t.Fatal("expected callback log to be recorded")
	}
}
