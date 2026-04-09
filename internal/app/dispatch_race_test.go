package app

import (
	"context"
	"sync"
	"testing"
	"time"

	"goahk/internal/actions"
	"goahk/internal/hotkey"
)

func TestDispatchHotkeyEvents_RaceCancellation(t *testing.T) {
	reg := actions.NewRegistry()
	_ = reg.Register("test.mark", func(_ actions.ActionContext, _ actions.Step) error { return nil })
	ex := actions.NewExecutor(reg)
	plans := map[string]actions.Plan{
		"hk":  {{Name: "test.mark"}},
		"hk2": {{Name: "test.mark"}},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	triggers := make(chan hotkey.TriggerEvent, 256)
	results := DispatchHotkeyEvents(ctx, triggers, plans, ex, actions.ActionContext{})

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				target := "hk"
				if j%2 == 0 {
					target = "hk2"
				}
				select {
				case triggers <- hotkey.TriggerEvent{BindingID: target}:
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	time.AfterFunc(10*time.Millisecond, cancel)
	wg.Wait()
	close(triggers)
	for range results {
	}
}
