package goahk

import "goahk/internal/input"

type inputService struct {
	ctx *Context
}

func (s inputService) SendText(text string) error {
	if s.ctx == nil || s.ctx.actionCtx == nil || s.ctx.actionCtx.Services.Input == nil {
		return nil
	}
	return s.ctx.actionCtx.Services.Input.SendText(s.ctx.Context(), text, input.SendOptions{})
}

func (s inputService) SendKeys(keys ...string) error {
	if s.ctx == nil || s.ctx.actionCtx == nil || s.ctx.actionCtx.Services.Input == nil {
		return nil
	}
	return s.ctx.actionCtx.Services.Input.SendKeys(s.ctx.Context(), input.Sequence{Tokens: tokensFromKeys(keys)}, input.SendOptions{})
}

func (s inputService) SendChord(keys ...string) error {
	if s.ctx == nil || s.ctx.actionCtx == nil || s.ctx.actionCtx.Services.Input == nil {
		return nil
	}
	return s.ctx.actionCtx.Services.Input.SendChord(s.ctx.Context(), input.Chord{Keys: keys}, input.SendOptions{})
}

func (s inputService) Paste(text string) error {
	if err := s.ctx.Clipboard.WriteText(text); err != nil {
		return err
	}
	return s.SendChord("ctrl", "v")
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
