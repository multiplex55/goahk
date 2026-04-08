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
		{name: "launch", in: Launch("notepad.exe"), want: program.StepSpec{Action: "process.launch", Params: map[string]any{"executable": "notepad.exe"}}},
		{name: "open", in: Open("www.chatgpt.com"), want: program.StepSpec{Action: "system.open", Params: map[string]any{"target": "www.chatgpt.com"}}},
		{name: "open url", in: OpenURL("https://chatgpt.com"), want: program.StepSpec{Action: "system.open", Params: map[string]any{"target": "https://chatgpt.com", "kind": "url"}}},
		{name: "open folder", in: OpenFolder(`C:\\Temp`), want: program.StepSpec{Action: "system.open", Params: map[string]any{"target": `C:\\Temp`, "kind": "folder"}}},
		{name: "start application", in: StartApplication(`C:\\Windows\\notepad.exe`), want: program.StepSpec{Action: "system.open", Params: map[string]any{"target": `C:\\Windows\\notepad.exe`, "kind": "application"}}},
		{name: "activate", in: ActivateWindow("title:Editor"), want: program.StepSpec{Action: "window.activate", Params: map[string]any{"matcher": "title:Editor"}}},
		{name: "send text", in: SendText("hello"), want: program.StepSpec{Action: "input.send_text", Params: map[string]any{"text": "hello"}}},
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
