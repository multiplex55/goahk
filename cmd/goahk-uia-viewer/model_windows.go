//go:build windows

package main

import "strings"

func uiaNodeLabel(nodeID, displayLabel, localizedControlType, controlType, name string) string {
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
	if typ == "" {
		return ""
	}
	return typ + " \"" + name + "\""
}
