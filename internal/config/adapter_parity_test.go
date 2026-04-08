package config_test

import (
	"reflect"
	"testing"

	"goahk/internal/actions"
	"goahk/internal/app"
	"goahk/internal/config"
	"goahk/internal/program"
)

func TestAdapterParity_MatrixClaims(t *testing.T) {
	t.Parallel()

	registry := actions.NewRegistry()

	t.Run("hotkey normalization and linear steps are equivalent", func(t *testing.T) {
		cfg := config.Config{Hotkeys: []config.HotkeyBinding{{
			ID:     "open-terminal",
			Hotkey: "shift+ctrl+v",
			Steps: []config.Step{
				{Action: "clipboard.read", Params: map[string]string{"var": "clip"}},
				{Action: "input.send_text", Params: map[string]string{"text": "{{clip}}"}},
			},
		}}}

		fromJSON, err := compileFromConfig(cfg, registry)
		if err != nil {
			t.Fatalf("compileFromConfig() error = %v", err)
		}
		fromBuilder, err := compileFromProgram(program.Program{Bindings: []program.BindingSpec{{
			ID:     "binding_1",
			Hotkey: "Ctrl+Shift+V",
			Steps: []program.StepSpec{
				{Action: "clipboard.read", Params: map[string]any{"var": "clip"}},
				{Action: "input.send_text", Params: map[string]any{"text": "{{clip}}"}},
			},
		}}}, registry)
		if err != nil {
			t.Fatalf("compileFromProgram(builder) error = %v", err)
		}

		if len(fromJSON) != 1 || len(fromBuilder) != 1 {
			t.Fatalf("unexpected binding counts: json=%d builder=%d", len(fromJSON), len(fromBuilder))
		}
		if fromJSON[0].Chord.String() != fromBuilder[0].Chord.String() {
			t.Fatalf("chord mismatch: json=%q builder=%q", fromJSON[0].Chord.String(), fromBuilder[0].Chord.String())
		}
		if !reflect.DeepEqual(fromJSON[0].Plan, fromBuilder[0].Plan) {
			t.Fatalf("plan mismatch: json=%#v builder=%#v", fromJSON[0].Plan, fromBuilder[0].Plan)
		}
	})

	t.Run("flow references compile equivalently", func(t *testing.T) {
		cfg := config.Config{
			Flows: []config.Flow{{
				ID: "paste-flow",
				Steps: []config.FlowStep{{
					Action: "input.send_text",
					Params: map[string]string{"text": "hello"},
				}},
			}},
			Hotkeys: []config.HotkeyBinding{{
				ID:     "paste",
				Hotkey: "Ctrl+Alt+P",
				Flow:   "paste-flow",
			}},
		}

		fromJSON, err := compileFromConfig(cfg, registry)
		if err != nil {
			t.Fatalf("compileFromConfig() error = %v", err)
		}
		fromBuilder, err := compileFromProgram(program.Program{
			Bindings: []program.BindingSpec{{ID: "binding_1", Hotkey: "Ctrl+Alt+P", Flow: "paste-flow"}},
			Options: program.Options{Flows: []program.FlowSpec{{
				ID:    "paste-flow",
				Steps: []program.FlowStepSpec{{Action: "input.send_text", Params: map[string]any{"text": "hello"}}},
			}}},
		}, registry)
		if err != nil {
			t.Fatalf("compileFromProgram(builder) error = %v", err)
		}

		if len(fromJSON) != 1 || len(fromBuilder) != 1 {
			t.Fatalf("unexpected binding counts: json=%d builder=%d", len(fromJSON), len(fromBuilder))
		}
		if fromJSON[0].Flow == nil || fromBuilder[0].Flow == nil {
			t.Fatalf("expected non-nil flows: json=%#v builder=%#v", fromJSON[0].Flow, fromBuilder[0].Flow)
		}
		if !reflect.DeepEqual(fromJSON[0].Flow.Steps, fromBuilder[0].Flow.Steps) {
			t.Fatalf("flow steps mismatch: json=%#v builder=%#v", fromJSON[0].Flow.Steps, fromBuilder[0].Flow.Steps)
		}
	})

	t.Run("uia selector mapping is equivalent", func(t *testing.T) {
		cfg := config.Config{UIASelectors: map[string]config.UIASelector{
			"searchBox": {AutomationID: "SearchBox", Name: "Search", ControlType: "Edit"},
		}}
		mapped, err := config.ToProgram(cfg)
		if err != nil {
			t.Fatalf("ToProgram() error = %v", err)
		}

		builderProgram := program.Program{Options: program.Options{UIASelectors: map[string]program.UIASelectorSpec{
			"searchBox": {AutomationID: "SearchBox", Name: "Search", ControlType: "Edit"},
		}}}

		if !reflect.DeepEqual(mapped.Options.UIASelectors, builderProgram.Options.UIASelectors) {
			t.Fatalf("uia selector mismatch: json=%#v builder=%#v", mapped.Options.UIASelectors, builderProgram.Options.UIASelectors)
		}
	})
}

func compileFromConfig(cfg config.Config, registry *actions.Registry) ([]app.RuntimeBinding, error) {
	p, err := config.ToProgram(cfg)
	if err != nil {
		return nil, err
	}
	return compileFromProgram(p, registry)
}

func compileFromProgram(p program.Program, registry *actions.Registry) ([]app.RuntimeBinding, error) {
	return app.CompileRuntimeBindingsFromProgram(p, registry)
}
