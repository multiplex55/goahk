package goahk

import (
	"strconv"

	"goahk/internal/program"
)

type Action struct {
	name   string
	params map[string]string
}

func (a Action) stepSpec() program.StepSpec {
	params := make(map[string]any, len(a.params))
	for k, v := range a.params {
		params[k] = v
	}
	return program.StepSpec{Action: a.name, Params: params}
}

func MessageBox(title, body string) Action {
	return Action{name: "system.message_box", params: map[string]string{"title": title, "body": body}}
}

func ClipboardWrite(text string) Action {
	return Action{name: "clipboard.write", params: map[string]string{"text": text}}
}

func Launch(executable string) Action {
	return Action{name: "process.launch", params: map[string]string{"executable": executable}}
}

func Open(target string) Action {
	return Action{name: "system.open", params: map[string]string{"target": target}}
}

func OpenURL(target string) Action {
	return Action{name: "system.open", params: map[string]string{"target": target, "kind": "url"}}
}

func OpenFolder(target string) Action {
	return Action{name: "system.open", params: map[string]string{"target": target, "kind": "folder"}}
}

func StartApplication(executable string) Action {
	return Action{name: "system.open", params: map[string]string{"target": executable, "kind": "application"}}
}

func ActivateWindow(matcher string) Action {
	return Action{name: "window.activate", params: map[string]string{"matcher": matcher}}
}

// ListOpenApplications stores window inventory JSON in metadata at saveAs.
func ListOpenApplications(saveAs string) Action {
	return Action{name: "window.list_open_applications", params: map[string]string{"save_as": saveAs}}
}

// ListOpenApplicationsWithOptions stores window inventory JSON with optional background and dedupe behavior.
func ListOpenApplicationsWithOptions(saveAs string, includeBackground bool, dedupeBy string) Action {
	params := map[string]string{"save_as": saveAs, "include_background": strconv.FormatBool(includeBackground)}
	if dedupeBy != "" {
		params["dedupe_by"] = dedupeBy
	}
	return Action{name: "window.list_open_applications", params: params}
}

func SendText(text string) Action {
	return Action{name: "input.send_text", params: map[string]string{"text": text}}
}

func Log(message string) Action {
	return Action{name: "system.log", params: map[string]string{"message": message}}
}

// Stop requests runtime shutdown after the current action step completes.
// Remaining steps in the same binding are skipped.
func Stop() Action {
	return Action{name: "runtime.stop", params: map[string]string{}}
}

// Quit is an alias for Stop.
func Quit() Action {
	return Stop()
}
