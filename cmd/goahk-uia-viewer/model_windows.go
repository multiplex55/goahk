//go:build windows

package main

import "strings"

func uiaNodeLabel(nodeID, displayLabel, localizedControlType, controlType, name string) string {
	if strings.TrimSpace(displayLabel) != "" {
		return displayLabel
	}
	if s := composeTypeNameLabel(localizedControlType, name); s != "" {
		return s
	}
	if s := composeTypeNameLabel(controlType, name); s != "" {
		return s
	}
	return nodeID
}

func composeTypeNameLabel(typ, name string) string {
	typ = strings.TrimSpace(typ)
	name = strings.TrimSpace(name)
	if typ == "" || name == "" {
		return ""
	}
	return typ + " \"" + name + "\""
}
