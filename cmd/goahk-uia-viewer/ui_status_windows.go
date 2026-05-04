//go:build windows

package main

func (ui *viewerUI) setStatus(text string) {
	ui.dispatcher.Queue(func() {
		if ui.statusText != nil {
			ui.statusText.SetText(text)
		}
	})
}
