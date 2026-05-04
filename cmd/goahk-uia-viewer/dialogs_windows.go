//go:build windows

package main

import (
	"fmt"

	"github.com/lxn/walk"
)

type walkDialogs struct{}

func (walkDialogs) PromptSetValue(defaultValue string) (string, bool, error) {
	dlg, err := walk.NewDialog(nil)
	if err != nil {
		return "", false, fmt.Errorf("create set value dialog: %w", err)
	}
	defer dlg.Dispose()
	dlg.SetTitle("Set Value")
	dlg.SetSize(walk.Size{Width: 420, Height: 140})

	edit, err := walk.NewLineEdit(dlg)
	if err != nil {
		return "", false, fmt.Errorf("create set value editor: %w", err)
	}
	edit.SetBounds(walk.Rectangle{X: 12, Y: 12, Width: 392, Height: 24})
	edit.SetText(defaultValue)

	okBtn, err := walk.NewPushButton(dlg)
	if err != nil {
		return "", false, fmt.Errorf("create set value ok button: %w", err)
	}
	okBtn.SetText("OK")
	okBtn.SetBounds(walk.Rectangle{X: 236, Y: 52, Width: 80, Height: 28})

	cancelBtn, err := walk.NewPushButton(dlg)
	if err != nil {
		return "", false, fmt.Errorf("create set value cancel button: %w", err)
	}
	cancelBtn.SetText("Cancel")
	cancelBtn.SetBounds(walk.Rectangle{X: 324, Y: 52, Width: 80, Height: 28})

	okBtn.Clicked().Attach(func() { dlg.Accept() })
	cancelBtn.Clicked().Attach(func() { dlg.Cancel() })

	result, err := dlg.Run()
	if err != nil {
		return "", false, fmt.Errorf("run set value dialog: %w", err)
	}
	if result != walk.DlgCmdOK {
		return "", false, nil
	}
	return edit.Text(), true, nil
}
