package uia

import (
	"context"
	"time"
)

type ActionDiagnostics struct {
	RetryCount        int
	Timeout           time.Duration
	SupportedPatterns []string
	MissingPattern    string
}

type ActionService interface {
	Find(context.Context, Selector, time.Duration, time.Duration) (Element, ActionDiagnostics, error)
	Invoke(context.Context, Selector, time.Duration, time.Duration) (ActionDiagnostics, error)
	ValueSet(context.Context, Selector, string, time.Duration, time.Duration) (ActionDiagnostics, error)
	ValueGet(context.Context, Selector, time.Duration, time.Duration) (string, ActionDiagnostics, error)
	Toggle(context.Context, Selector, time.Duration, time.Duration) (ActionDiagnostics, error)
	Expand(context.Context, Selector, time.Duration, time.Duration) (ActionDiagnostics, error)
	Select(context.Context, Selector, time.Duration, time.Duration) (ActionDiagnostics, error)
}
