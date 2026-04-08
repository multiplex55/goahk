package hotkey

import (
	"fmt"
	"sort"
	"strings"
)

type Modifiers uint8

const (
	ModCtrl Modifiers = 1 << iota
	ModAlt
	ModShift
	ModWin
)

type Chord struct {
	Modifiers Modifiers
	Key       string
}

func (c Chord) String() string {
	parts := make([]string, 0, 5)
	if c.Modifiers&ModCtrl != 0 {
		parts = append(parts, "Ctrl")
	}
	if c.Modifiers&ModAlt != 0 {
		parts = append(parts, "Alt")
	}
	if c.Modifiers&ModShift != 0 {
		parts = append(parts, "Shift")
	}
	if c.Modifiers&ModWin != 0 {
		parts = append(parts, "Win")
	}
	parts = append(parts, c.Key)
	return strings.Join(parts, "+")
}

func Parse(input string) (Chord, error) {
	if strings.TrimSpace(input) == "" {
		return Chord{}, fmt.Errorf("hotkey is empty")
	}
	tokens := strings.Split(input, "+")
	if len(tokens) == 1 {
		key, err := normalizeKey(normalizeToken(tokens[0]))
		if err != nil {
			return Chord{}, err
		}
		return Chord{Key: key}, nil
	}

	var chord Chord
	for i, raw := range tokens {
		tok := normalizeToken(raw)
		if tok == "" {
			return Chord{}, fmt.Errorf("empty token in hotkey: %q", input)
		}
		if mod, ok := parseModifier(tok); ok {
			if chord.Modifiers&mod != 0 {
				return Chord{}, fmt.Errorf("duplicate modifier %q", raw)
			}
			chord.Modifiers |= mod
			continue
		}
		if i != len(tokens)-1 {
			return Chord{}, fmt.Errorf("key must be the final token: %q", raw)
		}
		key, err := normalizeKey(tok)
		if err != nil {
			return Chord{}, err
		}
		chord.Key = key
	}

	if chord.Key == "" {
		return Chord{}, fmt.Errorf("missing non-modifier key")
	}
	return chord, nil
}

func normalizeToken(v string) string {
	return strings.ToLower(strings.TrimSpace(v))
}

func parseModifier(tok string) (Modifiers, bool) {
	switch tok {
	case "ctrl", "control":
		return ModCtrl, true
	case "alt", "option":
		return ModAlt, true
	case "shift":
		return ModShift, true
	case "win", "meta", "cmd", "command", "super":
		return ModWin, true
	default:
		return 0, false
	}
}

func normalizeKey(tok string) (string, error) {
	aliases := map[string]string{
		"esc":         "Escape",
		"escape":      "Escape",
		"enter":       "Enter",
		"return":      "Enter",
		"del":         "Delete",
		"delete":      "Delete",
		"space":       "Space",
		"spacebar":    "Space",
		"backspace":   "Backspace",
		"tab":         "Tab",
		"up":          "Up",
		"down":        "Down",
		"left":        "Left",
		"right":       "Right",
		"pageup":      "PageUp",
		"pagedown":    "PageDown",
		"home":        "Home",
		"end":         "End",
		"insert":      "Insert",
		"printscreen": "PrintScreen",
	}
	if val, ok := aliases[tok]; ok {
		return val, nil
	}

	if len(tok) == 1 {
		c := tok[0]
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') {
			return strings.ToUpper(tok), nil
		}
	}
	if strings.HasPrefix(tok, "f") {
		num := strings.TrimPrefix(tok, "f")
		switch num {
		case "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12":
			return "F" + num, nil
		}
	}

	return "", fmt.Errorf("unsupported key %q", tok)
}

func Sort(chords []Chord) {
	sort.Slice(chords, func(i, j int) bool { return chords[i].String() < chords[j].String() })
}
