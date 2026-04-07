package process

import "context"

type Request struct {
	Executable string
	Args       []string
	WorkingDir string
	Env        map[string]string
}

type Service interface {
	Launch(context.Context, Request) error
}

func NewService() Service {
	return newPlatformService()
}
