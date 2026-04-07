package flow

import (
	"context"
	"time"
)

func withTimeout(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout <= 0 {
		return parent, func() {}
	}
	return context.WithTimeout(parent, timeout)
}

func effectiveTimeout(flowTimeout, stepTimeout time.Duration) time.Duration {
	if stepTimeout > 0 {
		if flowTimeout <= 0 || stepTimeout < flowTimeout {
			return stepTimeout
		}
	}
	return flowTimeout
}
