//go:build windows

package main

func (ui *viewerUI) setStatus(text string) {
	if ui.statusText != nil {
		ui.statusText.SetText(text)
	}
}
