package config

import (
	"encoding/json"
	"fmt"
	"strings"

	"goahk/internal/uia"
)

func ParseUIASelector(params map[string]string, defs map[string]UIASelector) (uia.Selector, error) {
	ref := strings.TrimSpace(params["selector"])
	var sel uia.Selector
	if ref != "" {
		def, ok := defs[ref]
		if !ok {
			return uia.Selector{}, fmt.Errorf("unknown uia selector %q", ref)
		}
		sel = toRuntimeSelector(def)
	}
	if raw := strings.TrimSpace(params["selector_json"]); raw != "" {
		if err := json.Unmarshal([]byte(raw), &sel); err != nil {
			return uia.Selector{}, fmt.Errorf("decode selector_json: %w", err)
		}
	}
	if sel.AutomationID == "" {
		sel.AutomationID = strings.TrimSpace(params["automationId"])
	}
	if sel.Name == "" {
		sel.Name = strings.TrimSpace(params["name"])
	}
	if sel.ControlType == "" {
		sel.ControlType = strings.TrimSpace(params["controlType"])
	}
	if err := sel.Validate(); err != nil {
		return uia.Selector{}, err
	}
	return sel, nil
}

func EncodeSelectorJSON(sel uia.Selector) (string, error) {
	raw, err := json.Marshal(sel)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func toRuntimeSelector(sel UIASelector) uia.Selector {
	out := uia.Selector{
		AutomationID: strings.TrimSpace(sel.AutomationID),
		Name:         strings.TrimSpace(sel.Name),
		ControlType:  strings.TrimSpace(sel.ControlType),
	}
	for _, anc := range sel.Ancestors {
		out.Ancestors = append(out.Ancestors, toRuntimeSelector(anc))
	}
	return out
}
