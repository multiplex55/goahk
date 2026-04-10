package goahk

import (
	"fmt"

	"goahk/internal/input"
)

type inputService struct {
	ctx *Context
}

func (s inputService) SendText(text string) error {
	if s.ctx == nil || s.ctx.actionCtx == nil || s.ctx.actionCtx.Services.Input == nil {
		return fmt.Errorf("input send_text: %w", ErrInputServiceUnavailable)
	}
	return s.ctx.actionCtx.Services.Input.SendText(s.ctx.Context(), text, input.SendOptions{})
}

func (s inputService) SendKeys(keys ...string) error {
	if s.ctx == nil || s.ctx.actionCtx == nil || s.ctx.actionCtx.Services.Input == nil {
		return fmt.Errorf("input send_keys: %w", ErrInputServiceUnavailable)
	}
	return s.ctx.actionCtx.Services.Input.SendKeys(s.ctx.Context(), input.Sequence{Tokens: tokensFromKeys(keys)}, input.SendOptions{})
}

func (s inputService) SendChord(keys ...string) error {
	if s.ctx == nil || s.ctx.actionCtx == nil || s.ctx.actionCtx.Services.Input == nil {
		return fmt.Errorf("input send_chord: %w", ErrInputServiceUnavailable)
	}
	return s.ctx.actionCtx.Services.Input.SendChord(s.ctx.Context(), input.Chord{Keys: keys}, input.SendOptions{})
}

func (s inputService) MouseMoveAbsolute(x, y int) error {
	if s.ctx == nil || s.ctx.actionCtx == nil || s.ctx.actionCtx.Services.Input == nil {
		return fmt.Errorf("input mouse_move_absolute: %w", ErrInputServiceUnavailable)
	}
	return s.ctx.actionCtx.Services.Input.MoveAbsolute(s.ctx.Context(), x, y)
}

func (s inputService) MouseMoveRelative(dx, dy int) error {
	if s.ctx == nil || s.ctx.actionCtx == nil || s.ctx.actionCtx.Services.Input == nil {
		return fmt.Errorf("input mouse_move_relative: %w", ErrInputServiceUnavailable)
	}
	return s.ctx.actionCtx.Services.Input.MoveRelative(s.ctx.Context(), dx, dy)
}

func (s inputService) MousePosition() (input.MousePosition, error) {
	if s.ctx == nil || s.ctx.actionCtx == nil || s.ctx.actionCtx.Services.Input == nil {
		return input.MousePosition{}, fmt.Errorf("input mouse_position: %w", ErrInputServiceUnavailable)
	}
	return s.ctx.actionCtx.Services.Input.Position(s.ctx.Context())
}

func (s inputService) MouseButtonDown(button string) error {
	if s.ctx == nil || s.ctx.actionCtx == nil || s.ctx.actionCtx.Services.Input == nil {
		return fmt.Errorf("input mouse_button_down: %w", ErrInputServiceUnavailable)
	}
	return s.ctx.actionCtx.Services.Input.ButtonDown(s.ctx.Context(), button)
}

func (s inputService) MouseButtonUp(button string) error {
	if s.ctx == nil || s.ctx.actionCtx == nil || s.ctx.actionCtx.Services.Input == nil {
		return fmt.Errorf("input mouse_button_up: %w", ErrInputServiceUnavailable)
	}
	return s.ctx.actionCtx.Services.Input.ButtonUp(s.ctx.Context(), button)
}

func (s inputService) MouseClick(button string) error {
	if s.ctx == nil || s.ctx.actionCtx == nil || s.ctx.actionCtx.Services.Input == nil {
		return fmt.Errorf("input mouse_click: %w", ErrInputServiceUnavailable)
	}
	return s.ctx.actionCtx.Services.Input.Click(s.ctx.Context(), button)
}

func (s inputService) MouseDoubleClick(button string) error {
	if s.ctx == nil || s.ctx.actionCtx == nil || s.ctx.actionCtx.Services.Input == nil {
		return fmt.Errorf("input mouse_double_click: %w", ErrInputServiceUnavailable)
	}
	return s.ctx.actionCtx.Services.Input.DoubleClick(s.ctx.Context(), button)
}

func (s inputService) MouseWheel(delta int) error {
	if s.ctx == nil || s.ctx.actionCtx == nil || s.ctx.actionCtx.Services.Input == nil {
		return fmt.Errorf("input mouse_wheel: %w", ErrInputServiceUnavailable)
	}
	return s.ctx.actionCtx.Services.Input.Wheel(s.ctx.Context(), delta)
}

func (s inputService) MouseDrag(button string, startX, startY, endX, endY int) error {
	if s.ctx == nil || s.ctx.actionCtx == nil || s.ctx.actionCtx.Services.Input == nil {
		return fmt.Errorf("input mouse_drag: %w", ErrInputServiceUnavailable)
	}
	return s.ctx.actionCtx.Services.Input.Drag(s.ctx.Context(), button, startX, startY, endX, endY)
}

func (s inputService) Paste(text string) error {
	if s.ctx == nil || s.ctx.actionCtx == nil || s.ctx.actionCtx.Services.Input == nil {
		return fmt.Errorf("input paste: %w", ErrInputServiceUnavailable)
	}
	if err := s.ctx.Clipboard.WriteText(text); err != nil {
		return fmt.Errorf("input paste clipboard write: %w", err)
	}
	if err := s.SendChord("ctrl", "v"); err != nil {
		return fmt.Errorf("input paste send_chord: %w", err)
	}
	return nil
}

func tokensFromKeys(keys []string) []input.Token {
	tokens := make([]input.Token, 0, len(keys))
	for _, key := range keys {
		if key == "" {
			continue
		}
		tokens = append(tokens, input.Token{Raw: key, Keys: []string{key}})
	}
	return tokens
}
