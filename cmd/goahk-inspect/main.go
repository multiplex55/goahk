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
	windowActive func(context.Context) (window.Info, error)
	windowList   func(context.Context) ([]window.Info, error)
	uiaFocused   func(context.Context) (uia.Element, error)
	uiaUnder     func(context.Context) (uia.Element, error)
	uiaTree      func(context.Context, int) (*uia.Node, error)
}

func defaultDeps() deps {
	ny := func(name string) error { return fmt.Errorf("%s not wired", name) }
	return deps{
		windowActive: func(context.Context) (window.Info, error) { return window.Info{}, ny("window active") },
		windowList:   func(context.Context) ([]window.Info, error) { return nil, ny("window list") },
		uiaFocused:   func(context.Context) (uia.Element, error) { return uia.Element{}, ny("uia focused") },
		uiaUnder:     func(context.Context) (uia.Element, error) { return uia.Element{}, ny("uia under-cursor") },
		uiaTree:      func(context.Context, int) (*uia.Node, error) { return nil, ny("uia tree") },
	}
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
		w, err := d.windowActive(ctx)
		if err != nil {
			return err
		}
		return emit(out, format, w, func() string {
			return fmt.Sprintf("HWND: %s\nTitle: %s\nClass: %s\nPID: %d\nExe: %s\nActive: %t", w.HWND, w.Title, w.Class, w.PID, w.Exe, w.Active)
		})
	case "list":
		ws, err := d.windowList(ctx)
		if err != nil {
			return err
		}
		return emit(out, format, ws, func() string {
			lines := make([]string, 0, len(ws))
			for _, w := range ws {
				lines = append(lines, fmt.Sprintf("%s | %s | %s | pid=%d | active=%t", w.HWND, w.Title, w.Class, w.PID, w.Active))
			}
			return strings.Join(lines, "\n")
		})
	default:
		return fmt.Errorf("unknown window subcommand %q", action)
	}
}

func runUIA(ctx context.Context, action, format string, out io.Writer, args []string, d deps) error {
	switch action {
	case "focused":
		el, err := d.uiaFocused(ctx)
		if err != nil {
			return err
		}
		return emit(out, format, el, func() string { return uia.FormatElementText(el) })
	case "under-cursor":
		el, err := d.uiaUnder(ctx)
		if err != nil {
			return err
		}
		return emit(out, format, el, func() string { return uia.FormatElementText(el) })
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
		node, err := d.uiaTree(ctx, *depth)
		if err != nil {
			return err
		}
		return emit(out, format, node, func() string { return uia.FormatTreeText(node) })
	default:
		return fmt.Errorf("unknown uia subcommand %q", action)
	}
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
}
