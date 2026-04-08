package actions

import (
	"encoding/json"
	"fmt"
	"strings"

	"goahk/internal/shell/folders"
)

func runWindowListOpenFoldersAction(ctx ActionContext, step Step) error {
	if ctx.Services.FolderList == nil {
		return missingServiceError(step, "folders")
	}

	saveAs := strings.TrimSpace(step.Params["save_as"])
	if saveAs == "" {
		return fmt.Errorf("window.list_open_folders requires save_as")
	}

	items, err := ctx.Services.FolderList(ctx.Context)
	if err != nil {
		return err
	}
	if parseBoolDefault(step.Params["dedupe"], false) {
		items = dedupeFolderInfos(items)
	}

	raw, err := json.Marshal(items)
	if err != nil {
		return fmt.Errorf("window.list_open_folders marshal result: %w", err)
	}
	if ctx.Metadata == nil {
		ctx.Metadata = map[string]string{}
	}
	ctx.Metadata[saveAs] = string(raw)
	return nil
}

func dedupeFolderInfos(items []folders.FolderInfo) []folders.FolderInfo {
	seen := map[string]struct{}{}
	out := make([]folders.FolderInfo, 0, len(items))
	for _, item := range items {
		key := strings.ToLower(strings.TrimSpace(item.Path))
		if key == "" {
			key = strings.ToLower(strings.TrimSpace(item.HWND)) + "|" + item.Diagnostic
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, item)
	}
	return out
}
