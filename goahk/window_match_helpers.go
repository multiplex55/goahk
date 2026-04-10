package goahk

import "goahk/internal/window"

// MatchTitleExact matches a full window title (case-insensitive).
func MatchTitleExact(title string) window.Matcher {
	return window.MatchTitleExact(title)
}

// MatchTitleContains matches a title substring (case-insensitive).
func MatchTitleContains(title string) window.Matcher {
	return window.MatchTitleContains(title)
}

// MatchClass matches a window class name (case-insensitive).
func MatchClass(className string) window.Matcher {
	return window.MatchClass(className)
}

// MatchExe matches a process executable basename (case-insensitive).
func MatchExe(exe string) window.Matcher {
	return window.MatchExe(exe)
}
