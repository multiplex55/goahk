package hotkey_test

import (
	"context"
	"go/parser"
	"go/token"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"goahk/internal/actions"
	"goahk/internal/hotkey"
	rt "goahk/internal/runtime"
)

func TestHotkeyPackage_DoesNotImportSyntheticInputInternals(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	dir := filepath.Dir(file)
	matches, err := filepath.Glob(filepath.Join(dir, "*.go"))
	if err != nil {
		t.Fatalf("glob err: %v", err)
	}
	fset := token.NewFileSet()
	for _, path := range matches {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		parsed, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		for _, imp := range parsed.Imports {
			pkg := strings.Trim(imp.Path.Value, "\"")
			if pkg == "goahk/internal/input" {
				pos := fset.PositionFor(imp.Pos(), true)
				t.Fatalf("unexpected input import in trigger path: %s:%d", pos.Filename, pos.Line)
			}
		}
	}
}

func TestRapidTriggerBursts_RespectConcurrencyPolicy(t *testing.T) {
	tests := []struct {
		name         string
		policy       string
		triggerBurst int
		wantRuns     int
	}{
		{name: "serial", policy: "serial", triggerBurst: 8, wantRuns: 1},
		{name: "parallel", policy: "parallel", triggerBurst: 8, wantRuns: 8},
		{name: "queue-one", policy: "queue-one", triggerBurst: 8, wantRuns: 2},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			reg := actions.NewRegistry()
			release := make(chan struct{})
			started := make(chan struct{}, tc.triggerBurst+2)
			if err := reg.Register("test.block", func(ctx actions.ActionContext, _ actions.Step) error {
				started <- struct{}{}
				select {
				case <-release:
					return nil
				case <-ctx.Context.Done():
					return ctx.Context.Err()
				}
			}); err != nil {
				t.Fatalf("register action: %v", err)
			}

			bindings := map[string]actions.ExecutableBinding{
				"hk": {
					ID:     "hk",
					Kind:   actions.BindingKindPlan,
					Plan:   actions.Plan{{Name: "test.block"}},
					Policy: actions.BindingExecutionPolicy{Concurrency: tc.policy},
				},
			}
			events := make(chan hotkey.TriggerEvent, tc.triggerBurst+2)
			shutdown := make(chan struct{})
			handle := rt.DispatchHotkeyEventsWithBindingsHandle(context.Background(), shutdown, events, bindings, nil, actions.NewExecutor(reg), actions.ActionContext{}, nil, nil)

			for i := 0; i < tc.triggerBurst; i++ {
				events <- hotkey.TriggerEvent{BindingID: "hk"}
			}

			time.Sleep(30 * time.Millisecond)
			close(release)
			close(shutdown)

			got := 0
			for range handle.Results {
				got++
			}
			if got != tc.wantRuns {
				t.Fatalf("runs = %d, want %d", got, tc.wantRuns)
			}
		})
	}
}
