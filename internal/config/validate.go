package config

import (
	"fmt"
	"strings"
)

// Validate checks semantic constraints for runtime config.
func Validate(cfg Config) error {
	var errs []string

	ids := map[string]struct{}{}
	hotkeys := map[string]struct{}{}

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

		if len(hk.Steps) == 0 {
			errs = append(errs, fmt.Sprintf("hotkeys[%d].steps is required", i))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("invalid config: %s", strings.Join(errs, "; "))
	}
	return nil
}

func normalizeChord(chord string) string {
	return strings.ToLower(strings.Join(strings.Fields(chord), ""))
}
