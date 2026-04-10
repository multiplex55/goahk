package goahk

import (
	"time"

	"goahk/internal/actions"
)

type runtimeService struct {
	ctx *Context
}

func (s runtimeService) Stop() {
	if s.ctx == nil || s.ctx.actionCtx == nil {
		return
	}
	actions.RequestRuntimeStop(s.ctx.actionCtx, "runtime.stop")
}

func (s runtimeService) Sleep(duration time.Duration) bool {
	if duration <= 0 {
		return true
	}
	if s.ctx == nil {
		time.Sleep(duration)
		return true
	}
	return s.ctx.Sleep(duration)
}
