package folders

import "context"

type FolderInfo struct {
	Path       string `json:"path"`
	Title      string `json:"title"`
	PID        uint32 `json:"pid"`
	HWND       string `json:"hwnd"`
	Diagnostic string `json:"diagnostic,omitempty"`
}

type Service interface {
	ListOpenFolders(context.Context) ([]FolderInfo, error)
}

func NewService() Service {
	return newPlatformService()
}
