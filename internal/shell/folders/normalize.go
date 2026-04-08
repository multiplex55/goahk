package folders

import (
	"fmt"
	"sort"
	"strings"
)

type rawFolderInfo struct {
	Path       string
	Title      string
	PID        uint32
	HWND       string
	Diagnostic string
}

func normalizeAndDedupe(items []rawFolderInfo, dedupe bool) []FolderInfo {
	out := make([]FolderInfo, 0, len(items))
	for _, item := range items {
		path := strings.TrimSpace(item.Path)
		hwnd := normalizeHWND(item.HWND)
		title := strings.TrimSpace(item.Title)
		diagnostic := strings.TrimSpace(item.Diagnostic)
		if path == "" {
			if diagnostic == "" {
				continue
			}
			out = append(out, FolderInfo{Title: title, PID: item.PID, HWND: hwnd, Diagnostic: diagnostic})
			continue
		}
		out = append(out, FolderInfo{Path: path, Title: title, PID: item.PID, HWND: hwnd, Diagnostic: diagnostic})
	}

	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Path != out[j].Path {
			return out[i].Path < out[j].Path
		}
		if out[i].HWND != out[j].HWND {
			return out[i].HWND < out[j].HWND
		}
		if out[i].PID != out[j].PID {
			return out[i].PID < out[j].PID
		}
		return out[i].Title < out[j].Title
	})

	if !dedupe {
		return out
	}
	seen := map[string]struct{}{}
	deduped := make([]FolderInfo, 0, len(out))
	for _, item := range out {
		key := strings.ToLower(item.Path)
		if key == "" {
			key = fmt.Sprintf("diag|%s|%d|%s", item.Diagnostic, item.PID, item.HWND)
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		deduped = append(deduped, item)
	}
	return deduped
}

func normalizeHWND(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if strings.HasPrefix(raw, "0x") || strings.HasPrefix(raw, "0X") {
		return "0x" + strings.TrimPrefix(strings.TrimPrefix(raw, "0x"), "0X")
	}
	return raw
}
