package goahk

type clipboardService struct {
	ctx *Context
}

func (s clipboardService) ReadText() (string, error) {
	if s.ctx == nil || s.ctx.actionCtx == nil || s.ctx.actionCtx.Services.Clipboard == nil {
		return "", nil
	}
	return s.ctx.actionCtx.Services.Clipboard.ReadText(s.ctx.Context())
}

func (s clipboardService) WriteText(text string) error {
	if s.ctx == nil || s.ctx.actionCtx == nil || s.ctx.actionCtx.Services.Clipboard == nil {
		return nil
	}
	return s.ctx.actionCtx.Services.Clipboard.WriteText(s.ctx.Context(), text)
}

func (s clipboardService) AppendText(text string) error {
	if s.ctx == nil || s.ctx.actionCtx == nil || s.ctx.actionCtx.Services.Clipboard == nil {
		return nil
	}
	return s.ctx.actionCtx.Services.Clipboard.AppendText(s.ctx.Context(), text)
}

func (s clipboardService) PrependText(text string) error {
	if s.ctx == nil || s.ctx.actionCtx == nil || s.ctx.actionCtx.Services.Clipboard == nil {
		return nil
	}
	return s.ctx.actionCtx.Services.Clipboard.PrependText(s.ctx.Context(), text)
}
