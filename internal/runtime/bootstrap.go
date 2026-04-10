package runtime

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"goahk/internal/actions"
	"goahk/internal/clipboard"
	"goahk/internal/config"
	"goahk/internal/hotkey"
	"goahk/internal/input"
	"goahk/internal/process"
	"goahk/internal/program"
	"goahk/internal/services/messagebox"
	"goahk/internal/shell/folders"
	"goahk/internal/window"
)

type ProgramLoader func(context.Context, string) (program.Program, error)

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
	LoadProgram        ProgramLoader
	BuildRegistry      RegistryBuilder
	NewListener        ListenerFactory
	RecordResult       ResultRecorder
	LogDispatch        DispatchLogSink
	BaseActionCtx      actions.ActionContext
	StopGrace          time.Duration
	HardStopAfterGrace bool
}

func NewBootstrap() Bootstrap {
	windowProvider := window.NewOSProvider()
	folderSvc := folders.NewService()
	return Bootstrap{
		LoadProgram: func(_ context.Context, path string) (program.Program, error) {
			return config.LoadProgramFile(path)
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
			WindowActivate: func(ctx context.Context, matcher string) error {
				_, err := window.ActivateForeground(ctx, windowProvider, window.ParseMatcherString(matcher))
				return err
			},
			WindowActive: func(ctx context.Context) (window.Info, error) {
				return window.Active(ctx, windowProvider)
			},
			ActiveWindowTitle: func(ctx context.Context) (string, error) {
				active, err := window.Active(ctx, windowProvider)
				if err != nil {
					return "", err
				}
				return active.Title, nil
			},
			WindowList: func(ctx context.Context) ([]window.Info, error) {
				return window.Enumerate(ctx, windowProvider)
			},
			WindowBounds:   windowProvider.WindowBounds,
			WindowWorkArea: windowProvider.WorkAreaForWindow,
			WindowMove:     windowProvider.MoveWindow,
			WindowResize:   windowProvider.ResizeWindow,
			WindowMinimize: windowProvider.MinimizeWindow,
			WindowMaximize: windowProvider.MaximizeWindow,
			WindowRestore:  windowProvider.RestoreWindow,
			FolderList:     folderSvc.ListOpenFolders,
			Input:          input.NewService(),
		}},
		StopGrace:          300 * time.Millisecond,
		HardStopAfterGrace: false,
	}
}

func (b Bootstrap) Run(ctx context.Context, configPath string) error {
	if b.LoadProgram == nil {
		return fmt.Errorf("program loader is not configured")
	}
	if b.LogDispatch == nil {
		b.LogDispatch = func(context.Context, DispatchLogEntry) {}
	}

	b.LogDispatch(ctx, DispatchLogEntry{Event: "runtime_startup", Timestamp: time.Now().UTC()})
	p, err := b.LoadProgram(ctx, configPath)
	if err != nil {
		b.LogDispatch(ctx, DispatchLogEntry{Event: "runtime_load_program_failed", Error: err.Error(), Timestamp: time.Now().UTC()})
		return fmt.Errorf("load program: %w", err)
	}

	return b.RunProgram(ctx, p)
}

func (b Bootstrap) RunProgram(ctx context.Context, p program.Program) error {
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

	if err := program.Validate(p); err != nil {
		b.LogDispatch(ctx, DispatchLogEntry{Event: "runtime_validate_failed", Error: err.Error(), Timestamp: time.Now().UTC()})
		return fmt.Errorf("validate program: %w", err)
	}

	registry, err := b.BuildRegistry(ctx, p)
	if err != nil {
		b.LogDispatch(ctx, DispatchLogEntry{Event: "runtime_build_registry_failed", Error: err.Error(), Timestamp: time.Now().UTC()})
		return fmt.Errorf("build action registry: %w", err)
	}

	compiled, err := CompileRuntimeBindings(p, registry)
	if err != nil {
		b.LogDispatch(ctx, DispatchLogEntry{Event: "runtime_compile_failed", Error: err.Error(), Timestamp: time.Now().UTC()})
		return fmt.Errorf("compile runtime bindings: %w", err)
	}
	b.LogDispatch(ctx, DispatchLogEntry{Event: "binding_registration_summary", KnownCount: len(compiled), Timestamp: time.Now().UTC()})

	listener, err := b.NewListener(ctx)
	if err != nil {
		b.LogDispatch(ctx, DispatchLogEntry{Event: "runtime_listener_create_failed", Error: err.Error(), Timestamp: time.Now().UTC()})
		return fmt.Errorf("create windows listener: %w", err)
	}
	manager := hotkey.NewManager(listener)

	registered := make([]string, 0, len(compiled))
	for _, binding := range compiled {
		if err := manager.Register(binding.ID, binding.Chord); err != nil {
			_ = unregisterAll(manager, registered)
			_ = manager.Close()
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

	bindingDescriptors := buildExecutableBindings(compiled)
	controlByBindingID := make(map[string]RuntimeControlCommand, len(compiled))
	for _, binding := range compiled {
		switch RuntimeControlCommand(binding.ControlCommand) {
		case RuntimeControlStop, RuntimeControlHardStop, RuntimeControlSuspend, RuntimeControlReload:
			controlByBindingID[binding.ID] = RuntimeControlCommand(binding.ControlCommand)
		}
	}

	executor := actions.NewExecutor(registry)
	var hardStop atomic.Bool
	dispatch := DispatchHotkeyEventsWithBindingsHandle(runCtx, runCtx.Done(), manager.Events(), bindingDescriptors, controlByBindingID, executor, baseActionCtx, b.LogDispatch, func(ev runtimeControlEvent) {
		switch ev.Command {
		case RuntimeControlHardStop:
			hardStop.Store(true)
			cancel()
		case RuntimeControlStop:
			cancel()
		}
	})
	dispatchDone := make(chan struct{})
	go func() {
		defer close(dispatchDone)
		for result := range dispatch.Results {
			b.RecordResult(runCtx, result.BindingID, result.Execution)
		}
	}()

	loopErr := b.runLoop(runCtx, listener, managerErr)
	cancel()
	grace := b.StopGrace
	if grace <= 0 {
		grace = 300 * time.Millisecond
	}
	select {
	case <-dispatchDone:
	case <-time.After(grace):
		b.LogDispatch(ctx, DispatchLogEntry{Event: "shutdown_grace_timeout", Timestamp: time.Now().UTC()})
		if b.HardStopAfterGrace {
			dispatch.ForceTerminateAll()
		}
	}
	if hardStop.Load() {
		b.LogDispatch(ctx, DispatchLogEntry{Event: "job_forced_termination", Timestamp: time.Now().UTC()})
	}
	shutdownReason := "completed"
	if hardStop.Load() {
		shutdownReason = "runtime.hard_stop"
	} else if runCtx.Err() != nil {
		shutdownReason = runCtx.Err().Error()
	}
	b.LogDispatch(ctx, DispatchLogEntry{Event: "runtime_shutdown", Error: shutdownReason, Timestamp: time.Now().UTC()})
	_ = <-managerErr

	if err := unregisterAll(manager, registered); err != nil {
		_ = manager.Close()
		return fmt.Errorf("unregister hotkeys: %w", err)
	}
	if err := manager.Close(); err != nil {
		return fmt.Errorf("close hotkey manager: %w", err)
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
