package goahk

type Logger interface {
	Info(msg string, fields map[string]any)
}

type Option func(*App)

func WithLogger(logger Logger) Option {
	return func(a *App) { a.logger = logger }
}

func WithActionValidation(enabled bool) Option {
	return func(a *App) { a.validateActions = enabled }
}
