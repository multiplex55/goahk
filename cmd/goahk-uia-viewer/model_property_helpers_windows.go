//go:build windows

package main

import (
	"strconv"
	"strings"

	"goahk/internal/inspect"
)

func propertyValueFromList(properties []inspect.PropertyDTO, name string) string {
	for _, prop := range properties {
		if !strings.EqualFold(strings.TrimSpace(prop.Name), name) || prop.Value == nil {
			continue
		}
		if value := strings.TrimSpace(*prop.Value); value != "" {
			return value
		}
	}
	return ""
}

func propertyIntFromList(properties []inspect.PropertyDTO, name string) int {
	for _, prop := range properties {
		if !strings.EqualFold(strings.TrimSpace(prop.Name), name) || prop.Value == nil {
			continue
		}
		value := strings.TrimSpace(*prop.Value)
		if value == "" || strings.EqualFold(value, "<nil>") {
			continue
		}
		if parsed, err := strconv.Atoi(value); err == nil && parsed > 0 {
			return parsed
		}
	}
	return 0
}

func propertyValueFromMap(byName map[string]inspect.PropertyDTO, name string) string {
	p, ok := byName[name]
	if !ok || p.Value == nil {
		return ""
	}
	return strings.TrimSpace(*p.Value)
}
