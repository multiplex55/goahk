package actions

import (
	"fmt"
	"strings"

	"goahk/internal/process"
	"goahk/internal/services/messagebox"
)

func runMessageBoxAction(ctx ActionContext, step Step) error {
	if ctx.Services.MessageBox == nil {
		return missingServiceError(step, "message box")
	}
	body := strings.TrimSpace(step.Params["body"])
	if body == "" {
		body = strings.TrimSpace(step.Params["text"])
	}
	if body == "" {
		body = strings.TrimSpace(step.Params["message"])
	}
	if body == "" {
		return fmt.Errorf("system.message_box requires body")
	}
	req := messagebox.Request{
		Title:   step.Params["title"],
		Body:    body,
		Icon:    defaultString(step.Params["icon"], "info"),
		Options: defaultString(step.Params["options"], "ok"),
	}
	return ctx.Services.MessageBox.Show(ctx.Context, req)
}

func runClipboardWriteAction(ctx ActionContext, step Step) error {
	if ctx.Services.Clipboard == nil {
		return missingServiceError(step, "clipboard")
	}
	text, ok := step.Params["text"]
	if !ok {
		return fmt.Errorf("clipboard.write requires text")
	}
	writer := func() error { return ctx.Services.Clipboard.WriteText(ctx.Context, text) }
	if withRestore(step) {
		return runWithClipboardRestore(ctx, writer)
	}
	return writer()
}

func runProcessLaunchAction(ctx ActionContext, step Step) error {
	if ctx.Services.Process == nil {
		return missingServiceError(step, "process")
	}
	executable := strings.TrimSpace(step.Params["executable"])
	if executable == "" {
		executable = strings.TrimSpace(step.Params["path"])
	}
	if executable == "" {
		return fmt.Errorf("process.launch requires executable")
	}
	request := process.Request{
		Executable: executable,
		Args:       splitArgs(step.Params["args"]),
		WorkingDir: strings.TrimSpace(step.Params["working_dir"]),
		Env:        parseEnv(step.Params["env"]),
	}
	return ctx.Services.Process.Launch(ctx.Context, request)
}

func splitArgs(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	return strings.Fields(raw)
}

func parseEnv(raw string) map[string]string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	out := map[string]string{}
	for _, token := range strings.Split(raw, ";") {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}
		parts := strings.SplitN(token, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		if key == "" {
			continue
		}
		out[key] = strings.TrimSpace(parts[1])
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func missingServiceError(step Step, service string) error {
	return fmt.Errorf("%s: %s service unavailable", step.Name, service)
}
