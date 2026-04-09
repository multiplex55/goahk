package goahk

import (
	"fmt"
	"reflect"
	"strings"
)

// BindingBuilder is an intermediate fluent value produced by App.On.
type BindingBuilder struct {
	app    *App
	hotkey string
}

// On starts a binding chain for a hotkey.
//
// Use On("Ctrl+H").Do(...) when you prefer a fluent style.
func (a *App) On(hotkey string) *BindingBuilder {
	return &BindingBuilder{app: a, hotkey: normalizeHotkey(hotkey)}
}

// Bind adds a hotkey binding in one call and returns the same app for chaining.
func (a *App) Bind(hotkey string, steps ...stepSpecProvider) *App {
	return a.On(hotkey).Do(steps...)
}

// Do completes an App.On(...).Do(...) chain by attaching action steps.
//
// Invalid wiring is reported by App.Run as a concise build error.
func (b *BindingBuilder) Do(steps ...stepSpecProvider) *App {
	copied := make([]stepSpecProvider, len(steps))
	copy(copied, steps)
	b.app.bindings = append(b.app.bindings, bindingSpec{hotkey: b.hotkey, steps: copied})
	b.app.recordBindingWiringErrors(b.hotkey, copied)
	return b.app
}

func (a *App) recordBindingWiringErrors(hotkey string, steps []stepSpecProvider) {
	if strings.TrimSpace(hotkey) == "" {
		a.buildErrors = append(a.buildErrors, "binding hotkey cannot be empty")
	}
	if len(steps) == 0 {
		a.buildErrors = append(a.buildErrors, fmt.Sprintf("binding %q must include at least one action", hotkey))
		return
	}
	for i, step := range steps {
		if isNilStep(step) {
			a.buildErrors = append(a.buildErrors, fmt.Sprintf("binding %q step %d is nil", hotkey, i+1))
			continue
		}
		if cb, ok := step.(callbackStep); ok && cb.fn == nil {
			a.buildErrors = append(a.buildErrors, fmt.Sprintf("binding %q step %d callback is nil", hotkey, i+1))
			continue
		}
		spec := step.stepSpec()
		if strings.TrimSpace(spec.Action) == "" {
			a.buildErrors = append(a.buildErrors, fmt.Sprintf("binding %q step %d has an empty action name", hotkey, i+1))
		}
	}
}

func isNilStep(step stepSpecProvider) bool {
	if step == nil {
		return true
	}
	v := reflect.ValueOf(step)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return v.IsNil()
	default:
		return false
	}
}

func normalizeHotkey(in string) string {
	in = strings.TrimSpace(in)
	if in == "" {
		return in
	}
	tokens := strings.Split(in, "+")
	for i := range tokens {
		tokens[i] = strings.TrimSpace(tokens[i])
	}
	if len(tokens) == 1 {
		return normalizeKey(tokens[0])
	}
	mods := make([]string, 0, len(tokens)-1)
	seen := map[string]bool{}
	for _, tok := range tokens[:len(tokens)-1] {
		mod := normalizeModifier(tok)
		if mod == "" {
			continue
		}
		if !seen[mod] {
			mods = append(mods, mod)
			seen[mod] = true
		}
	}
	key := normalizeKey(tokens[len(tokens)-1])
	if len(mods) == 0 {
		return key
	}
	order := []string{"Ctrl", "Alt", "Shift", "Win"}
	ordered := make([]string, 0, len(mods)+1)
	for _, candidate := range order {
		if seen[candidate] {
			ordered = append(ordered, candidate)
		}
	}
	ordered = append(ordered, key)
	return strings.Join(ordered, "+")
}

func normalizeModifier(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "ctrl", "control":
		return "Ctrl"
	case "alt", "option":
		return "Alt"
	case "shift":
		return "Shift"
	case "win", "meta", "cmd", "command", "super":
		return "Win"
	default:
		return ""
	}
}

func normalizeKey(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return v
	}
	lower := strings.ToLower(v)
	aliases := map[string]string{"esc": "Escape", "escape": "Escape"}
	if alias, ok := aliases[lower]; ok {
		return alias
	}
	if len(v) == 1 {
		return strings.ToUpper(v)
	}
	return strings.ToUpper(v[:1]) + strings.ToLower(v[1:])
}

func bindingID(idx int) string { return fmt.Sprintf("binding_%d", idx+1) }
