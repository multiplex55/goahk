package runtime

import (
	"context"
	"go/build/constraint"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"goahk/internal/clipboard"
	"goahk/internal/input"
	"goahk/internal/services/messagebox"
	"goahk/internal/window"
)

func TestServiceContractMatrix_StubAndFakeBackends(t *testing.T) {
	t.Parallel()

	t.Run("clipboard fake contract", func(t *testing.T) {
		t.Parallel()
		svc := clipboard.NewService(clipboardMatrixFakeBackend{text: "line\x00\x00"})
		got, err := svc.ReadText(context.Background())
		if err != nil {
			t.Fatalf("ReadText() error = %v", err)
		}
		if got != "line" {
			t.Fatalf("ReadText() = %q, want line", got)
		}
	})

	if runtime.GOOS == "windows" {
		t.Skip("stub service assertions apply only to non-Windows builds")
	}

	t.Run("messagebox stub", func(t *testing.T) {
		t.Parallel()
		err := messagebox.NewService().Show(context.Background(), messagebox.Request{Body: "hello"})
		if err == nil {
			t.Fatal("expected non-windows messagebox service to return an unsupported-platform error")
		}
	})

	t.Run("window stub", func(t *testing.T) {
		t.Parallel()
		_, err := window.NewOSProvider().ActiveWindow(context.Background())
		if err == nil {
			t.Fatal("expected non-windows window provider to return an unsupported-platform error")
		}
	})

	t.Run("input stub", func(t *testing.T) {
		t.Parallel()
		err := input.NewService().SendText(context.Background(), "hello", input.SendOptions{})
		if err == nil {
			t.Fatal("expected non-windows input service to return an unsupported-platform error")
		}
		if !input.IsUnsupported(err) {
			t.Fatalf("expected IsUnsupported(err) to be true, got err=%v", err)
		}
	})
}

func TestBootstrapDefaults_UseRuntimePlatformServices(t *testing.T) {
	t.Parallel()

	b := NewBootstrap()
	if b.NewListener == nil {
		t.Fatal("NewListener must be wired to a runtime listener factory")
	}
	if b.BaseActionCtx.Services.MessageBox == nil {
		t.Fatal("MessageBox service must be configured")
	}
	if b.BaseActionCtx.Services.Clipboard == nil {
		t.Fatal("Clipboard service must be configured")
	}
	if b.BaseActionCtx.Services.Input == nil {
		t.Fatal("Input service must be configured")
	}
}

type clipboardMatrixFakeBackend struct{ text string }

func (b clipboardMatrixFakeBackend) ReadText(context.Context) (string, error) { return b.text, nil }
func (b clipboardMatrixFakeBackend) WriteText(context.Context, string) error  { return nil }

func TestBuildTagCoverage_TargetServicesHaveWindowsAndStubFiles(t *testing.T) {
	t.Parallel()

	type target struct {
		name    string
		dir     string
		windows string
		stub    string
	}

	targets := []target{
		{name: "hotkey", dir: "../hotkey", windows: "win32_listener_backend_windows.go", stub: "win32_listener_backend_stub.go"},
		{name: "messagebox", dir: "../services/messagebox", windows: "service_windows.go", stub: "service_stub.go"},
		{name: "clipboard", dir: "../clipboard", windows: "backend_windows.go", stub: "backend_stub.go"},
		{name: "window", dir: "../window", windows: "provider_windows.go", stub: "provider_stub.go"},
		{name: "input", dir: "../input", windows: "service_windows.go", stub: "service_stub.go"},
	}

	for _, tc := range targets {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			windowsPath := filepath.Clean(filepath.Join(tc.dir, tc.windows))
			stubPath := filepath.Clean(filepath.Join(tc.dir, tc.stub))
			assertFileHasBuildExpr(t, windowsPath, "windows")
			assertFileHasBuildExpr(t, stubPath, "!windows")
		})
	}
}

func assertFileHasBuildExpr(t *testing.T, path string, want string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	lines := strings.Split(string(data), "\n")

	var exprLine string
	for _, line := range lines[:min(10, len(lines))] {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//go:build ") {
			exprLine = strings.TrimPrefix(trimmed, "//go:build ")
			break
		}
	}
	if exprLine == "" {
		t.Fatalf("%s missing //go:build expression", path)
	}

	expr, err := constraint.Parse("//go:build " + exprLine)
	if err != nil {
		t.Fatalf("constraint.Parse(%q) error = %v", exprLine, err)
	}

	ok := expr.Eval(func(tag string) bool { return tag == "windows" })
	if want == "windows" && !ok {
		t.Fatalf("%s build expression %q does not evaluate true for windows", path, exprLine)
	}
	if want == "!windows" && ok {
		t.Fatalf("%s build expression %q unexpectedly evaluates true for windows", path, exprLine)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
