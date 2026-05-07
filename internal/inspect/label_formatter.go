package inspect

import "strings"

func formatDisplayLabel(name, localizedControlType, controlType string) string {
	control := strings.TrimSpace(localizedControlType)
	if control == "" {
		control = fallbackControlTypeLabel(controlType)
	}
	if control == "" {
		control = "element"
	}
	escapedName := strings.ReplaceAll(name, `"`, `\"`)
	return control + ` "` + escapedName + `"`
}

func fallbackControlTypeLabel(controlType string) string {
	trimmed := strings.TrimSpace(controlType)
	if trimmed == "" {
		return ""
	}
	lower := strings.ToLower(trimmed)
	if mapped, ok := uiaControlTypeFallbackLabels[lower]; ok {
		return mapped
	}
	return trimmed
}

var uiaControlTypeFallbackLabels = map[string]string{
	"button":             "button",
	"edit":               "edit",
	"pane":               "pane",
	"window":             "window",
	"controltype.pane":   "pane",
	"controltype.button": "button",
	"controltype.edit":   "edit",
	"controltype.window": "window",
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
