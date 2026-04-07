//go:build windows

package process

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type windowsService struct{}

func newPlatformService() Service {
	return windowsService{}
}

func (windowsService) Launch(ctx context.Context, req Request) error {
	exe := strings.TrimSpace(req.Executable)
	if exe == "" {
		return fmt.Errorf("process: executable is required")
	}
	cmd := exec.CommandContext(ctx, exe, req.Args...)
	if dir := strings.TrimSpace(req.WorkingDir); dir != "" {
		cmd.Dir = dir
	}
	if len(req.Env) > 0 {
		cmd.Env = mergeEnv(req.Env)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("process launch %q: %w", exe, err)
	}
	if cmd.Process != nil {
		_ = cmd.Process.Release()
	}
	return nil
}

func mergeEnv(overrides map[string]string) []string {
	values := map[string]string{}
	for _, item := range os.Environ() {
		parts := strings.SplitN(item, "=", 2)
		if len(parts) != 2 {
			continue
		}
		values[parts[0]] = parts[1]
	}
	for k, v := range overrides {
		if strings.TrimSpace(k) == "" {
			continue
		}
		values[k] = v
	}
	out := make([]string, 0, len(values))
	for k, v := range values {
		out = append(out, k+"="+v)
	}
	return out
}
