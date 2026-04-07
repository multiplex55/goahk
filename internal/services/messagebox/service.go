package messagebox

import "context"

type Request struct {
	Title   string
	Body    string
	Icon    string
	Options string
}

type Service interface {
	Show(context.Context, Request) error
}

func NewService() Service {
	return newPlatformService()
}
