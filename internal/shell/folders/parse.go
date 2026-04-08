package folders

import (
	"encoding/json"
	"fmt"
)

type powershellFolderInfo struct {
	Path       string `json:"path"`
	Title      string `json:"title"`
	PID        uint32 `json:"pid"`
	HWND       string `json:"hwnd"`
	Diagnostic string `json:"diagnostic"`
}

func parsePowerShellFolderResults(raw []byte) ([]rawFolderInfo, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	var list []powershellFolderInfo
	if err := json.Unmarshal(raw, &list); err != nil {
		var single powershellFolderInfo
		if errSingle := json.Unmarshal(raw, &single); errSingle != nil {
			return nil, fmt.Errorf("folders: parse explorer results: %w", err)
		}
		list = []powershellFolderInfo{single}
	}
	out := make([]rawFolderInfo, 0, len(list))
	for _, item := range list {
		out = append(out, rawFolderInfo{
			Path:       item.Path,
			Title:      item.Title,
			PID:        item.PID,
			HWND:       item.HWND,
			Diagnostic: item.Diagnostic,
		})
	}
	return out, nil
}
