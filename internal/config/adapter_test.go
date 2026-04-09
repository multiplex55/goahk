package config_test

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"goahk/internal/actions"
	"goahk/internal/app"
	"goahk/internal/config"
	"goahk/internal/program"
)

func TestToProgram_MinimalConfigMapsExpectedProgram(t *testing.T) {
	cfg := config.Config{
		Hotkeys: []config.HotkeyBinding{{
			ID:     "open-terminal",
			Hotkey: "Ctrl+Alt+T",
			Steps:  []config.Step{{Action: "input.send_keys", Params: map[string]string{"sequence": "Win+R"}}},
		}},
	}

	got, err := config.ToProgram(cfg)
	if err != nil {
		t.Fatalf("ToProgram() error = %v", err)
	}

	want := program.Program{
		Bindings: []program.BindingSpec{{
			ID:                "open-terminal",
			Hotkey:            "Ctrl+Alt+T",
			ConcurrencyPolicy: program.ConcurrencyPolicySerial,
			Steps: []program.StepSpec{{
				Action: "input.send_keys",
				Params: map[string]any{"sequence": "Win+R"},
			}},
		}},
		Options: program.Options{
			Flows:        []program.FlowSpec{},
			UIASelectors: map[string]program.UIASelectorSpec{},
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ToProgram() mismatch:\n got: %#v\nwant: %#v", got, want)
	}
}

func TestToProgram_Parity_JSONAndProgramCompileEquivalence(t *testing.T) {
	cfg, err := config.LoadFile(filepath.Join("..", "..", "testdata", "config", "valid_minimal.json"))
	if err != nil {
		t.Fatalf("LoadFile() error = %v", err)
	}

	fromConfig, err := config.ToProgram(cfg)
	if err != nil {
		t.Fatalf("ToProgram() error = %v", err)
	}
	fromCode := program.Program{
		Bindings: []program.BindingSpec{{
			ID:     "open-terminal",
			Hotkey: "Ctrl+Alt+T",
			Steps: []program.StepSpec{{
				Action: "input.send_keys",
				Params: map[string]any{"sequence": "Win+R"},
			}},
		}},
		Options: program.Options{Flows: []program.FlowSpec{}, UIASelectors: map[string]program.UIASelectorSpec{}},
	}

	registry := actions.NewRegistry()

	compiledFromConfig, err := app.CompileRuntimeBindingsFromProgram(fromConfig, registry)
	if err != nil {
		t.Fatalf("CompileRuntimeBindingsFromProgram(config) error = %v", err)
	}
	compiledFromCode, err := app.CompileRuntimeBindingsFromProgram(fromCode, registry)
	if err != nil {
		t.Fatalf("CompileRuntimeBindingsFromProgram(code) error = %v", err)
	}

	if len(compiledFromConfig) != len(compiledFromCode) {
		t.Fatalf("compiled length mismatch: %d vs %d", len(compiledFromConfig), len(compiledFromCode))
	}
	for i := range compiledFromConfig {
		if compiledFromConfig[i].ID != compiledFromCode[i].ID {
			t.Fatalf("binding id mismatch at %d: %q vs %q", i, compiledFromConfig[i].ID, compiledFromCode[i].ID)
		}
		if compiledFromConfig[i].Chord != compiledFromCode[i].Chord {
			t.Fatalf("binding chord mismatch at %d: %q vs %q", i, compiledFromConfig[i].Chord.String(), compiledFromCode[i].Chord.String())
		}
		if !reflect.DeepEqual(compiledFromConfig[i].Plan, compiledFromCode[i].Plan) {
			t.Fatalf("binding plan mismatch at %d: %#v vs %#v", i, compiledFromConfig[i].Plan, compiledFromCode[i].Plan)
		}
	}
}

func TestToProgram_Errors_MalformedPayloadAndMissingFields(t *testing.T) {
	t.Run("malformed flow step payload", func(t *testing.T) {
		cfg := config.Config{
			Flows: []config.Flow{{
				ID: "flow-1",
				Steps: []config.FlowStep{{
					Action: "system.log",
					If:     &config.FlowIf{},
				}},
			}},
			Hotkeys: []config.HotkeyBinding{{
				ID:     "h1",
				Hotkey: "Ctrl+1",
				Flow:   "flow-1",
			}},
		}
		_, err := config.ToProgram(cfg)
		if err == nil {
			t.Fatal("ToProgram() error = nil, want failure")
		}
		if !strings.Contains(err.Error(), "must set only one") {
			t.Fatalf("ToProgram() error = %q, want malformed payload details", err)
		}
	})

	t.Run("missing required field", func(t *testing.T) {
		cfg := config.Config{
			Hotkeys: []config.HotkeyBinding{{Hotkey: "Ctrl+1", Steps: []config.Step{{Action: "system.log"}}}},
		}
		_, err := config.ToProgram(cfg)
		if err == nil {
			t.Fatal("ToProgram() error = nil, want failure")
		}
		if !strings.Contains(err.Error(), "hotkeys[0].id is required") {
			t.Fatalf("ToProgram() error = %q, missing required-field message", err)
		}
	})
}
