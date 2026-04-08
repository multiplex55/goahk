package runtime

import (
	"context"
	"fmt"

	"goahk/internal/actions"
	appinternal "goahk/internal/app"
	"goahk/internal/clipboard"
	"goahk/internal/config"
	"goahk/internal/hotkey"
	"goahk/internal/process"
	"goahk/internal/program"
	"goahk/internal/services/messagebox"
)

type ConfigLoader func(context.Context, string) (config.Config, error)

type RegistryBuilder func(context.Context, program.Program) (*actions.Registry, error)

type Listener interface {
	Register(registrationID int, chord hotkey.Chord) error
	Unregister(registrationID int) error
	Events() <-chan hotkey.ListenerEvent
	Run(context.Context) error
	Close() error
}

type ListenerFactory func(context.Context) (Listener, error)

type ResultRecorder func(context.Context, string, actions.ExecutionResult)

type Bootstrap struct {
	LoadConfig    ConfigLoader
	BuildRegistry RegistryBuilder
	NewListener   ListenerFactory
	RecordResult  ResultRecorder
	LogDispatch   DispatchLogSink
	BaseActionCtx actions.ActionContext
}

func NewBootstrap() Bootstrap {
	return Bootstrap{
		LoadConfig: func(_ context.Context, path string) (config.Config, error) {
			return config.LoadFile(path)
		},
		BuildRegistry: func(_ context.Context, _ program.Program) (*actions.Registry, error) {
			return actions.NewRegistry(), nil
		},
		NewListener:  NewWindowsListener,
		RecordResult: func(context.Context, string, actions.ExecutionResult) {},
		LogDispatch:  func(context.Context, DispatchLogEntry) {},
		BaseActionCtx: actions.ActionContext{Services: actions.Services{
			MessageBox: messagebox.NewService(),
			Clipboard:  clipboard.NewService(nil),
			Process:    process.NewService(),
		}},
	}
}

func (b Bootstrap) Run(ctx context.Context, configPath string) error {
	if b.LoadConfig == nil {
		return fmt.Errorf("config loader is not configured")
	}
	if b.BuildRegistry == nil {
		return fmt.Errorf("registry builder is not configured")
	}
	if b.NewListener == nil {
		return fmt.Errorf("listener factory is not configured")
	}
	if b.RecordResult == nil {
		b.RecordResult = func(context.Context, string, actions.ExecutionResult) {}
	}
	if b.LogDispatch == nil {
		b.LogDispatch = func(context.Context, DispatchLogEntry) {}
	}

	cfg, err := b.LoadConfig(ctx, configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	p, err := config.ToProgram(cfg)
	if err != nil {
		return fmt.Errorf("adapt config to program: %w", err)
	}
	if err := program.Validate(p); err != nil {
		return fmt.Errorf("validate program: %w", err)
	}

	registry, err := b.BuildRegistry(ctx, p)
	if err != nil {
		return fmt.Errorf("build action registry: %w", err)
	}

	compiled, err := appinternal.CompileRuntimeBindingsFromProgram(p, registry)
	if err != nil {
		return fmt.Errorf("compile runtime bindings: %w", err)
	}

	listener, err := b.NewListener(ctx)
	if err != nil {
		return fmt.Errorf("create windows listener: %w", err)
	}
	manager := hotkey.NewManager(listener)

	registered := make([]string, 0, len(compiled))
	for _, binding := range compiled {
		if err := manager.Register(binding.ID, binding.Chord); err != nil {
			_ = unregisterAll(manager, registered)
			_ = listener.Close()
			return fmt.Errorf("register binding %q: %w", binding.ID, err)
		}
		registered = append(registered, binding.ID)
	}

	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	baseActionCtx := b.BaseActionCtx
	prevStop := baseActionCtx.Stop
	baseActionCtx.Stop = func(reason string) {
		if prevStop != nil {
			prevStop(reason)
		}
		cancel()
	}

	managerErr := make(chan error, 1)
	go func() {
		err := manager.Run(runCtx)
		if err != nil && err != context.Canceled {
			managerErr <- err
			return
		}
		managerErr <- nil
	}()

	plansByBindingID := make(map[string]actions.Plan, len(compiled))
	for _, binding := range compiled {
		if binding.Flow == nil {
			plansByBindingID[binding.ID] = binding.Plan
		}
	}

	executor := actions.NewExecutor(registry)
	results := DispatchHotkeyEvents(runCtx, runCtx.Done(), manager.Events(), plansByBindingID, executor, baseActionCtx, b.LogDispatch)
	dispatchDone := make(chan struct{})
	go func() {
		defer close(dispatchDone)
		for result := range results {
			b.RecordResult(runCtx, result.BindingID, result.Execution)
		}
	}()

	loopErr := b.runLoop(runCtx, listener, managerErr)
	cancel()
	<-dispatchDone
	_ = <-managerErr

	if err := unregisterAll(manager, registered); err != nil {
		_ = listener.Close()
		return fmt.Errorf("unregister hotkeys: %w", err)
	}
	if err := listener.Close(); err != nil {
		return fmt.Errorf("close listener: %w", err)
	}
	if loopErr != nil {
		return loopErr
	}
	return nil
}

func (b Bootstrap) runLoop(ctx context.Context, listener Listener, managerErr <-chan error) error {
	loopErr := make(chan error, 1)
	go func() {
		err := listener.Run(ctx)
		if err != nil && err != context.Canceled {
			loopErr <- fmt.Errorf("run windows message loop: %w", err)
			return
		}
		loopErr <- nil
	}()

	select {
	case <-ctx.Done():
		return nil
	case err := <-managerErr:
		if err != nil {
			return fmt.Errorf("run hotkey manager: %w", err)
		}
		return nil
	case err := <-loopErr:
		return err
	}
}

func unregisterAll(manager *hotkey.Manager, bindingIDs []string) error {
	var first error
	for i := len(bindingIDs) - 1; i >= 0; i-- {
		if err := manager.Unregister(bindingIDs[i]); err != nil && first == nil {
			first = err
		}
	}
	return first
}
