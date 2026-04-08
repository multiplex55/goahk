package process

import "context"

type Request struct {
	Executable string
	Args       []string
	WorkingDir string
	Env        map[string]string
	OpenTarget string
	OpenKind   OpenKind
}

type Service interface {
	Launch(context.Context, Request) error
}

type OpenKind string

const (
	OpenKindDefault     OpenKind = ""
	OpenKindURL         OpenKind = "url"
	OpenKindFolder      OpenKind = "folder"
	OpenKindApplication OpenKind = "application"
)

func NewService() Service {
	return newPlatformService()
}
