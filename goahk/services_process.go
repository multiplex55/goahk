package goahk

import "goahk/internal/process"

type processService struct {
	ctx *Context
}

func (s processService) Launch(executable string, args ...string) error {
	if s.ctx == nil || s.ctx.actionCtx == nil || s.ctx.actionCtx.Services.Process == nil {
		return nil
	}
	return s.ctx.actionCtx.Services.Process.Launch(s.ctx.Context(), process.Request{Executable: executable, Args: args})
}

func (s processService) Open(target string) error {
	if s.ctx == nil || s.ctx.actionCtx == nil || s.ctx.actionCtx.Services.Process == nil {
		return nil
	}
	return s.ctx.actionCtx.Services.Process.Launch(s.ctx.Context(), processRequestForOpen(target))
}
