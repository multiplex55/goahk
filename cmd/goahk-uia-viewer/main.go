package main

import "goahk/internal/inspect"

func main() {
	_ = NewViewerApp(inspect.NewService())
	// Wails runtime bootstrap is intentionally deferred to follow-up changes.
}
