package inspect

import "strings"

func formatDisplayLabel(name, localizedControlType, controlType string) string {
	control := strings.TrimSpace(localizedControlType)
	if control == "" {
		control = "element"
	}
	trimmedName := strings.TrimSpace(name)
	escapedName := strings.ReplaceAll(trimmedName, `"`, `\"`)
	_ = controlType // retained for API compatibility at call sites
	return control + ` "` + escapedName + `"`
}

func buildDebugMeta(el *uiaElement) DebugMetaDTO {
	if el == nil {
		return DebugMetaDTO{}
	}
	return DebugMetaDTO{
		ClassName:    strings.TrimSpace(el.ClassName),
		HWND:         strings.TrimSpace(el.HWND),
		AutomationID: strings.TrimSpace(el.AutomationID),
		RuntimeID:    strings.TrimSpace(el.RuntimeID),
	}
}

func selectorResolution(suggestions []SelectorCandidate) SelectorResolutionDTO {
	if len(suggestions) == 0 {
		return SelectorResolutionDTO{}
	}
	best := suggestions[0]
	resolution := SelectorResolutionDTO{Best: &best}
	if len(suggestions) > 1 {
		resolution.Alternates = append([]SelectorCandidate(nil), suggestions[1:]...)
	}
	return resolution
}
