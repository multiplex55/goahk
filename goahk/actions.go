package goahk

import (
	"strconv"
	"strings"

	"goahk/internal/program"
)

type stepSpecProvider interface {
	stepSpec() program.StepSpec
}

// Action represents a declarative runtime action.
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

// MessageBox shows a system modal dialog with title and body text.
func MessageBox(title, body string) Action {
	return Action{name: "system.message_box", params: map[string]string{"title": title, "body": body}}
}

// ClipboardWrite replaces clipboard text with text.
func ClipboardWrite(text string) Action {
	return Action{name: "clipboard.write", params: map[string]string{"text": text}}
}

// ClipboardAppend appends text to current clipboard contents.
func ClipboardAppend(text string) Action {
	return Action{name: "clipboard.append", params: map[string]string{"text": text}}
}

// ClipboardPrepend prepends text to current clipboard contents.
func ClipboardPrepend(text string) Action {
	return Action{name: "clipboard.prepend", params: map[string]string{"text": text}}
}

// ClipboardRead stores clipboard text in metadata at saveAs.
func ClipboardRead(saveAs string) Action {
	return Action{name: "clipboard.read", params: map[string]string{"save_as": saveAs}}
}

// Launch starts an executable process.
func Launch(executable string) Action {
	return Action{name: "process.launch", params: map[string]string{"executable": executable}}
}

// Open opens a path or URL with the OS default handler.
func Open(target string) Action {
	return Action{name: "system.open", params: map[string]string{"target": target}}
}

// OpenURL opens a URL with the default browser.
func OpenURL(target string) Action {
	return Action{name: "system.open", params: map[string]string{"target": target, "kind": "url"}}
}

// OpenFolder opens a folder path in the default file explorer.
func OpenFolder(target string) Action {
	return Action{name: "system.open", params: map[string]string{"target": target, "kind": "folder"}}
}

// StartApplication opens an application executable path.
func StartApplication(executable string) Action {
	return Action{name: "system.open", params: map[string]string{"target": executable, "kind": "application"}}
}

// ActivateWindow focuses the first window matching matcher.
func ActivateWindow(matcher string) Action {
	return Action{name: "window.activate", params: map[string]string{"matcher": matcher}}
}

// CopyActiveWindowTitle copies the active window title to the clipboard.
func CopyActiveWindowTitle() Action {
	return Action{name: "window.copy_active_title_to_clipboard", params: map[string]string{}}
}

// ListOpenApplications stores window inventory JSON in metadata at saveAs.
func ListOpenApplications(saveAs string) Action {
	return Action{name: "window.list_open_applications", params: map[string]string{"save_as": saveAs}}
}

// ListOpenFolders stores open folder inventory JSON in metadata at saveAs.
func ListOpenFolders(saveAs string) Action {
	return Action{name: "window.list_open_folders", params: map[string]string{"save_as": saveAs}}
}

// ListOpenApplicationsWithOptions stores window inventory JSON with optional background and dedupe behavior.
func ListOpenApplicationsWithOptions(saveAs string, includeBackground bool, dedupeBy string) Action {
	params := map[string]string{"save_as": saveAs, "include_background": strconv.FormatBool(includeBackground)}
	if dedupeBy != "" {
		params["dedupe_by"] = dedupeBy
	}
	return Action{name: "window.list_open_applications", params: params}
}

// SendText sends literal text as keyboard input.
func SendText(text string) Action {
	return Action{name: "input.send_text", params: map[string]string{"text": text}}
}

// SendKeys encodes a key sequence using input.send_keys syntax (for example "ctrl+c {enter}").
func SendKeys(sequence string) Action {
	return Action{name: "input.send_keys", params: map[string]string{"sequence": sequence}}
}

// SendChord encodes either a preformatted chord ("ctrl+v") or individual keys ("ctrl", "v").
func SendChord(chordOrKeys ...string) Action {
	return Action{name: "input.send_chord", params: map[string]string{"chord": encodeChord(chordOrKeys...)}}
}

// Log records a message with action/system metadata.
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

func encodeChord(chordOrKeys ...string) string {
	if len(chordOrKeys) == 0 {
		return ""
	}
	if len(chordOrKeys) == 1 {
		return chordOrKeys[0]
	}
	return strings.Join(chordOrKeys, "+")
}
