package actions

import (
	"encoding/json"
	"fmt"
	"strings"

	"goahk/internal/window"
)

type windowInventoryEntry struct {
	Title  string `json:"title"`
	Exe    string `json:"exe"`
	PID    uint32 `json:"pid"`
	HWND   string `json:"hwnd"`
	Class  string `json:"class"`
	Active bool   `json:"active"`
}

func runWindowListOpenApplicationsAction(ctx ActionContext, step Step) error {
	if ctx.Services.WindowList == nil {
		return missingServiceError(step, "window")
	}

	saveAs := strings.TrimSpace(step.Params["save_as"])
	if saveAs == "" {
		return fmt.Errorf("window.list_open_applications requires save_as")
	}

	windows, err := ctx.Services.WindowList(ctx.Context)
	if err != nil {
		return err
	}

	includeBackground := parseBoolDefault(step.Params["include_background"], false)
	dedupeBy := strings.ToLower(strings.TrimSpace(step.Params["dedupe_by"]))
	if dedupeBy == "" {
		dedupeBy = "process"
	}
	if dedupeBy != "process" && dedupeBy != "window" {
		return fmt.Errorf("window.list_open_applications dedupe_by must be one of process,window")
	}

	filtered := filterWindowInventory(windows, includeBackground)
	entries := dedupeWindowInventory(filtered, dedupeBy)

	raw, err := json.Marshal(entries)
	if err != nil {
		return fmt.Errorf("window.list_open_applications marshal result: %w", err)
	}
	if ctx.Metadata == nil {
		ctx.Metadata = map[string]string{}
	}
	ctx.Metadata[saveAs] = string(raw)
	return nil
}

func filterWindowInventory(windows []window.Info, includeBackground bool) []window.Info {
	out := make([]window.Info, 0, len(windows))
	for _, win := range windows {
		if strings.TrimSpace(win.Title) == "" {
			continue
		}
		if !includeBackground && !win.Active {
			continue
		}
		if isSystemShellWindow(win) {
			continue
		}
		out = append(out, win)
	}
	return out
}

func dedupeWindowInventory(windows []window.Info, dedupeBy string) []windowInventoryEntry {
	seen := map[string]struct{}{}
	entries := make([]windowInventoryEntry, 0, len(windows))
	for _, win := range windows {
		key := win.Exe + "|" + fmt.Sprintf("%d", win.PID)
		if dedupeBy == "window" {
			key = key + "|" + win.HWND.String()
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		entries = append(entries, windowInventoryEntry{
			Title:  win.Title,
			Exe:    win.Exe,
			PID:    win.PID,
			HWND:   win.HWND.String(),
			Class:  win.Class,
			Active: win.Active,
		})
	}
	return entries
}

func isSystemShellWindow(win window.Info) bool {
	className := strings.ToLower(strings.TrimSpace(win.Class))
	exe := strings.ToLower(strings.TrimSpace(win.Exe))
	title := strings.ToLower(strings.TrimSpace(win.Title))

	systemClasses := map[string]struct{}{
		"progman":                    {},
		"workerw":                    {},
		"shell_traywnd":              {},
		"shell_secondarytraywnd":     {},
		"button":                     {},
		"windows.ui.core.corewindow": {},
	}
	if _, ok := systemClasses[className]; ok {
		return true
	}
	if exe == "explorer.exe" && (title == "program manager" || strings.Contains(title, "taskbar")) {
		return true
	}
	return false
}
