package goahk

import (
	"fmt"
	"strings"
)

type BindingBuilder struct {
	app    *App
	hotkey string
}

func (a *App) On(hotkey string) *BindingBuilder {
	return &BindingBuilder{app: a, hotkey: normalizeHotkey(hotkey)}
}

func (a *App) Bind(hotkey string, actions ...Action) *App {
	return a.On(hotkey).Do(actions...)
}

func (b *BindingBuilder) Do(actions ...Action) *App {
	copied := make([]Action, len(actions))
	copy(copied, actions)
	b.app.bindings = append(b.app.bindings, bindingSpec{hotkey: b.hotkey, actions: copied})
	return b.app
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
