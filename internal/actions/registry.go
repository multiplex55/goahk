package actions

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"goahk/internal/clipboard"
	"goahk/internal/input"
	"goahk/internal/uia"
)

type Registry struct {
	handlers map[string]Handler
}

var cooldowns sync.Map

func NewRegistry() *Registry {
	r := &Registry{handlers: map[string]Handler{}}
	r.registerBuiltins()
	return r
}

func (r *Registry) Register(name string, handler Handler) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("action name is required")
	}
	if handler == nil {
		return fmt.Errorf("handler for %q is nil", name)
	}
	if _, exists := r.handlers[name]; exists {
		return fmt.Errorf("action %q already registered", name)
	}
	r.handlers[name] = handler
	return nil
}

func (r *Registry) MustRegister(name string, handler Handler) {
	if err := r.Register(name, handler); err != nil {
		panic(err)
	}
}

func (r *Registry) Lookup(name string) (Handler, bool) {
	h, ok := r.handlers[name]
	return h, ok
}

func (r *Registry) registerBuiltins() {
	r.MustRegister("system.log", func(ctx ActionContext, step Step) error {
		if ctx.Logger == nil {
			ctx.Logger = NoopLogger{}
		}
		msg := step.Params["message"]
		if msg == "" {
			msg = "system.log"
		}
		ctx.Logger.Info(msg, map[string]any{"action": step.Name, "binding_id": ctx.BindingID})
		return nil
	})

	r.MustRegister("system.message_box", func(ctx ActionContext, step Step) error {
		if ctx.Services.MessageBox == nil {
			return nil
		}
		title := step.Params["title"]
		text := step.Params["text"]
		if text == "" {
			text = step.Params["message"]
		}
		return ctx.Services.MessageBox(ctx.Context, title, text)
	})

	r.MustRegister("clipboard.read", func(ctx ActionContext, step Step) error {
		if ctx.Services.Clipboard == nil {
			return nil
		}
		text, err := ctx.Services.Clipboard.ReadText(ctx.Context)
		if err != nil {
			return err
		}
		if key := strings.TrimSpace(step.Params["save_as"]); key != "" {
			if ctx.Metadata == nil {
				ctx.Metadata = map[string]string{}
			}
			ctx.Metadata[key] = text
		}
		return nil
	})

	r.MustRegister("clipboard.write", func(ctx ActionContext, step Step) error {
		if ctx.Services.Clipboard == nil {
			return nil
		}
		writer := func() error { return ctx.Services.Clipboard.WriteText(ctx.Context, step.Params["text"]) }
		if withRestore(step) {
			return runWithClipboardRestore(ctx, writer)
		}
		return writer()
	})

	r.MustRegister("clipboard.append", func(ctx ActionContext, step Step) error {
		if ctx.Services.Clipboard == nil {
			return nil
		}
		runner := func() error { return ctx.Services.Clipboard.AppendText(ctx.Context, step.Params["text"]) }
		if withRestore(step) {
			return runWithClipboardRestore(ctx, runner)
		}
		return runner()
	})

	r.MustRegister("clipboard.prepend", func(ctx ActionContext, step Step) error {
		if ctx.Services.Clipboard == nil {
			return nil
		}
		runner := func() error { return ctx.Services.Clipboard.PrependText(ctx.Context, step.Params["text"]) }
		if withRestore(step) {
			return runWithClipboardRestore(ctx, runner)
		}
		return runner()
	})

	r.MustRegister("process.launch", func(ctx ActionContext, step Step) error {
		if ctx.Services.ProcessLaunch == nil {
			return nil
		}
		var args []string
		if raw := strings.TrimSpace(step.Params["args"]); raw != "" {
			args = strings.Fields(raw)
		}
		return ctx.Services.ProcessLaunch(ctx.Context, step.Params["path"], args)
	})

	r.MustRegister("window.activate", func(ctx ActionContext, step Step) error {
		if ctx.Services.WindowActivate == nil {
			return nil
		}
		return ctx.Services.WindowActivate(ctx.Context, step.Params["matcher"])
	})

	r.MustRegister("window.copy_active_title_to_clipboard", func(ctx ActionContext, step Step) error {
		if ctx.Services.ActiveWindowTitle == nil || ctx.Services.Clipboard == nil {
			return nil
		}
		title, err := ctx.Services.ActiveWindowTitle(ctx.Context)
		if err != nil {
			return err
		}
		return ctx.Services.Clipboard.WriteText(ctx.Context, title)
	})

	r.MustRegister("input.send_text", func(ctx ActionContext, step Step) error {
		if ctx.Services.Input == nil {
			return nil
		}
		opts, err := guardInputAction(ctx, step)
		if err != nil {
			return err
		}
		text := step.Params["text"]
		if parseBoolDefault(step.Params["decode_escapes"], false) {
			text, err = input.DecodeEscapes(text)
			if err != nil {
				return err
			}
		}
		return ctx.Services.Input.SendText(ctx.Context, text, opts)
	})

	r.MustRegister("input.send_keys", func(ctx ActionContext, step Step) error {
		if ctx.Services.Input == nil {
			return nil
		}
		opts, err := guardInputAction(ctx, step)
		if err != nil {
			return err
		}
		raw := strings.TrimSpace(step.Params["sequence"])
		if raw == "" {
			raw = strings.TrimSpace(step.Params["keys"])
		}
		if raw == "" {
			return fmt.Errorf("input.send_keys requires sequence")
		}
		seq, err := input.ParseSequence(raw)
		if err != nil {
			return err
		}
		return ctx.Services.Input.SendKeys(ctx.Context, seq, opts)
	})

	r.MustRegister("input.send_chord", func(ctx ActionContext, step Step) error {
		if ctx.Services.Input == nil {
			return nil
		}
		opts, err := guardInputAction(ctx, step)
		if err != nil {
			return err
		}
		raw := strings.TrimSpace(step.Params["chord"])
		if raw == "" {
			return fmt.Errorf("input.send_chord requires chord")
		}
		tokens, err := input.TokenizeSequence(raw)
		if err != nil {
			return err
		}
		if len(tokens) != 1 {
			return fmt.Errorf("input.send_chord expects a single token")
		}
		return ctx.Services.Input.SendChord(ctx.Context, input.Chord{Keys: tokens[0].Keys}, opts)
	})

	r.MustRegister("uia.find", func(ctx ActionContext, step Step) error {
		if ctx.Services.UIA == nil {
			return nil
		}
		sel, timeout, retry, err := parseUIAActionInputs(step)
		if err != nil {
			return err
		}
		el, diag, err := ctx.Services.UIA.Find(ctx.Context, sel, timeout, retry)
		logUIADiagnostics(ctx, step, sel, timeout, diag)
		if err != nil {
			return err
		}
		if key := strings.TrimSpace(step.Params["save_as"]); key != "" {
			if ctx.Metadata == nil {
				ctx.Metadata = map[string]string{}
			}
			raw, _ := json.Marshal(el)
			ctx.Metadata[key] = string(raw)
		}
		return nil
	})

	r.MustRegister("uia.invoke", func(ctx ActionContext, step Step) error {
		return runUIAOp(ctx, step, func(sel uia.Selector, timeout, retry time.Duration) (uia.ActionDiagnostics, error) {
			return ctx.Services.UIA.Invoke(ctx.Context, sel, timeout, retry)
		})
	})
	r.MustRegister("uia.value_set", func(ctx ActionContext, step Step) error {
		value := step.Params["value"]
		return runUIAOp(ctx, step, func(sel uia.Selector, timeout, retry time.Duration) (uia.ActionDiagnostics, error) {
			return ctx.Services.UIA.ValueSet(ctx.Context, sel, value, timeout, retry)
		})
	})
	r.MustRegister("uia.value_get", func(ctx ActionContext, step Step) error {
		if ctx.Services.UIA == nil {
			return nil
		}
		sel, timeout, retry, err := parseUIAActionInputs(step)
		if err != nil {
			return err
		}
		value, diag, err := ctx.Services.UIA.ValueGet(ctx.Context, sel, timeout, retry)
		logUIADiagnostics(ctx, step, sel, timeout, diag)
		if err != nil {
			return err
		}
		if key := strings.TrimSpace(step.Params["save_as"]); key != "" {
			if ctx.Metadata == nil {
				ctx.Metadata = map[string]string{}
			}
			ctx.Metadata[key] = value
		}
		return nil
	})
	r.MustRegister("uia.toggle", func(ctx ActionContext, step Step) error {
		return runUIAOp(ctx, step, func(sel uia.Selector, timeout, retry time.Duration) (uia.ActionDiagnostics, error) {
			return ctx.Services.UIA.Toggle(ctx.Context, sel, timeout, retry)
		})
	})
	r.MustRegister("uia.expand", func(ctx ActionContext, step Step) error {
		return runUIAOp(ctx, step, func(sel uia.Selector, timeout, retry time.Duration) (uia.ActionDiagnostics, error) {
			return ctx.Services.UIA.Expand(ctx.Context, sel, timeout, retry)
		})
	})
	r.MustRegister("uia.select", func(ctx ActionContext, step Step) error {
		return runUIAOp(ctx, step, func(sel uia.Selector, timeout, retry time.Duration) (uia.ActionDiagnostics, error) {
			return ctx.Services.UIA.Select(ctx.Context, sel, timeout, retry)
		})
	})
}

func withRestore(step Step) bool {
	raw := strings.TrimSpace(step.Params["with_restore"])
	if raw == "" {
		return false
	}
	v, err := strconv.ParseBool(raw)
	return err == nil && v
}

func runWithClipboardRestore(ctx ActionContext, run func() error) error {
	prior, err := ctx.Services.Clipboard.ReadText(ctx.Context)
	if err != nil {
		return err
	}
	if err := run(); err != nil {
		return err
	}
	return ctx.Services.Clipboard.WriteText(ctx.Context, clipboard.NormalizeWriteText(prior))
}

func guardInputAction(ctx ActionContext, step Step) (input.SendOptions, error) {
	delayMS, err := parseIntParam(step.Params, "delay_ms")
	if err != nil {
		return input.SendOptions{}, err
	}
	cooldownMS, err := parseIntParam(step.Params, "cooldown_ms")
	if err != nil {
		return input.SendOptions{}, err
	}
	if cooldownMS > 0 {
		key := strings.Join([]string{ctx.BindingID, step.Name, serializeParams(step.Params)}, "|")
		now := time.Now()
		if priorRaw, ok := cooldowns.Load(key); ok {
			prior := priorRaw.(time.Time)
			if now.Sub(prior) < time.Duration(cooldownMS)*time.Millisecond {
				return input.SendOptions{}, nil
			}
		}
		cooldowns.Store(key, now)
	}
	if ctx.Metadata == nil {
		ctx.Metadata = map[string]string{}
	}
	suppress := parseBoolDefault(step.Params["suppress_reentrancy"], true)
	if suppress {
		ctx.Metadata["suppress_reentrancy"] = "true"
	}
	return input.SendOptions{DelayBefore: time.Duration(delayMS) * time.Millisecond, SuppressReentrancy: suppress}, nil
}

func parseIntParam(params map[string]string, key string) (int, error) {
	raw := strings.TrimSpace(params[key])
	if raw == "" {
		return 0, nil
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v < 0 {
		return 0, fmt.Errorf("invalid %s: %q", key, raw)
	}
	return v, nil
}

func parseBoolDefault(raw string, defaultValue bool) bool {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return defaultValue
	}
	v, err := strconv.ParseBool(raw)
	if err != nil {
		return defaultValue
	}
	return v
}

func serializeParams(params map[string]string) string {
	if len(params) == 0 {
		return ""
	}
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	// stable order for cooldown key
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[j] < keys[i] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+"="+params[k])
	}
	return strings.Join(parts, ",")
}

func parseUIAActionInputs(step Step) (uia.Selector, time.Duration, time.Duration, error) {
	raw := strings.TrimSpace(step.Params["selector_json"])
	if raw == "" {
		return uia.Selector{}, 0, 0, fmt.Errorf("%s requires selector_json", step.Name)
	}
	var sel uia.Selector
	if err := json.Unmarshal([]byte(raw), &sel); err != nil {
		return uia.Selector{}, 0, 0, fmt.Errorf("decode selector_json: %w", err)
	}
	if err := sel.Validate(); err != nil {
		return uia.Selector{}, 0, 0, err
	}
	timeoutMS, err := parseIntParam(step.Params, "timeout_ms")
	if err != nil {
		return uia.Selector{}, 0, 0, err
	}
	retryMS, err := parseIntParam(step.Params, "retry_interval_ms")
	if err != nil {
		return uia.Selector{}, 0, 0, err
	}
	return sel, time.Duration(timeoutMS) * time.Millisecond, time.Duration(retryMS) * time.Millisecond, nil
}

func runUIAOp(ctx ActionContext, step Step, op func(uia.Selector, time.Duration, time.Duration) (uia.ActionDiagnostics, error)) error {
	if ctx.Services.UIA == nil {
		return nil
	}
	sel, timeout, retry, err := parseUIAActionInputs(step)
	if err != nil {
		return err
	}
	diag, err := op(sel, timeout, retry)
	logUIADiagnostics(ctx, step, sel, timeout, diag)
	return err
}

func logUIADiagnostics(ctx ActionContext, step Step, sel uia.Selector, timeout time.Duration, diag uia.ActionDiagnostics) {
	if ctx.Logger == nil {
		ctx.Logger = NoopLogger{}
	}
	ctx.Logger.Info("uia.action", map[string]any{
		"action":             step.Name,
		"selector":           sel,
		"retry_count":        diag.RetryCount,
		"timeout":            timeout.String(),
		"supported_patterns": diag.SupportedPatterns,
		"missing_pattern":    diag.MissingPattern,
	})
}
