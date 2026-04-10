package goahk

import "goahk/internal/actions"

type Logger interface {
	Info(msg string, fields map[string]any)
}

type noopLogger struct{}

func (noopLogger) Info(string, map[string]any) {}

type actionLoggerAdapter struct {
	logger Logger
}

func (l actionLoggerAdapter) Info(msg string, fields map[string]any) {
	if l.logger == nil {
		return
	}
	l.logger.Info(msg, copyLogFields(fields))
}

func (l actionLoggerAdapter) Error(msg string, fields map[string]any) {
	if l.logger == nil {
		return
	}
	out := copyLogFields(fields)
	if _, ok := out["level"]; !ok {
		out["level"] = "error"
	}
	l.logger.Info(msg, out)
}

func copyLogFields(fields map[string]any) map[string]any {
	if len(fields) == 0 {
		return map[string]any{}
	}
	out := make(map[string]any, len(fields))
	for key, value := range fields {
		out[key] = value
	}
	return out
}

type Option func(*App)

func WithLogger(logger Logger) Option {
	return func(a *App) { a.logger = logger }
}

func (a *App) configuredLogger() Logger {
	if a == nil || a.logger == nil {
		return noopLogger{}
	}
	return a.logger
}

func (a *App) actionLogger() actions.Logger {
	return actionLoggerAdapter{logger: a.configuredLogger()}
}

func WithActionValidation(enabled bool) Option {
	return func(a *App) { a.validateActions = enabled }
}
