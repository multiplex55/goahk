package config

import (
	"fmt"
	"strings"

	"goahk/internal/program"
)

// Validate checks semantic constraints for runtime config.
func Validate(cfg Config) error {
	var errs []string

	ids := map[string]struct{}{}
	hotkeys := map[string]struct{}{}
	flows := map[string]Flow{}

	for i, f := range cfg.Flows {
		id := strings.ToLower(strings.TrimSpace(f.ID))
		if id == "" {
			errs = append(errs, fmt.Sprintf("flows[%d].id is required", i))
			continue
		}
		if _, ok := flows[id]; ok {
			errs = append(errs, fmt.Sprintf("flows[%d].id duplicates %q", i, f.ID))
		} else {
			flows[id] = f
		}
		if len(f.Steps) == 0 {
			errs = append(errs, fmt.Sprintf("flows[%d].steps is required", i))
		}
	}

	for i, hk := range cfg.Hotkeys {
		if strings.TrimSpace(hk.ID) == "" {
			errs = append(errs, fmt.Sprintf("hotkeys[%d].id is required", i))
		} else {
			key := strings.ToLower(strings.TrimSpace(hk.ID))
			if _, exists := ids[key]; exists {
				errs = append(errs, fmt.Sprintf("hotkeys[%d].id duplicates %q", i, hk.ID))
			}
			ids[key] = struct{}{}
		}

		if strings.TrimSpace(hk.Hotkey) == "" {
			errs = append(errs, fmt.Sprintf("hotkeys[%d].hotkey is required", i))
		} else {
			chord := normalizeChord(hk.Hotkey)
			if _, exists := hotkeys[chord]; exists {
				errs = append(errs, fmt.Sprintf("hotkeys[%d].hotkey duplicates %q", i, hk.Hotkey))
			}
			hotkeys[chord] = struct{}{}
		}

		hasSteps := len(hk.Steps) > 0
		hasFlow := strings.TrimSpace(hk.Flow) != ""
		if !hasSteps && !hasFlow {
			errs = append(errs, fmt.Sprintf("hotkeys[%d] requires steps or flow", i))
		}
		if hasSteps && hasFlow {
			errs = append(errs, fmt.Sprintf("hotkeys[%d] cannot set both steps and flow", i))
		}
		if hasFlow {
			if _, ok := flows[strings.ToLower(strings.TrimSpace(hk.Flow))]; !ok {
				errs = append(errs, fmt.Sprintf("hotkeys[%d].flow references unknown flow %q", i, hk.Flow))
			}
		}

		for j, step := range hk.Steps {
			if strings.HasPrefix(step.Action, "uia.") {
				if ref := strings.TrimSpace(step.Params["selector"]); ref != "" {
					if _, ok := cfg.UIASelectors[ref]; !ok {
						errs = append(errs, fmt.Sprintf("hotkeys[%d].steps[%d].selector references unknown uia selector %q", i, j, ref))
					}
				}
			}
		}
		policy := strings.ToLower(strings.TrimSpace(hk.ConcurrencyPolicy))
		if policy == "" {
			policy = string(program.DefaultConcurrencyPolicy())
		}
		if !isAllowedPolicy(policy) {
			errs = append(errs, fmt.Sprintf("hotkeys[%d].concurrencyPolicy has unsupported value %q", i, hk.ConcurrencyPolicy))
		}
	}
	for name, sel := range cfg.UIASelectors {
		if err := validateUIASelector(sel); err != nil {
			errs = append(errs, fmt.Sprintf("uiaSelectors.%s %s", name, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("invalid config: %s", strings.Join(errs, "; "))
	}
	return nil
}

func validateUIASelector(sel UIASelector) error {
	if strings.TrimSpace(sel.AutomationID) == "" && strings.TrimSpace(sel.Name) == "" && strings.TrimSpace(sel.ControlType) == "" {
		return fmt.Errorf("must specify at least one of automationId, name, controlType")
	}
	for i, anc := range sel.Ancestors {
		if err := validateUIASelector(anc); err != nil {
			return fmt.Errorf("ancestors[%d] %w", i, err)
		}
	}
	return nil
}

func normalizeChord(chord string) string {
	return strings.ToLower(strings.Join(strings.Fields(chord), ""))
}

func isAllowedPolicy(policy string) bool {
	switch policy {
	case string(program.ConcurrencyPolicySerial), string(program.ConcurrencyPolicyReplace), string(program.ConcurrencyPolicyParallel), string(program.ConcurrencyPolicyQueueOne), string(program.ConcurrencyPolicyDrop):
		return true
	default:
		return false
	}
}
