package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"goahk/internal/uia"
	"goahk/internal/window"
)

func main() {
	if err := run(context.Background(), os.Args[1:], os.Stdout, os.Stderr, defaultDeps()); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

type deps struct {
	window windowProvider
	uia    uiaProvider
}

type windowProvider interface {
	Active(context.Context) (window.Info, error)
	List(context.Context) ([]window.Info, error)
}

type uiaProvider interface {
	Focused(context.Context) (uia.Element, error)
	UnderCursor(context.Context) (uia.Element, error)
	ActiveWindowTree(context.Context, int) (*uia.Node, error)
}

func defaultDeps() deps {
	osWindow := window.NewOSProvider()
	return deps{
		window: osWindowProvider{provider: osWindow},
		uia:    uia.NewOSInspectProvider(),
	}
}

type osWindowProvider struct {
	provider *window.OSProvider
}

func (p osWindowProvider) Active(ctx context.Context) (window.Info, error) {
	if p.provider == nil {
		return window.Info{}, errors.New("window provider is not configured")
	}
	return window.Active(ctx, p.provider)
}

func (p osWindowProvider) List(ctx context.Context) ([]window.Info, error) {
	if p.provider == nil {
		return nil, errors.New("window provider is not configured")
	}
	return window.Enumerate(ctx, p.provider)
}

func run(ctx context.Context, args []string, stdout, stderr io.Writer, d deps) error {
	format := "text"
	args, formatErr := parseGlobal(args, &format)
	if formatErr != nil {
		return formatErr
	}
	if len(args) < 2 {
		printUsage(stderr)
		return errors.New("missing command")
	}

	scope, action := args[0], args[1]
	switch scope {
	case "window":
		return runWindow(ctx, action, format, stdout, d)
	case "uia":
		return runUIA(ctx, action, format, stdout, args[2:], d)
	default:
		printUsage(stderr)
		return fmt.Errorf("unknown scope %q", scope)
	}
}

func parseGlobal(args []string, format *string) ([]string, error) {
	out := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		if args[i] == "--format" {
			if i+1 >= len(args) {
				return nil, errors.New("--format requires a value")
			}
			*format = args[i+1]
			i++
			continue
		}
		if strings.HasPrefix(args[i], "--format=") {
			*format = strings.TrimPrefix(args[i], "--format=")
			continue
		}
		out = append(out, args[i])
	}
	if *format != "text" && *format != "json" {
		return nil, fmt.Errorf("unsupported format %q", *format)
	}
	return out, nil
}

func runWindow(ctx context.Context, action, format string, out io.Writer, d deps) error {
	switch action {
	case "active":
		w, err := d.window.Active(ctx)
		if err != nil {
			return mapOpError("window active", err)
		}
		return emit(out, format, w, func() string { return formatWindowText(w) })
	case "list":
		ws, err := d.window.List(ctx)
		if err != nil {
			return mapOpError("window list", err)
		}
		return emit(out, format, ws, func() string { return formatWindowListText(ws) })
	default:
		return fmt.Errorf("unknown window subcommand %q", action)
	}
}

func runUIA(ctx context.Context, action, format string, out io.Writer, args []string, d deps) error {
	switch action {
	case "focused":
		copyBestSelector, parseErr := parseCopyBestSelectorFlag(args)
		if parseErr != nil {
			return parseErr
		}
		el, err := d.uia.Focused(ctx)
		if err != nil {
			return mapOpError("uia focused", err)
		}
		result := uia.BuildInspectResult(el)
		return emit(out, format, result, func() string {
			return formatUIAText(result, copyBestSelector)
		})
	case "under-cursor":
		copyBestSelector, parseErr := parseCopyBestSelectorFlag(args)
		if parseErr != nil {
			return parseErr
		}
		el, err := d.uia.UnderCursor(ctx)
		if err != nil {
			return mapOpError("uia under-cursor", err)
		}
		result := uia.BuildInspectResult(el)
		return emit(out, format, result, func() string {
			return formatUIAText(result, copyBestSelector)
		})
	case "tree":
		fs := flag.NewFlagSet("uia tree", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		activeWindow := fs.Bool("active-window", false, "inspect active window")
		depth := fs.Int("depth", 3, "maximum depth")
		if err := fs.Parse(args); err != nil {
			return err
		}
		if !*activeWindow {
			return errors.New("uia tree currently requires --active-window")
		}
		node, err := d.uia.ActiveWindowTree(ctx, *depth)
		if err != nil {
			return mapOpError("uia tree", err)
		}
		return emit(out, format, node, func() string { return uia.FormatTreeText(node) })
	default:
		return fmt.Errorf("unknown uia subcommand %q", action)
	}
}

func mapOpError(operation string, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, window.ErrUnsupportedPlatform) || errors.Is(err, uia.ErrUnsupportedPlatform) {
		return fmt.Errorf("%s: unsupported platform", operation)
	}
	if errors.Is(err, uia.ErrInspectUnavailable) {
		return fmt.Errorf("%s: ui automation backend unavailable", operation)
	}
	return fmt.Errorf("%s: %w", operation, err)
}

func emit(w io.Writer, format string, value any, text func() string) error {
	if format == "json" {
		b, err := json.MarshalIndent(value, "", "  ")
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(w, string(b))
		return err
	}
	_, err := fmt.Fprintln(w, text())
	return err
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "usage: goahk-inspect [--format text|json] <window|uia> <subcommand>")
	fmt.Fprintln(w, "  window active")
	fmt.Fprintln(w, "  window list")
	fmt.Fprintln(w, "  uia focused")
	fmt.Fprintln(w, "  uia under-cursor")
	fmt.Fprintln(w, "  uia tree --active-window [--depth N]")
	fmt.Fprintln(w, "  uia <focused|under-cursor> [--copy-best-selector]")
}

func formatWindowText(w window.Info) string {
	return strings.Join([]string{
		fmt.Sprintf("HWND: %s", w.HWND),
		fmt.Sprintf("Title: %s", w.Title),
		fmt.Sprintf("Class: %s", w.Class),
		fmt.Sprintf("PID: %d", w.PID),
		fmt.Sprintf("Exe: %s", emptyOrMissing(w.Exe)),
		fmt.Sprintf("ProcessPath: %s", emptyOrMissing(w.ProcessPath)),
		fmt.Sprintf("ProcessPathStatus: %s", emptyOrMissing(w.ProcessPathStatus)),
		fmt.Sprintf("ProcessPathError: %s", emptyOrMissing(w.ProcessPathError)),
		fmt.Sprintf("Active: %t", w.Active),
		fmt.Sprintf("Visible: %s", optionalBool(w.Visible)),
		fmt.Sprintf("Minimized: %s", optionalBool(w.Minimized)),
		fmt.Sprintf("Maximized: %s", optionalBool(w.Maximized)),
		fmt.Sprintf("Rect: %s", formatRect(w.Rect)),
	}, "\n")
}

func formatWindowListText(ws []window.Info) string {
	lines := make([]string, 0, len(ws))
	for _, w := range ws {
		lines = append(lines, fmt.Sprintf("%s | %s | %s | pid=%d | active=%t | visible=%s | minimized=%s | maximized=%s | rect=%s | exe=%s | processPath=%s | processPathStatus=%s",
			w.HWND, w.Title, w.Class, w.PID, w.Active, optionalBool(w.Visible), optionalBool(w.Minimized), optionalBool(w.Maximized), formatRect(w.Rect), emptyOrMissing(w.Exe), emptyOrMissing(w.ProcessPath), emptyOrMissing(w.ProcessPathStatus)))
	}
	return strings.Join(lines, "\n")
}

func parseCopyBestSelectorFlag(args []string) (bool, error) {
	fs := flag.NewFlagSet("uia inspect", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	copyBestSelector := fs.Bool("copy-best-selector", false, "emit best selector JSON at end for easy copy")
	if err := fs.Parse(args); err != nil {
		return false, err
	}
	return *copyBestSelector, nil
}

func formatUIAText(result uia.InspectResult, copyBestSelector bool) string {
	text := uia.FormatElementText(result.Element)
	if copyBestSelector && result.BestSelector != nil {
		raw, _ := json.Marshal(result.BestSelector)
		text += fmt.Sprintf("\nBestSelectorJSON: %s", string(raw))
	}
	return text
}

func optionalBool(v *bool) string {
	if v == nil {
		return "(unknown)"
	}
	return fmt.Sprintf("%t", *v)
}

func formatRect(rect *window.Rect) string {
	if rect == nil {
		return "(unknown)"
	}
	return fmt.Sprintf("%d,%d,%d,%d", rect.Left, rect.Top, rect.Right, rect.Bottom)
}

func emptyOrMissing(v string) string {
	if strings.TrimSpace(v) == "" {
		return "(missing)"
	}
	return v
}
