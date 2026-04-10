package goahk

import "fmt"

func ExampleNewApp() {
	app := NewApp()
	app.Bind("Escape", ControlStop())
	fmt.Println("configured")
	// Output: configured
}

func ExampleApp_Bind() {
	app := NewApp().
		Bind("1", MessageBox("Hotkey pressed", "You hit 1"), Stop()).
		Bind("Escape", ControlStop())

	fmt.Println(len(app.toProgram().Bindings))
	// Output: 2
}

func Example_messageBoxAndStop() {
	app := NewApp().
		Bind("F1", MessageBox("Heads up", "Task completed"), Stop()).
		Bind("Escape", ControlStop())

	fmt.Println(app.toProgram().Bindings[0].Steps[0].Action)
	// Output: system.message_box
}

func Example_callbackComposition() {
	app := NewApp().Bind(
		"Ctrl+Shift+M",
		Log("before callback"),
		Func(func(ctx *Context) error {
			ctx.Vars["message"] = "from callback"
			return nil
		}),
		Log("after callback"),
	)

	_, _, callbacks := app.runtimeArtifacts()
	fmt.Println(len(callbacks))
	// Output: 1
}

func Example_sharedAppStateAndPerTriggerVars() {
	app := NewApp().Bind("Ctrl+R", Func(func(ctx *Context) error {
		count := 1
		if raw, ok := ctx.AppState.Get("count"); ok {
			fmt.Sscanf(raw, "%d", &count)
			count++
		}
		ctx.AppState.Set("count", fmt.Sprintf("%d", count))
		ctx.Vars["count"] = fmt.Sprintf("%d", count) // per-trigger metadata
		return nil
	}))

	fmt.Println(len(app.toProgram().Bindings))
	// Output: 1
}
