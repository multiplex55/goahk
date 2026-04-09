package runtime

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"goahk/internal/actions"
	"goahk/internal/hotkey"
	"goahk/internal/testutil"
)

func TestDispatchLogGolden_StopControlPath(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	events := make(chan hotkey.TriggerEvent, 2)
	shutdown := make(chan struct{})

	var mu sync.Mutex
	lines := []string{}
	sink := func(_ context.Context, entry DispatchLogEntry) {
		if entry.Event != "dispatch_startup" && entry.Event != "control_command_received" {
			return
		}
		mu.Lock()
		defer mu.Unlock()
		lines = append(lines, entry.Event+":"+entry.BindingID)
	}

	control := map[string]RuntimeControlCommand{"esc": RuntimeControlStop}
	_ = DispatchHotkeyEvents(ctx, shutdown, events, map[string]actions.Plan{"esc": {}}, control, actions.NewExecutor(actions.NewRegistry()), actions.ActionContext{}, sink, func(ev runtimeControlEvent) {
		if ev.Command == RuntimeControlStop {
			cancel()
		}
	})

	events <- hotkey.TriggerEvent{BindingID: "esc", Chord: hotkey.Chord{Key: "Escape"}}
	<-ctx.Done()
	close(shutdown)
	select {
	case <-time.After(40 * time.Millisecond):
	}

	mu.Lock()
	got := strings.Join(lines, "\n") + "\n"
	mu.Unlock()
	testutil.AssertGolden(t, "testdata/golden/runtime/dispatch_stop_events.txt", got)
}

func TestDispatchLogGolden_ReplacePolicyPath(t *testing.T) {
	reg := actions.NewRegistry()
	firstStarted := make(chan struct{}, 1)
	releaseSecond := make(chan struct{})
	runCount := 0
	reg.MustRegister("test.replace_block", func(ctx actions.ActionContext, _ actions.Step) error {
		runCount++
		if runCount == 1 {
			firstStarted <- struct{}{}
			<-ctx.Context.Done()
			return ctx.Context.Err()
		}
		select {
		case <-releaseSecond:
			return nil
		case <-ctx.Context.Done():
			return ctx.Context.Err()
		}
	})

	events := make(chan hotkey.TriggerEvent, 4)
	shutdown := make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var mu sync.Mutex
	lines := []string{}
	sink := func(_ context.Context, entry DispatchLogEntry) {
		switch entry.Event {
		case "policy_replace_admit_latest", "policy_replace_cancel_running":
			mu.Lock()
			lines = append(lines, entry.Event+":"+entry.BindingID)
			mu.Unlock()
		}
	}

	bindings := map[string]actions.ExecutableBinding{
		"replace": {
			ID:     "replace",
			Kind:   actions.BindingKindPlan,
			Plan:   actions.Plan{{Name: "test.replace_block"}},
			Policy: actions.BindingExecutionPolicy{Concurrency: "replace"},
		},
	}
	_ = DispatchHotkeyEventsWithBindingsHandle(ctx, shutdown, events, bindings, nil, actions.NewExecutor(reg), actions.ActionContext{}, sink, nil).Results

	events <- hotkey.TriggerEvent{BindingID: "replace"}
	select {
	case <-firstStarted:
	case <-time.After(time.Second):
		t.Fatal("first run did not start")
	}
	events <- hotkey.TriggerEvent{BindingID: "replace"}
	time.Sleep(40 * time.Millisecond)
	close(releaseSecond)
	close(shutdown)
	select {
	case <-time.After(40 * time.Millisecond):
	}

	mu.Lock()
	got := strings.Join(lines, "\n") + "\n"
	mu.Unlock()
	testutil.AssertGolden(t, "testdata/golden/runtime/dispatch_replace_policy_events.txt", got)
}
