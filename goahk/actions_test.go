package goahk

import (
	"reflect"
	"testing"

	"goahk/internal/program"
)

func TestActionConstructorsSerializeStepSpecs(t *testing.T) {
	tests := []struct {
		name string
		in   Action
		want program.StepSpec
	}{
		{name: "message box", in: MessageBox("T", "Body"), want: program.StepSpec{Action: "system.message_box", Params: map[string]any{"title": "T", "body": "Body"}}},
		{name: "clipboard write", in: ClipboardWrite("abc"), want: program.StepSpec{Action: "clipboard.write", Params: map[string]any{"text": "abc"}}},
		{name: "clipboard append", in: ClipboardAppend("abc"), want: program.StepSpec{Action: "clipboard.append", Params: map[string]any{"text": "abc"}}},
		{name: "clipboard prepend", in: ClipboardPrepend("abc"), want: program.StepSpec{Action: "clipboard.prepend", Params: map[string]any{"text": "abc"}}},
		{name: "clipboard read", in: ClipboardRead("clip"), want: program.StepSpec{Action: "clipboard.read", Params: map[string]any{"save_as": "clip"}}},
		{name: "launch", in: Launch("notepad.exe"), want: program.StepSpec{Action: "process.launch", Params: map[string]any{"executable": "notepad.exe"}}},
		{name: "open", in: Open("www.chatgpt.com"), want: program.StepSpec{Action: "system.open", Params: map[string]any{"target": "www.chatgpt.com"}}},
		{name: "open url", in: OpenURL("https://chatgpt.com"), want: program.StepSpec{Action: "system.open", Params: map[string]any{"target": "https://chatgpt.com", "kind": "url"}}},
		{name: "open folder", in: OpenFolder(`C:\\Temp`), want: program.StepSpec{Action: "system.open", Params: map[string]any{"target": `C:\\Temp`, "kind": "folder"}}},
		{name: "start application", in: StartApplication(`C:\\Windows\\notepad.exe`), want: program.StepSpec{Action: "system.open", Params: map[string]any{"target": `C:\\Windows\\notepad.exe`, "kind": "application"}}},
		{name: "activate", in: ActivateWindow("title:Editor"), want: program.StepSpec{Action: "window.activate", Params: map[string]any{"matcher": "title:Editor"}}},
		{name: "copy active title", in: CopyActiveWindowTitle(), want: program.StepSpec{Action: "window.copy_active_title_to_clipboard", Params: map[string]any{}}},
		{name: "list open applications", in: ListOpenApplications("apps"), want: program.StepSpec{Action: "window.list_open_applications", Params: map[string]any{"save_as": "apps"}}},
		{name: "list open applications with options", in: ListOpenApplicationsWithOptions("apps", true, "window"), want: program.StepSpec{Action: "window.list_open_applications", Params: map[string]any{"save_as": "apps", "include_background": "true", "dedupe_by": "window"}}},
		{name: "list open folders", in: ListOpenFolders("folders"), want: program.StepSpec{Action: "window.list_open_folders", Params: map[string]any{"save_as": "folders"}}},
		{name: "send text", in: SendText("hello"), want: program.StepSpec{Action: "input.send_text", Params: map[string]any{"text": "hello"}}},
		{name: "send keys", in: SendKeys("ctrl+c {enter}"), want: program.StepSpec{Action: "input.send_keys", Params: map[string]any{"sequence": "ctrl+c {enter}"}}},
		{name: "send chord encoded", in: SendChord("ctrl+v"), want: program.StepSpec{Action: "input.send_chord", Params: map[string]any{"chord": "ctrl+v"}}},
		{name: "send chord keys", in: SendChord("ctrl", "v"), want: program.StepSpec{Action: "input.send_chord", Params: map[string]any{"chord": "ctrl+v"}}},
		{name: "log", in: Log("ok"), want: program.StepSpec{Action: "system.log", Params: map[string]any{"message": "ok"}}},
		{name: "stop", in: Stop(), want: program.StepSpec{Action: "runtime.stop", Params: map[string]any{}}},
		{name: "quit alias", in: Quit(), want: program.StepSpec{Action: "runtime.stop", Params: map[string]any{}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.in.stepSpec(); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("stepSpec = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestActionConstructorsEdgeCases(t *testing.T) {
	tests := []struct {
		name string
		in   Action
		want program.StepSpec
	}{
		{name: "clipboard append empty", in: ClipboardAppend(""), want: program.StepSpec{Action: "clipboard.append", Params: map[string]any{"text": ""}}},
		{name: "clipboard prepend empty", in: ClipboardPrepend(""), want: program.StepSpec{Action: "clipboard.prepend", Params: map[string]any{"text": ""}}},
		{name: "clipboard read empty save_as", in: ClipboardRead(""), want: program.StepSpec{Action: "clipboard.read", Params: map[string]any{"save_as": ""}}},
		{name: "list open folders empty save_as", in: ListOpenFolders(""), want: program.StepSpec{Action: "window.list_open_folders", Params: map[string]any{"save_as": ""}}},
		{name: "send keys empty", in: SendKeys(""), want: program.StepSpec{Action: "input.send_keys", Params: map[string]any{"sequence": ""}}},
		{name: "send chord no keys", in: SendChord(), want: program.StepSpec{Action: "input.send_chord", Params: map[string]any{"chord": ""}}},
		{name: "send chord empty keys", in: SendChord("", ""), want: program.StepSpec{Action: "input.send_chord", Params: map[string]any{"chord": "+"}}},
		{name: "list open apps no dedupe", in: ListOpenApplicationsWithOptions("apps", false, ""), want: program.StepSpec{Action: "window.list_open_applications", Params: map[string]any{"save_as": "apps", "include_background": "false"}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.in.stepSpec(); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("stepSpec = %#v, want %#v", got, tt.want)
			}
		})
	}
}
