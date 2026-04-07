package clipboard

import "context"

type Backend interface {
	ReadText(context.Context) (string, error)
	WriteText(context.Context, string) error
}

type Service interface {
	ReadText(context.Context) (string, error)
	WriteText(context.Context, string) error
	AppendText(context.Context, string) error
	PrependText(context.Context, string) error
}

type service struct {
	backend Backend
}

func NewService(backend Backend) Service {
	if backend == nil {
		backend = NewPlatformBackend()
	}
	return &service{backend: backend}
}

func (s *service) ReadText(ctx context.Context) (string, error) {
	text, err := s.backend.ReadText(ctx)
	if err != nil {
		return "", err
	}
	return NormalizeReadText(text), nil
}

func (s *service) WriteText(ctx context.Context, text string) error {
	return s.backend.WriteText(ctx, NormalizeWriteText(text))
}

func (s *service) AppendText(ctx context.Context, suffix string) error {
	current, err := s.ReadText(ctx)
	if err != nil {
		return err
	}
	return s.WriteText(ctx, AppendText(current, suffix))
}

func (s *service) PrependText(ctx context.Context, prefix string) error {
	current, err := s.ReadText(ctx)
	if err != nil {
		return err
	}
	return s.WriteText(ctx, PrependText(prefix, current))
}
