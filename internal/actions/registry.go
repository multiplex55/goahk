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
	handlers         map[string]Handler
	callbacks        map[string]CallbackHandler
	bindingCallbacks map[string]CallbackHandler
}

var cooldowns sync.Map

const CallbackActionName = "goahk.callback"

type CallbackHandler func(CallbackContext) error

func NewRegistry() *Registry {
	r := &Registry{
		handlers:         map[string]Handler{},
		callbacks:        map[string]CallbackHandler{},
		bindingCallbacks: map[string]CallbackHandler{},
	}
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

func (r *Registry) RegisterCallback(name string, callback CallbackHandler) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("callback name is required")
	}
	if callback == nil {
		return fmt.Errorf("callback for %q is nil", name)
	}
	if _, exists := r.callbacks[name]; exists {
		return fmt.Errorf("callback %q already registered", name)
	}
	r.callbacks[name] = callback
	return nil
}

func (r *Registry) MustRegisterCallback(name string, callback CallbackHandler) {
	if err := r.RegisterCallback(name, callback); err != nil {
		panic(err)
	}
}

func (r *Registry) BindCallback(bindingID string, callback CallbackHandler) error {
	bindingID = strings.TrimSpace(bindingID)
	if bindingID == "" {
		return fmt.Errorf("binding id is required")
	}
	if callback == nil {
		return fmt.Errorf("binding callback for %q is nil", bindingID)
	}
	if _, exists := r.bindingCallbacks[bindingID]; exists {
		return fmt.Errorf("binding callback %q already registered", bindingID)
	}
	r.bindingCallbacks[bindingID] = callback
	return nil
}

func (r *Registry) ResolveBindingCallback(bindingID, ref string) (CallbackHandler, bool) {
	if named := strings.TrimSpace(ref); named != "" {
		if cb, ok := r.callbacks[named]; ok {
			return cb, true
		}
	}
	if cb, ok := r.bindingCallbacks[strings.TrimSpace(bindingID)]; ok {
		return cb, true
	}
	return nil, false
}

func (r *Registry) resolve(step Step, ctx ActionContext) (Handler, bool) {
	if handler, ok := r.Lookup(step.Name); ok {
		return handler, true
	}
	if !strings.EqualFold(step.Name, CallbackActionName) {
		return nil, false
	}
	if cb, ok := r.ResolveBindingCallback(ctx.BindingID, step.Params["callback_ref"]); ok {
		return adaptCallbackHandler(cb), true
	}
	return nil, false
}

func adaptCallbackHandler(callback CallbackHandler) Handler {
	return func(actionCtx ActionContext, _ Step) error {
		return callback(NewCallbackContext(&actionCtx))
	}
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

	r.MustRegister("system.message_box", runMessageBoxAction)

	r.MustRegister("runtime.stop", func(ctx ActionContext, _ Step) error {
		RequestRuntimeStop(&ctx, "runtime.stop")
		return nil
	})
	r.MustRegister("system.stop", func(ctx ActionContext, step Step) error {
		handler, _ := r.Lookup("runtime.stop")
		return handler(ctx, step)
	})

	r.MustRegister("clipboard.read", func(ctx ActionContext, step Step) error {
		if ctx.Services.Clipboard == nil {
			return missingServiceError(step, "clipboard")
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

	r.MustRegister("clipboard.write", runClipboardWriteAction)

	r.MustRegister("clipboard.append", func(ctx ActionContext, step Step) error {
		if ctx.Services.Clipboard == nil {
			return missingServiceError(step, "clipboard")
		}
		runner := func() error { return ctx.Services.Clipboard.AppendText(ctx.Context, step.Params["text"]) }
		if withRestore(step) {
			return runWithClipboardRestore(ctx, runner)
		}
		return runner()
	})

	r.MustRegister("clipboard.prepend", func(ctx ActionContext, step Step) error {
		if ctx.Services.Clipboard == nil {
			return missingServiceError(step, "clipboard")
		}
		runner := func() error { return ctx.Services.Clipboard.PrependText(ctx.Context, step.Params["text"]) }
		if withRestore(step) {
			return runWithClipboardRestore(ctx, runner)
		}
		return runner()
	})

	r.MustRegister("process.launch", runProcessLaunchAction)
	r.MustRegister("system.open", runSystemOpenAction)

	r.MustRegister("window.activate", func(ctx ActionContext, step Step) error {
		if ctx.Services.WindowActivate == nil {
			return missingServiceError(step, "window")
		}
		return ctx.Services.WindowActivate(ctx.Context, step.Params["matcher"])
	})
	r.MustRegister("window.move", func(ctx ActionContext, step Step) error {
		if ctx.Services.WindowMove == nil {
			return missingServiceError(step, "window")
		}
		target, err := resolveWindowForMatcher(ctx, step)
		if err != nil {
			return err
		}
		x, err := parseInt(step.Params, "x")
		if err != nil {
			return err
		}
		y, err := parseInt(step.Params, "y")
		if err != nil {
			return err
		}
		return ctx.Services.WindowMove(ctx.Context, target.HWND, x, y)
	})
	r.MustRegister("window.resize", func(ctx ActionContext, step Step) error {
		if ctx.Services.WindowResize == nil {
			return missingServiceError(step, "window")
		}
		target, err := resolveWindowForMatcher(ctx, step)
		if err != nil {
			return err
		}
		width, err := parseInt(step.Params, "width")
		if err != nil {
			return err
		}
		height, err := parseInt(step.Params, "height")
		if err != nil {
			return err
		}
		return ctx.Services.WindowResize(ctx.Context, target.HWND, width, height)
	})
	r.MustRegister("window.minimize", func(ctx ActionContext, step Step) error {
		if ctx.Services.WindowMinimize == nil {
			return missingServiceError(step, "window")
		}
		target, err := resolveWindowForMatcher(ctx, step)
		if err != nil {
			return err
		}
		return ctx.Services.WindowMinimize(ctx.Context, target.HWND)
	})
	r.MustRegister("window.maximize", func(ctx ActionContext, step Step) error {
		if ctx.Services.WindowMaximize == nil {
			return missingServiceError(step, "window")
		}
		target, err := resolveWindowForMatcher(ctx, step)
		if err != nil {
			return err
		}
		return ctx.Services.WindowMaximize(ctx.Context, target.HWND)
	})
	r.MustRegister("window.restore", func(ctx ActionContext, step Step) error {
		if ctx.Services.WindowRestore == nil {
			return missingServiceError(step, "window")
		}
		target, err := resolveWindowForMatcher(ctx, step)
		if err != nil {
			return err
		}
		return ctx.Services.WindowRestore(ctx.Context, target.HWND)
	})

	r.MustRegister("window.copy_active_title_to_clipboard", func(ctx ActionContext, step Step) error {
		if ctx.Services.ActiveWindowTitle == nil {
			return missingServiceError(step, "window")
		}
		if ctx.Services.Clipboard == nil {
			return missingServiceError(step, "clipboard")
		}
		title, err := ctx.Services.ActiveWindowTitle(ctx.Context)
		if err != nil {
			return err
		}
		return ctx.Services.Clipboard.WriteText(ctx.Context, title)
	})

	r.MustRegister("window.list_open_applications", runWindowListOpenApplicationsAction)
	r.MustRegister("window.list_open_folders", runWindowListOpenFoldersAction)

	r.MustRegister("input.send_text", func(ctx ActionContext, step Step) error {
		if ctx.Services.Input == nil {
			return missingServiceError(step, "input")
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
			return missingServiceError(step, "input")
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
			return missingServiceError(step, "input")
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
	r.MustRegister("input.mouse_move_absolute", func(ctx ActionContext, step Step) error {
		if ctx.Services.Input == nil {
			return missingServiceError(step, "input")
		}
		x, err := parseInt(step.Params, "x")
		if err != nil {
			return err
		}
		y, err := parseInt(step.Params, "y")
		if err != nil {
			return err
		}
		return ctx.Services.Input.MoveAbsolute(ctx.Context, x, y)
	})
	r.MustRegister("input.mouse_move_relative", func(ctx ActionContext, step Step) error {
		if ctx.Services.Input == nil {
			return missingServiceError(step, "input")
		}
		dx, err := parseInt(step.Params, "dx")
		if err != nil {
			return err
		}
		dy, err := parseInt(step.Params, "dy")
		if err != nil {
			return err
		}
		return ctx.Services.Input.MoveRelative(ctx.Context, dx, dy)
	})
	r.MustRegister("input.mouse_button_down", func(ctx ActionContext, step Step) error {
		if ctx.Services.Input == nil {
			return missingServiceError(step, "input")
		}
		return ctx.Services.Input.ButtonDown(ctx.Context, step.Params["button"])
	})
	r.MustRegister("input.mouse_button_up", func(ctx ActionContext, step Step) error {
		if ctx.Services.Input == nil {
			return missingServiceError(step, "input")
		}
		return ctx.Services.Input.ButtonUp(ctx.Context, step.Params["button"])
	})
	r.MustRegister("input.mouse_click", func(ctx ActionContext, step Step) error {
		if ctx.Services.Input == nil {
			return missingServiceError(step, "input")
		}
		return ctx.Services.Input.Click(ctx.Context, step.Params["button"])
	})
	r.MustRegister("input.mouse_double_click", func(ctx ActionContext, step Step) error {
		if ctx.Services.Input == nil {
			return missingServiceError(step, "input")
		}
		return ctx.Services.Input.DoubleClick(ctx.Context, step.Params["button"])
	})
	r.MustRegister("input.mouse_wheel", func(ctx ActionContext, step Step) error {
		if ctx.Services.Input == nil {
			return missingServiceError(step, "input")
		}
		delta, err := parseInt(step.Params, "delta")
		if err != nil {
			return err
		}
		return ctx.Services.Input.Wheel(ctx.Context, delta)
	})
	r.MustRegister("input.mouse_drag", func(ctx ActionContext, step Step) error {
		if ctx.Services.Input == nil {
			return missingServiceError(step, "input")
		}
		startX, err := parseInt(step.Params, "start_x")
		if err != nil {
			return err
		}
		startY, err := parseInt(step.Params, "start_y")
		if err != nil {
			return err
		}
		endX, err := parseInt(step.Params, "end_x")
		if err != nil {
			return err
		}
		endY, err := parseInt(step.Params, "end_y")
		if err != nil {
			return err
		}
		return ctx.Services.Input.Drag(ctx.Context, step.Params["button"], startX, startY, endX, endY)
	})
	r.MustRegister("input.mouse_get_position", func(ctx ActionContext, step Step) error {
		if ctx.Services.Input == nil {
			return missingServiceError(step, "input")
		}
		pos, err := ctx.Services.Input.Position(ctx.Context)
		if err != nil {
			return err
		}
		saveAs := strings.TrimSpace(step.Params["save_as"])
		if saveAs == "" {
			return nil
		}
		if ctx.Metadata == nil {
			ctx.Metadata = map[string]string{}
		}
		ctx.Metadata[saveAs+"_x"] = strconv.Itoa(pos.X)
		ctx.Metadata[saveAs+"_y"] = strconv.Itoa(pos.Y)
		return nil
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
