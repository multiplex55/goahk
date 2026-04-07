package actions

import (
	"fmt"
	"strconv"
	"strings"

	"goahk/internal/clipboard"
)

type Registry struct {
	handlers map[string]Handler
}

func NewRegistry() *Registry {
	r := &Registry{handlers: map[string]Handler{}}
	r.registerBuiltins()
	return r
}

func (r *Registry) Register(name string, handler Handler) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("action name is required")
	}
	if handler == nil {
		return fmt.Errorf("handler for %q is nil", name)
	}
	if _, exists := r.handlers[name]; exists {
		return fmt.Errorf("action %q already registered", name)
	}
	r.handlers[name] = handler
	return nil
}

func (r *Registry) MustRegister(name string, handler Handler) {
	if err := r.Register(name, handler); err != nil {
		panic(err)
	}
}

func (r *Registry) Lookup(name string) (Handler, bool) {
	h, ok := r.handlers[name]
	return h, ok
}

func (r *Registry) registerBuiltins() {
	r.MustRegister("system.log", func(ctx ActionContext, step Step) error {
		if ctx.Logger == nil {
			ctx.Logger = NoopLogger{}
		}
		msg := step.Params["message"]
		if msg == "" {
			msg = "system.log"
		}
		ctx.Logger.Info(msg, map[string]any{"action": step.Name, "binding_id": ctx.BindingID})
		return nil
	})

	r.MustRegister("system.message_box", func(ctx ActionContext, step Step) error {
		if ctx.Services.MessageBox == nil {
			return nil
		}
		title := step.Params["title"]
		text := step.Params["text"]
		if text == "" {
			text = step.Params["message"]
		}
		return ctx.Services.MessageBox(ctx.Context, title, text)
	})

	r.MustRegister("clipboard.read", func(ctx ActionContext, step Step) error {
		if ctx.Services.Clipboard == nil {
			return nil
		}
		text, err := ctx.Services.Clipboard.ReadText(ctx.Context)
		if err != nil {
			return err
		}
		if key := strings.TrimSpace(step.Params["save_as"]); key != "" {
			if ctx.Metadata == nil {
				ctx.Metadata = map[string]string{}
			}
			ctx.Metadata[key] = text
		}
		return nil
	})

	r.MustRegister("clipboard.write", func(ctx ActionContext, step Step) error {
		if ctx.Services.Clipboard == nil {
			return nil
		}
		writer := func() error { return ctx.Services.Clipboard.WriteText(ctx.Context, step.Params["text"]) }
		if withRestore(step) {
			return runWithClipboardRestore(ctx, writer)
		}
		return writer()
	})

	r.MustRegister("clipboard.append", func(ctx ActionContext, step Step) error {
		if ctx.Services.Clipboard == nil {
			return nil
		}
		runner := func() error { return ctx.Services.Clipboard.AppendText(ctx.Context, step.Params["text"]) }
		if withRestore(step) {
			return runWithClipboardRestore(ctx, runner)
		}
		return runner()
	})

	r.MustRegister("clipboard.prepend", func(ctx ActionContext, step Step) error {
		if ctx.Services.Clipboard == nil {
			return nil
		}
		runner := func() error { return ctx.Services.Clipboard.PrependText(ctx.Context, step.Params["text"]) }
		if withRestore(step) {
			return runWithClipboardRestore(ctx, runner)
		}
		return runner()
	})

	r.MustRegister("process.launch", func(ctx ActionContext, step Step) error {
		if ctx.Services.ProcessLaunch == nil {
			return nil
		}
		var args []string
		if raw := strings.TrimSpace(step.Params["args"]); raw != "" {
			args = strings.Fields(raw)
		}
		return ctx.Services.ProcessLaunch(ctx.Context, step.Params["path"], args)
	})

	r.MustRegister("window.activate", func(ctx ActionContext, step Step) error {
		if ctx.Services.WindowActivate == nil {
			return nil
		}
		return ctx.Services.WindowActivate(ctx.Context, step.Params["matcher"])
	})

	r.MustRegister("window.copy_active_title_to_clipboard", func(ctx ActionContext, step Step) error {
		if ctx.Services.ActiveWindowTitle == nil || ctx.Services.Clipboard == nil {
			return nil
		}
		title, err := ctx.Services.ActiveWindowTitle(ctx.Context)
		if err != nil {
			return err
		}
		return ctx.Services.Clipboard.WriteText(ctx.Context, title)
	})

	r.MustRegister("input.send_text", func(ctx ActionContext, step Step) error {
		if ctx.Services.InputSendText == nil {
			return nil
		}
		return ctx.Services.InputSendText(ctx.Context, step.Params["text"])
	})
}

func withRestore(step Step) bool {
	raw := strings.TrimSpace(step.Params["with_restore"])
	if raw == "" {
		return false
	}
	v, err := strconv.ParseBool(raw)
	return err == nil && v
}

func runWithClipboardRestore(ctx ActionContext, run func() error) error {
	prior, err := ctx.Services.Clipboard.ReadText(ctx.Context)
	if err != nil {
		return err
	}
	if err := run(); err != nil {
		return err
	}
	return ctx.Services.Clipboard.WriteText(ctx.Context, clipboard.NormalizeWriteText(prior))
}
