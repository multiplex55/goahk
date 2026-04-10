package input

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	keyeventfKeyUp   uint32 = 0x0002
	keyeventfUnicode uint32 = 0x0004

	mouseeventfLeftDown   uint32 = 0x0002
	mouseeventfLeftUp     uint32 = 0x0004
	mouseeventfRightDown  uint32 = 0x0008
	mouseeventfRightUp    uint32 = 0x0010
	mouseeventfMiddleDown uint32 = 0x0020
	mouseeventfMiddleUp   uint32 = 0x0040
	mouseeventfWheel      uint32 = 0x0800
)

const (
	vkShift   uint16 = 0x10
	vkControl uint16 = 0x11
	vkMenu    uint16 = 0x12
	vkLWin    uint16 = 0x5B
)

type keyInput struct {
	vk    uint16
	scan  uint16
	flags uint32
}

type mouseInput struct {
	flags uint32
	data  uint32
}

func buildTextKeyInputs(text string) []keyInput {
	if strings.TrimSpace(text) == "" {
		return nil
	}
	out := make([]keyInput, 0, len([]rune(text))*2)
	for _, r := range text {
		s := uint16(r)
		out = append(out,
			keyInput{scan: s, flags: keyeventfUnicode},
			keyInput{scan: s, flags: keyeventfUnicode | keyeventfKeyUp},
		)
	}
	return out
}

func buildChordKeyInputs(keys []string) ([]keyInput, error) {
	if len(keys) == 0 {
		return nil, fmt.Errorf("%w: chord must contain at least one key", ErrInvalidInputArgument)
	}
	normalized := make([]string, 0, len(keys))
	for _, key := range keys {
		k := strings.ToLower(strings.TrimSpace(key))
		if k == "" {
			return nil, fmt.Errorf("%w: chord contains an empty key", ErrInvalidInputArgument)
		}
		normalized = append(normalized, k)
	}

	mods := normalized[:len(normalized)-1]
	main := normalized[len(normalized)-1]
	out := make([]keyInput, 0, len(mods)*2+2)

	for _, mod := range mods {
		vk, ok := modifierVK(mod)
		if !ok {
			return nil, fmt.Errorf("%w: unsupported modifier %q", ErrInvalidInputArgument, mod)
		}
		out = append(out, keyInput{vk: vk})
	}

	mainDown, mainUp, err := keyDownUp(main)
	if err != nil {
		return nil, err
	}
	out = append(out, mainDown, mainUp)

	for i := len(mods) - 1; i >= 0; i-- {
		vk, _ := modifierVK(mods[i])
		out = append(out, keyInput{vk: vk, flags: keyeventfKeyUp})
	}
	return out, nil
}

func keyDownUp(key string) (keyInput, keyInput, error) {
	if vk, ok := namedKeyVK(key); ok {
		down := keyInput{vk: vk}
		up := keyInput{vk: vk, flags: keyeventfKeyUp}
		return down, up, nil
	}
	runes := []rune(key)
	if len(runes) != 1 {
		return keyInput{}, keyInput{}, fmt.Errorf("%w: unsupported key %q", ErrInvalidInputArgument, key)
	}
	scan := uint16(runes[0])
	return keyInput{scan: scan, flags: keyeventfUnicode}, keyInput{scan: scan, flags: keyeventfUnicode | keyeventfKeyUp}, nil
}

func modifierVK(k string) (uint16, bool) {
	switch k {
	case "ctrl", "control":
		return vkControl, true
	case "alt":
		return vkMenu, true
	case "shift":
		return vkShift, true
	case "win", "lwin", "rwin", "meta":
		return vkLWin, true
	default:
		return 0, false
	}
}

func namedKeyVK(key string) (uint16, bool) {
	switch key {
	case "enter", "return":
		return 0x0D, true
	case "tab":
		return 0x09, true
	case "esc", "escape":
		return 0x1B, true
	case "space":
		return 0x20, true
	case "backspace":
		return 0x08, true
	case "delete", "del":
		return 0x2E, true
	case "up":
		return 0x26, true
	case "down":
		return 0x28, true
	case "left":
		return 0x25, true
	case "right":
		return 0x27, true
	case "home":
		return 0x24, true
	case "end":
		return 0x23, true
	case "pgup", "pageup":
		return 0x21, true
	case "pgdn", "pagedown":
		return 0x22, true
	}
	if strings.HasPrefix(key, "f") {
		n, err := strconv.Atoi(strings.TrimPrefix(key, "f"))
		if err == nil && n >= 1 && n <= 24 {
			return uint16(0x70 + n - 1), true
		}
	}
	r := []rune(key)
	if len(r) == 1 {
		ch := r[0]
		if ch >= 'a' && ch <= 'z' {
			return uint16(ch - 'a' + 'A'), true
		}
		if ch >= '0' && ch <= '9' {
			return uint16(ch), true
		}
	}
	return 0, false
}

func buildMouseButtonInputs(button, mode string) ([]mouseInput, error) {
	down, up, err := mouseButtonFlags(button)
	if err != nil {
		return nil, err
	}
	switch mode {
	case "down":
		return []mouseInput{{flags: down}}, nil
	case "up":
		return []mouseInput{{flags: up}}, nil
	case "click":
		return []mouseInput{{flags: down}, {flags: up}}, nil
	case "double":
		return []mouseInput{{flags: down}, {flags: up}, {flags: down}, {flags: up}}, nil
	default:
		return nil, fmt.Errorf("%w: unsupported mouse mode %q", ErrInvalidInputArgument, mode)
	}
}

func mouseButtonFlags(button string) (down uint32, up uint32, err error) {
	switch button {
	case MouseButtonLeft:
		return mouseeventfLeftDown, mouseeventfLeftUp, nil
	case MouseButtonRight:
		return mouseeventfRightDown, mouseeventfRightUp, nil
	case MouseButtonMiddle:
		return mouseeventfMiddleDown, mouseeventfMiddleUp, nil
	default:
		return 0, 0, fmt.Errorf("%w: unsupported mouse button %q", ErrInvalidInputArgument, button)
	}
}
