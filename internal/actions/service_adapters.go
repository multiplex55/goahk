package actions

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
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

func runSystemOpenAction(ctx ActionContext, step Step) error {
	if ctx.Services.Process == nil {
		return missingServiceError(step, "process")
	}

	target := strings.TrimSpace(step.Params["target"])
	if target == "" {
		return fmt.Errorf("system.open requires target")
	}

	kind := strings.ToLower(strings.TrimSpace(step.Params["kind"]))
	if kind == "" {
		kind = "auto"
	}

	switch kind {
	case "url":
		normalized, err := normalizeURLTarget(target)
		if err != nil {
			return err
		}
		return ctx.Services.Process.Launch(ctx.Context, process.Request{
			OpenTarget: normalized,
			OpenKind:   process.OpenKindURL,
		})
	case "folder":
		normalized, err := normalizeFolderTarget(target)
		if err != nil {
			return err
		}
		return ctx.Services.Process.Launch(ctx.Context, process.Request{
			OpenTarget: normalized,
			OpenKind:   process.OpenKindFolder,
		})
	case "application":
		exe, err := normalizeApplicationTarget(target)
		if err != nil {
			return err
		}
		return ctx.Services.Process.Launch(ctx.Context, process.Request{
			Executable: exe,
			Args:       splitArgs(step.Params["args"]),
			WorkingDir: strings.TrimSpace(step.Params["working_dir"]),
			Env:        parseEnv(step.Params["env"]),
			OpenKind:   process.OpenKindApplication,
		})
	case "auto":
		if normalized, err := normalizeURLTarget(target); err == nil {
			return ctx.Services.Process.Launch(ctx.Context, process.Request{
				OpenTarget: normalized,
				OpenKind:   process.OpenKindURL,
			})
		}
		if normalized, err := normalizeFolderTarget(target); err == nil {
			return ctx.Services.Process.Launch(ctx.Context, process.Request{
				OpenTarget: normalized,
				OpenKind:   process.OpenKindFolder,
			})
		}
		exe, err := normalizeApplicationTarget(target)
		if err != nil {
			return fmt.Errorf("system.open could not infer kind for target %q", target)
		}
		return ctx.Services.Process.Launch(ctx.Context, process.Request{
			Executable: exe,
			Args:       splitArgs(step.Params["args"]),
			WorkingDir: strings.TrimSpace(step.Params["working_dir"]),
			Env:        parseEnv(step.Params["env"]),
			OpenKind:   process.OpenKindApplication,
		})
	default:
		return fmt.Errorf("system.open kind must be one of auto,url,folder,application")
	}
}

func normalizeURLTarget(target string) (string, error) {
	target = strings.TrimSpace(target)
	if target == "" {
		return "", fmt.Errorf("system.open url target is required")
	}
	candidate := target
	if !strings.Contains(candidate, "://") && looksLikeHost(candidate) {
		candidate = "https://" + candidate
	}
	parsed, err := url.ParseRequestURI(candidate)
	if err != nil {
		return "", fmt.Errorf("system.open invalid url %q: %w", target, err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("system.open invalid url %q: only http/https are supported", target)
	}
	if strings.TrimSpace(parsed.Host) == "" {
		return "", fmt.Errorf("system.open invalid url %q: host is required", target)
	}
	return parsed.String(), nil
}

func looksLikeHost(value string) bool {
	if strings.ContainsAny(value, " \t\r\n\\") {
		return false
	}
	return strings.Contains(value, ".")
}

func normalizeFolderTarget(target string) (string, error) {
	resolved := resolveFolderAlias(strings.TrimSpace(target))
	if !filepath.IsAbs(resolved) {
		return "", fmt.Errorf("system.open folder target must be an absolute path")
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return "", fmt.Errorf("system.open folder %q: %w", resolved, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("system.open folder %q is not a directory", resolved)
	}
	return resolved, nil
}

func resolveFolderAlias(target string) string {
	switch strings.ToLower(target) {
	case "downloads":
		home := os.Getenv("USERPROFILE")
		if strings.TrimSpace(home) == "" {
			home, _ = os.UserHomeDir()
		}
		if strings.TrimSpace(home) == "" {
			return target
		}
		return filepath.Join(home, "Downloads")
	default:
		return target
	}
}

func normalizeApplicationTarget(target string) (string, error) {
	target = strings.TrimSpace(target)
	if target == "" {
		return "", fmt.Errorf("system.open application target is required")
	}
	if !filepath.IsAbs(target) {
		return "", fmt.Errorf("system.open application target must be an absolute .exe path")
	}
	if strings.ToLower(filepath.Ext(target)) != ".exe" {
		return "", fmt.Errorf("system.open application target must be an absolute .exe path")
	}
	return target, nil
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
