package program

import (
	"fmt"
	"strings"

	"goahk/internal/hotkey"
)

func Validate(p Program) error {
	seenIDs := map[string]struct{}{}
	parsed := make([]hotkey.Binding, 0, len(p.Bindings))
	for _, b := range p.Bindings {
		idKey := strings.ToLower(strings.TrimSpace(b.ID))
		if _, exists := seenIDs[idKey]; exists {
			return fmt.Errorf("duplicate binding id %q", b.ID)
		}
		seenIDs[idKey] = struct{}{}

		binding, err := hotkey.ParseBinding(b.ID, b.Hotkey)
		if err != nil {
			return err
		}
		parsed = append(parsed, binding)

		for j, step := range b.Steps {
			if strings.TrimSpace(step.Action) == "" {
				return fmt.Errorf("binding %q binding/actions[%d]/name: action name is required", b.ID, j)
			}
		}
	}
	if err := hotkey.DetectConflicts(parsed); err != nil {
		return err
	}

	flowIDs := map[string]struct{}{}
	for _, f := range p.Options.Flows {
		flowIDs[strings.ToLower(strings.TrimSpace(f.ID))] = struct{}{}
	}
	for _, b := range p.Bindings {
		if ref := strings.ToLower(strings.TrimSpace(b.Flow)); ref != "" {
			if _, ok := flowIDs[ref]; !ok {
				return fmt.Errorf("binding %q references unknown flow %q", b.ID, b.Flow)
			}
		}
	}
	return nil
}
